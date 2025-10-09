package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// WorkerConfig holds worker configuration
type WorkerConfig struct {
	Concurrency    int           `json:"concurrency"`     // Number of concurrent workers
	PollInterval   time.Duration `json:"poll_interval"`   // How often to poll for jobs
	JobTimeout     time.Duration `json:"job_timeout"`     // Maximum time to process a job
	ShutdownTimeout time.Duration `json:"shutdown_timeout"` // Time to wait for graceful shutdown
}

// DefaultWorkerConfig returns default worker configuration
func DefaultWorkerConfig() *WorkerConfig {
	return &WorkerConfig{
		Concurrency:     5,
		PollInterval:    1 * time.Second,
		JobTimeout:      5 * time.Minute,
		ShutdownTimeout: 30 * time.Second,
	}
}

// Worker processes jobs from a queue
type Worker struct {
	queue    *JobQueue
	config   *WorkerConfig
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	handlers map[string]JobHandler
	metrics  *WorkerMetrics
}

// WorkerMetrics holds worker performance metrics
type WorkerMetrics struct {
	JobsProcessed    int64         `json:"jobs_processed"`
	JobsSucceeded    int64         `json:"jobs_succeeded"`
	JobsFailed       int64         `json:"jobs_failed"`
	JobsRetried      int64         `json:"jobs_retried"`
	AverageJobTime   time.Duration `json:"average_job_time"`
	TotalJobTime     time.Duration `json:"total_job_time"`
	LastJobTime      time.Time     `json:"last_job_time"`
	StartTime        time.Time     `json:"start_time"`
	Uptime           time.Duration `json:"uptime"`
}

// NewWorker creates a new worker
func NewWorker(queue *JobQueue, config *WorkerConfig) *Worker {
	if config == nil {
		config = DefaultWorkerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		queue:    queue,
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
		handlers: make(map[string]JobHandler),
		metrics: &WorkerMetrics{
			StartTime: time.Now(),
		},
	}
}

// RegisterHandler registers a job handler
func (w *Worker) RegisterHandler(handler JobHandler) {
	w.handlers[handler.GetJobType()] = handler
	w.queue.RegisterHandler(handler)
}

// Start starts the worker
func (w *Worker) Start() {
	log.Printf("Starting worker with %d concurrent workers", w.config.Concurrency)

	for i := 0; i < w.config.Concurrency; i++ {
		w.wg.Add(1)
		go w.workerLoop(i)
	}
}

// Stop stops the worker gracefully
func (w *Worker) Stop() {
	log.Println("Stopping worker...")
	w.cancel()

	// Wait for all workers to finish
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Worker stopped gracefully")
	case <-time.After(w.config.ShutdownTimeout):
		log.Println("Worker shutdown timeout exceeded")
	}
}

// workerLoop is the main worker loop
func (w *Worker) workerLoop(workerID int) {
	defer w.wg.Done()

	log.Printf("Worker %d started", workerID)

	for {
		select {
		case <-w.ctx.Done():
			log.Printf("Worker %d stopping", workerID)
			return
		default:
			// Try to get a job
			job, err := w.queue.Dequeue(w.ctx)
			if err != nil {
				log.Printf("Worker %d: Error dequeuing job: %v", workerID, err)
				time.Sleep(w.config.PollInterval)
				continue
			}

			if job == nil {
				// No jobs available, wait and try again
				time.Sleep(w.config.PollInterval)
				continue
			}

			// Process the job
			w.processJob(workerID, job)
		}
	}
}

// processJob processes a single job
func (w *Worker) processJob(workerID int, job *Job) {
	startTime := time.Now()
	log.Printf("Worker %d: Processing job %s of type %s", workerID, job.ID, job.Type)

	// Create job context with timeout
	jobCtx, cancel := context.WithTimeout(w.ctx, w.config.JobTimeout)
	defer cancel()

	// Get handler for job type
	handler, exists := w.handlers[job.Type]
	if !exists {
		log.Printf("Worker %d: No handler found for job type %s", workerID, job.Type)
		w.queue.Fail(jobCtx, job.ID, fmt.Sprintf("No handler found for job type: %s", job.Type))
		w.updateMetrics(false, false, startTime)
		return
	}

	// Process the job
	err := handler.Handle(jobCtx, job)
	processingTime := time.Since(startTime)

	if err != nil {
		log.Printf("Worker %d: Job %s failed: %v", workerID, job.ID, err)
		w.queue.Fail(jobCtx, job.ID, err.Error())
		w.updateMetrics(false, true, startTime)
	} else {
		log.Printf("Worker %d: Job %s completed successfully in %v", workerID, job.ID, processingTime)
		w.queue.Complete(jobCtx, job.ID, map[string]interface{}{
			"processing_time": processingTime.String(),
			"worker_id":       workerID,
		})
		w.updateMetrics(true, false, startTime)
	}
}

// updateMetrics updates worker metrics
func (w *Worker) updateMetrics(succeeded, retried bool, startTime time.Time) {
	processingTime := time.Since(startTime)

	w.metrics.JobsProcessed++
	w.metrics.TotalJobTime += processingTime
	w.metrics.AverageJobTime = w.metrics.TotalJobTime / time.Duration(w.metrics.JobsProcessed)
	w.metrics.LastJobTime = time.Now()
	w.metrics.Uptime = time.Since(w.metrics.StartTime)

	if succeeded {
		w.metrics.JobsSucceeded++
	} else {
		w.metrics.JobsFailed++
	}

	if retried {
		w.metrics.JobsRetried++
	}
}

// GetMetrics returns worker metrics
func (w *Worker) GetMetrics() *WorkerMetrics {
	metrics := *w.metrics
	metrics.Uptime = time.Since(metrics.StartTime)
	return &metrics
}

// GetHealthStatus returns worker health status
func (w *Worker) GetHealthStatus() map[string]interface{} {
	metrics := w.GetMetrics()

	status := map[string]interface{}{
		"status":           "healthy",
		"uptime":           metrics.Uptime.String(),
		"jobs_processed":   metrics.JobsProcessed,
		"jobs_succeeded":   metrics.JobsSucceeded,
		"jobs_failed":      metrics.JobsFailed,
		"jobs_retried":     metrics.JobsRetried,
		"average_job_time": metrics.AverageJobTime.String(),
		"last_job_time":    metrics.LastJobTime,
		"concurrency":      w.config.Concurrency,
	}

	// Check if worker is healthy
	if metrics.JobsProcessed > 0 {
		successRate := float64(metrics.JobsSucceeded) / float64(metrics.JobsProcessed) * 100
		status["success_rate"] = successRate

		if successRate < 80 {
			status["status"] = "warning"
			status["warning"] = "Low success rate detected"
		}
	}

	// Check if worker is responsive
	if time.Since(metrics.LastJobTime) > 5*time.Minute && metrics.JobsProcessed > 0 {
		status["status"] = "warning"
		status["warning"] = "Worker appears to be idle"
	}

	return status
}

// WorkerPool manages multiple workers
type WorkerPool struct {
	workers []*Worker
	queues  map[string]*JobQueue
	redis   *redis.Client
	config  *WorkerConfig
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(redis *redis.Client, config *WorkerConfig) *WorkerPool {
	if config == nil {
		config = DefaultWorkerConfig()
	}

	return &WorkerPool{
		workers: make([]*Worker, 0),
		queues:  make(map[string]*JobQueue),
		redis:   redis,
		config:  config,
	}
}

// AddQueue adds a queue to the worker pool
func (wp *WorkerPool) AddQueue(queueName string) *JobQueue {
	queue := NewJobQueue(wp.redis, queueName)
	wp.queues[queueName] = queue
	return queue
}

// GetQueue gets a queue by name
func (wp *WorkerPool) GetQueue(queueName string) *JobQueue {
	return wp.queues[queueName]
}

// StartWorker starts a worker for a specific queue
func (wp *WorkerPool) StartWorker(queueName string) error {
	queue, exists := wp.queues[queueName]
	if !exists {
		return fmt.Errorf("queue not found: %s", queueName)
	}

	worker := NewWorker(queue, wp.config)
	wp.workers = append(wp.workers, worker)
	worker.Start()

	return nil
}

// RegisterHandler registers a handler for a specific queue
func (wp *WorkerPool) RegisterHandler(queueName string, handler JobHandler) error {
	queue, exists := wp.queues[queueName]
	if !exists {
		return fmt.Errorf("queue not found: %s", queueName)
	}

	queue.RegisterHandler(handler)
	return nil
}

// StartAll starts all workers
func (wp *WorkerPool) StartAll() {
	for queueName := range wp.queues {
		wp.StartWorker(queueName)
	}
}

// StopAll stops all workers
func (wp *WorkerPool) StopAll() {
	for _, worker := range wp.workers {
		worker.Stop()
	}
}

// GetPoolMetrics returns metrics for all workers
func (wp *WorkerPool) GetPoolMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	
	totalJobsProcessed := int64(0)
	totalJobsSucceeded := int64(0)
	totalJobsFailed := int64(0)
	totalJobsRetried := int64(0)
	
	for i, worker := range wp.workers {
		workerMetrics := worker.GetMetrics()
		metrics[fmt.Sprintf("worker_%d", i)] = workerMetrics
		
		totalJobsProcessed += workerMetrics.JobsProcessed
		totalJobsSucceeded += workerMetrics.JobsSucceeded
		totalJobsFailed += workerMetrics.JobsFailed
		totalJobsRetried += workerMetrics.JobsRetried
	}
	
	metrics["total"] = map[string]interface{}{
		"jobs_processed": totalJobsProcessed,
		"jobs_succeeded": totalJobsSucceeded,
		"jobs_failed":    totalJobsFailed,
		"jobs_retried":   totalJobsRetried,
		"success_rate":   func() float64 {
			if totalJobsProcessed > 0 {
				return float64(totalJobsSucceeded) / float64(totalJobsProcessed) * 100
			}
			return 0
		}(),
	}
	
	return metrics
}

// GetPoolHealthStatus returns health status for all workers
func (wp *WorkerPool) GetPoolHealthStatus() map[string]interface{} {
	status := make(map[string]interface{})
	
	healthyWorkers := 0
	totalWorkers := len(wp.workers)
	
	for i, worker := range wp.workers {
		workerStatus := worker.GetHealthStatus()
		status[fmt.Sprintf("worker_%d", i)] = workerStatus
		
		if workerStatus["status"] == "healthy" {
			healthyWorkers++
		}
	}
	
	overallStatus := "healthy"
	if healthyWorkers < totalWorkers {
		overallStatus = "warning"
	}
	
	status["overall"] = map[string]interface{}{
		"status":           overallStatus,
		"healthy_workers":  healthyWorkers,
		"total_workers":    totalWorkers,
		"health_percentage": float64(healthyWorkers) / float64(totalWorkers) * 100,
	}
	
	return status
}
