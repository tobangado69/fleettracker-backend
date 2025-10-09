package jobs

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
)

// JobAPI provides HTTP API for job management
type JobAPI struct {
	manager *Manager
}

// NewJobAPI creates a new job API
func NewJobAPI(manager *Manager) *JobAPI {
	return &JobAPI{
		manager: manager,
	}
}

// EnqueueJobRequest represents a request to enqueue a job
type EnqueueJobRequest struct {
	Type      string                 `json:"type" binding:"required"`
	Data      map[string]interface{} `json:"data"`
	Priority  JobPriority            `json:"priority"`
	MaxRetries int                   `json:"max_retries"`
	Tags      []string               `json:"tags"`
}

// EnqueueJobResponse represents the response after enqueuing a job
type EnqueueJobResponse struct {
	JobID    string    `json:"job_id"`
	Status   string    `json:"status"`
	Enqueued time.Time `json:"enqueued"`
}

// JobStatusResponse represents job status information
type JobStatusResponse struct {
	Job    *Job    `json:"job"`
	Status string  `json:"status"`
}

// QueueStatsResponse represents queue statistics
type QueueStatsResponse struct {
	Stats map[string]interface{} `json:"stats"`
}

// WorkerMetricsResponse represents worker metrics
type WorkerMetricsResponse struct {
	Metrics *WorkerMetrics `json:"metrics"`
}

// ScheduledJobRequest represents a request to create a scheduled job
type ScheduledJobRequest struct {
	Name      string                 `json:"name" binding:"required"`
	JobType   string                 `json:"job_type" binding:"required"`
	Data      map[string]interface{} `json:"data"`
	Schedule  string                 `json:"schedule" binding:"required"`
	Priority  JobPriority            `json:"priority"`
	IsActive  bool                   `json:"is_active"`
}

// EnqueueJobHandler handles job enqueue requests
func (ja *JobAPI) EnqueueJobHandler(c *gin.Context) {
	var req EnqueueJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Get user and company information
	userID, _ := c.Get("user_id")
	companyID, _ := c.Get("company_id")

	// Create job
	job := &Job{
		Type:      req.Type,
		Data:      req.Data,
		Priority:  req.Priority,
		MaxRetries: req.MaxRetries,
		Tags:      req.Tags,
		CompanyID: companyID.(string),
		UserID:    userID.(string),
	}

	// Enqueue job
	err := ja.manager.EnqueueJob(c.Request.Context(), job)
	if err != nil {
		middleware.AbortWithInternal(c, "Job operation failed", err)
		return
	}

	response := EnqueueJobResponse{
		JobID:    job.ID,
		Status:   string(job.Status),
		Enqueued: job.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetJobStatusHandler handles job status requests
func (ja *JobAPI) GetJobStatusHandler(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		middleware.AbortWithBadRequest(c, "job ID is required")
		return
	}

	job, err := ja.manager.GetJobStatus(c.Request.Context(), jobID)
	if err != nil {
		middleware.AbortWithNotFound(c, err.Error())
		return
	}

	response := JobStatusResponse{
		Job:    job,
		Status: string(job.Status),
	}

	c.JSON(http.StatusOK, response)
}

// GetJobsByStatusHandler handles requests for jobs by status
func (ja *JobAPI) GetJobsByStatusHandler(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		middleware.AbortWithBadRequest(c, "status is required")
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		middleware.AbortWithBadRequest(c, "invalid limit parameter")
		return
	}

	jobs, err := ja.manager.GetJobsByStatus(c.Request.Context(), JobStatus(status), int(limit))
	if err != nil {
		middleware.AbortWithInternal(c, "Job operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

// GetQueueStatsHandler handles queue statistics requests
func (ja *JobAPI) GetQueueStatsHandler(c *gin.Context) {
	stats, err := ja.manager.GetQueueStats(c.Request.Context())
	if err != nil {
		middleware.AbortWithInternal(c, "Job operation failed", err)
		return
	}

	response := QueueStatsResponse{Stats: stats}
	c.JSON(http.StatusOK, response)
}

// GetWorkerMetricsHandler handles worker metrics requests
func (ja *JobAPI) GetWorkerMetricsHandler(c *gin.Context) {
	metrics := ja.manager.GetWorkerMetrics()
	response := WorkerMetricsResponse{Metrics: metrics}
	c.JSON(http.StatusOK, response)
}

// GetWorkerHealthHandler handles worker health requests
func (ja *JobAPI) GetWorkerHealthHandler(c *gin.Context) {
	// Get worker health from metrics
	metrics := ja.manager.GetWorkerMetrics()
	health := gin.H{
		"status": "healthy",
		"metrics": metrics,
	}
	c.JSON(http.StatusOK, health)
}

// CancelJobHandler handles job cancellation requests
func (ja *JobAPI) CancelJobHandler(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		middleware.AbortWithBadRequest(c, "job ID is required")
		return
	}

	err := ja.manager.CancelJob(c.Request.Context(), jobID)
	if err != nil {
		middleware.AbortWithInternal(c, "Job operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job cancelled successfully"})
}

// ResetJobHandler handles job reset requests
func (ja *JobAPI) ResetJobHandler(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		middleware.AbortWithBadRequest(c, "job ID is required")
		return
	}

	// Retry the job
	err := ja.manager.RetryJob(c.Request.Context(), jobID)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to retry job", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job reset successfully"})
}

// CreateScheduledJobHandler handles scheduled job creation requests
func (ja *JobAPI) CreateScheduledJobHandler(c *gin.Context) {
	var req ScheduledJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Get user and company information
	userID, _ := c.Get("user_id")
	companyID, _ := c.Get("company_id")

	// Create scheduled job
	scheduledJob := &ScheduledJob{
		Name:      req.Name,
		JobType:   req.JobType,
		Data:      req.Data,
		Schedule:  req.Schedule,
		Priority:  req.Priority,
		IsActive:  req.IsActive,
		CompanyID: companyID.(string),
		UserID:    userID.(string),
	}

	err := ja.manager.UpdateScheduledJob(scheduledJob)
	if err != nil {
		middleware.AbortWithInternal(c, "Job operation failed", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Scheduled job created successfully",
		"job_id":  scheduledJob.ID,
	})
}

// GetScheduledJobsHandler handles scheduled jobs list requests
func (ja *JobAPI) GetScheduledJobsHandler(c *gin.Context) {
	jobs := ja.manager.GetScheduledJobs()
	c.JSON(http.StatusOK, gin.H{"scheduled_jobs": jobs})
}

// GetScheduledJobHandler handles individual scheduled job requests
func (ja *JobAPI) GetScheduledJobHandler(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		middleware.AbortWithBadRequest(c, "job ID is required")
		return
	}

	// Get all scheduled jobs and find the one we want
	allJobs := ja.manager.GetScheduledJobs()
	var job *ScheduledJob
	for _, j := range allJobs {
		if j.ID == jobID {
			job = j
			break
		}
	}
	if job == nil {
		middleware.AbortWithNotFound(c, "Scheduled job not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"scheduled_job": job})
}

// UpdateScheduledJobHandler handles scheduled job update requests
func (ja *JobAPI) UpdateScheduledJobHandler(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		middleware.AbortWithBadRequest(c, "job ID is required")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Get the existing scheduled job
	allJobs := ja.manager.GetScheduledJobs()
	var existingJob *ScheduledJob
	for _, j := range allJobs {
		if j.ID == jobID {
			existingJob = j
			break
		}
	}
	if existingJob == nil {
		middleware.AbortWithNotFound(c, "Scheduled job not found")
		return
	}
	
	// Apply updates (simplified - in production you'd use a proper update mechanism)
	if name, ok := updates["name"].(string); ok {
		existingJob.Name = name
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		existingJob.IsActive = isActive
	}
	
	err := ja.manager.UpdateScheduledJob(existingJob)
	if err != nil {
		middleware.AbortWithInternal(c, "Job operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduled job updated successfully"})
}

// DeleteScheduledJobHandler handles scheduled job deletion requests
func (ja *JobAPI) DeleteScheduledJobHandler(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		middleware.AbortWithBadRequest(c, "job ID is required")
		return
	}

	err := ja.manager.DeleteScheduledJob(jobID)
	if err != nil {
		middleware.AbortWithInternal(c, "Job operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduled job deleted successfully"})
}

// GetSchedulerStatsHandler handles scheduler statistics requests
func (ja *JobAPI) GetSchedulerStatsHandler(c *gin.Context) {
	// Get basic scheduler stats from scheduled jobs
	allJobs := ja.manager.GetScheduledJobs()
	activeCount := 0
	for _, job := range allJobs {
		if job.IsActive {
			activeCount++
		}
	}
	
	stats := gin.H{
		"total_scheduled_jobs": len(allJobs),
		"active_jobs": activeCount,
		"inactive_jobs": len(allJobs) - activeCount,
	}
	c.JSON(http.StatusOK, gin.H{"scheduler_stats": stats})
}

// CleanupJobsHandler handles job cleanup requests
func (ja *JobAPI) CleanupJobsHandler(c *gin.Context) {
	olderThanStr := c.DefaultQuery("older_than", "24h")
	olderThan, err := time.ParseDuration(olderThanStr)
	if err != nil {
		middleware.AbortWithBadRequest(c, "invalid older_than parameter")
		return
	}

	count, err := ja.manager.PurgeCompletedJobs(c.Request.Context(), olderThan)
	if err != nil {
		middleware.AbortWithInternal(c, "Job operation failed", err)
		return
	}
	
	// Also cleanup failed jobs
	failedCount, err := ja.manager.PurgeFailedJobs(c.Request.Context(), olderThan)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to cleanup failed jobs", err)
		return
	}
	
	count += failedCount

	c.JSON(http.StatusOK, gin.H{
		"message": "Job cleanup completed successfully",
		"count": count,
	})
}

// GetJobMetricsHandler returns comprehensive job metrics
func (ja *JobAPI) GetJobMetricsHandler(c *gin.Context) {
	metrics := ja.manager.GetMetrics()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetJobTypeMetricsHandler returns metrics per job type
func (ja *JobAPI) GetJobTypeMetricsHandler(c *gin.Context) {
	metrics := ja.manager.GetJobTypeMetrics()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetExecutionHistoryHandler returns job execution history
func (ja *JobAPI) GetExecutionHistoryHandler(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		middleware.AbortWithBadRequest(c, "invalid limit parameter")
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		middleware.AbortWithBadRequest(c, "invalid offset parameter")
		return
	}

	history, err := ja.manager.GetExecutionHistoryFromRedis(c.Request.Context(), limit, offset)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get execution history", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
		"count":   len(history),
	})
}

// GetFailedJobsHandler returns recent failed jobs
func (ja *JobAPI) GetFailedJobsHandler(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		middleware.AbortWithBadRequest(c, "invalid limit parameter")
		return
	}

	failed := ja.manager.GetFailedJobsHistory(limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    failed,
		"count":   len(failed),
	})
}

// GetJobAlertsHandler returns job system alerts
func (ja *JobAPI) GetJobAlertsHandler(c *gin.Context) {
	alerts := ja.manager.GetFailureAlerts(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    alerts,
		"count":   len(alerts),
	})
}

// GetPrometheusMetricsHandler returns metrics in Prometheus format
func (ja *JobAPI) GetPrometheusMetricsHandler(c *gin.Context) {
	metrics := ja.manager.ExportPrometheusMetrics()
	c.String(http.StatusOK, metrics)
}

// AdjustPrioritiesHandler adjusts job priorities
func (ja *JobAPI) AdjustPrioritiesHandler(c *gin.Context) {
	adjusted, err := ja.manager.AdjustJobPriorities(c.Request.Context())
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to adjust priorities", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Job priorities adjusted",
		"count":   adjusted,
	})
}

// GetPurgeStatsHandler returns purgeable job statistics
func (ja *JobAPI) GetPurgeStatsHandler(c *gin.Context) {
	olderThanStr := c.DefaultQuery("older_than", "7d")
	olderThan, err := time.ParseDuration(olderThanStr)
	if err != nil {
		middleware.AbortWithBadRequest(c, "invalid older_than parameter")
		return
	}

	stats, err := ja.manager.GetPurgeStats(c.Request.Context(), olderThan)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get purge stats", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// CleanupDuplicatesHandler cleans up old duplicate markers
func (ja *JobAPI) CleanupDuplicatesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Duplicate markers are cleaned up automatically via TTL (15 min)",
	})
}

// SetupJobRoutes sets up job management routes
func SetupJobRoutes(r *gin.RouterGroup, api *JobAPI) {
	jobs := r.Group("/jobs")
	{
		// Job management
		jobs.POST("/enqueue", api.EnqueueJobHandler)
		jobs.GET("/:id", api.GetJobStatusHandler)
		jobs.GET("/status/:status", api.GetJobsByStatusHandler)
		jobs.DELETE("/:id", api.CancelJobHandler)
		jobs.POST("/:id/reset", api.ResetJobHandler)
		
		// Queue management
		jobs.GET("/stats", api.GetQueueStatsHandler)
		jobs.POST("/cleanup", api.CleanupJobsHandler)
		
		// Worker management
		jobs.GET("/worker/metrics", api.GetWorkerMetricsHandler)
		jobs.GET("/worker/health", api.GetWorkerHealthHandler)
		
		// Scheduled jobs
		scheduled := jobs.Group("/scheduled")
		{
			scheduled.POST("", api.CreateScheduledJobHandler)
			scheduled.GET("", api.GetScheduledJobsHandler)
			scheduled.GET("/:id", api.GetScheduledJobHandler)
			scheduled.PUT("/:id", api.UpdateScheduledJobHandler)
			scheduled.DELETE("/:id", api.DeleteScheduledJobHandler)
			scheduled.GET("/stats", api.GetSchedulerStatsHandler)
		}
		
		// Enhanced monitoring endpoints
		monitoring := jobs.Group("/monitoring")
		{
			monitoring.GET("/metrics", api.GetJobMetricsHandler)
			monitoring.GET("/type-metrics", api.GetJobTypeMetricsHandler)
			monitoring.GET("/history", api.GetExecutionHistoryHandler)
			monitoring.GET("/failed", api.GetFailedJobsHandler)
			monitoring.GET("/alerts", api.GetJobAlertsHandler)
			monitoring.GET("/prometheus", api.GetPrometheusMetricsHandler)
		}
		
		// Performance optimization endpoints
		optimization := jobs.Group("/optimization")
		{
			optimization.POST("/adjust-priorities", api.AdjustPrioritiesHandler)
			optimization.GET("/purge-stats", api.GetPurgeStatsHandler)
			optimization.POST("/cleanup-duplicates", api.CleanupDuplicatesHandler)
		}
	}
}
