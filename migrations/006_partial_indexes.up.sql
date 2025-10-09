-- Partial Indexes for Filtered Queries
-- Created: 2025-10-08
-- Description: Smaller, faster indexes for specific filtered queries

-- =============================================================================
-- ACTIVE/STATUS PARTIAL INDEXES
-- =============================================================================

-- Active vehicles only (60-80% of queries use is_active = true)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_active_only 
ON vehicles(company_id, status, created_at DESC) 
WHERE is_active = true;

-- Available vehicles (for assignment operations)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_available 
ON vehicles(company_id, make, model) 
WHERE status = 'available' AND is_active = true;

-- Active drivers only
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_active_only 
ON drivers(company_id, overall_score DESC) 
WHERE is_active = true AND status = 'active';

-- Available drivers (no vehicle assigned)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_unassigned 
ON drivers(company_id, overall_score DESC, sim_expiry) 
WHERE status = 'available' AND vehicle_id IS NULL;

-- Active geofences only
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_active_only 
ON geofences(company_id, type, created_at DESC) 
WHERE is_active = true;

-- =============================================================================
-- TIME-BASED PARTIAL INDEXES
-- =============================================================================

-- Recent GPS tracks (last 24 hours - real-time tracking)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_last_24h 
ON gps_tracks(vehicle_id, timestamp DESC, speed, heading) 
WHERE timestamp >= NOW() - INTERVAL '24 hours';

-- Recent trips (last 7 days - active operations)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_last_7days 
ON trips(company_id, vehicle_id, start_time DESC) 
WHERE start_time >= NOW() - INTERVAL '7 days';

-- Ongoing trips
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_ongoing 
ON trips(company_id, vehicle_id, driver_id, start_time DESC) 
WHERE status = 'active' OR status = 'in_progress';

-- Recent driver events (last 30 days)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_driver_events_last_30days 
ON driver_events(driver_id, event_type, timestamp DESC) 
WHERE timestamp >= NOW() - INTERVAL '30 days';

-- Upcoming maintenance (next 30 days)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_upcoming_maintenance 
ON vehicles(company_id, next_service_date) 
WHERE next_service_date BETWEEN NOW() AND NOW() + INTERVAL '30 days';

-- =============================================================================
-- COMPLIANCE & EXPIRY PARTIAL INDEXES
-- =============================================================================

-- Expiring SIM cards (next 60 days)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_sim_expiring_soon 
ON drivers(company_id, sim_expiry, name) 
WHERE sim_expiry BETWEEN NOW() AND NOW() + INTERVAL '60 days';

-- Expired SIM cards
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_sim_expired 
ON drivers(company_id, name, phone) 
WHERE sim_expiry < NOW();

-- Vehicles needing inspection
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_inspection_due 
ON vehicles(company_id, license_plate, inspection_date) 
WHERE inspection_date < NOW() + INTERVAL '30 days';

-- Insurance expiring soon
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_insurance_expiring 
ON vehicles(company_id, license_plate, insurance_expiry) 
WHERE insurance_expiry BETWEEN NOW() AND NOW() + INTERVAL '30 days';

-- =============================================================================
-- PAYMENT STATUS PARTIAL INDEXES
-- =============================================================================

-- Unpaid invoices
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_unpaid 
ON invoices(company_id, due_date, invoice_date DESC) 
WHERE status = 'unpaid' OR status = 'partial';

-- Overdue invoices (critical for collections)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_overdue_critical 
ON invoices(company_id, due_date, amount) 
WHERE status IN ('unpaid', 'partial') AND due_date < NOW();

-- Pending payments
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payments_pending 
ON payments(company_id, invoice_id, created_at DESC) 
WHERE status = 'pending';

-- Failed payments (for retry)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payments_failed 
ON payments(company_id, invoice_id, created_at DESC, retry_count) 
WHERE status = 'failed' AND retry_count < 3;

-- =============================================================================
-- PERFORMANCE & SAFETY PARTIAL INDEXES
-- =============================================================================

-- Speeding violations (speed > 80 km/h for Indonesian highways)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_speeding_violations 
ON gps_tracks(vehicle_id, driver_id, timestamp DESC, speed) 
WHERE speed > 80;

-- Harsh braking events (high deceleration)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_driver_events_harsh_braking 
ON driver_events(driver_id, vehicle_id, timestamp DESC) 
WHERE event_type = 'harsh_braking';

-- Rapid acceleration events
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_driver_events_rapid_acceleration 
ON driver_events(driver_id, vehicle_id, timestamp DESC) 
WHERE event_type = 'rapid_acceleration';

-- Critical driver events
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_driver_events_critical_only 
ON driver_events(driver_id, vehicle_id, timestamp DESC, event_type) 
WHERE severity = 'critical' OR severity = 'high';

-- Low performance drivers (needs training)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_low_performance 
ON drivers(company_id, name, overall_score) 
WHERE overall_score < 70 AND is_active = true;

-- =============================================================================
-- FUEL EFFICIENCY PARTIAL INDEXES
-- =============================================================================

-- Abnormal fuel consumption (potential theft detection)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fuel_logs_abnormal 
ON fuel_logs(vehicle_id, fuel_date DESC, amount, cost) 
WHERE amount > 100 OR cost > 2000000; -- > 100L or > 2M IDR

-- Fuel logs with efficiency data
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fuel_logs_with_efficiency 
ON fuel_logs(vehicle_id, fuel_date DESC, amount, odometer_reading) 
WHERE fuel_efficiency > 0;

-- =============================================================================
-- TRIP STATUS PARTIAL INDEXES
-- =============================================================================

-- Incomplete trips (data quality check)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_incomplete 
ON trips(vehicle_id, start_time DESC) 
WHERE end_time IS NULL OR total_distance = 0;

-- Long trips (> 8 hours, potential driver fatigue)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_long_duration 
ON trips(vehicle_id, driver_id, start_time DESC) 
WHERE status = 'active' 
  AND start_time < NOW() - INTERVAL '8 hours';

-- =============================================================================
-- USER & AUTHENTICATION PARTIAL INDEXES
-- =============================================================================

-- Active users only
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_active_only 
ON users(company_id, role, created_at DESC) 
WHERE is_active = true AND deleted_at IS NULL;

-- Locked accounts (security monitoring)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_locked 
ON users(company_id, email, locked_until) 
WHERE locked_until > NOW();

-- Verified users
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_verified 
ON users(company_id, role) 
WHERE is_verified = true AND is_active = true;

-- =============================================================================
-- SOFT DELETE OPTIMIZATION
-- =============================================================================

-- Non-deleted records (most queries filter deleted_at IS NULL)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_not_deleted 
ON vehicles(company_id, status, updated_at DESC) 
WHERE deleted_at IS NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_not_deleted 
ON drivers(company_id, status, updated_at DESC) 
WHERE deleted_at IS NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_not_deleted 
ON users(company_id, email) 
WHERE deleted_at IS NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_companies_not_deleted 
ON companies(status, created_at DESC) 
WHERE deleted_at IS NULL;

-- =============================================================================
-- MAINTENANCE OPTIMIZATION
-- =============================================================================

-- Update table statistics after creating indexes
ANALYZE vehicles;
ANALYZE drivers;
ANALYZE gps_tracks;
ANALYZE trips;
ANALYZE driver_events;
ANALYZE fuel_logs;
ANALYZE maintenance_logs;
ANALYZE payments;
ANALYZE invoices;
ANALYZE geofences;
ANALYZE users;
ANALYZE sessions;

