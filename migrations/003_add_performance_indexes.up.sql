-- Add performance indexes for better query optimization
-- This migration adds indexes for frequently queried fields

-- Vehicle table indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_company_id_status ON vehicles(company_id, status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_company_id_is_active ON vehicles(company_id, is_active);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_company_id_driver_id ON vehicles(company_id, driver_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_make_model ON vehicles(make, model);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_year ON vehicles(year);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_fuel_type ON vehicles(fuel_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_license_plate_lower ON vehicles(LOWER(license_plate));
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_vin ON vehicles(vin);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_inspection_date ON vehicles(inspection_date) WHERE inspection_date IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_last_updated_at ON vehicles(last_updated_at);

-- GPS tracks table indexes (most critical for performance)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_vehicle_id_timestamp ON gps_tracks(vehicle_id, timestamp DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_driver_id_timestamp ON gps_tracks(driver_id, timestamp DESC) WHERE driver_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_trip_id_timestamp ON gps_tracks(trip_id, timestamp DESC) WHERE trip_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_timestamp ON gps_tracks(timestamp DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_vehicle_timestamp_range ON gps_tracks(vehicle_id, timestamp) WHERE timestamp >= NOW() - INTERVAL '30 days';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_speed ON gps_tracks(speed) WHERE speed > 0;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_accuracy ON gps_tracks(accuracy) WHERE accuracy > 0;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_moving ON gps_tracks(moving, timestamp) WHERE moving = true;

-- Driver table indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_company_id_status ON drivers(company_id, status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_company_id_is_active ON drivers(company_id, is_active);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_vehicle_id ON drivers(vehicle_id) WHERE vehicle_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_nik ON drivers(nik);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_sim_number ON drivers(sim_number);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_medical_checkup_date ON drivers(medical_checkup_date) WHERE medical_checkup_date IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_training_expiry ON drivers(training_expiry) WHERE training_expiry IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_performance_score ON drivers(overall_score) WHERE overall_score > 0;

-- Trip table indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_vehicle_id_status ON trips(vehicle_id, status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_driver_id_status ON trips(driver_id, status) WHERE driver_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_company_id ON trips(company_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_start_time ON trips(start_time);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_end_time ON trips(end_time);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_status ON trips(status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_vehicle_start_end ON trips(vehicle_id, start_time, end_time);

-- Driver events table indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_driver_events_driver_id_timestamp ON driver_events(driver_id, timestamp DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_driver_events_vehicle_id_timestamp ON driver_events(vehicle_id, timestamp DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_driver_events_event_type ON driver_events(event_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_driver_events_severity ON driver_events(severity);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_driver_events_timestamp ON driver_events(timestamp DESC);

-- Payment and invoice indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payments_company_id_status ON payments(company_id, status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payments_invoice_id ON payments(invoice_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payments_payment_date ON payments(payment_date);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payments_payment_method ON payments(payment_method);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_company_id_status ON invoices(company_id, status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_invoice_date ON invoices(invoice_date);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_due_date ON invoices(due_date);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_invoice_number ON invoices(invoice_number);

-- User and session indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_company_id ON users(company_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_is_active ON users(is_active);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_user_expires ON sessions(user_id, expires_at);

-- Maintenance and fuel logs indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_maintenance_logs_vehicle_id ON maintenance_logs(vehicle_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_maintenance_logs_company_id ON maintenance_logs(company_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_maintenance_logs_maintenance_date ON maintenance_logs(maintenance_date);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_maintenance_logs_type ON maintenance_logs(maintenance_type);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fuel_logs_vehicle_id ON fuel_logs(vehicle_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fuel_logs_company_id ON fuel_logs(company_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fuel_logs_fuel_date ON fuel_logs(fuel_date);

-- Geofence indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_company_id ON geofences(company_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_is_active ON geofences(is_active);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_type ON geofences(type);

-- Vehicle history indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicle_history_vehicle_id ON vehicle_history(vehicle_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicle_history_company_id ON vehicle_history(company_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicle_history_event_date ON vehicle_history(event_date);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicle_history_event_type ON vehicle_history(event_type);

-- Performance log indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_performance_logs_driver_id ON performance_logs(driver_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_performance_logs_vehicle_id ON performance_logs(vehicle_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_performance_logs_log_date ON performance_logs(log_date);

-- Audit log indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_company_id ON audit_logs(company_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Password reset token indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_password_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);

-- Company indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_companies_npwp ON companies(npwp);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_companies_is_active ON companies(is_active);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_companies_status ON companies(status);

-- Subscription indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_company_id ON subscriptions(company_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_expires_at ON subscriptions(expires_at);

-- Composite indexes for common query patterns
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_company_active_gps ON vehicles(company_id, is_active, is_gps_enabled);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_vehicle_timestamp_accuracy ON gps_tracks(vehicle_id, timestamp DESC, accuracy) WHERE accuracy <= 50;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_company_active_vehicle ON drivers(company_id, is_active, vehicle_id) WHERE vehicle_id IS NOT NULL;
