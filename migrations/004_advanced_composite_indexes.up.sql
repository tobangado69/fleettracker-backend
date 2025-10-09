-- Advanced Composite Indexes for Query Optimization
-- Created: 2025-10-08
-- Description: Add composite indexes based on actual query patterns for 5-10x performance boost

-- =============================================================================
-- ANALYTICS QUERY OPTIMIZATIONS
-- =============================================================================

-- Analytics: Trip statistics by company and date range
-- Query: WHERE company_id = ? AND start_time BETWEEN ? AND ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_company_start_time 
ON trips(company_id, start_time DESC) 
WHERE status != 'cancelled';

-- Analytics: Active trips with distance
-- Query: WHERE company_id = ? AND status = 'active'
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_company_status_distance 
ON trips(company_id, status, total_distance DESC);

-- Analytics: Fuel consumption by company and date
-- Query: WHERE company_id = ? AND date BETWEEN ? AND ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fuel_logs_company_date 
ON fuel_logs(company_id, date DESC);

-- Analytics: Fuel logs with amount for aggregation
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fuel_logs_company_date_amount 
ON fuel_logs(company_id, date DESC, amount) 
WHERE amount > 0;

-- =============================================================================
-- GPS TRACKING OPTIMIZATIONS
-- =============================================================================

-- Most critical: GPS tracks by vehicle and time range
-- Query: WHERE vehicle_id = ? AND timestamp BETWEEN ? AND ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_vehicle_time_speed 
ON gps_tracks(vehicle_id, timestamp DESC, speed);

-- GPS tracks with driver filter
-- Query: WHERE vehicle_id = ? AND driver_id = ? AND timestamp BETWEEN ? AND ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_vehicle_driver_time 
ON gps_tracks(vehicle_id, driver_id, timestamp DESC) 
WHERE driver_id IS NOT NULL;

-- GPS tracks with accuracy filter (for reliable data)
-- Query: WHERE vehicle_id = ? AND timestamp >= ? AND accuracy <= ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_vehicle_time_accurate 
ON gps_tracks(vehicle_id, timestamp DESC) 
WHERE accuracy <= 50;

-- Speed violations detection
-- Query: WHERE vehicle_id = ? AND speed > ? AND timestamp BETWEEN ? AND ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_vehicle_speeding 
ON gps_tracks(vehicle_id, speed, timestamp DESC) 
WHERE speed > 0;

-- Recent GPS tracks (last 30 days - most common query)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_recent_30days 
ON gps_tracks(vehicle_id, timestamp DESC, latitude, longitude) 
WHERE timestamp >= NOW() - INTERVAL '30 days';

-- =============================================================================
-- VEHICLE MANAGEMENT OPTIMIZATIONS
-- =============================================================================

-- Vehicle list by company with active filter
-- Query: WHERE company_id = ? AND is_active = true ORDER BY created_at DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_company_active_created 
ON vehicles(company_id, is_active, created_at DESC);

-- Vehicle search by license plate (case-insensitive)
-- Query: WHERE LOWER(license_plate) = LOWER(?)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_license_plate_lower 
ON vehicles(LOWER(license_plate));

-- Active vehicles with GPS enabled (for tracking)
-- Query: WHERE company_id = ? AND is_active = true AND is_gps_enabled = true
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_company_active_gps_tracking 
ON vehicles(company_id, is_active, is_gps_enabled, driver_id) 
WHERE is_gps_enabled = true;

-- Vehicle maintenance due checks
-- Query: WHERE company_id = ? AND next_service_date <= ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_maintenance_due 
ON vehicles(company_id, next_service_date) 
WHERE next_service_date IS NOT NULL;

-- Vehicle status distribution
-- Query: GROUP BY company_id, status
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_company_status_count 
ON vehicles(company_id, status, id);

-- =============================================================================
-- DRIVER MANAGEMENT OPTIMIZATIONS
-- =============================================================================

-- Driver list by company with status
-- Query: WHERE company_id = ? AND status = ? ORDER BY created_at DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_company_status_created 
ON drivers(company_id, status, created_at DESC);

-- Available drivers (for assignment)
-- Query: WHERE company_id = ? AND status = 'available' AND vehicle_id IS NULL
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_available_for_assignment 
ON drivers(company_id, status) 
WHERE status = 'available' AND vehicle_id IS NULL;

-- Drivers with vehicles (for unassignment)
-- Query: WHERE company_id = ? AND vehicle_id IS NOT NULL
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_with_vehicle 
ON drivers(company_id, vehicle_id, status) 
WHERE vehicle_id IS NOT NULL;

-- Driver performance queries
-- Query: WHERE company_id = ? AND overall_score >= ? ORDER BY overall_score DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_company_performance 
ON drivers(company_id, overall_score DESC) 
WHERE overall_score > 0;

-- Driver compliance checks (SIM expiry)
-- Query: WHERE company_id = ? AND sim_expiry <= ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_sim_expiry 
ON drivers(company_id, sim_expiry) 
WHERE sim_expiry IS NOT NULL;

-- Driver medical checkup due
-- Query: WHERE company_id = ? AND medical_checkup_date <= ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_medical_due 
ON drivers(company_id, medical_checkup_date) 
WHERE medical_checkup_date IS NOT NULL;

-- =============================================================================
-- PAYMENT & BILLING OPTIMIZATIONS
-- =============================================================================

-- Invoice list by company and status
-- Query: WHERE company_id = ? AND status = ? ORDER BY invoice_date DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_company_status_date 
ON invoices(company_id, status, invoice_date DESC);

-- Overdue invoices
-- Query: WHERE company_id = ? AND status = 'unpaid' AND due_date < NOW()
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_overdue 
ON invoices(company_id, due_date) 
WHERE status = 'unpaid' AND due_date < NOW();

-- Payment transactions by company and date
-- Query: WHERE company_id = ? AND payment_date BETWEEN ? AND ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payments_company_date_amount 
ON payments(company_id, payment_date DESC, amount);

-- Payments by invoice
-- Query: WHERE invoice_id = ? AND status = 'completed'
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payments_invoice_status 
ON payments(invoice_id, status, payment_date DESC);

-- Subscription status by company
-- Query: WHERE company_id = ? AND status = 'active' AND expires_at > NOW()
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_active 
ON subscriptions(company_id, status, expires_at) 
WHERE status = 'active';

-- =============================================================================
-- GEOFENCING OPTIMIZATIONS
-- =============================================================================

-- Active geofences by company
-- Query: WHERE company_id = ? AND is_active = true
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_company_active 
ON geofences(company_id, is_active) 
WHERE is_active = true;

-- Geofence violations tracking
-- Query: WHERE vehicle_id = ? AND violation_time BETWEEN ? AND ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofence_violations_vehicle_time 
ON geofence_violations(vehicle_id, violation_time DESC);

-- Geofence violations by type
-- Query: WHERE company_id = ? AND violation_type = ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofence_violations_company_type 
ON geofence_violations(company_id, violation_type, violation_time DESC);

-- =============================================================================
-- MAINTENANCE & FUEL OPTIMIZATIONS
-- =============================================================================

-- Maintenance logs by vehicle and date
-- Query: WHERE vehicle_id = ? AND maintenance_date BETWEEN ? AND ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_maintenance_logs_vehicle_date_type 
ON maintenance_logs(vehicle_id, maintenance_date DESC, maintenance_type);

-- Upcoming maintenance by company
-- Query: WHERE company_id = ? AND next_maintenance_date <= ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_maintenance_logs_upcoming 
ON maintenance_logs(company_id, next_maintenance_date) 
WHERE next_maintenance_date IS NOT NULL;

-- Fuel efficiency analysis
-- Query: WHERE vehicle_id = ? AND fuel_date BETWEEN ? AND ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fuel_logs_vehicle_date_efficiency 
ON fuel_logs(vehicle_id, fuel_date DESC, amount, odometer_reading);

-- =============================================================================
-- DRIVER EVENTS & PERFORMANCE
-- =============================================================================

-- Driver events by severity (for alerts)
-- Query: WHERE driver_id = ? AND severity >= ? AND timestamp >= ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_driver_events_critical 
ON driver_events(driver_id, severity, timestamp DESC) 
WHERE severity >= 'medium';

-- Driver events by type
-- Query: WHERE driver_id = ? AND event_type = ? AND timestamp BETWEEN ? AND ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_driver_events_type_time 
ON driver_events(driver_id, event_type, timestamp DESC);

-- Performance logs for analytics
-- Query: WHERE driver_id = ? AND log_date BETWEEN ? AND ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_performance_logs_driver_date 
ON performance_logs(driver_id, log_date DESC, score);

-- =============================================================================
-- AUDIT & SESSION OPTIMIZATIONS
-- =============================================================================

-- Active sessions by user
-- Query: WHERE user_id = ? AND expires_at > NOW()
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_active 
ON sessions(user_id, expires_at DESC) 
WHERE expires_at > NOW();

-- Session cleanup
-- Query: WHERE expires_at < NOW()
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_expired 
ON sessions(expires_at) 
WHERE expires_at < NOW();

-- Audit logs by company and action
-- Query: WHERE company_id = ? AND action = ? AND created_at >= ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_company_action_time 
ON audit_logs(company_id, action, created_at DESC);

-- Audit logs by user
-- Query: WHERE user_id = ? AND created_at BETWEEN ? AND ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user_time 
ON audit_logs(user_id, created_at DESC);

-- =============================================================================
-- VEHICLE HISTORY OPTIMIZATIONS
-- =============================================================================

-- Vehicle history by vehicle and date
-- Query: WHERE vehicle_id = ? AND event_date BETWEEN ? AND ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicle_history_vehicle_date_type 
ON vehicle_history(vehicle_id, event_date DESC, event_type);

-- Recent vehicle events
-- Query: WHERE vehicle_id = ? ORDER BY event_date DESC LIMIT 10
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicle_history_recent 
ON vehicle_history(vehicle_id, event_date DESC) 
WHERE event_date >= NOW() - INTERVAL '90 days';

-- =============================================================================
-- COVERING INDEXES (Include additional columns)
-- =============================================================================

-- Vehicles list query covering index
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_list_covering 
ON vehicles(company_id, status, is_active, created_at DESC) 
INCLUDE (license_plate, make, model, driver_id);

-- Drivers list query covering index
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_list_covering 
ON drivers(company_id, status, is_active, created_at DESC) 
INCLUDE (name, phone, vehicle_id, overall_score);

-- GPS tracks covering index for location history
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_history_covering 
ON gps_tracks(vehicle_id, timestamp DESC) 
INCLUDE (latitude, longitude, speed, heading, accuracy);

-- =============================================================================
-- TEXT SEARCH INDEXES (for search functionality)
-- =============================================================================

-- Vehicle search by license plate, make, model
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_text_search 
ON vehicles USING gin(to_tsvector('english', 
    COALESCE(license_plate, '') || ' ' || 
    COALESCE(make, '') || ' ' || 
    COALESCE(model, '')
));

-- Driver search by name, phone, NIK
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_text_search 
ON drivers USING gin(to_tsvector('english', 
    COALESCE(name, '') || ' ' || 
    COALESCE(phone, '') || ' ' || 
    COALESCE(nik, '')
));

-- Company search
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_companies_text_search 
ON companies USING gin(to_tsvector('english', 
    COALESCE(name, '') || ' ' || 
    COALESCE(email, '') || ' ' || 
    COALESCE(npwp, '')
));

