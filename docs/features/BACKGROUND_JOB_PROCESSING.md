# Background Job Processing Documentation

## Overview

The FleetTracker Pro backend now includes a comprehensive background job processing system that handles heavy operations asynchronously, improving API response times and system scalability. The system is built on Redis for distributed job queuing and includes job scheduling, monitoring, and management capabilities.

## Architecture

### Core Components

1. **Job Queue** (`internal/common/jobs/queue.go`) - Redis-based job queue with priority support
2. **Job Workers** (`internal/common/jobs/worker.go`) - Concurrent job processing with retry logic
3. **Job Handlers** (`internal/common/jobs/handlers.go`) - Specific handlers for different job types
4. **Job Scheduler** (`internal/common/jobs/scheduler.go`) - Cron-like job scheduling
5. **Job API** (`internal/common/jobs/api.go`) - HTTP API for job management

## Job Queue System

### Job Structure

```go
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
```

### Job Statuses

- **Pending**: Job is queued and waiting to be processed
- **Processing**: Job is currently being executed
- **Completed**: Job finished successfully
- **Failed**: Job failed permanently (exceeded max retries)
- **Retrying**: Job failed but will be retried
- **Cancelled**: Job was cancelled before execution

### Job Priorities

- **Low** (1): Background tasks, cleanup operations
- **Normal** (5): Regular business operations
- **High** (10): Important user requests
- **Critical** (20): System-critical operations

## Job Types and Handlers

### 1. Email Notification Jobs

**Type**: `email_notification`

**Purpose**: Send email notifications to users

**Data Structure**:
```json
{
    "to": "user@example.com",
    "subject": "Fleet Update",
    "body": "Your fleet status has been updated"
}
```

**Handler**: `EmailNotificationJob`

**Use Cases**:
- User registration confirmations
- Password reset emails
- Fleet status notifications
- Maintenance reminders
- Report delivery notifications

### 2. Report Generation Jobs

**Type**: `report_generation`

**Purpose**: Generate various types of reports

**Data Structure**:
```json
{
    "report_type": "fleet_summary",
    "start_date": "2024-01-01",
    "end_date": "2024-01-31"
}
```

**Supported Report Types**:
- `fleet_summary`: Fleet overview with vehicle and trip statistics
- `driver_performance`: Driver performance analysis
- `fuel_consumption`: Fuel usage and efficiency reports
- `maintenance`: Maintenance logs and schedules

**Handler**: `ReportGenerationJob`

**Use Cases**:
- Monthly fleet reports
- Driver performance evaluations
- Fuel efficiency analysis
- Maintenance scheduling reports

### 3. Data Export Jobs

**Type**: `data_export`

**Purpose**: Export data in various formats

**Data Structure**:
```json
{
    "export_type": "vehicles",
    "format": "csv"
}
```

**Supported Export Types**:
- `vehicles`: Vehicle data export
- `drivers`: Driver information export
- `trips`: Trip history export
- `gps_tracks`: GPS tracking data export

**Supported Formats**:
- `csv`: Comma-separated values
- `json`: JSON format
- `xlsx`: Excel format

**Handler**: `DataExportJob`

**Use Cases**:
- Data backup and archival
- Third-party system integration
- Compliance reporting
- Data analysis and analytics

### 4. Maintenance Reminder Jobs

**Type**: `maintenance_reminder`

**Purpose**: Create maintenance reminders for vehicles

**Data Structure**:
```json
{
    "vehicle_id": "vehicle-uuid",
    "maintenance_type": "routine_check"
}
```

**Handler**: `MaintenanceReminderJob`

**Use Cases**:
- Scheduled maintenance notifications
- Service interval reminders
- Inspection due alerts
- Repair scheduling

### 5. Data Cleanup Jobs

**Type**: `data_cleanup`

**Purpose**: Clean up old data to maintain system performance

**Data Structure**:
```json
{
    "cleanup_type": "gps_tracks",
    "older_than_days": 90
}
```

**Supported Cleanup Types**:
- `gps_tracks`: Remove old GPS tracking data
- `audit_logs`: Clean up old audit logs
- `driver_events`: Remove old driver events

**Handler**: `DataCleanupJob`

**Use Cases**:
- Database maintenance
- Storage optimization
- Compliance data retention
- Performance optimization

## Job Scheduling

### Scheduled Job Structure

```go
type ScheduledJob struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    JobType     string                 `json:"job_type"`
    Data        map[string]interface{} `json:"data"`
    Schedule    string                 `json:"schedule"`
    Priority    JobPriority            `json:"priority"`
    IsActive    bool                   `json:"is_active"`
    LastRun     *time.Time             `json:"last_run,omitempty"`
    NextRun     time.Time              `json:"next_run"`
    CompanyID   string                 `json:"company_id,omitempty"`
    UserID      string                 `json:"user_id,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}
```

### Schedule Formats

**Predefined Schedules**:
- `@hourly`: Every hour
- `@daily`: Every day at midnight
- `@weekly`: Every week on Sunday
- `@monthly`: Every month on the 1st

**Duration-based Schedules**:
- `1h`: Every hour
- `30m`: Every 30 minutes
- `1d`: Every day
- `1w`: Every week

### Default Scheduled Jobs

The system automatically creates these scheduled jobs:

1. **Daily Data Cleanup**
   - **Schedule**: `@daily`
   - **Type**: `data_cleanup`
   - **Purpose**: Clean up GPS tracks older than 90 days

2. **Weekly Maintenance Check**
   - **Schedule**: `@weekly`
   - **Type**: `maintenance_reminder`
   - **Purpose**: Check for vehicles due for maintenance

3. **Monthly Fleet Report**
   - **Schedule**: `@monthly`
   - **Type**: `report_generation`
   - **Purpose**: Generate monthly fleet summary reports

## Worker System

### Worker Configuration

```go
type WorkerConfig struct {
    Concurrency    int           `json:"concurrency"`     // Number of concurrent workers
    PollInterval   time.Duration `json:"poll_interval"`   // How often to poll for jobs
    JobTimeout     time.Duration `json:"job_timeout"`     // Maximum time to process a job
    ShutdownTimeout time.Duration `json:"shutdown_timeout"` // Time to wait for graceful shutdown
}
```

**Default Configuration**:
- **Concurrency**: 5 workers
- **Poll Interval**: 1 second
- **Job Timeout**: 5 minutes
- **Shutdown Timeout**: 30 seconds

### Worker Metrics

```go
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
```

### Retry Logic

Jobs automatically retry on failure with exponential backoff:

1. **First retry**: 1 second delay
2. **Second retry**: 4 seconds delay
3. **Third retry**: 9 seconds delay
4. **Maximum retries**: 3 (configurable)

## Job Management API

### Admin-Only Endpoints

All job management endpoints require admin role and are available under `/api/v1/admin/jobs/`.

#### Job Management

**Enqueue Job**:
```
POST /api/v1/admin/jobs/enqueue
```

**Request Body**:
```json
{
    "type": "email_notification",
    "data": {
        "to": "user@example.com",
        "subject": "Test Email",
        "body": "This is a test email"
    },
    "priority": 5,
    "max_retries": 3,
    "tags": ["test", "notification"]
}
```

**Response**:
```json
{
    "job_id": "job_1234567890",
    "status": "pending",
    "enqueued": "2024-01-15T10:30:00Z"
}
```

**Get Job Status**:
```
GET /api/v1/admin/jobs/{id}
```

**Response**:
```json
{
    "job": {
        "id": "job_1234567890",
        "type": "email_notification",
        "status": "completed",
        "created_at": "2024-01-15T10:30:00Z",
        "completed_at": "2024-01-15T10:30:05Z",
        "result": {
            "processing_time": "5s",
            "worker_id": 1
        }
    },
    "status": "completed"
}
```

**Get Jobs by Status**:
```
GET /api/v1/admin/jobs/status/{status}?limit=50
```

**Cancel Job**:
```
DELETE /api/v1/admin/jobs/{id}
```

**Reset Job**:
```
POST /api/v1/admin/jobs/{id}/reset
```

#### Queue Management

**Get Queue Statistics**:
```
GET /api/v1/admin/jobs/stats
```

**Response**:
```json
{
    "stats": {
        "pending": 5,
        "processing": 2,
        "completed": 150,
        "failed": 3,
        "total": 160
    }
}
```

**Cleanup Old Jobs**:
```
POST /api/v1/admin/jobs/cleanup?older_than=24h
```

#### Worker Management

**Get Worker Metrics**:
```
GET /api/v1/admin/jobs/worker/metrics
```

**Response**:
```json
{
    "metrics": {
        "jobs_processed": 1000,
        "jobs_succeeded": 950,
        "jobs_failed": 50,
        "jobs_retried": 25,
        "average_job_time": "2.5s",
        "total_job_time": "2500s",
        "last_job_time": "2024-01-15T10:30:00Z",
        "start_time": "2024-01-15T09:00:00Z",
        "uptime": "1h30m"
    }
}
```

**Get Worker Health**:
```
GET /api/v1/admin/jobs/worker/health
```

**Response**:
```json
{
    "status": "healthy",
    "uptime": "1h30m",
    "jobs_processed": 1000,
    "jobs_succeeded": 950,
    "jobs_failed": 50,
    "jobs_retried": 25,
    "average_job_time": "2.5s",
    "last_job_time": "2024-01-15T10:30:00Z",
    "concurrency": 5,
    "success_rate": 95.0
}
```

#### Scheduled Jobs Management

**Create Scheduled Job**:
```
POST /api/v1/admin/jobs/scheduled
```

**Request Body**:
```json
{
    "name": "Daily Report Generation",
    "job_type": "report_generation",
    "data": {
        "report_type": "fleet_summary",
        "start_date": "2024-01-01",
        "end_date": "2024-01-31"
    },
    "schedule": "@daily",
    "priority": 5,
    "is_active": true
}
```

**Get Scheduled Jobs**:
```
GET /api/v1/admin/jobs/scheduled
```

**Update Scheduled Job**:
```
PUT /api/v1/admin/jobs/scheduled/{id}
```

**Delete Scheduled Job**:
```
DELETE /api/v1/admin/jobs/scheduled/{id}
```

**Get Scheduler Statistics**:
```
GET /api/v1/admin/jobs/scheduled/stats
```

**Response**:
```json
{
    "scheduler_stats": {
        "total_scheduled_jobs": 5,
        "active_jobs": 4,
        "inactive_jobs": 1,
        "next_run": "2024-01-15T11:00:00Z"
    }
}
```

## Integration with Services

### Service Integration

The job processing system is integrated with existing services to offload heavy operations:

**Analytics Service**:
- Report generation jobs for complex analytics
- Data export jobs for large datasets
- Scheduled report generation

**Vehicle Service**:
- Maintenance reminder jobs
- Vehicle data export jobs
- Bulk vehicle operations

**Driver Service**:
- Driver performance report jobs
- Driver data export jobs
- Driver event processing

**Payment Service**:
- Invoice generation jobs
- Payment notification jobs
- Financial report generation

**Tracking Service**:
- GPS data cleanup jobs
- Location history export jobs
- Geofence violation processing

### Example Usage

**Enqueue a Report Generation Job**:
```go
job := &jobs.Job{
    Type: "report_generation",
    Data: map[string]interface{}{
        "report_type": "fleet_summary",
        "start_date": "2024-01-01",
        "end_date": "2024-01-31",
    },
    Priority: jobs.JobPriorityNormal,
    CompanyID: companyID,
    UserID: userID,
}

err := jobQueue.Enqueue(ctx, job)
if err != nil {
    return fmt.Errorf("failed to enqueue report job: %w", err)
}
```

**Create a Scheduled Job**:
```go
scheduledJob := &jobs.ScheduledJob{
    Name: "Weekly Maintenance Check",
    JobType: "maintenance_reminder",
    Data: map[string]interface{}{
        "maintenance_type": "routine_check",
    },
    Schedule: "@weekly",
    Priority: jobs.JobPriorityNormal,
    IsActive: true,
    CompanyID: companyID,
    UserID: userID,
}

err := jobScheduler.AddScheduledJob(scheduledJob)
if err != nil {
    return fmt.Errorf("failed to create scheduled job: %w", err)
}
```

## Performance and Scalability

### Redis-Based Architecture

**Benefits**:
- **Distributed**: Works across multiple server instances
- **Persistent**: Jobs survive server restarts
- **Scalable**: Can handle thousands of concurrent jobs
- **Reliable**: Built-in retry and error handling

**Redis Data Structures**:
- **Sorted Sets**: Priority-based job queuing
- **Hash Maps**: Job data storage
- **Sets**: Job status tracking (processing, completed, failed)

### Performance Characteristics

**Throughput**:
- **Job Processing**: 100-500 jobs per minute per worker
- **Queue Operations**: Sub-millisecond enqueue/dequeue
- **Redis Operations**: 1-2ms per operation

**Memory Usage**:
- **Job Storage**: ~1KB per job
- **Worker Overhead**: ~10MB per worker
- **Redis Memory**: Scales with job volume

**Latency**:
- **Job Enqueue**: <5ms
- **Job Processing**: Varies by job type (1s-5min)
- **Status Updates**: <1ms

## Monitoring and Observability

### Metrics Collection

**Queue Metrics**:
- Total jobs processed
- Jobs by status (pending, processing, completed, failed)
- Average processing time
- Queue depth and backlog

**Worker Metrics**:
- Worker health and uptime
- Success/failure rates
- Processing throughput
- Resource utilization

**Scheduler Metrics**:
- Scheduled job execution rates
- Next run times
- Schedule accuracy
- Missed executions

### Health Monitoring

**Health Checks**:
- Redis connectivity
- Worker availability
- Queue processing status
- Scheduled job execution

**Alerting**:
- High failure rates (>20%)
- Queue backlog (>100 jobs)
- Worker unavailability
- Redis connection issues

### Logging

**Job Execution Logs**:
- Job start/completion times
- Processing duration
- Error messages and stack traces
- Retry attempts and reasons

**System Logs**:
- Worker startup/shutdown
- Queue operations
- Scheduler events
- Performance metrics

## Error Handling and Recovery

### Error Types

**Job Errors**:
- **Validation Errors**: Invalid job data
- **Processing Errors**: Handler execution failures
- **Timeout Errors**: Jobs exceeding time limits
- **Resource Errors**: Memory or CPU exhaustion

**System Errors**:
- **Redis Errors**: Connection or operation failures
- **Worker Errors**: Worker process crashes
- **Scheduler Errors**: Schedule parsing or execution failures

### Recovery Mechanisms

**Automatic Retry**:
- Exponential backoff strategy
- Configurable retry limits
- Error-specific retry logic

**Graceful Degradation**:
- Fallback to synchronous processing
- Queue bypass for critical operations
- Error notification systems

**Data Consistency**:
- Transactional job operations
- Idempotent job handlers
- State recovery mechanisms

## Security Considerations

### Access Control

**Admin-Only Access**:
- All job management endpoints require admin role
- Job data isolation by company
- User-specific job tracking

**Data Protection**:
- Sensitive data encryption in job payloads
- Secure job data transmission
- Audit logging for all job operations

### Job Security

**Input Validation**:
- Job data sanitization
- Type checking and validation
- Size limits and constraints

**Execution Security**:
- Sandboxed job execution
- Resource limits and quotas
- Malicious job detection

## Best Practices

### Job Design

**Idempotent Operations**:
- Jobs should be safe to retry
- Use unique identifiers for operations
- Implement proper state checking

**Error Handling**:
- Provide clear error messages
- Log detailed error information
- Implement proper cleanup on failure

**Resource Management**:
- Set appropriate timeouts
- Limit memory and CPU usage
- Clean up temporary resources

### Queue Management

**Priority Assignment**:
- Use appropriate priority levels
- Consider user impact and urgency
- Balance system resources

**Batch Operations**:
- Group related operations
- Use bulk processing when possible
- Optimize for throughput

### Monitoring

**Proactive Monitoring**:
- Set up alerts for critical metrics
- Monitor queue health regularly
- Track performance trends

**Performance Optimization**:
- Profile job execution times
- Optimize slow-running jobs
- Scale workers based on load

## Troubleshooting

### Common Issues

**High Queue Backlog**:
- Check worker availability
- Increase worker concurrency
- Optimize slow jobs
- Scale Redis resources

**Job Failures**:
- Review error logs
- Check job data validity
- Verify handler implementation
- Test job retry logic

**Scheduler Issues**:
- Verify schedule format
- Check system time accuracy
- Monitor scheduler health
- Review scheduled job configuration

### Debugging Tools

**Job Inspection**:
- Use job status endpoints
- Review job execution logs
- Check Redis data directly
- Monitor worker metrics

**Performance Analysis**:
- Profile job execution times
- Monitor Redis performance
- Track worker resource usage
- Analyze queue patterns

## Future Enhancements

### Planned Features

**Advanced Scheduling**:
- Cron expression support
- Timezone-aware scheduling
- Conditional job execution
- Job dependency management

**Enhanced Monitoring**:
- Real-time dashboards
- Advanced alerting rules
- Performance analytics
- Capacity planning tools

**Job Templates**:
- Predefined job configurations
- Template-based job creation
- Parameterized job execution
- Job workflow management

**Distributed Processing**:
- Multi-region job distribution
- Load balancing across workers
- Fault tolerance improvements
- Cross-region job replication

The FleetTracker Pro background job processing system provides a robust, scalable, and maintainable solution for handling heavy operations asynchronously, significantly improving system performance and user experience.
