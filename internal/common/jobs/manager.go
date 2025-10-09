package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// Manager coordinates job queue, workers, and scheduler
type Manager struct {
	db           *gorm.DB
	redis        *redis.Client
	queue        *JobQueue
	worker       *Worker
	scheduler    *JobScheduler
	handlers     []JobHandler
	metrics      *JobMetrics
	deduplicator *JobDeduplicator
	priorityAdjuster *JobPriorityAdjuster
	purger       *JobPurger
}

// ManagerConfig holds manager configuration
type ManagerConfig struct {
	QueueName        string
	WorkerConcurrency int
	PollInterval     time.Duration
	JobTimeout       time.Duration
}

// DefaultManagerConfig returns default manager configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		QueueName:         "fleettracker:jobs",
		WorkerConcurrency: 5,
		PollInterval:      1 * time.Second,
		JobTimeout:        5 * time.Minute,
	}
}

// NewManager creates a new job manager
func NewManager(db *gorm.DB, redis *redis.Client, config *ManagerConfig) *Manager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	// Create job queue
	queue := NewJobQueue(redis, config.QueueName)

	// Create worker
	workerConfig := &WorkerConfig{
		Concurrency:     config.WorkerConcurrency,
		PollInterval:    config.PollInterval,
		JobTimeout:      config.JobTimeout,
		ShutdownTimeout: 30 * time.Second,
	}
	worker := NewWorker(queue, workerConfig)

	// Create scheduler
	scheduler := NewJobScheduler(redis, queue)

	// Create metrics tracker
	metrics := NewJobMetrics(redis)

	// Create deduplicator (15 minute TTL for deduplication)
	deduplicator := NewJobDeduplicator(redis, config.QueueName, 15*time.Minute)

	// Create priority adjuster
	priorityAdjuster := NewJobPriorityAdjuster(redis, queue)

	// Create purger
	purger := NewJobPurger(redis, queue)

	return &Manager{
		db:               db,
		redis:            redis,
		queue:            queue,
		worker:           worker,
		scheduler:        scheduler,
		handlers:         []JobHandler{},
		metrics:          metrics,
		deduplicator:     deduplicator,
		priorityAdjuster: priorityAdjuster,
		purger:           purger,
	}
}

// RegisterHandler registers a job handler
func (m *Manager) RegisterHandler(handler JobHandler) {
	m.handlers = append(m.handlers, handler)
	m.worker.RegisterHandler(handler)
	m.queue.RegisterHandler(handler)
}

// RegisterAllHandlers registers all built-in job handlers
func (m *Manager) RegisterAllHandlers() {
	log.Println("Registering job handlers...")

	// Report generation jobs
	reportHandler := NewReportGenerationJob(m.db)
	m.RegisterHandler(reportHandler)
	log.Printf("Registered handler: %s", reportHandler.GetJobType())

	// Data cleanup jobs
	cleanupHandler := NewDataCleanupJob(m.db)
	m.RegisterHandler(cleanupHandler)
	log.Printf("Registered handler: %s", cleanupHandler.GetJobType())

	log.Printf("Registered %d job handlers", len(m.handlers))
}

// SetupScheduledJobs sets up recurring scheduled jobs
func (m *Manager) SetupScheduledJobs() error {
	log.Println("Setting up scheduled jobs...")

	// Daily analytics aggregation
	err := m.scheduler.AddScheduledJob(&ScheduledJob{
		Name:     "Daily Analytics Aggregation",
		JobType:  "analytics_aggregation",
		Schedule: "@daily",
		Data: map[string]interface{}{
			"period": "daily",
		},
		Priority: JobPriorityNormal,
		IsActive: true,
	})
	if err != nil {
		return fmt.Errorf("failed to schedule daily analytics: %w", err)
	}

	// Monthly invoice generation
	err = m.scheduler.AddScheduledJob(&ScheduledJob{
		Name:     "Monthly Invoice Generation",
		JobType:  "invoice_generation",
		Schedule: "@monthly",
		Data: map[string]interface{}{
			"billing_period": "monthly",
		},
		Priority: JobPriorityHigh,
		IsActive: true,
	})
	if err != nil {
		return fmt.Errorf("failed to schedule monthly invoices: %w", err)
	}

	// Weekly data cleanup
	err = m.scheduler.AddScheduledJob(&ScheduledJob{
		Name:     "Weekly Data Cleanup",
		JobType:  "data_cleanup",
		Schedule: "@weekly",
		Data: map[string]interface{}{
			"cleanup_type": "weekly",
			"retention_days": 90,
		},
		Priority: JobPriorityLow,
		IsActive: true,
	})
	if err != nil {
		return fmt.Errorf("failed to schedule weekly cleanup: %w", err)
	}

	// Daily report generation
	err = m.scheduler.AddScheduledJob(&ScheduledJob{
		Name:     "Daily Fleet Report",
		JobType:  "report_generation",
		Schedule: "@daily",
		Data: map[string]interface{}{
			"report_type": "fleet_summary",
			"period": "daily",
		},
		Priority: JobPriorityNormal,
		IsActive: true,
	})
	if err != nil {
		return fmt.Errorf("failed to schedule daily report: %w", err)
	}

	// Hourly notification processing
	err = m.scheduler.AddScheduledJob(&ScheduledJob{
		Name:     "Hourly Notification Processing",
		JobType:  "notification",
		Schedule: "@hourly",
		Data: map[string]interface{}{
			"notification_type": "scheduled",
		},
		Priority: JobPriorityNormal,
		IsActive: true,
	})
	if err != nil {
		return fmt.Errorf("failed to schedule hourly notifications: %w", err)
	}

	log.Println("Scheduled jobs configured successfully")
	return nil
}

// Start starts the job manager (worker and scheduler)
func (m *Manager) Start() error {
	log.Println("Starting job manager...")

	// Register all handlers
	m.RegisterAllHandlers()

	// Setup scheduled jobs
	if err := m.SetupScheduledJobs(); err != nil {
		return fmt.Errorf("failed to setup scheduled jobs: %w", err)
	}

	// Start worker
	m.worker.Start()

	// Start scheduler
	m.scheduler.Start()

	log.Println("Job manager started successfully")
	return nil
}

// Stop stops the job manager gracefully
func (m *Manager) Stop() {
	log.Println("Stopping job manager...")

	// Stop scheduler
	m.scheduler.Stop()

	// Stop worker
	m.worker.Stop()

	log.Println("Job manager stopped")
}

// EnqueueJob enqueues a new job with deduplication check
func (m *Manager) EnqueueJob(ctx context.Context, job *Job) error {
	// Check for duplicates
	isDuplicate, err := m.deduplicator.IsDuplicate(ctx, job)
	if err != nil {
		log.Printf("Warning: deduplication check failed: %v", err)
	} else if isDuplicate {
		return fmt.Errorf("duplicate job detected: job with same fingerprint already exists")
	}

	// Enqueue the job
	if err := m.queue.Enqueue(ctx, job); err != nil {
		return err
	}

	// Mark as processed for deduplication
	if err := m.deduplicator.MarkAsProcessed(ctx, job); err != nil {
		log.Printf("Warning: failed to mark job as processed: %v", err)
	}

	// Record metrics
	m.metrics.RecordJobEnqueued(job.Type)

	return nil
}

// GetJobStatus returns the status of a job
func (m *Manager) GetJobStatus(ctx context.Context, jobID string) (*Job, error) {
	return m.queue.GetJob(ctx, jobID)
}

// GetJobsByStatus returns jobs by status  
func (m *Manager) GetJobsByStatus(ctx context.Context, status JobStatus, limit int) ([]*Job, error) {
	return m.queue.GetJobsByStatus(ctx, status, int64(limit))
}

// CancelJob cancels a pending job
func (m *Manager) CancelJob(ctx context.Context, jobID string) error {
	return m.queue.Cancel(ctx, jobID)
}

// RetryJob retries a failed job
func (m *Manager) RetryJob(ctx context.Context, jobID string) error {
	job, err := m.queue.GetJob(ctx, jobID)
	if err != nil {
		return err
	}
	job.RetryCount = 0
	job.Status = JobStatusPending
	return m.queue.Enqueue(ctx, job)
}

// GetWorkerMetrics returns worker metrics
func (m *Manager) GetWorkerMetrics() *WorkerMetrics {
	return m.worker.GetMetrics()
}

// GetQueueStats returns queue statistics
func (m *Manager) GetQueueStats(ctx context.Context) (map[string]interface{}, error) {
	return m.queue.GetQueueStats(ctx)
}

// PurgeCompletedJobs removes completed jobs older than the specified duration
func (m *Manager) PurgeCompletedJobs(ctx context.Context, olderThan time.Duration) (int, error) {
	return m.purger.PurgeCompletedJobs(ctx, olderThan)
}

// PurgeFailedJobs removes failed jobs older than the specified duration
func (m *Manager) PurgeFailedJobs(ctx context.Context, olderThan time.Duration) (int, error) {
	return m.purger.PurgeFailedJobs(ctx, olderThan)
}

// GetScheduledJobs returns all scheduled jobs
func (m *Manager) GetScheduledJobs() []*ScheduledJob {
	return m.scheduler.GetScheduledJobs()
}

// UpdateScheduledJob updates a scheduled job
func (m *Manager) UpdateScheduledJob(job *ScheduledJob) error {
	return m.scheduler.AddScheduledJob(job)
}

// DeleteScheduledJob deletes a scheduled job
func (m *Manager) DeleteScheduledJob(jobID string) error {
	return m.scheduler.RemoveScheduledJob(jobID)
}

// Helper methods for specific job types

// EnqueueReportGeneration enqueues a report generation job
func (m *Manager) EnqueueReportGeneration(ctx context.Context, reportType string, companyID string, data map[string]interface{}) (*Job, error) {
	job := &Job{
		Type:      "report_generation",
		CompanyID: companyID,
		Priority:  JobPriorityNormal,
		MaxRetries: 3,
		Data: map[string]interface{}{
			"report_type": reportType,
		},
	}

	// Merge additional data
	for k, v := range data {
		job.Data[k] = v
	}

	if err := m.EnqueueJob(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

// EnqueueInvoiceGeneration enqueues an invoice generation job
func (m *Manager) EnqueueInvoiceGeneration(ctx context.Context, companyID string, billingPeriod string) (*Job, error) {
	job := &Job{
		Type:      "invoice_generation",
		CompanyID: companyID,
		Priority:  JobPriorityHigh,
		MaxRetries: 3,
		Data: map[string]interface{}{
			"billing_period": billingPeriod,
		},
	}

	if err := m.EnqueueJob(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

// EnqueueDataCleanup enqueues a data cleanup job
func (m *Manager) EnqueueDataCleanup(ctx context.Context, cleanupType string, retentionDays int) (*Job, error) {
	job := &Job{
		Type:      "data_cleanup",
		Priority:  JobPriorityLow,
		MaxRetries: 2,
		Data: map[string]interface{}{
			"cleanup_type":   cleanupType,
			"retention_days": retentionDays,
		},
	}

	if err := m.EnqueueJob(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

// EnqueueNotification enqueues a notification job
func (m *Manager) EnqueueNotification(ctx context.Context, notificationType string, recipientID string, data map[string]interface{}) (*Job, error) {
	job := &Job{
		Type:      "notification",
		UserID:    recipientID,
		Priority:  JobPriorityHigh,
		MaxRetries: 5,
		Data: map[string]interface{}{
			"notification_type": notificationType,
		},
	}

	// Merge additional data
	for k, v := range data {
		job.Data[k] = v
	}

	if err := m.EnqueueJob(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

// GetMetrics returns comprehensive job metrics
func (m *Manager) GetMetrics() *JobMetricsStats {
	return m.metrics.GetStats()
}

// GetJobTypeMetrics returns metrics per job type
func (m *Manager) GetJobTypeMetrics() map[string]*JobTypeMetrics {
	return m.metrics.GetJobTypeMetrics()
}

// GetExecutionHistory returns recent job execution history
func (m *Manager) GetExecutionHistory(limit int) []*JobExecution {
	return m.metrics.GetExecutionHistory(limit)
}

// GetExecutionHistoryFromRedis retrieves execution history from Redis
func (m *Manager) GetExecutionHistoryFromRedis(ctx context.Context, limit int, offset int) ([]*JobExecution, error) {
	return m.metrics.GetExecutionHistoryFromRedis(ctx, limit, offset)
}

// GetFailedJobsHistory returns recent failed jobs
func (m *Manager) GetFailedJobsHistory(limit int) []*JobExecution {
	return m.metrics.GetFailedJobs(limit)
}

// GetFailureAlerts returns jobs that need attention
func (m *Manager) GetFailureAlerts(ctx context.Context) []*JobAlert {
	return m.metrics.GetFailureAlerts(ctx)
}

// ExportPrometheusMetrics exports job metrics in Prometheus format
func (m *Manager) ExportPrometheusMetrics() string {
	return m.metrics.ExportPrometheusMetrics()
}

// AdjustJobPriorities adjusts priorities for pending jobs
func (m *Manager) AdjustJobPriorities(ctx context.Context) (int, error) {
	return m.priorityAdjuster.AdjustAllPriorities(ctx)
}

// GetPurgeStats returns statistics about purgeable jobs
func (m *Manager) GetPurgeStats(ctx context.Context, olderThan time.Duration) (map[string]interface{}, error) {
	return m.purger.GetPurgeStats(ctx, olderThan)
}

// CheckDuplicate checks if a job is a duplicate
func (m *Manager) CheckDuplicate(ctx context.Context, job *Job) (bool, error) {
	return m.deduplicator.IsDuplicate(ctx, job)
}

