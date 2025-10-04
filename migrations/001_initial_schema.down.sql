-- FleetTracker Pro Initial Schema Rollback
-- Created: 2025-10-04
-- Description: Drop all tables in reverse dependency order

-- Drop tables in reverse order of dependencies

-- Payment & Subscription tables
DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS subscriptions CASCADE;

-- GPS Tracking tables
DROP TABLE IF EXISTS vehicle_history CASCADE;
DROP TABLE IF EXISTS geofences CASCADE;
DROP TABLE IF EXISTS trips CASCADE;
DROP TABLE IF EXISTS gps_tracks CASCADE;

-- Driver tables
DROP TABLE IF EXISTS performance_logs CASCADE;
DROP TABLE IF EXISTS driver_events CASCADE;
DROP TABLE IF EXISTS drivers CASCADE;

-- Vehicle tables
DROP TABLE IF EXISTS fuel_logs CASCADE;
DROP TABLE IF EXISTS maintenance_logs CASCADE;
DROP TABLE IF EXISTS vehicles CASCADE;

-- User & Auth tables
DROP TABLE IF NOT EXISTS password_reset_tokens CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Core tables
DROP TABLE IF EXISTS companies CASCADE;

-- Drop extensions (only if not used by other databases)
-- DROP EXTENSION IF EXISTS postgis;
-- DROP EXTENSION IF EXISTS "uuid-ossp";

-- Success message
DO $$
BEGIN
    RAISE NOTICE '🔄 FleetTracker Pro schema rolled back successfully';
    RAISE NOTICE '✅ All tables dropped';
END
$$;

