package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// ScheduledJob represents a scheduled job
type ScheduledJob struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	JobType     string                 `json:"job_type"`
	Data        map[string]interface{} `json:"data"`
	Schedule    string                 `json:"schedule"`    // Cron-like schedule
	Priority    JobPriority            `json:"priority"`
	IsActive    bool                   `json:"is_active"`
	LastRun     *time.Time             `json:"last_run,omitempty"`
	NextRun     time.Time              `json:"next_run"`
	CompanyID   string                 `json:"company_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// JobScheduler manages scheduled jobs
type JobScheduler struct {
	redis       *redis.Client
	queue       *JobQueue
	scheduledJobs map[string]*ScheduledJob
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	ticker      *time.Ticker
}

// NewJobScheduler creates a new job scheduler
func NewJobScheduler(redis *redis.Client, queue *JobQueue) *JobScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &JobScheduler{
		redis:         redis,
		queue:         queue,
		scheduledJobs: make(map[string]*ScheduledJob),
		ctx:           ctx,
		cancel:        cancel,
		ticker:        time.NewTicker(1 * time.Minute), // Check every minute
	}
}

// Start starts the job scheduler
func (js *JobScheduler) Start() {
	log.Println("Starting job scheduler...")
	
	// Load scheduled jobs from Redis
	js.loadScheduledJobs()
	
	// Start the scheduler loop
	go js.schedulerLoop()
}

// Stop stops the job scheduler
func (js *JobScheduler) Stop() {
	log.Println("Stopping job scheduler...")
	js.cancel()
	js.ticker.Stop()
}

// AddScheduledJob adds a new scheduled job
func (js *JobScheduler) AddScheduledJob(job *ScheduledJob) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()
	
	// Set default values
	if job.ID == "" {
		job.ID = fmt.Sprintf("scheduled_%d", time.Now().UnixNano())
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	job.UpdatedAt = time.Now()
	
	// Calculate next run time
	nextRun, err := js.calculateNextRun(job.Schedule)
	if err != nil {
		return fmt.Errorf("invalid schedule: %w", err)
	}
	job.NextRun = nextRun
	
	// Store in memory
	js.scheduledJobs[job.ID] = job
	
	// Store in Redis
	return js.saveScheduledJob(job)
}

// UpdateScheduledJob updates an existing scheduled job
func (js *JobScheduler) UpdateScheduledJob(jobID string, updates map[string]interface{}) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()
	
	job, exists := js.scheduledJobs[jobID]
	if !exists {
		return fmt.Errorf("scheduled job not found: %s", jobID)
	}
	
	// Update fields
	if schedule, ok := updates["schedule"].(string); ok {
		job.Schedule = schedule
		nextRun, err := js.calculateNextRun(schedule)
		if err != nil {
			return fmt.Errorf("invalid schedule: %w", err)
		}
		job.NextRun = nextRun
	}
	
	if isActive, ok := updates["is_active"].(bool); ok {
		job.IsActive = isActive
	}
	
	if data, ok := updates["data"].(map[string]interface{}); ok {
		job.Data = data
	}
	
	if priority, ok := updates["priority"].(JobPriority); ok {
		job.Priority = priority
	}
	
	job.UpdatedAt = time.Now()
	
	// Save to Redis
	return js.saveScheduledJob(job)
}

// RemoveScheduledJob removes a scheduled job
func (js *JobScheduler) RemoveScheduledJob(jobID string) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()
	
	// Remove from memory
	delete(js.scheduledJobs, jobID)
	
	// Remove from Redis
	key := fmt.Sprintf("scheduled_job:%s", jobID)
	return js.redis.Del(js.ctx, key).Err()
}

// GetScheduledJob gets a scheduled job by ID
func (js *JobScheduler) GetScheduledJob(jobID string) (*ScheduledJob, error) {
	js.mutex.RLock()
	defer js.mutex.RUnlock()
	
	job, exists := js.scheduledJobs[jobID]
	if !exists {
		return nil, fmt.Errorf("scheduled job not found: %s", jobID)
	}
	
	// Return a copy
	jobCopy := *job
	return &jobCopy, nil
}

// GetScheduledJobs returns all scheduled jobs
func (js *JobScheduler) GetScheduledJobs() []*ScheduledJob {
	js.mutex.RLock()
	defer js.mutex.RUnlock()
	
	jobs := make([]*ScheduledJob, 0, len(js.scheduledJobs))
	for _, job := range js.scheduledJobs {
		jobCopy := *job
		jobs = append(jobs, &jobCopy)
	}
	
	return jobs
}

// schedulerLoop is the main scheduler loop
func (js *JobScheduler) schedulerLoop() {
	for {
		select {
		case <-js.ctx.Done():
			return
		case <-js.ticker.C:
			js.checkScheduledJobs()
		}
	}
}

// checkScheduledJobs checks for jobs that need to be executed
func (js *JobScheduler) checkScheduledJobs() {
	js.mutex.RLock()
	now := time.Now()
	
	var jobsToRun []*ScheduledJob
	for _, job := range js.scheduledJobs {
		if job.IsActive && job.NextRun.Before(now) {
			jobsToRun = append(jobsToRun, job)
		}
	}
	js.mutex.RUnlock()
	
	// Execute jobs that are due
	for _, job := range jobsToRun {
		js.executeScheduledJob(job)
	}
}

// executeScheduledJob executes a scheduled job
func (js *JobScheduler) executeScheduledJob(scheduledJob *ScheduledJob) {
	log.Printf("Executing scheduled job: %s", scheduledJob.Name)
	
	// Create a regular job from the scheduled job
	job := &Job{
		Type:      scheduledJob.JobType,
		Data:      scheduledJob.Data,
		Priority:  scheduledJob.Priority,
		CompanyID: scheduledJob.CompanyID,
		UserID:    scheduledJob.UserID,
		Tags:      []string{"scheduled", scheduledJob.ID},
	}
	
	// Enqueue the job
	err := js.queue.Enqueue(js.ctx, job)
	if err != nil {
		log.Printf("Failed to enqueue scheduled job %s: %v", scheduledJob.Name, err)
		return
	}
	
	// Update scheduled job
	js.mutex.Lock()
	now := time.Now()
	scheduledJob.LastRun = &now
	
	// Calculate next run time
	nextRun, err := js.calculateNextRun(scheduledJob.Schedule)
	if err != nil {
		log.Printf("Failed to calculate next run for job %s: %v", scheduledJob.Name, err)
		scheduledJob.IsActive = false
	} else {
		scheduledJob.NextRun = nextRun
	}
	
	scheduledJob.UpdatedAt = now
	js.mutex.Unlock()
	
	// Save updated job
	js.saveScheduledJob(scheduledJob)
}

// calculateNextRun calculates the next run time based on schedule
func (js *JobScheduler) calculateNextRun(schedule string) (time.Time, error) {
	now := time.Now()
	
	// Simple schedule parsing (can be extended for more complex cron expressions)
	switch schedule {
	case "@hourly":
		return now.Add(1 * time.Hour).Truncate(time.Hour), nil
	case "@daily":
		return now.Add(24 * time.Hour).Truncate(24 * time.Hour), nil
	case "@weekly":
		return now.Add(7 * 24 * time.Hour).Truncate(24 * time.Hour), nil
	case "@monthly":
		return now.AddDate(0, 1, 0).Truncate(24 * time.Hour), nil
	default:
		// Try to parse as duration (e.g., "1h", "30m", "1d")
		duration, err := time.ParseDuration(schedule)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid schedule format: %s", schedule)
		}
		return now.Add(duration), nil
	}
}

// loadScheduledJobs loads scheduled jobs from Redis
func (js *JobScheduler) loadScheduledJobs() {
	keys, err := js.redis.Keys(js.ctx, "scheduled_job:*").Result()
	if err != nil {
		log.Printf("Failed to load scheduled jobs: %v", err)
		return
	}
	
	for _, key := range keys {
		data, err := js.redis.Get(js.ctx, key).Result()
		if err != nil {
			continue
		}
		
		var job ScheduledJob
		if err := json.Unmarshal([]byte(data), &job); err != nil {
			continue
		}
		
		js.scheduledJobs[job.ID] = &job
	}
	
	log.Printf("Loaded %d scheduled jobs", len(js.scheduledJobs))
}

// saveScheduledJob saves a scheduled job to Redis
func (js *JobScheduler) saveScheduledJob(job *ScheduledJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal scheduled job: %w", err)
	}
	
	key := fmt.Sprintf("scheduled_job:%s", job.ID)
	return js.redis.Set(js.ctx, key, data, 0).Err() // No expiration
}

// InitializeDefaultScheduledJobs initializes default scheduled jobs
func (js *JobScheduler) InitializeDefaultScheduledJobs() {
	// Daily data cleanup job
	cleanupJob := &ScheduledJob{
		Name:      "Daily Data Cleanup",
		JobType:   "data_cleanup",
		Schedule:  "@daily",
		Priority:  JobPriorityLow,
		IsActive:  true,
		Data: map[string]interface{}{
			"cleanup_type":    "gps_tracks",
			"older_than_days": 90,
		},
		CompanyID: "system",
		UserID:    "system",
	}
	js.AddScheduledJob(cleanupJob)
	
	// Weekly maintenance reminder job
	maintenanceJob := &ScheduledJob{
		Name:      "Weekly Maintenance Check",
		JobType:   "maintenance_reminder",
		Schedule:  "@weekly",
		Priority:  JobPriorityNormal,
		IsActive:  true,
		Data: map[string]interface{}{
			"maintenance_type": "routine_check",
		},
		CompanyID: "system",
		UserID:    "system",
	}
	js.AddScheduledJob(maintenanceJob)
	
	// Monthly report generation job
	reportJob := &ScheduledJob{
		Name:      "Monthly Fleet Report",
		JobType:   "report_generation",
		Schedule:  "@monthly",
		Priority:  JobPriorityNormal,
		IsActive:  true,
		Data: map[string]interface{}{
			"report_type": "fleet_summary",
			"start_date":  time.Now().AddDate(0, -1, 0).Format("2006-01-02"),
			"end_date":    time.Now().Format("2006-01-02"),
		},
		CompanyID: "system",
		UserID:    "system",
	}
	js.AddScheduledJob(reportJob)
	
	log.Println("Initialized default scheduled jobs")
}

// GetSchedulerStats returns scheduler statistics
func (js *JobScheduler) GetSchedulerStats() map[string]interface{} {
	js.mutex.RLock()
	defer js.mutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_scheduled_jobs": len(js.scheduledJobs),
		"active_jobs":          0,
		"inactive_jobs":        0,
		"next_run":             nil,
	}
	
	var nextRun *time.Time
	activeCount := 0
	
	for _, job := range js.scheduledJobs {
		if job.IsActive {
			activeCount++
			if nextRun == nil || job.NextRun.Before(*nextRun) {
				nextRun = &job.NextRun
			}
		}
	}
	
	stats["active_jobs"] = activeCount
	stats["inactive_jobs"] = len(js.scheduledJobs) - activeCount
	if nextRun != nil {
		stats["next_run"] = *nextRun
	}
	
	return stats
}
