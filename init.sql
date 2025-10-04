-- FleetTracker Pro Database Initialization
-- PostgreSQL with PostGIS for main application data
-- Indonesian Fleet Management SaaS Application

-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create database user if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'fleettracker') THEN
        CREATE ROLE fleettracker WITH LOGIN PASSWORD 'password123';
    END IF;
END
$$;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE fleettracker TO fleettracker;

-- Companies table for multi-tenancy
CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    address TEXT,
    phone VARCHAR(20),
    email VARCHAR(255),
    subscription_plan VARCHAR(50) DEFAULT 'starter',
    subscription_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    settings JSONB DEFAULT '{}',
    
    -- Indonesian specific fields
    npwp VARCHAR(20), -- Indonesian tax number
    siup VARCHAR(50), -- Indonesian business license
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    province VARCHAR(100),
    postal_code VARCHAR(10),
    country VARCHAR(2) DEFAULT 'ID'
);

-- Users table with Better Auth compatibility
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    role VARCHAR(50) NOT NULL DEFAULT 'driver',
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    profile_image TEXT,
    language VARCHAR(10) DEFAULT 'id',
    timezone VARCHAR(50) DEFAULT 'Asia/Jakarta',
    last_login TIMESTAMPTZ,
    
    -- Indonesian specific fields
    nik VARCHAR(16), -- Indonesian ID number
    birth_date DATE,
    birth_place VARCHAR(100),
    religion VARCHAR(50),
    marital_status VARCHAR(20),
    emergency_contact_name VARCHAR(255),
    emergency_contact_phone VARCHAR(20)
);

-- Sessions table for Better Auth
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    is_active BOOLEAN DEFAULT TRUE
);

-- Vehicles table
CREATE TABLE IF NOT EXISTS vehicles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    license_plate VARCHAR(20) NOT NULL,
    brand VARCHAR(100),
    model VARCHAR(100),
    year INTEGER,
    vehicle_type VARCHAR(50) NOT NULL,
    fuel_capacity DECIMAL(8,2),
    device_id VARCHAR(100), -- GPS device identifier
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Indonesian specific fields
    stnk_number VARCHAR(50), -- Indonesian vehicle registration
    bpkb_number VARCHAR(50), -- Indonesian vehicle ownership
    color VARCHAR(50),
    chassis_number VARCHAR(100),
    engine_number VARCHAR(100),
    insurance_company VARCHAR(100),
    insurance_number VARCHAR(100),
    insurance_expiry DATE,
    
    UNIQUE(company_id, license_plate)
);

-- Drivers table
CREATE TABLE IF NOT EXISTS drivers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    license_number VARCHAR(50) NOT NULL,
    license_type VARCHAR(10) NOT NULL, -- A, B, C, etc.
    license_expiry DATE,
    experience_years INTEGER DEFAULT 0,
    assigned_vehicle_id UUID REFERENCES vehicles(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Indonesian specific fields
    sim_class VARCHAR(10), -- Indonesian driving license class
    sim_expiry DATE,
    address TEXT,
    blood_type VARCHAR(5),
    medical_check_date DATE,
    medical_check_expiry DATE
);

-- Trips table
CREATE TABLE IF NOT EXISTS trips (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    start_location GEOMETRY(POINT, 4326),
    end_location GEOMETRY(POINT, 4326),
    planned_route GEOMETRY(LINESTRING, 4326),
    actual_route GEOMETRY(LINESTRING, 4326),
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    estimated_arrival TIMESTAMPTZ,
    distance_km DECIMAL(8,2),
    fuel_consumed DECIMAL(6,2),
    status VARCHAR(20) DEFAULT 'planned',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Indonesian specific fields
    purpose VARCHAR(255), -- Trip purpose
    passenger_count INTEGER DEFAULT 0,
    cargo_weight DECIMAL(8,2),
    toll_cost DECIMAL(10,2),
    parking_cost DECIMAL(10,2)
);

-- Driver events table for behavior monitoring
CREATE TABLE IF NOT EXISTS driver_events (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    trip_id UUID REFERENCES trips(id) ON DELETE SET NULL,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL, -- 'harsh_braking', 'speeding', 'rapid_acceleration'
    severity VARCHAR(20) DEFAULT 'medium', -- 'low', 'medium', 'high'
    location GEOMETRY(POINT, 4326),
    speed_at_event DECIMAL(5,2),
    timestamp TIMESTAMPTZ NOT NULL,
    metadata JSONB, -- Additional event-specific data
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    INDEX (driver_id, timestamp),
    INDEX (event_type, timestamp),
    INDEX (company_id, timestamp),
    INDEX (severity, timestamp)
);

-- Subscriptions table for Indonesian payment integration
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    plan_name VARCHAR(50) NOT NULL,
    vehicle_count INTEGER NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'IDR',
    status VARCHAR(20) DEFAULT 'active',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    payment_method VARCHAR(50), -- 'qris', 'bank_transfer', 'e_wallet'
    payment_id VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Indonesian specific fields
    invoice_number VARCHAR(100),
    tax_invoice_number VARCHAR(100),
    payment_reference VARCHAR(255)
);

-- Payments table for transaction tracking
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'IDR',
    payment_method VARCHAR(50) NOT NULL,
    payment_status VARCHAR(20) DEFAULT 'pending',
    payment_reference VARCHAR(255),
    external_payment_id VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    paid_at TIMESTAMPTZ,
    
    -- Indonesian specific fields
    bank_code VARCHAR(10), -- For bank transfers
    account_number VARCHAR(50),
    e_wallet_type VARCHAR(20), -- 'gopay', 'ovo', 'dana', 'shopeepay'
    qris_string TEXT
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_company_id ON users(company_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

CREATE INDEX IF NOT EXISTS idx_vehicles_company_id ON vehicles(company_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_license_plate ON vehicles(license_plate);
CREATE INDEX IF NOT EXISTS idx_vehicles_status ON vehicles(status);
CREATE INDEX IF NOT EXISTS idx_vehicles_device_id ON vehicles(device_id);

CREATE INDEX IF NOT EXISTS idx_drivers_company_id ON drivers(company_id);
CREATE INDEX IF NOT EXISTS idx_drivers_user_id ON drivers(user_id);
CREATE INDEX IF NOT EXISTS idx_drivers_assigned_vehicle_id ON drivers(assigned_vehicle_id);
CREATE INDEX IF NOT EXISTS idx_drivers_is_active ON drivers(is_active);

CREATE INDEX IF NOT EXISTS idx_trips_company_id ON trips(company_id);
CREATE INDEX IF NOT EXISTS idx_trips_vehicle_id ON trips(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_trips_driver_id ON trips(driver_id);
CREATE INDEX IF NOT EXISTS idx_trips_status ON trips(status);
CREATE INDEX IF NOT EXISTS idx_trips_start_time ON trips(start_time);

CREATE INDEX IF NOT EXISTS idx_driver_events_company_id ON driver_events(company_id);
CREATE INDEX IF NOT EXISTS idx_driver_events_vehicle_id ON driver_events(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_driver_events_driver_id ON driver_events(driver_id);
CREATE INDEX IF NOT EXISTS idx_driver_events_timestamp ON driver_events(timestamp);

CREATE INDEX IF NOT EXISTS idx_subscriptions_company_id ON subscriptions(company_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_end_date ON subscriptions(end_date);

CREATE INDEX IF NOT EXISTS idx_payments_company_id ON payments(company_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(payment_status);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at);

-- Create spatial indexes for PostGIS
CREATE INDEX IF NOT EXISTS idx_trips_start_location ON trips USING GIST (start_location);
CREATE INDEX IF NOT EXISTS idx_trips_end_location ON trips USING GIST (end_location);
CREATE INDEX IF NOT EXISTS idx_driver_events_location ON driver_events USING GIST (location);

-- Insert default company for development
INSERT INTO companies (id, name, email, phone, address, subscription_plan, subscription_expires_at) 
VALUES (
    uuid_generate_v4(),
    'FleetTracker Pro Demo Company',
    'demo@fleettracker.id',
    '+6281234567890',
    'Jakarta, Indonesia',
    'professional',
    NOW() + INTERVAL '1 year'
) ON CONFLICT DO NOTHING;

-- Insert default admin user for development
INSERT INTO users (id, email, name, role, company_id, password_hash, phone, language, timezone)
SELECT 
    uuid_generate_v4(),
    'admin@fleettracker.id',
    'FleetTracker Admin',
    'fleet_manager',
    c.id,
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj4JqyYh6U8q', -- password: admin123
    '+6281234567890',
    'id',
    'Asia/Jakarta'
FROM companies c 
WHERE c.name = 'FleetTracker Pro Demo Company'
ON CONFLICT (email) DO NOTHING;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_companies_updated_at BEFORE UPDATE ON companies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_vehicles_updated_at BEFORE UPDATE ON vehicles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_drivers_updated_at BEFORE UPDATE ON drivers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_trips_updated_at BEFORE UPDATE ON trips FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON subscriptions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Grant permissions to fleettracker user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO fleettracker;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO fleettracker;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO fleettracker;

-- Success message
DO $$
BEGIN
    RAISE NOTICE 'FleetTracker Pro database initialized successfully!';
    RAISE NOTICE 'Default admin user: admin@fleettracker.id';
    RAISE NOTICE 'Default password: admin123';
END
$$;
