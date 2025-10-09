package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// JobMetrics tracks comprehensive job system metrics
type JobMetrics struct {
	redis *redis.Client
	mu    sync.RWMutex

	// Job counters
	jobsEnqueued   int64
	jobsProcessed  int64
	jobsSucceeded  int64
	jobsFailed     int64
	jobsRetried    int64
	jobsCancelled  int64

	// Performance metrics
	totalProcessingTime time.Duration
	avgProcessingTime   time.Duration
	minProcessingTime   time.Duration
	maxProcessingTime   time.Duration

	// Queue metrics
	queueDepth      int64
	processingCount int64

	// Per-type metrics
	jobTypeMetrics map[string]*JobTypeMetrics

	// Execution history
	executionHistory []*JobExecution
	maxHistorySize   int

	startTime time.Time
}

// JobTypeMetrics tracks metrics per job type
type JobTypeMetrics struct {
	Type              string        `json:"type"`
	Enqueued          int64         `json:"enqueued"`
	Processed         int64         `json:"processed"`
	Succeeded         int64         `json:"succeeded"`
	Failed            int64         `json:"failed"`
	AvgProcessingTime time.Duration `json:"avg_processing_time"`
	TotalTime         time.Duration `json:"total_time"`
}

// JobExecution represents a job execution record
type JobExecution struct {
	JobID         string        `json:"job_id"`
	JobType       string        `json:"job_type"`
	Status        JobStatus     `json:"status"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Duration      time.Duration `json:"duration"`
	Error         string        `json:"error,omitempty"`
	RetryCount    int           `json:"retry_count"`
	CompanyID     string        `json:"company_id,omitempty"`
	UserID        string        `json:"user_id,omitempty"`
}

// NewJobMetrics creates a new job metrics tracker
func NewJobMetrics(redis *redis.Client) *JobMetrics {
	return &JobMetrics{
		redis:            redis,
		jobTypeMetrics:   make(map[string]*JobTypeMetrics),
		executionHistory: make([]*JobExecution, 0, 1000),
		maxHistorySize:   1000,
		startTime:        time.Now(),
		minProcessingTime: time.Hour * 24, // Initialize with large value
	}
}

// RecordJobEnqueued records a job being enqueued
func (jm *JobMetrics) RecordJobEnqueued(jobType string) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jm.jobsEnqueued++
	jm.queueDepth++

	if _, exists := jm.jobTypeMetrics[jobType]; !exists {
		jm.jobTypeMetrics[jobType] = &JobTypeMetrics{Type: jobType}
	}
	jm.jobTypeMetrics[jobType].Enqueued++
}

// RecordJobStarted records a job starting processing
func (jm *JobMetrics) RecordJobStarted() {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jm.processingCount++
	jm.queueDepth--
}

// RecordJobCompleted records a job completion
func (jm *JobMetrics) RecordJobCompleted(job *Job, duration time.Duration, err error) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jm.jobsProcessed++
	jm.processingCount--
	jm.totalProcessingTime += duration

	// Update type-specific metrics
	if typeMetrics, exists := jm.jobTypeMetrics[job.Type]; exists {
		typeMetrics.Processed++
		typeMetrics.TotalTime += duration
		if typeMetrics.Processed > 0 {
			typeMetrics.AvgProcessingTime = typeMetrics.TotalTime / time.Duration(typeMetrics.Processed)
		}
	}

	// Update success/failure counters
	if err != nil {
		jm.jobsFailed++
		if typeMetrics, exists := jm.jobTypeMetrics[job.Type]; exists {
			typeMetrics.Failed++
		}
	} else {
		jm.jobsSucceeded++
		if typeMetrics, exists := jm.jobTypeMetrics[job.Type]; exists {
			typeMetrics.Succeeded++
		}
	}

	// Update processing time stats
	if jm.jobsProcessed > 0 {
		jm.avgProcessingTime = jm.totalProcessingTime / time.Duration(jm.jobsProcessed)
	}
	if duration < jm.minProcessingTime {
		jm.minProcessingTime = duration
	}
	if duration > jm.maxProcessingTime {
		jm.maxProcessingTime = duration
	}

	// Record execution history
	execution := &JobExecution{
		JobID:      job.ID,
		JobType:    job.Type,
		Status:     job.Status,
		Duration:   duration,
		RetryCount: job.RetryCount,
		CompanyID:  job.CompanyID,
		UserID:     job.UserID,
		EndTime:    time.Now(),
	}
	if job.StartedAt != nil {
		execution.StartTime = *job.StartedAt
	}
	if err != nil {
		execution.Error = err.Error()
	}

	jm.addExecutionHistory(execution)
}

// RecordJobRetried records a job retry
func (jm *JobMetrics) RecordJobRetried(jobType string) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jm.jobsRetried++
}

// RecordJobCancelled records a job cancellation
func (jm *JobMetrics) RecordJobCancelled(jobType string) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jm.jobsCancelled++
	jm.queueDepth--
}

// GetStats returns comprehensive job metrics
func (jm *JobMetrics) GetStats() *JobMetricsStats {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	successRate := 0.0
	if jm.jobsProcessed > 0 {
		successRate = float64(jm.jobsSucceeded) / float64(jm.jobsProcessed) * 100
	}

	failureRate := 0.0
	if jm.jobsProcessed > 0 {
		failureRate = float64(jm.jobsFailed) / float64(jm.jobsProcessed) * 100
	}

	retryRate := 0.0
	if jm.jobsProcessed > 0 {
		retryRate = float64(jm.jobsRetried) / float64(jm.jobsProcessed) * 100
	}

	return &JobMetricsStats{
		JobsEnqueued:        jm.jobsEnqueued,
		JobsProcessed:       jm.jobsProcessed,
		JobsSucceeded:       jm.jobsSucceeded,
		JobsFailed:          jm.jobsFailed,
		JobsRetried:         jm.jobsRetried,
		JobsCancelled:       jm.jobsCancelled,
		SuccessRate:         successRate,
		FailureRate:         failureRate,
		RetryRate:           retryRate,
		QueueDepth:          jm.queueDepth,
		ProcessingCount:     jm.processingCount,
		AvgProcessingTime:   jm.avgProcessingTime,
		MinProcessingTime:   jm.minProcessingTime,
		MaxProcessingTime:   jm.maxProcessingTime,
		TotalProcessingTime: jm.totalProcessingTime,
		Uptime:              time.Since(jm.startTime),
		StartTime:           jm.startTime,
	}
}

// GetJobTypeMetrics returns metrics per job type
func (jm *JobMetrics) GetJobTypeMetrics() map[string]*JobTypeMetrics {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]*JobTypeMetrics)
	for k, v := range jm.jobTypeMetrics {
		result[k] = &JobTypeMetrics{
			Type:              v.Type,
			Enqueued:          v.Enqueued,
			Processed:         v.Processed,
			Succeeded:         v.Succeeded,
			Failed:            v.Failed,
			AvgProcessingTime: v.AvgProcessingTime,
			TotalTime:         v.TotalTime,
		}
	}
	return result
}

// GetExecutionHistory returns recent job execution history
func (jm *JobMetrics) GetExecutionHistory(limit int) []*JobExecution {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	if limit <= 0 || limit > len(jm.executionHistory) {
		limit = len(jm.executionHistory)
	}

	// Return most recent executions
	start := len(jm.executionHistory) - limit
	if start < 0 {
		start = 0
	}

	result := make([]*JobExecution, limit)
	copy(result, jm.executionHistory[start:])
	return result
}

// GetFailedJobs returns recent failed job executions
func (jm *JobMetrics) GetFailedJobs(limit int) []*JobExecution {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	failed := make([]*JobExecution, 0, limit)
	for i := len(jm.executionHistory) - 1; i >= 0 && len(failed) < limit; i-- {
		if jm.executionHistory[i].Status == JobStatusFailed {
			failed = append(failed, jm.executionHistory[i])
		}
	}
	return failed
}

// ExportPrometheusMetrics exports metrics in Prometheus format
func (jm *JobMetrics) ExportPrometheusMetrics() string {
	stats := jm.GetStats()
	typeMetrics := jm.GetJobTypeMetrics()

	metrics := fmt.Sprintf(`# HELP jobs_enqueued_total Total number of jobs enqueued
# TYPE jobs_enqueued_total counter
jobs_enqueued_total %d

# HELP jobs_processed_total Total number of jobs processed
# TYPE jobs_processed_total counter
jobs_processed_total %d

# HELP jobs_succeeded_total Total number of jobs that succeeded
# TYPE jobs_succeeded_total counter
jobs_succeeded_total %d

# HELP jobs_failed_total Total number of jobs that failed
# TYPE jobs_failed_total counter
jobs_failed_total %d

# HELP jobs_retried_total Total number of job retries
# TYPE jobs_retried_total counter
jobs_retried_total %d

# HELP jobs_cancelled_total Total number of jobs cancelled
# TYPE jobs_cancelled_total counter
jobs_cancelled_total %d

# HELP jobs_queue_depth Current number of jobs in queue
# TYPE jobs_queue_depth gauge
jobs_queue_depth %d

# HELP jobs_processing_count Current number of jobs being processed
# TYPE jobs_processing_count gauge
jobs_processing_count %d

# HELP jobs_success_rate Job success rate percentage
# TYPE jobs_success_rate gauge
jobs_success_rate %.2f

# HELP jobs_failure_rate Job failure rate percentage
# TYPE jobs_failure_rate gauge
jobs_failure_rate %.2f

# HELP jobs_retry_rate Job retry rate percentage
# TYPE jobs_retry_rate gauge
jobs_retry_rate %.2f

# HELP jobs_avg_processing_seconds Average job processing time in seconds
# TYPE jobs_avg_processing_seconds gauge
jobs_avg_processing_seconds %.2f

# HELP jobs_min_processing_seconds Minimum job processing time in seconds
# TYPE jobs_min_processing_seconds gauge
jobs_min_processing_seconds %.2f

# HELP jobs_max_processing_seconds Maximum job processing time in seconds
# TYPE jobs_max_processing_seconds gauge
jobs_max_processing_seconds %.2f

`,
		stats.JobsEnqueued,
		stats.JobsProcessed,
		stats.JobsSucceeded,
		stats.JobsFailed,
		stats.JobsRetried,
		stats.JobsCancelled,
		stats.QueueDepth,
		stats.ProcessingCount,
		stats.SuccessRate,
		stats.FailureRate,
		stats.RetryRate,
		stats.AvgProcessingTime.Seconds(),
		stats.MinProcessingTime.Seconds(),
		stats.MaxProcessingTime.Seconds(),
	)

	// Add per-type metrics
	for jobType, typeMetrics := range typeMetrics {
		metrics += fmt.Sprintf(`# HELP jobs_by_type_total Jobs by type
# TYPE jobs_by_type_total counter
jobs_by_type_total{type="%s",status="enqueued"} %d
jobs_by_type_total{type="%s",status="processed"} %d
jobs_by_type_total{type="%s",status="succeeded"} %d
jobs_by_type_total{type="%s",status="failed"} %d

# HELP jobs_by_type_avg_duration_seconds Average processing time by job type
# TYPE jobs_by_type_avg_duration_seconds gauge
jobs_by_type_avg_duration_seconds{type="%s"} %.2f

`,
			jobType, typeMetrics.Enqueued,
			jobType, typeMetrics.Processed,
			jobType, typeMetrics.Succeeded,
			jobType, typeMetrics.Failed,
			jobType, typeMetrics.AvgProcessingTime.Seconds(),
		)
	}

	return metrics
}

// addExecutionHistory adds an execution to history (internal, not thread-safe)
func (jm *JobMetrics) addExecutionHistory(execution *JobExecution) {
	// Maintain circular buffer
	if len(jm.executionHistory) >= jm.maxHistorySize {
		// Remove oldest
		jm.executionHistory = jm.executionHistory[1:]
	}
	jm.executionHistory = append(jm.executionHistory, execution)

	// Also persist to Redis for durability
	jm.persistExecutionToRedis(execution)
}

// persistExecutionToRedis saves execution history to Redis
func (jm *JobMetrics) persistExecutionToRedis(execution *JobExecution) {
	ctx := context.Background()
	key := fmt.Sprintf("job:history:%s", execution.JobID)

	data, err := json.Marshal(execution)
	if err != nil {
		return
	}

	// Store with 7-day expiration
	jm.redis.Set(ctx, key, data, 7*24*time.Hour)

	// Add to sorted set for time-based queries
	historyKey := "job:history:timeline"
	jm.redis.ZAdd(ctx, historyKey, &redis.Z{
		Score:  float64(execution.EndTime.Unix()),
		Member: execution.JobID,
	})

	// Trim old entries (keep last 10000)
	jm.redis.ZRemRangeByRank(ctx, historyKey, 0, -10001)
}

// GetExecutionHistoryFromRedis retrieves execution history from Redis
func (jm *JobMetrics) GetExecutionHistoryFromRedis(ctx context.Context, limit int, offset int) ([]*JobExecution, error) {
	historyKey := "job:history:timeline"

	// Get job IDs from sorted set (most recent first)
	jobIDs, err := jm.redis.ZRevRange(ctx, historyKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get job history IDs: %w", err)
	}

	executions := make([]*JobExecution, 0, len(jobIDs))
	for _, jobID := range jobIDs {
		key := fmt.Sprintf("job:history:%s", jobID)
		data, err := jm.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var execution JobExecution
		if err := json.Unmarshal([]byte(data), &execution); err != nil {
			continue
		}
		executions = append(executions, &execution)
	}

	return executions, nil
}

// GetFailureAlerts returns jobs that need attention
func (jm *JobMetrics) GetFailureAlerts(ctx context.Context) []*JobAlert {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	alerts := make([]*JobAlert, 0)

	// Check for high failure rate
	if jm.jobsProcessed > 10 {
		failureRate := float64(jm.jobsFailed) / float64(jm.jobsProcessed) * 100
		if failureRate > 20 { // Alert if > 20% failure rate
			alerts = append(alerts, &JobAlert{
				Severity:  "warning",
				Message:   fmt.Sprintf("High job failure rate: %.2f%%", failureRate),
				Timestamp: time.Now(),
				Details: map[string]interface{}{
					"failure_rate":   failureRate,
					"jobs_failed":    jm.jobsFailed,
					"jobs_processed": jm.jobsProcessed,
				},
			})
		}
	}

	// Check for stuck jobs (processing for too long)
	if jm.processingCount > 10 {
		alerts = append(alerts, &JobAlert{
			Severity:  "warning",
			Message:   fmt.Sprintf("High number of jobs in processing: %d", jm.processingCount),
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"processing_count": jm.processingCount,
			},
		})
	}

	// Check for queue backlog
	if jm.queueDepth > 100 {
		alerts = append(alerts, &JobAlert{
			Severity:  "critical",
			Message:   fmt.Sprintf("Job queue backlog: %d jobs pending", jm.queueDepth),
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"queue_depth": jm.queueDepth,
			},
		})
	}

	return alerts
}

// Reset resets all metrics
func (jm *JobMetrics) Reset() {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jm.jobsEnqueued = 0
	jm.jobsProcessed = 0
	jm.jobsSucceeded = 0
	jm.jobsFailed = 0
	jm.jobsRetried = 0
	jm.jobsCancelled = 0
	jm.totalProcessingTime = 0
	jm.avgProcessingTime = 0
	jm.minProcessingTime = time.Hour * 24
	jm.maxProcessingTime = 0
	jm.queueDepth = 0
	jm.processingCount = 0
	jm.jobTypeMetrics = make(map[string]*JobTypeMetrics)
	jm.executionHistory = make([]*JobExecution, 0, jm.maxHistorySize)
	jm.startTime = time.Now()
}

// JobMetricsStats represents job metrics statistics
type JobMetricsStats struct {
	JobsEnqueued        int64         `json:"jobs_enqueued"`
	JobsProcessed       int64         `json:"jobs_processed"`
	JobsSucceeded       int64         `json:"jobs_succeeded"`
	JobsFailed          int64         `json:"jobs_failed"`
	JobsRetried         int64         `json:"jobs_retried"`
	JobsCancelled       int64         `json:"jobs_cancelled"`
	SuccessRate         float64       `json:"success_rate"`
	FailureRate         float64       `json:"failure_rate"`
	RetryRate           float64       `json:"retry_rate"`
	QueueDepth          int64         `json:"queue_depth"`
	ProcessingCount     int64         `json:"processing_count"`
	AvgProcessingTime   time.Duration `json:"avg_processing_time"`
	MinProcessingTime   time.Duration `json:"min_processing_time"`
	MaxProcessingTime   time.Duration `json:"max_processing_time"`
	TotalProcessingTime time.Duration `json:"total_processing_time"`
	Uptime              time.Duration `json:"uptime"`
	StartTime           time.Time     `json:"start_time"`
}

// JobAlert represents a job system alert
type JobAlert struct {
	Severity  string                 `json:"severity"`  // info, warning, critical
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details"`
}

