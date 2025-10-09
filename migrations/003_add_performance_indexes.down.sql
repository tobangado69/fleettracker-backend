-- Remove performance indexes
-- This migration removes all indexes added in the up migration

-- Vehicle table indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_company_id_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_company_id_is_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_company_id_driver_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_make_model;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_year;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_fuel_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_license_plate_lower;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_vin;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_inspection_date;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_last_updated_at;

-- GPS tracks table indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_vehicle_id_timestamp;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_driver_id_timestamp;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_trip_id_timestamp;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_timestamp;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_vehicle_timestamp_range;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_speed;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_accuracy;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_moving;

-- Driver table indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_company_id_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_company_id_is_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_vehicle_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_nik;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_sim_number;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_medical_checkup_date;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_training_expiry;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_performance_score;

-- Trip table indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_trips_vehicle_id_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_trips_driver_id_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_trips_company_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_trips_start_time;
DROP INDEX CONCURRENTLY IF EXISTS idx_trips_end_time;
DROP INDEX CONCURRENTLY IF EXISTS idx_trips_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_trips_vehicle_start_end;

-- Driver events table indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_driver_events_driver_id_timestamp;
DROP INDEX CONCURRENTLY IF EXISTS idx_driver_events_vehicle_id_timestamp;
DROP INDEX CONCURRENTLY IF EXISTS idx_driver_events_event_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_driver_events_severity;
DROP INDEX CONCURRENTLY IF EXISTS idx_driver_events_timestamp;

-- Payment and invoice indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_payments_company_id_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_payments_invoice_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_payments_payment_date;
DROP INDEX CONCURRENTLY IF EXISTS idx_payments_payment_method;

DROP INDEX CONCURRENTLY IF EXISTS idx_invoices_company_id_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_invoices_invoice_date;
DROP INDEX CONCURRENTLY IF EXISTS idx_invoices_due_date;
DROP INDEX CONCURRENTLY IF EXISTS idx_invoices_invoice_number;

-- User and session indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_users_company_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_email;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_role;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_is_active;

DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_token;
DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_expires_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_user_expires;

-- Maintenance and fuel logs indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_maintenance_logs_vehicle_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_maintenance_logs_company_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_maintenance_logs_maintenance_date;
DROP INDEX CONCURRENTLY IF EXISTS idx_maintenance_logs_type;

DROP INDEX CONCURRENTLY IF EXISTS idx_fuel_logs_vehicle_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_fuel_logs_company_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_fuel_logs_fuel_date;

-- Geofence indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_geofences_company_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_geofences_is_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_geofences_type;

-- Vehicle history indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicle_history_vehicle_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicle_history_company_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicle_history_event_date;
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicle_history_event_type;

-- Performance log indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_performance_logs_driver_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_performance_logs_vehicle_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_performance_logs_log_date;

-- Audit log indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_logs_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_logs_company_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_logs_action;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_logs_created_at;

-- Password reset token indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_password_reset_tokens_token;
DROP INDEX CONCURRENTLY IF EXISTS idx_password_reset_tokens_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_password_reset_tokens_expires_at;

-- Company indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_companies_npwp;
DROP INDEX CONCURRENTLY IF EXISTS idx_companies_is_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_companies_status;

-- Subscription indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_subscriptions_company_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_subscriptions_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_subscriptions_expires_at;

-- Composite indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_company_active_gps;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_vehicle_timestamp_accuracy;
DROP INDEX CONCURRENTLY IF EXISTS idx_drivers_company_active_vehicle;
