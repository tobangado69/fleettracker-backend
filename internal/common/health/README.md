# Health Check & Monitoring System

Production-ready health check system with Kubernetes probes, dependency monitoring, and Prometheus metrics.

## Features

### ✅ Health Check Endpoints
- **`GET /health`** - Basic health check (liveness)
- **`GET /health/live`** - Kubernetes liveness probe
- **`GET /health/ready`** - Kubernetes readiness probe with dependency checks
- **`GET /health/detailed`** - Comprehensive health status

### ✅ Dependency Monitoring
- Database connectivity check
- Redis connectivity check
- Latency measurement
- Status classification (healthy/unhealthy/degraded)

### ✅ System Metrics
- Memory usage tracking
- Goroutine count monitoring
- CPU count detection
- Service uptime tracking

### ✅ Prometheus Metrics
- **`GET /metrics`** - Prometheus text format
- **`GET /metrics/json`** - JSON format for dashboards

## Quick Start

### Initialize Health Checker

```go
import "github.com/tobangado69/fleettracker-pro/backend/internal/common/health"

// Create health checker
healthChecker := health.NewHealthChecker(db, redisClient, "FleetTracker Pro API", "1.0.0")

// Create handlers
healthHandler := health.NewHandler(healthChecker)
metricsHandler := health.NewMetricsHandler(healthChecker)

// Setup routes
health.SetupHealthRoutes(r, healthHandler)
health.SetupMetricsRoutes(r, metricsHandler)
```

## API Endpoints

### 1. Basic Health Check

```bash
GET /health
```

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-08T10:30:00Z",
  "service": "FleetTracker Pro API",
  "version": "1.0.0",
  "uptime": "2h 30m 15s"
}
```

**Use Case:** Load balancers, simple monitoring  
**Performance:** <1ms response time  
**Dependencies:** None checked

---

### 2. Liveness Probe (Kubernetes)

```bash
GET /health/live
```

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-08T10:30:00Z",
  "service": "FleetTracker Pro API",
  "version": "1.0.0"
}
```

**Use Case:** Kubernetes liveness probe  
**Meaning:** Process is responsive, restart if not  
**Performance:** <1ms response time

---

### 3. Readiness Probe (Kubernetes)

```bash
GET /health/ready
```

**Response (200 OK - Healthy):**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-08T10:30:00Z",
  "service": "FleetTracker Pro API",
  "version": "1.0.0",
  "uptime": "2h 30m 15s",
  "dependencies": {
    "database": {
      "status": "healthy",
      "latency_ms": 2,
      "message": "connected"
    },
    "redis": {
      "status": "healthy",
      "latency_ms": 1,
      "message": "connected"
    }
  },
  "system": {
    "memory_usage_mb": 256,
    "memory_alloc_mb": 128,
    "goroutine_count": 45,
    "cpu_count": 8
  }
}
```

**Response (503 Service Unavailable - Unhealthy):**
```json
{
  "status": "unhealthy",
  "timestamp": "2025-01-08T10:30:00Z",
  "service": "FleetTracker Pro API",
  "version": "1.0.0",
  "uptime": "2h 30m 15s",
  "dependencies": {
    "database": {
      "status": "unhealthy",
      "latency_ms": 2000,
      "error": "database ping failed: connection timeout"
    },
    "redis": {
      "status": "healthy",
      "latency_ms": 1,
      "message": "connected"
    }
  },
  "errors": [
    "database: database ping failed: connection timeout"
  ]
}
```

**Response (200 OK - Degraded):**
```json
{
  "status": "degraded",
  "timestamp": "2025-01-08T10:30:00Z",
  "dependencies": {
    "database": {
      "status": "healthy",
      "latency_ms": 2
    },
    "redis": {
      "status": "unhealthy",
      "error": "redis ping failed: connection refused"
    }
  },
  "errors": [
    "redis: redis ping failed: connection refused"
  ]
}
```

**Use Case:** Kubernetes readiness probe  
**Meaning:** Service is ready to accept traffic  
**Performance:** ~5-10ms response time (includes dependency checks)

**Status Meanings:**
- **healthy** - All dependencies operational
- **degraded** - Service works but Redis is down (caching disabled)
- **unhealthy** - Critical dependency (database) is down

---

### 4. Detailed Health Check

```bash
GET /health/detailed
```

Same as `/health/ready` but always returns 200 OK even if unhealthy.

**Use Case:** Operations dashboard, debugging

---

### 5. Prometheus Metrics

```bash
GET /metrics
```

**Response (200 OK - Text Format):**
```prometheus
# HELP fleettracker_up Service up status (1 = up, 0 = down)
# TYPE fleettracker_up gauge
fleettracker_up 1

# HELP fleettracker_uptime_seconds Service uptime in seconds
# TYPE fleettracker_uptime_seconds counter
fleettracker_uptime_seconds 9015.000000

# HELP fleettracker_memory_usage_bytes Memory usage in bytes
# TYPE fleettracker_memory_usage_bytes gauge
fleettracker_memory_usage_bytes 268435456

# HELP fleettracker_memory_alloc_bytes Allocated memory in bytes
# TYPE fleettracker_memory_alloc_bytes gauge
fleettracker_memory_alloc_bytes 134217728

# HELP fleettracker_goroutines Current number of goroutines
# TYPE fleettracker_goroutines gauge
fleettracker_goroutines 45

# HELP fleettracker_cpu_count Number of CPUs
# TYPE fleettracker_cpu_count gauge
fleettracker_cpu_count 8
```

**Use Case:** Prometheus monitoring, Grafana dashboards  
**Scrape Interval:** Recommended 15-30 seconds

---

### 6. JSON Metrics

```bash
GET /metrics/json
```

**Response (200 OK):**
```json
{
  "timestamp": "2025-01-08T10:30:00Z",
  "service": "FleetTracker Pro API",
  "version": "1.0.0",
  "uptime": "2h 30m 15s",
  "memory": {
    "alloc_mb": 128,
    "total_alloc_mb": 512,
    "sys_mb": 256,
    "num_gc": 42
  },
  "goroutines": 45,
  "cpu_count": 8
}
```

**Use Case:** Custom dashboards, JavaScript applications

---

## Kubernetes Configuration

### Deployment YAML

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fleettracker-api
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: api
        image: fleettracker-api:1.0.0
        ports:
        - containerPort: 8080
        
        # Liveness probe: Restart if not responsive
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        
        # Readiness probe: Remove from load balancer if not ready
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          successThreshold: 1
          failureThreshold: 3
```

### Service YAML

```yaml
apiVersion: v1
kind: Service
metadata:
  name: fleettracker-api
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/path: "/metrics"
    prometheus.io/port: "8080"
spec:
  selector:
    app: fleettracker-api
  ports:
  - port: 80
    targetPort: 8080
```

---

## Prometheus Configuration

### prometheus.yml

```yaml
scrape_configs:
  - job_name: 'fleettracker-api'
    scrape_interval: 15s
    static_configs:
      - targets: ['fleettracker-api:8080']
    metrics_path: '/metrics'
```

### Grafana Dashboard Query Examples

```promql
# Request rate
rate(fleettracker_uptime_seconds[5m])

# Memory usage
fleettracker_memory_usage_bytes / 1024 / 1024

# Goroutine count
fleettracker_goroutines

# Service availability
fleettracker_up
```

---

## Monitoring & Alerting

### Recommended Alerts

#### Service Down
```promql
fleettracker_up == 0
```
**Severity:** Critical  
**Action:** Immediate investigation

#### High Memory Usage
```promql
fleettracker_memory_alloc_bytes > 1073741824  # 1GB
```
**Severity:** Warning  
**Action:** Check for memory leaks

#### High Goroutine Count
```promql
fleettracker_goroutines > 1000
```
**Severity:** Warning  
**Action:** Check for goroutine leaks

---

## Status Meanings

### Healthy ✅
All systems operational:
- Database: Connected, <1s latency
- Redis: Connected, <500ms latency
- HTTP Status: **200 OK**

### Degraded ⚠️
Service operational but degraded:
- Database: Healthy
- Redis: Down (caching disabled)
- HTTP Status: **200 OK** (still serves traffic)

### Unhealthy ❌
Critical systems down:
- Database: Down or unreachable
- HTTP Status: **503 Service Unavailable**
- Action: Remove from load balancer

---

## Development

### Running Tests

```bash
# Run all health check tests
go test ./internal/common/health/... -v -cover

# Run with race detection
go test ./internal/common/health/... -race

# Benchmark
go test ./internal/common/health/... -bench=.
```

### Test Coverage

```bash
go test ./internal/common/health/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Current coverage: **26.5%+** (critical paths tested)

---

## Performance

### Response Times
- `/health`: **<1ms**
- `/health/live`: **<1ms**
- `/health/ready`: **~5-10ms** (includes DB/Redis ping)
- `/metrics`: **<2ms**

### Resource Usage
- Memory: **~5MB** for health checker
- CPU: **<0.1%** average
- Goroutines: **+0** (no additional goroutines)

---

## Troubleshooting

### Issue: Readiness probe fails but service works

**Cause:** Database or Redis connectivity issues  
**Check:**
```bash
curl http://localhost:8080/health/detailed
```

**Solution:**
1. Check database connectivity
2. Check Redis connectivity
3. Review errors array in response

---

### Issue: High memory usage reported

**Cause:** Normal Go memory behavior or actual leak  
**Check:**
```bash
curl http://localhost:8080/metrics/json | jq '.memory'
```

**Solution:**
1. Check `num_gc` - should increase over time
2. Monitor `alloc_mb` vs `sys_mb`
3. Force GC: Not recommended in production

---

### Issue: Service marked as unhealthy but database works

**Cause:** Timeout or network issues  
**Check:**
```bash
# Check database latency
curl http://localhost:8080/health/ready | jq '.dependencies.database'
```

**Solution:**
1. Increase timeout threshold
2. Check network latency
3. Review database connection pool settings

---

## Best Practices

### 1. Use Appropriate Probes

```yaml
# Liveness: Restart if process is stuck
livenessProbe:
  httpGet:
    path: /health/live
  periodSeconds: 10
  failureThreshold: 3

# Readiness: Remove from LB if not ready
readinessProbe:
  httpGet:
    path: /health/ready
  periodSeconds: 5
  failureThreshold: 2
```

### 2. Set Appropriate Timeouts

```yaml
# Give dependencies time to respond
readinessProbe:
  timeoutSeconds: 3  # DB + Redis check
  
livenessProbe:
  timeoutSeconds: 5  # Just process check
```

### 3. Monitor Trends

Track over time:
- Memory usage trends
- Goroutine count trends
- Response latency trends
- Error rates

### 4. Alert on Degraded State

Even if service is degraded (Redis down), monitor:
```promql
# Alert if degraded for > 5 minutes
count(health_status{status="degraded"}) > 0
```

---

## Integration Examples

### Load Balancer Health Check

**HAProxy:**
```conf
backend fleettracker
  option httpchk GET /health
  http-check expect status 200
  server api1 10.0.0.1:8080 check inter 5s
  server api2 10.0.0.2:8080 check inter 5s
```

**Nginx:**
```nginx
upstream fleettracker {
  server 10.0.0.1:8080;
  server 10.0.0.2:8080;
}

location /health {
  proxy_pass http://fleettracker/health;
  proxy_connect_timeout 1s;
  proxy_read_timeout 1s;
}
```

---

## License

MIT License - Part of FleetTracker Pro

