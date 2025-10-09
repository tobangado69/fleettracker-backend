package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusRetrying   JobStatus = "retrying"
	JobStatusCancelled  JobStatus = "cancelled"
)

// JobPriority represents the priority of a job
type JobPriority int

const (
	JobPriorityLow    JobPriority = 1
	JobPriorityNormal JobPriority = 5
	JobPriorityHigh   JobPriority = 10
	JobPriorityCritical JobPriority = 20
)

// Job represents a background job
type Job struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Data        map[string]interface{} `json:"data"`
	Priority    JobPriority            `json:"priority"`
	Status      JobStatus              `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	Error       string                 `json:"error,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
	CompanyID   string                 `json:"company_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

// JobHandler defines the interface for job handlers
type JobHandler interface {
	Handle(ctx context.Context, job *Job) error
	GetJobType() string
}

// JobQueue provides job queue functionality
type JobQueue struct {
	redis         *redis.Client
	handlers      map[string]JobHandler
	queueName     string
	processingSet string
	failedSet     string
	completedSet  string
}

// NewJobQueue creates a new job queue
func NewJobQueue(redis *redis.Client, queueName string) *JobQueue {
	return &JobQueue{
		redis:         redis,
		handlers:      make(map[string]JobHandler),
		queueName:     queueName,
		processingSet: fmt.Sprintf("%s:processing", queueName),
		failedSet:     fmt.Sprintf("%s:failed", queueName),
		completedSet:  fmt.Sprintf("%s:completed", queueName),
	}
}

// RegisterHandler registers a job handler
func (jq *JobQueue) RegisterHandler(handler JobHandler) {
	jq.handlers[handler.GetJobType()] = handler
}

// Enqueue adds a job to the queue
func (jq *JobQueue) Enqueue(ctx context.Context, job *Job) error {
	// Set default values
	if job.ID == "" {
		job.ID = fmt.Sprintf("job_%d", time.Now().UnixNano())
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.Status == "" {
		job.Status = JobStatusPending
	}
	if job.Priority == 0 {
		job.Priority = JobPriorityNormal
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}

	// Serialize job
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Add to priority queue (higher priority = higher score)
	score := float64(job.Priority)
	err = jq.redis.ZAdd(ctx, jq.queueName, &redis.Z{
		Score:  score,
		Member: job.ID,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to add job to queue: %w", err)
	}

	// Store job data
	jobKey := fmt.Sprintf("%s:job:%s", jq.queueName, job.ID)
	err = jq.redis.Set(ctx, jobKey, jobData, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to store job data: %w", err)
	}

	return nil
}

// Dequeue gets the next job from the queue
func (jq *JobQueue) Dequeue(ctx context.Context) (*Job, error) {
	// Get highest priority job
	result, err := jq.redis.ZPopMax(ctx, jq.queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No jobs available
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) == 0 {
		return nil, nil // No jobs available
	}

	jobID := result[0].Member.(string)

	// Get job data
	jobKey := fmt.Sprintf("%s:job:%s", jq.queueName, jobID)
	jobData, err := jq.redis.Get(ctx, jobKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("job data not found: %s", jobID)
		}
		return nil, fmt.Errorf("failed to get job data: %w", err)
	}

	// Deserialize job
	var job Job
	err = json.Unmarshal([]byte(jobData), &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Move to processing set
	err = jq.redis.ZAdd(ctx, jq.processingSet, &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: jobID,
	}).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to move job to processing: %w", err)
	}

	// Update job status
	job.Status = JobStatusProcessing
	now := time.Now()
	job.StartedAt = &now

	// Update job data
	updatedJobData, err := json.Marshal(job)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated job: %w", err)
	}

	err = jq.redis.Set(ctx, jobKey, updatedJobData, 24*time.Hour).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to update job data: %w", err)
	}

	return &job, nil
}

// Complete marks a job as completed
func (jq *JobQueue) Complete(ctx context.Context, jobID string, result map[string]interface{}) error {
	jobKey := fmt.Sprintf("%s:job:%s", jq.queueName, jobID)
	
	// Get current job data
	jobData, err := jq.redis.Get(ctx, jobKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get job data: %w", err)
	}

	var job Job
	err = json.Unmarshal([]byte(jobData), &job)
	if err != nil {
		return fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Update job status
	job.Status = JobStatusCompleted
	now := time.Now()
	job.CompletedAt = &now
	job.Result = result

	// Update job data
	updatedJobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal updated job: %w", err)
	}

	err = jq.redis.Set(ctx, jobKey, updatedJobData, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to update job data: %w", err)
	}

	// Move to completed set
	err = jq.redis.ZAdd(ctx, jq.completedSet, &redis.Z{
		Score:  float64(now.Unix()),
		Member: jobID,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to move job to completed: %w", err)
	}

	// Remove from processing set
	jq.redis.ZRem(ctx, jq.processingSet, jobID)

	return nil
}

// Fail marks a job as failed
func (jq *JobQueue) Fail(ctx context.Context, jobID string, errorMsg string) error {
	jobKey := fmt.Sprintf("%s:job:%s", jq.queueName, jobID)
	
	// Get current job data
	jobData, err := jq.redis.Get(ctx, jobKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get job data: %w", err)
	}

	var job Job
	err = json.Unmarshal([]byte(jobData), &job)
	if err != nil {
		return fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Check if we should retry
	if job.RetryCount < job.MaxRetries {
		// Retry the job
		job.RetryCount++
		job.Status = JobStatusRetrying
		job.Error = errorMsg

		// Update job data
		updatedJobData, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("failed to marshal updated job: %w", err)
		}

		err = jq.redis.Set(ctx, jobKey, updatedJobData, 24*time.Hour).Err()
		if err != nil {
			return fmt.Errorf("failed to update job data: %w", err)
		}

		// Re-enqueue with exponential backoff
		backoffDelay := time.Duration(job.RetryCount*job.RetryCount) * time.Second
		score := float64(time.Now().Add(backoffDelay).Unix())
		
		err = jq.redis.ZAdd(ctx, jq.queueName, &redis.Z{
			Score:  score,
			Member: jobID,
		}).Err()
		if err != nil {
			return fmt.Errorf("failed to re-enqueue job: %w", err)
		}
	} else {
		// Mark as permanently failed
		job.Status = JobStatusFailed
		job.Error = errorMsg
		now := time.Now()
		job.CompletedAt = &now

		// Update job data
		updatedJobData, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("failed to marshal updated job: %w", err)
		}

		err = jq.redis.Set(ctx, jobKey, updatedJobData, 24*time.Hour).Err()
		if err != nil {
			return fmt.Errorf("failed to update job data: %w", err)
		}

		// Move to failed set
		err = jq.redis.ZAdd(ctx, jq.failedSet, &redis.Z{
			Score:  float64(now.Unix()),
			Member: jobID,
		}).Err()
		if err != nil {
			return fmt.Errorf("failed to move job to failed: %w", err)
		}
	}

	// Remove from processing set
	jq.redis.ZRem(ctx, jq.processingSet, jobID)

	return nil
}

// GetJob retrieves a job by ID
func (jq *JobQueue) GetJob(ctx context.Context, jobID string) (*Job, error) {
	jobKey := fmt.Sprintf("%s:job:%s", jq.queueName, jobID)
	
	jobData, err := jq.redis.Get(ctx, jobKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}
		return nil, fmt.Errorf("failed to get job data: %w", err)
	}

	var job Job
	err = json.Unmarshal([]byte(jobData), &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// GetJobsByStatus retrieves jobs by status
func (jq *JobQueue) GetJobsByStatus(ctx context.Context, status JobStatus, limit int64) ([]*Job, error) {
	var setKey string
	switch status {
	case JobStatusPending:
		setKey = jq.queueName
	case JobStatusProcessing:
		setKey = jq.processingSet
	case JobStatusCompleted:
		setKey = jq.completedSet
	case JobStatusFailed:
		setKey = jq.failedSet
	default:
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	// Get job IDs
	results, err := jq.redis.ZRevRange(ctx, setKey, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get job IDs: %w", err)
	}

	var jobs []*Job
	for _, jobID := range results {
		job, err := jq.GetJob(ctx, jobID)
		if err != nil {
			continue // Skip invalid jobs
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetQueueStats returns queue statistics
func (jq *JobQueue) GetQueueStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get counts for each status
	pendingCount, err := jq.redis.ZCard(ctx, jq.queueName).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending count: %w", err)
	}

	processingCount, err := jq.redis.ZCard(ctx, jq.processingSet).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get processing count: %w", err)
	}

	completedCount, err := jq.redis.ZCard(ctx, jq.completedSet).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get completed count: %w", err)
	}

	failedCount, err := jq.redis.ZCard(ctx, jq.failedSet).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get failed count: %w", err)
	}

	stats["pending"] = pendingCount
	stats["processing"] = processingCount
	stats["completed"] = completedCount
	stats["failed"] = failedCount
	stats["total"] = pendingCount + processingCount + completedCount + failedCount

	return stats, nil
}

// Cancel cancels a job
func (jq *JobQueue) Cancel(ctx context.Context, jobID string) error {
	jobKey := fmt.Sprintf("%s:job:%s", jq.queueName, jobID)
	
	// Get current job data
	jobData, err := jq.redis.Get(ctx, jobKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get job data: %w", err)
	}

	var job Job
	err = json.Unmarshal([]byte(jobData), &job)
	if err != nil {
		return fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Only cancel pending or retrying jobs
	if job.Status != JobStatusPending && job.Status != JobStatusRetrying {
		return fmt.Errorf("cannot cancel job with status: %s", job.Status)
	}

	// Update job status
	job.Status = JobStatusCancelled
	now := time.Now()
	job.CompletedAt = &now

	// Update job data
	updatedJobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal updated job: %w", err)
	}

	err = jq.redis.Set(ctx, jobKey, updatedJobData, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to update job data: %w", err)
	}

	// Remove from queue
	jq.redis.ZRem(ctx, jq.queueName, jobID)
	jq.redis.ZRem(ctx, jq.processingSet, jobID)

	return nil
}

// Cleanup removes old completed and failed jobs
func (jq *JobQueue) Cleanup(ctx context.Context, olderThan time.Duration) error {
	cutoffTime := time.Now().Add(-olderThan)
	cutoffScore := float64(cutoffTime.Unix())

	// Cleanup completed jobs
	completedJobs, err := jq.redis.ZRangeByScore(ctx, jq.completedSet, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%.0f", cutoffScore),
	}).Result()
	if err != nil {
		return fmt.Errorf("failed to get old completed jobs: %w", err)
	}

	for _, jobID := range completedJobs {
		jobKey := fmt.Sprintf("%s:job:%s", jq.queueName, jobID)
		jq.redis.Del(ctx, jobKey)
		jq.redis.ZRem(ctx, jq.completedSet, jobID)
	}

	// Cleanup failed jobs
	failedJobs, err := jq.redis.ZRangeByScore(ctx, jq.failedSet, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%.0f", cutoffScore),
	}).Result()
	if err != nil {
		return fmt.Errorf("failed to get old failed jobs: %w", err)
	}

	for _, jobID := range failedJobs {
		jobKey := fmt.Sprintf("%s:job:%s", jq.queueName, jobID)
		jq.redis.Del(ctx, jobKey)
		jq.redis.ZRem(ctx, jq.failedSet, jobID)
	}

	return nil
}
