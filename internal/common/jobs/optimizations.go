package jobs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// JobDeduplicator prevents duplicate job execution
type JobDeduplicator struct {
	redis  *redis.Client
	prefix string
	ttl    time.Duration
}

// NewJobDeduplicator creates a new job deduplicator
func NewJobDeduplicator(redis *redis.Client, prefix string, ttl time.Duration) *JobDeduplicator {
	return &JobDeduplicator{
		redis:  redis,
		prefix: prefix,
		ttl:    ttl,
	}
}

// IsDuplicate checks if a job is a duplicate
func (jd *JobDeduplicator) IsDuplicate(ctx context.Context, job *Job) (bool, error) {
	fingerprint := jd.generateJobFingerprint(job)
	key := fmt.Sprintf("%s:dedup:%s", jd.prefix, fingerprint)

	// Check if fingerprint exists
	exists, err := jd.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check duplicate: %w", err)
	}

	return exists > 0, nil
}

// MarkAsProcessed marks a job as processed to prevent duplicates
func (jd *JobDeduplicator) MarkAsProcessed(ctx context.Context, job *Job) error {
	fingerprint := jd.generateJobFingerprint(job)
	key := fmt.Sprintf("%s:dedup:%s", jd.prefix, fingerprint)

	// Store fingerprint with TTL
	err := jd.redis.Set(ctx, key, job.ID, jd.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to mark job as processed: %w", err)
	}

	return nil
}

// RemoveDuplicate removes a job from duplicate tracking
func (jd *JobDeduplicator) RemoveDuplicate(ctx context.Context, job *Job) error {
	fingerprint := jd.generateJobFingerprint(job)
	key := fmt.Sprintf("%s:dedup:%s", jd.prefix, fingerprint)

	err := jd.redis.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to remove duplicate marker: %w", err)
	}

	return nil
}

// generateJobFingerprint generates a unique fingerprint for a job
func (jd *JobDeduplicator) generateJobFingerprint(job *Job) string {
	// Create fingerprint from job type, company ID, and data
	data := map[string]interface{}{
		"type":       job.Type,
		"company_id": job.CompanyID,
		"data":       job.Data,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// CleanupOldDuplicates removes old duplicate markers
func (jd *JobDeduplicator) CleanupOldDuplicates(ctx context.Context) (int, error) {
	pattern := fmt.Sprintf("%s:dedup:*", jd.prefix)

	keys, err := jd.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get duplicate keys: %w", err)
	}

	cleaned := 0
	for _, key := range keys {
		// Check TTL
		ttl, err := jd.redis.TTL(ctx, key).Result()
		if err != nil {
			continue
		}

		// Remove expired keys
		if ttl <= 0 {
			jd.redis.Del(ctx, key)
			cleaned++
		}
	}

	return cleaned, nil
}

// JobPriorityAdjuster dynamically adjusts job priorities
type JobPriorityAdjuster struct {
	redis  *redis.Client
	queue  *JobQueue
}

// NewJobPriorityAdjuster creates a new priority adjuster
func NewJobPriorityAdjuster(redis *redis.Client, queue *JobQueue) *JobPriorityAdjuster {
	return &JobPriorityAdjuster{
		redis: redis,
		queue: queue,
	}
}

// AdjustPriority adjusts job priority based on waiting time and retry count
func (jpa *JobPriorityAdjuster) AdjustPriority(ctx context.Context, job *Job) JobPriority {
	// Base priority
	priority := job.Priority

	// Increase priority for jobs waiting too long
	waitTime := time.Since(job.CreatedAt)
	if waitTime > 30*time.Minute {
		priority += JobPriority(2)
	}
	if waitTime > 1*time.Hour {
		priority += JobPriority(3)
	}

	// Decrease priority for retried jobs (to give fresh jobs a chance)
	if job.RetryCount > 0 {
		priority -= JobPriority(job.RetryCount)
	}

	// Ensure priority stays within bounds
	if priority < JobPriorityLow {
		priority = JobPriorityLow
	}
	if priority > JobPriorityCritical {
		priority = JobPriorityCritical
	}

	return priority
}

// AdjustAllPriorities adjusts priorities for all pending jobs
func (jpa *JobPriorityAdjuster) AdjustAllPriorities(ctx context.Context) (int, error) {
	// Get all pending jobs
	jobs, err := jpa.queue.GetJobsByStatus(ctx, JobStatusPending, 1000)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending jobs: %w", err)
	}

	adjusted := 0
	for _, job := range jobs {
		oldPriority := job.Priority
		newPriority := jpa.AdjustPriority(ctx, job)

		if newPriority != oldPriority {
			job.Priority = newPriority
			// Note: Priority is updated in memory, will take effect on next dequeue
			adjusted++
		}
	}

	return adjusted, nil
}

// JobPurger handles cleanup of old jobs
type JobPurger struct {
	redis *redis.Client
	queue *JobQueue
}

// NewJobPurger creates a new job purger
func NewJobPurger(redis *redis.Client, queue *JobQueue) *JobPurger {
	return &JobPurger{
		redis: redis,
		queue: queue,
	}
}

// PurgeCompletedJobs removes completed jobs older than the specified duration
func (jp *JobPurger) PurgeCompletedJobs(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-olderThan)
	pattern := jp.queue.completedSet

	// Get all completed job IDs
	jobIDs, err := jp.redis.SMembers(ctx, pattern).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get completed jobs: %w", err)
	}

	purged := 0
	for _, jobID := range jobIDs {
		// Get job to check completion time
		job, err := jp.queue.GetJob(ctx, jobID)
		if err != nil || job.CompletedAt == nil {
			continue
		}

		// Check if job is old enough to purge
		if job.CompletedAt.Before(cutoffTime) {
			// Remove from completed set
			jp.redis.SRem(ctx, pattern, jobID)

			// Remove job data
			jobKey := fmt.Sprintf("%s:job:%s", jp.queue.queueName, jobID)
			jp.redis.Del(ctx, jobKey)

			purged++
		}
	}

	return purged, nil
}

// PurgeFailedJobs removes failed jobs older than the specified duration
func (jp *JobPurger) PurgeFailedJobs(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-olderThan)
	pattern := jp.queue.failedSet

	// Get all failed job IDs
	jobIDs, err := jp.redis.SMembers(ctx, pattern).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get failed jobs: %w", err)
	}

	purged := 0
	for _, jobID := range jobIDs {
		// Get job to check failure time
		job, err := jp.queue.GetJob(ctx, jobID)
		if err != nil || job.CompletedAt == nil {
			continue
		}

		// Check if job is old enough to purge
		if job.CompletedAt.Before(cutoffTime) {
			// Remove from failed set
			jp.redis.SRem(ctx, pattern, jobID)

			// Remove job data
			jobKey := fmt.Sprintf("%s:job:%s", jp.queue.queueName, jobID)
			jp.redis.Del(ctx, jobKey)

			purged++
		}
	}

	return purged, nil
}

// PurgeAllOldJobs removes all old jobs (completed and failed)
func (jp *JobPurger) PurgeAllOldJobs(ctx context.Context, olderThan time.Duration) (int, error) {
	completedCount, err := jp.PurgeCompletedJobs(ctx, olderThan)
	if err != nil {
		return 0, err
	}

	failedCount, err := jp.PurgeFailedJobs(ctx, olderThan)
	if err != nil {
		return completedCount, err
	}

	return completedCount + failedCount, nil
}

// GetPurgeStats returns statistics about purgeable jobs
func (jp *JobPurger) GetPurgeStats(ctx context.Context, olderThan time.Duration) (map[string]interface{}, error) {
	cutoffTime := time.Now().Add(-olderThan)

	// Count purgeable completed jobs
	completedJobIDs, _ := jp.redis.SMembers(ctx, jp.queue.completedSet).Result()
	purgeableCompleted := 0
	for _, jobID := range completedJobIDs {
		job, err := jp.queue.GetJob(ctx, jobID)
		if err == nil && job.CompletedAt != nil && job.CompletedAt.Before(cutoffTime) {
			purgeableCompleted++
		}
	}

	// Count purgeable failed jobs
	failedJobIDs, _ := jp.redis.SMembers(ctx, jp.queue.failedSet).Result()
	purgeableFailed := 0
	for _, jobID := range failedJobIDs {
		job, err := jp.queue.GetJob(ctx, jobID)
		if err == nil && job.CompletedAt != nil && job.CompletedAt.Before(cutoffTime) {
			purgeableFailed++
		}
	}

	return map[string]interface{}{
		"purgeable_completed": purgeableCompleted,
		"purgeable_failed":    purgeableFailed,
		"total_purgeable":     purgeableCompleted + purgeableFailed,
		"cutoff_time":         cutoffTime,
	}, nil
}

