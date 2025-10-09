-- Rollback Advanced Composite Indexes

-- Analytics indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_trips_company_start_time;
DROP INDEX CONCURRENTLY IF EXISTS idx_trips_company_status_distance;
DROP INDEX CONCURRENTLY IF EXISTS idx_fuel_logs_company_date;
DROP INDEX CONCURRENTLY IF EXISTS idx_fuel_logs_company_date_amount;

-- GPS tracking indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_vehicle_time_speed;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_vehicle_driver_time;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_vehicle_time_accurate;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_vehicle_speeding;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_recent_30days;

-- Vehicle management indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_company_active_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_license_plate_lower;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_company_active_gps_tracking;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_maintenance_due;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_company_status_count;

-- Driver management indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_company_status_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_available_for_assignment;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_with_vehicle;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_company_performance;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_sim_expiry;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_medical_due;

-- Payment & billing indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_invoices_company_status_date;
DROP INDEX CONCURRENTLY IF EXISTS idx_invoices_overdue;
DROP INDEX CONCURRENTLY IF EXISTS idx_payments_company_date_amount;
DROP INDEX CONCURRENTLY IF EXISTS idx_payments_invoice_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_subscriptions_active;

-- Geofencing indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_geofences_company_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_geofence_violations_vehicle_time;
DROP INDEX CONCURRENTLY IF EXISTS idx_geofence_violations_company_type;

-- Maintenance & fuel indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_maintenance_logs_vehicle_date_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_maintenance_logs_upcoming;
DROP INDEX CONCURRENTLY IF EXISTS idx_fuel_logs_vehicle_date_efficiency;

-- Driver events & performance indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_driver_events_critical;
DROP INDEX CONCURRENTLY IF EXISTS idx_driver_events_type_time;
DROP INDEX CONCURRENTLY IF EXISTS idx_performance_logs_driver_date;

-- Audit & session indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_expired;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_logs_company_action_time;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_logs_user_time;

-- Vehicle history indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicle_history_vehicle_date_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicle_history_recent;

-- Covering indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_list_covering;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_list_covering;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_history_covering;

-- Text search indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_text_search;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_text_search;
DROP INDEX CONCURRENTLY IF EXISTS idx_companies_text_search;

