-- FleetTracker Pro Initial Schema Migration
-- Created: 2025-10-04
-- Description: Create all core tables for Indonesian fleet management

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS postgis;

-- =============================================================================
-- CORE TABLES
-- =============================================================================

-- Companies table (Multi-tenant fleet management companies)
CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    address TEXT,
    city VARCHAR(100),
    province VARCHAR(100),
    postal_code VARCHAR(10),
    country VARCHAR(100) DEFAULT 'Indonesia',
    
    -- Indonesian Compliance Fields
    npwp VARCHAR(20) UNIQUE,           -- Tax ID (Nomor Pokok Wajib Pajak)
    siup VARCHAR(50),                  -- Business License (Surat Izin Usaha Perdagangan)
    skt VARCHAR(50),                   -- Tax Certificate
    pkp BOOLEAN DEFAULT FALSE,         -- VAT Registered (Pengusaha Kena Pajak)
    company_type VARCHAR(50),          -- PT, CV, UD, etc.
    
    -- Business Information
    industry VARCHAR(100),
    fleet_size INTEGER DEFAULT 0,
    max_vehicles INTEGER DEFAULT 100,
    subscription_tier VARCHAR(50) DEFAULT 'basic',
    
    -- Status and Settings
    status VARCHAR(20) DEFAULT 'active',
    is_active BOOLEAN DEFAULT TRUE,
    settings JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Users table (System users: admin, manager, operator)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    
    -- Personal Information
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    avatar VARCHAR(500),
    
    -- Indonesian Fields
    nik VARCHAR(16) UNIQUE,            -- Indonesian ID (Nomor Induk Kependudukan)
    address TEXT,
    city VARCHAR(100),
    province VARCHAR(100),
    postal_code VARCHAR(10),
    
    -- Role and Permissions
    role VARCHAR(50) NOT NULL DEFAULT 'operator',  -- admin, manager, operator
    permissions JSONB DEFAULT '{}',
    
    -- Account Status
    status VARCHAR(20) DEFAULT 'active',
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMPTZ,
    
    -- Security
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ DEFAULT NOW(),
    must_change_password BOOLEAN DEFAULT FALSE,
    
    -- Preferences
    language VARCHAR(10) DEFAULT 'id',
    timezone VARCHAR(50) DEFAULT 'Asia/Jakarta',
    notifications_enabled BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Sessions table (JWT session management)
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    refresh_token TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    refresh_expires_at TIMESTAMPTZ,
    ip_address INET,
    user_agent TEXT,
    device_info JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    last_accessed_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Audit logs table (Activity tracking)
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100),
    entity_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Password reset tokens
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- =============================================================================
-- FLEET MANAGEMENT TABLES
-- =============================================================================

-- Vehicles table
CREATE TABLE IF NOT EXISTS vehicles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    license_plate VARCHAR(20) NOT NULL UNIQUE,
    manufacturer VARCHAR(100),
    model VARCHAR(100),
    year INTEGER,
    color VARCHAR(50),
    vehicle_type VARCHAR(50) NOT NULL,  -- truck, van, car, motorcycle
    fuel_type VARCHAR(20) DEFAULT 'gasoline',
    fuel_capacity DECIMAL(8,2),
    
    -- Device Information
    device_id VARCHAR(100) UNIQUE,
    device_imei VARCHAR(20),
    sim_card_number VARCHAR(20),
    
    -- Status and Ownership
    status VARCHAR(20) DEFAULT 'active',  -- active, inactive, maintenance, retired
    ownership_type VARCHAR(20) DEFAULT 'owned',  -- owned, leased, rented
    purchase_date DATE,
    purchase_price DECIMAL(15,2),
    
    -- Insurance
    insurance_company VARCHAR(255),
    insurance_policy_number VARCHAR(100),
    insurance_expiry_date DATE,
    
    -- Registration (Indonesian STNK/BPKB)
    stnk_number VARCHAR(50),
    stnk_expiry_date DATE,
    bpkb_number VARCHAR(50),
    kir_expiry_date DATE,  -- KIR (vehicle worthiness test)
    
    -- Technical Details
    engine_number VARCHAR(50),
    chassis_number VARCHAR(50),
    odometer_reading DECIMAL(10,2) DEFAULT 0,
    last_service_date DATE,
    next_service_date DATE,
    next_service_odometer DECIMAL(10,2),
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Vehicle maintenance logs
CREATE TABLE IF NOT EXISTS maintenance_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    maintenance_type VARCHAR(50) NOT NULL,  -- service, repair, inspection
    description TEXT,
    cost DECIMAL(12,2),
    odometer_reading DECIMAL(10,2),
    performed_by VARCHAR(255),
    workshop_name VARCHAR(255),
    next_service_date DATE,
    parts_replaced JSONB DEFAULT '[]',
    labor_hours DECIMAL(5,2),
    notes TEXT,
    maintenance_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Vehicle fuel logs
CREATE TABLE IF NOT EXISTS fuel_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    fuel_type VARCHAR(20) NOT NULL,
    quantity DECIMAL(8,2) NOT NULL,  -- liters
    cost DECIMAL(12,2) NOT NULL,     -- IDR
    price_per_liter DECIMAL(8,2),
    odometer_reading DECIMAL(10,2),
    location VARCHAR(255),
    gas_station VARCHAR(255),
    filled_by VARCHAR(255),
    payment_method VARCHAR(50),
    receipt_number VARCHAR(100),
    notes TEXT,
    fuel_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Drivers table
CREATE TABLE IF NOT EXISTS drivers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- Personal Information
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20) NOT NULL,
    date_of_birth DATE,
    photo VARCHAR(500),
    
    -- Indonesian ID
    nik VARCHAR(16) UNIQUE NOT NULL,
    address TEXT,
    city VARCHAR(100),
    province VARCHAR(100),
    postal_code VARCHAR(10),
    
    -- Driver License (SIM)
    license_number VARCHAR(50) NOT NULL,
    license_type VARCHAR(10) NOT NULL,  -- A, B1, B2, C
    license_expiry_date DATE NOT NULL,
    
    -- Employment
    hire_date DATE,
    employment_type VARCHAR(20) DEFAULT 'permanent',  -- permanent, contract, freelance
    salary DECIMAL(15,2),
    bank_account VARCHAR(50),
    bank_name VARCHAR(100),
    
    -- Health & Safety
    blood_type VARCHAR(5),
    emergency_contact_name VARCHAR(255),
    emergency_contact_phone VARCHAR(20),
    emergency_contact_relationship VARCHAR(50),
    medical_checkup_date DATE,
    medical_checkup_expiry DATE,
    
    -- Performance
    total_trips INTEGER DEFAULT 0,
    total_distance DECIMAL(12,2) DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0,
    violations_count INTEGER DEFAULT 0,
    
    -- Status
    status VARCHAR(20) DEFAULT 'active',
    is_available BOOLEAN DEFAULT TRUE,
    current_vehicle_id UUID REFERENCES vehicles(id) ON DELETE SET NULL,
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Driver events (violations, achievements)
CREATE TABLE IF NOT EXISTS driver_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,  -- violation, achievement, training
    severity VARCHAR(20),  -- low, medium, high, critical
    description TEXT,
    points INTEGER DEFAULT 0,
    location VARCHAR(255),
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    event_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Driver performance logs
CREATE TABLE IF NOT EXISTS performance_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    total_trips INTEGER DEFAULT 0,
    total_distance DECIMAL(12,2) DEFAULT 0,
    total_duration INTEGER DEFAULT 0,  -- minutes
    average_speed DECIMAL(5,2),
    max_speed DECIMAL(5,2),
    harsh_braking_count INTEGER DEFAULT 0,
    rapid_acceleration_count INTEGER DEFAULT 0,
    speeding_count INTEGER DEFAULT 0,
    idle_time INTEGER DEFAULT 0,  -- minutes
    fuel_efficiency DECIMAL(6,3),
    safety_score DECIMAL(5,2),
    punctuality_score DECIMAL(5,2),
    overall_rating DECIMAL(3,2),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- =============================================================================
-- GPS TRACKING TABLES
-- =============================================================================

-- GPS tracking data
CREATE TABLE IF NOT EXISTS gps_tracks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    driver_id UUID REFERENCES drivers(id) ON DELETE SET NULL,
    trip_id UUID,
    
    -- GPS Coordinates
    latitude DECIMAL(10,8) NOT NULL,
    longitude DECIMAL(11,8) NOT NULL,
    altitude DECIMAL(8,2),
    heading DECIMAL(5,2),  -- degrees 0-360
    speed DECIMAL(5,2),    -- km/h
    
    -- Location Information
    location VARCHAR(255),
    address TEXT,
    city VARCHAR(100),
    province VARCHAR(100),
    country VARCHAR(100) DEFAULT 'Indonesia',
    
    -- Data Quality
    accuracy DECIMAL(5,2),  -- meters
    satellites INTEGER DEFAULT 0,
    hdop DECIMAL(3,1),      -- Horizontal Dilution of Precision
    
    -- Vehicle Status
    ignition_on BOOLEAN DEFAULT FALSE,
    engine_on BOOLEAN DEFAULT FALSE,
    moving BOOLEAN DEFAULT FALSE,
    idle_time INTEGER DEFAULT 0,  -- seconds
    
    -- Fuel Information
    fuel_level DECIMAL(5,2),
    fuel_consumption DECIMAL(8,2),
    
    -- Distance and Odometer
    distance DECIMAL(10,2),
    odometer DECIMAL(10,2),
    
    -- Additional Data
    battery_voltage DECIMAL(4,2),
    gsm_signal INTEGER,
    gps_quality INTEGER,
    event_data JSONB DEFAULT '{}',
    
    -- Timestamp
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Trips table
CREATE TABLE IF NOT EXISTS trips (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    driver_id UUID REFERENCES drivers(id) ON DELETE SET NULL,
    
    -- Trip Details
    trip_number VARCHAR(50),
    trip_type VARCHAR(50),  -- delivery, pickup, service, personal
    purpose TEXT,
    
    -- Location
    start_location VARCHAR(255),
    start_latitude DECIMAL(10,8),
    start_longitude DECIMAL(11,8),
    end_location VARCHAR(255),
    end_latitude DECIMAL(10,8),
    end_longitude DECIMAL(11,8),
    
    -- Time
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    planned_duration INTEGER,  -- minutes
    actual_duration INTEGER,   -- minutes
    
    -- Distance and Fuel
    distance DECIMAL(10,2),
    start_odometer DECIMAL(10,2),
    end_odometer DECIMAL(10,2),
    average_speed DECIMAL(5,2),
    max_speed DECIMAL(5,2),
    idle_time INTEGER DEFAULT 0,
    
    -- Fuel Information
    fuel_consumed DECIMAL(8,2) DEFAULT 0,
    fuel_efficiency DECIMAL(5,2) DEFAULT 0,
    start_fuel_level DECIMAL(5,2),
    end_fuel_level DECIMAL(5,2),
    
    -- Violations and Events
    violations INTEGER DEFAULT 0,
    harsh_braking INTEGER DEFAULT 0,
    rapid_acceleration INTEGER DEFAULT 0,
    sharp_cornering INTEGER DEFAULT 0,
    speeding_events INTEGER DEFAULT 0,
    
    -- Route Information
    route_data JSONB DEFAULT '{}',
    route_optimized BOOLEAN DEFAULT FALSE,
    
    -- Status
    status VARCHAR(20) DEFAULT 'in_progress',  -- planned, in_progress, completed, cancelled
    notes TEXT,
    weather_data JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Geofences table
CREATE TABLE IF NOT EXISTS geofences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    geofence_type VARCHAR(50) NOT NULL,  -- circle, polygon
    center_latitude DECIMAL(10,8),
    center_longitude DECIMAL(11,8),
    radius DECIMAL(10,2),  -- meters (for circle)
    polygon_points JSONB,  -- for polygon
    is_active BOOLEAN DEFAULT TRUE,
    alert_on_enter BOOLEAN DEFAULT FALSE,
    alert_on_exit BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Vehicle history table
CREATE TABLE IF NOT EXISTS vehicle_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    history_type VARCHAR(50) NOT NULL,  -- maintenance, repair, accident, inspection
    description TEXT NOT NULL,
    cost DECIMAL(12,2),
    odometer_reading DECIMAL(10,2),
    performed_by VARCHAR(255),
    workshop_name VARCHAR(255),
    parts_replaced JSONB DEFAULT '[]',
    labor_hours DECIMAL(5,2),
    severity VARCHAR(20),
    downtime_hours INTEGER,
    notes TEXT,
    attachments JSONB DEFAULT '[]',
    history_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- =============================================================================
-- PAYMENT & SUBSCRIPTION TABLES
-- =============================================================================

-- Subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    plan_name VARCHAR(100) NOT NULL,
    plan_tier VARCHAR(50) NOT NULL,  -- basic, professional, enterprise
    billing_cycle VARCHAR(20) NOT NULL,  -- monthly, yearly
    price DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'IDR',
    max_vehicles INTEGER NOT NULL,
    max_drivers INTEGER NOT NULL,
    features JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',  -- active, cancelled, expired, suspended
    start_date DATE NOT NULL,
    end_date DATE,
    auto_renew BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Payments table
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
    invoice_id UUID,
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'IDR',
    payment_method VARCHAR(50) NOT NULL,  -- bank_transfer, qris, credit_card
    payment_status VARCHAR(20) DEFAULT 'pending',  -- pending, completed, failed, refunded
    payment_date TIMESTAMPTZ,
    transaction_id VARCHAR(255),
    payment_proof VARCHAR(500),
    bank_name VARCHAR(100),
    account_number VARCHAR(50),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Invoices table
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
    invoice_number VARCHAR(100) UNIQUE NOT NULL,
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    tax_amount DECIMAL(12,2) DEFAULT 0,
    total_amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'IDR',
    status VARCHAR(20) DEFAULT 'unpaid',  -- unpaid, paid, overdue, cancelled
    paid_date DATE,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- =============================================================================
-- INDEXES FOR PERFORMANCE
-- =============================================================================

-- Companies
CREATE INDEX IF NOT EXISTS idx_companies_status ON companies(status) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_companies_npwp ON companies(npwp) WHERE npwp IS NOT NULL;

-- Users
CREATE INDEX IF NOT EXISTS idx_users_company ON users(company_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status) WHERE status = 'active';

-- Sessions
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

-- Vehicles
CREATE INDEX IF NOT EXISTS idx_vehicles_company ON vehicles(company_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_license ON vehicles(license_plate);
CREATE INDEX IF NOT EXISTS idx_vehicles_status ON vehicles(status);
CREATE INDEX IF NOT EXISTS idx_vehicles_device ON vehicles(device_id) WHERE device_id IS NOT NULL;

-- Drivers
CREATE INDEX IF NOT EXISTS idx_drivers_company ON drivers(company_id);
CREATE INDEX IF NOT EXISTS idx_drivers_status ON drivers(status);
CREATE INDEX IF NOT EXISTS idx_drivers_vehicle ON drivers(current_vehicle_id) WHERE current_vehicle_id IS NOT NULL;

-- GPS Tracks
CREATE INDEX IF NOT EXISTS idx_gps_vehicle_time ON gps_tracks(vehicle_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_gps_driver_time ON gps_tracks(driver_id, timestamp DESC) WHERE driver_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_gps_trip ON gps_tracks(trip_id) WHERE trip_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_gps_timestamp ON gps_tracks(timestamp DESC);

-- Trips
CREATE INDEX IF NOT EXISTS idx_trips_company ON trips(company_id);
CREATE INDEX IF NOT EXISTS idx_trips_vehicle ON trips(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_trips_driver ON trips(driver_id) WHERE driver_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_trips_status ON trips(status);
CREATE INDEX IF NOT EXISTS idx_trips_start_time ON trips(start_time DESC);

-- Audit Logs
CREATE INDEX IF NOT EXISTS idx_audit_company ON audit_logs(company_id);
CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_logs(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at DESC);

-- Success message
DO $$
BEGIN
    RAISE NOTICE '✅ FleetTracker Pro initial schema created successfully!';
    RAISE NOTICE '✅ All tables, indexes, and constraints are in place';
    RAISE NOTICE '🇮🇩 Ready for Indonesian fleet management data';
END
$$;

