-- Rollback Partial Indexes

-- Active/Status partial indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_active_only;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_available;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_active_only;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_unassigned;
DROP INDEX CONCURRENTLY IF EXISTS idx_geofences_active_only;

-- Time-based partial indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_last_24h;
DROP INDEX CONCURRENTLY IF EXISTS idx_trips_last_7days;
DROP INDEX CONCURRENTLY IF EXISTS idx_trips_ongoing;
DROP INDEX CONCURRENTLY IF EXISTS idx_driver_events_last_30days;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_upcoming_maintenance;

-- Compliance & expiry partial indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_sim_expiring_soon;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_sim_expired;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_inspection_due;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_insurance_expiring;

-- Payment status partial indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_invoices_unpaid;
DROP INDEX CONCURRENTLY IF EXISTS idx_invoices_overdue_critical;
DROP INDEX CONCURRENTLY IF EXISTS idx_payments_pending;
DROP INDEX CONCURRENTLY IF EXISTS idx_payments_failed;

-- Performance & safety partial indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_speeding_violations;
DROP INDEX CONCURRENTLY IF EXISTS idx_driver_events_harsh_braking;
DROP INDEX CONCURRENTLY IF EXISTS idx_driver_events_rapid_acceleration;
DROP INDEX CONCURRENTLY IF EXISTS idx_driver_events_critical_only;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_low_performance;

-- Fuel efficiency partial indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_fuel_logs_abnormal;
DROP INDEX CONCURRENTLY IF EXISTS idx_fuel_logs_with_efficiency;

-- Trip status partial indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_trips_incomplete;
DROP INDEX CONCURRENTLY IF EXISTS idx_trips_long_duration;

-- User & authentication partial indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_users_active_only;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_locked;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_verified;

-- Soft delete optimization
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_not_deleted;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_not_deleted;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_not_deleted;
DROP INDEX CONCURRENTLY IF EXISTS idx_companies_not_deleted;

