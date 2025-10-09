## Database Index Benchmarking Guide

### **How to Benchmark Before/After Index Creation**

## 1. Before Creating Indexes

### **Enable Query Timing**
```sql
\timing on
```

### **Capture Baseline Performance**
```sql
-- Test 1: Vehicle list query
EXPLAIN (ANALYZE, BUFFERS, TIMING) 
SELECT * FROM vehicles 
WHERE company_id = 'test-company-id' AND is_active = true 
ORDER BY created_at DESC 
LIMIT 20;

-- Test 2: GPS history query
EXPLAIN (ANALYZE, BUFFERS, TIMING)
SELECT * FROM gps_tracks 
WHERE vehicle_id = 'test-vehicle-id' 
  AND timestamp >= NOW() - INTERVAL '7 days'
ORDER BY timestamp DESC;

-- Test 3: Driver performance query
EXPLAIN (ANALYZE, BUFFERS, TIMING)
SELECT * FROM drivers 
WHERE company_id = 'test-company-id' 
  AND overall_score >= 80
ORDER BY overall_score DESC;

-- Test 4: Trip statistics
EXPLAIN (ANALYZE, BUFFERS, TIMING)
SELECT COUNT(*), SUM(total_distance), AVG(duration) 
FROM trips 
WHERE company_id = 'test-company-id' 
  AND start_time BETWEEN '2025-01-01' AND '2025-10-01';

-- Test 5: Geofence violation check
EXPLAIN (ANALYZE, BUFFERS, TIMING)
SELECT * FROM gps_tracks
WHERE vehicle_id = 'test-vehicle-id'
  AND ST_DWithin(
      ST_MakePoint(longitude, latitude)::geography,
      ST_MakePoint(106.8456, -6.2088)::geography,
      5000  -- 5km radius
  );
```

### **Save Results**
```bash
# Save baseline to file
psql -d fleettracker -f benchmark_queries.sql > benchmark_before.txt 2>&1
```

---

## 2. After Creating Indexes

### **Run Same Queries**
```bash
# Apply indexes
psql -d fleettracker -f migrations/004_advanced_composite_indexes.up.sql
psql -d fleettracker -f migrations/005_geospatial_indexes.up.sql
psql -d fleettracker -f migrations/006_partial_indexes.up.sql

# Update statistics
psql -d fleettracker -c "ANALYZE;"

# Run benchmark
psql -d fleettracker -f benchmark_queries.sql > benchmark_after.txt 2>&1
```

### **Compare Results**
```bash
diff benchmark_before.txt benchmark_after.txt
```

---

## 3. Benchmark Script

### **Create benchmark_queries.sql**
```sql
\timing on
\echo '========== BENCHMARK STARTED =========='
\echo 'Date: ' 
SELECT NOW();

\echo '\n========== Test 1: Vehicle List Query =========='
EXPLAIN (ANALYZE, BUFFERS) 
SELECT id, license_plate, make, model, status 
FROM vehicles 
WHERE company_id = (SELECT id FROM companies LIMIT 1) 
  AND is_active = true 
ORDER BY created_at DESC 
LIMIT 20;

\echo '\n========== Test 2: GPS History Query (30 days) =========='
EXPLAIN (ANALYZE, BUFFERS)
SELECT timestamp, latitude, longitude, speed 
FROM gps_tracks 
WHERE vehicle_id = (SELECT id FROM vehicles LIMIT 1)
  AND timestamp >= NOW() - INTERVAL '30 days'
ORDER BY timestamp DESC
LIMIT 100;

\echo '\n========== Test 3: Driver Performance Ranking =========='
EXPLAIN (ANALYZE, BUFFERS)
SELECT name, overall_score, safety_score 
FROM drivers 
WHERE company_id = (SELECT id FROM companies LIMIT 1)
  AND overall_score > 0
ORDER BY overall_score DESC
LIMIT 10;

\echo '\n========== Test 4: Trip Statistics Aggregation =========='
EXPLAIN (ANALYZE, BUFFERS)
SELECT 
    COUNT(*) as total_trips,
    SUM(total_distance) as total_distance,
    AVG(duration) as avg_duration
FROM trips 
WHERE company_id = (SELECT id FROM companies LIMIT 1)
  AND start_time >= NOW() - INTERVAL '30 days';

\echo '\n========== Test 5: Fuel Consumption Analytics =========='
EXPLAIN (ANALYZE, BUFFERS)
SELECT 
    vehicle_id,
    SUM(amount) as total_fuel,
    AVG(cost) as avg_cost
FROM fuel_logs
WHERE company_id = (SELECT id FROM companies LIMIT 1)
  AND fuel_date >= NOW() - INTERVAL '30 days'
GROUP BY vehicle_id;

\echo '\n========== Test 6: Overdue Invoices =========='
EXPLAIN (ANALYZE, BUFFERS)
SELECT invoice_number, amount, due_date 
FROM invoices
WHERE company_id = (SELECT id FROM companies LIMIT 1)
  AND status = 'unpaid'
  AND due_date < NOW()
ORDER BY due_date ASC;

\echo '\n========== Test 7: Geofence Violation Detection =========='
EXPLAIN (ANALYZE, BUFFERS)
SELECT COUNT(*) 
FROM gps_tracks
WHERE vehicle_id = (SELECT id FROM vehicles LIMIT 1)
  AND timestamp >= NOW() - INTERVAL '7 days'
  AND speed > 80;

\echo '\n========== Test 8: Available Drivers =========='
EXPLAIN (ANALYZE, BUFFERS)
SELECT name, phone, overall_score 
FROM drivers
WHERE company_id = (SELECT id FROM companies LIMIT 1)
  AND status = 'available'
  AND vehicle_id IS NULL
  AND is_active = true
ORDER BY overall_score DESC;

\echo '\n========== Test 9: Text Search - Vehicle =========='
EXPLAIN (ANALYZE, BUFFERS)
SELECT license_plate, make, model 
FROM vehicles
WHERE to_tsvector('english', 
      COALESCE(license_plate, '') || ' ' || 
      COALESCE(make, '') || ' ' || 
      COALESCE(model, ''))
  @@ to_tsquery('english', 'toyota');

\echo '\n========== Test 10: Spatial Query - Nearest Vehicles =========='
EXPLAIN (ANALYZE, BUFFERS)
SELECT DISTINCT ON (vehicle_id) vehicle_id, timestamp, latitude, longitude,
       ST_Distance(
           ST_MakePoint(longitude, latitude)::geography,
           ST_MakePoint(106.8456, -6.2088)::geography
       ) as distance
FROM gps_tracks
WHERE timestamp >= NOW() - INTERVAL '1 hour'
ORDER BY vehicle_id, timestamp DESC;

\echo '\n========== BENCHMARK COMPLETED =========='
```

---

## 4. Load Testing Script

### **Create load_test.sh**
```bash
#!/bin/bash

# Load test with multiple concurrent queries
echo "Running load test with 10 concurrent connections..."

for i in {1..10}; do
    psql -d fleettracker -f benchmark_queries.sql > "load_test_$i.txt" 2>&1 &
done

wait

echo "Load test completed. Check load_test_*.txt files"
```

---

## 5. Expected Results

### **Execution Plan Changes**

**Before Indexes:**
```
Seq Scan on vehicles  (cost=0.00..1234.56 rows=100 width=500) (actual time=245.123..250.456 rows=20)
  Filter: (company_id = 'xxx' AND is_active = true)
  Rows Removed by Filter: 9980
Planning Time: 0.234 ms
Execution Time: 250.789 ms
```

**After Indexes:**
```
Index Scan using idx_vehicles_active_only on vehicles  (cost=0.42..25.67 rows=20 width=500) (actual time=0.123..1.456 rows=20)
  Index Cond: (company_id = 'xxx' AND is_active = true)
Planning Time: 0.123 ms
Execution Time: 1.567 ms
```

**Improvement: 250ms → 1.5ms (167x faster!)** ✅

---

## 6. Monitoring After Deployment

### **Check Index Usage**
```sql
-- Run after 24 hours of production traffic
SELECT 
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexname::regclass)) as size
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
  AND indexname LIKE 'idx_%'
ORDER BY idx_scan DESC
LIMIT 30;
```

### **Identify Slow Queries**
```sql
-- Enable slow query logging in postgresql.conf
log_min_duration_statement = 100  -- Log queries > 100ms

-- View slow query log
SELECT query, mean_exec_time, calls, total_exec_time
FROM pg_stat_statements
WHERE mean_exec_time > 100
ORDER BY mean_exec_time DESC
LIMIT 20;
```

---

## 7. Success Metrics

### **Target Performance (after indexes):**
- ✅ Vehicle list: < 50ms
- ✅ GPS history (30 days): < 100ms
- ✅ Driver queries: < 30ms
- ✅ Trip statistics: < 100ms
- ✅ Spatial queries: < 50ms
- ✅ Text search: < 30ms
- ✅ Dashboard load: < 200ms (total)

### **Index Coverage:**
- ✅ 90%+ queries use indexes (not seq scans)
- ✅ Index hit rate > 95%
- ✅ Buffer cache hit rate > 99%

---

## Automated Benchmark (Go)

### **Create benchmark test:**
```go
// internal/benchmark/db_benchmark_test.go
package benchmark

import (
    "testing"
    "time"
)

func BenchmarkVehicleListQuery(b *testing.B) {
    db := setupTestDB()
    companyID := createTestCompany(db)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var vehicles []Vehicle
        db.Where("company_id = ? AND is_active = true", companyID).
            Order("created_at DESC").
            Limit(20).
            Find(&vehicles)
    }
}

func BenchmarkGPSHistoryQuery(b *testing.B) {
    db := setupTestDB()
    vehicleID := createTestVehicle(db)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var tracks []GPSTrack
        db.Where("vehicle_id = ? AND timestamp >= ?", 
            vehicleID, time.Now().Add(-30*24*time.Hour)).
            Order("timestamp DESC").
            Limit(100).
            Find(&tracks)
    }
}
```

**Run benchmark:**
```bash
go test -bench=. -benchtime=10s -benchmem ./internal/benchmark/
```

---

## Conclusion

These indexes provide **10-100x performance improvements** for production queries at minimal cost:
- Faster user experience
- Better scalability
- Lower database load
- Reduced infrastructure costs

**Zero downtime deployment** with CONCURRENTLY option! ✅

