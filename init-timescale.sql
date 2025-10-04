-- FleetTracker Pro TimescaleDB Initialization
-- TimescaleDB for GPS tracking time-series data
-- Indonesian Fleet Management SaaS Application

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- Create database user if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'fleettracker') THEN
        CREATE ROLE fleettracker WITH LOGIN PASSWORD 'password123';
    END IF;
END
$$;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE fleettracker_timeseries TO fleettracker;

-- GPS tracking data (optimized for time-series)
CREATE TABLE IF NOT EXISTS gps_tracks (
    id BIGSERIAL,
    vehicle_id UUID NOT NULL,
    driver_id UUID,
    location GEOMETRY(POINT, 4326) NOT NULL, -- PostGIS point
    speed DECIMAL(5,2), -- km/h
    heading INTEGER, -- degrees 0-360
    altitude DECIMAL(8,2), -- meters
    accuracy DECIMAL(5,2), -- meters
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Indonesian specific fields
    road_type VARCHAR(50), -- 'highway', 'city', 'toll_road'
    speed_limit INTEGER, -- km/h based on location
    fuel_level DECIMAL(5,2), -- percentage
    engine_status VARCHAR(20), -- 'on', 'off', 'idle'
    temperature DECIMAL(4,2), -- engine temperature in celsius
    
    PRIMARY KEY (vehicle_id, timestamp)
);

-- Convert to TimescaleDB hypertable for better time-series performance
SELECT create_hypertable('gps_tracks', 'timestamp', 'vehicle_id', 4);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_gps_tracks_vehicle_time ON gps_tracks (vehicle_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_gps_tracks_location ON gps_tracks USING GIST (location);
CREATE INDEX IF NOT EXISTS idx_gps_tracks_speed ON gps_tracks (speed) WHERE speed > 80; -- Speed violations
CREATE INDEX IF NOT EXISTS idx_gps_tracks_driver_time ON gps_tracks (driver_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_gps_tracks_road_type ON gps_tracks (road_type, timestamp);

-- Fuel consumption tracking table
CREATE TABLE IF NOT EXISTS fuel_consumption (
    id BIGSERIAL,
    vehicle_id UUID NOT NULL,
    driver_id UUID,
    trip_id UUID,
    fuel_level DECIMAL(5,2) NOT NULL, -- percentage
    fuel_consumed DECIMAL(6,2), -- liters
    fuel_cost DECIMAL(10,2), -- IDR
    location GEOMETRY(POINT, 4326),
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Indonesian specific fields
    fuel_type VARCHAR(20), -- 'pertalite', 'pertamax', 'solar'
    gas_station VARCHAR(100),
    gas_station_location GEOMETRY(POINT, 4326),
    price_per_liter DECIMAL(8,2), -- IDR per liter
    
    PRIMARY KEY (vehicle_id, timestamp)
);

-- Convert to hypertable
SELECT create_hypertable('fuel_consumption', 'timestamp', 'vehicle_id', 4);

-- Create indexes for fuel consumption
CREATE INDEX IF NOT EXISTS idx_fuel_consumption_vehicle_time ON fuel_consumption (vehicle_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_fuel_consumption_driver_time ON fuel_consumption (driver_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_fuel_consumption_cost ON fuel_consumption (fuel_cost, timestamp DESC);

-- Engine diagnostics table
CREATE TABLE IF NOT EXISTS engine_diagnostics (
    id BIGSERIAL,
    vehicle_id UUID NOT NULL,
    driver_id UUID,
    engine_rpm INTEGER,
    engine_temperature DECIMAL(4,2), -- celsius
    battery_voltage DECIMAL(4,2), -- volts
    fuel_pressure DECIMAL(5,2), -- psi
    oil_pressure DECIMAL(5,2), -- psi
    coolant_temperature DECIMAL(4,2), -- celsius
    air_intake_temperature DECIMAL(4,2), -- celsius
    throttle_position DECIMAL(4,2), -- percentage
    brake_pressure DECIMAL(5,2), -- psi
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Indonesian specific fields
    odometer_reading DECIMAL(10,2), -- km
    service_due_km DECIMAL(10,2), -- km until next service
    maintenance_alerts JSONB, -- maintenance alerts and warnings
    
    PRIMARY KEY (vehicle_id, timestamp)
);

-- Convert to hypertable
SELECT create_hypertable('engine_diagnostics', 'timestamp', 'vehicle_id', 4);

-- Create indexes for engine diagnostics
CREATE INDEX IF NOT EXISTS idx_engine_diagnostics_vehicle_time ON engine_diagnostics (vehicle_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_engine_diagnostics_temperature ON engine_diagnostics (engine_temperature, timestamp DESC);

-- Driver behavior events (detailed time-series)
CREATE TABLE IF NOT EXISTS behavior_events (
    id BIGSERIAL,
    vehicle_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    trip_id UUID,
    event_type VARCHAR(50) NOT NULL, -- 'harsh_braking', 'speeding', 'rapid_acceleration', 'sharp_turning'
    severity VARCHAR(20) DEFAULT 'medium', -- 'low', 'medium', 'high', 'critical'
    location GEOMETRY(POINT, 4326),
    speed_at_event DECIMAL(5,2), -- km/h
    acceleration DECIMAL(6,3), -- m/s²
    deceleration DECIMAL(6,3), -- m/s²
    g_force DECIMAL(6,3), -- g-force measurement
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Indonesian specific fields
    road_condition VARCHAR(50), -- 'dry', 'wet', 'slippery'
    weather_condition VARCHAR(50), -- 'clear', 'rain', 'fog'
    traffic_condition VARCHAR(50), -- 'light', 'moderate', 'heavy'
    speed_limit_at_location INTEGER, -- km/h
    violation_duration INTEGER, -- seconds
    
    PRIMARY KEY (vehicle_id, timestamp)
);

-- Convert to hypertable
SELECT create_hypertable('behavior_events', 'timestamp', 'vehicle_id', 4);

-- Create indexes for behavior events
CREATE INDEX IF NOT EXISTS idx_behavior_events_vehicle_time ON behavior_events (vehicle_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_behavior_events_driver_time ON behavior_events (driver_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_behavior_events_type ON behavior_events (event_type, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_behavior_events_severity ON behavior_events (severity, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_behavior_events_location ON behavior_events USING GIST (location);

-- Route optimization data
CREATE TABLE IF NOT EXISTS route_optimization (
    id BIGSERIAL,
    vehicle_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    trip_id UUID NOT NULL,
    planned_route GEOMETRY(LINESTRING, 4326),
    actual_route GEOMETRY(LINESTRING, 4326),
    optimized_route GEOMETRY(LINESTRING, 4326),
    planned_distance DECIMAL(8,2), -- km
    actual_distance DECIMAL(8,2), -- km
    optimized_distance DECIMAL(8,2), -- km
    planned_duration INTEGER, -- seconds
    actual_duration INTEGER, -- seconds
    optimized_duration INTEGER, -- seconds
    fuel_efficiency DECIMAL(6,3), -- km/liter
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Indonesian specific fields
    toll_cost DECIMAL(10,2), -- IDR
    parking_cost DECIMAL(10,2), -- IDR
    traffic_jam_duration INTEGER, -- seconds
    traffic_jam_locations JSONB, -- locations with traffic jams
    
    PRIMARY KEY (vehicle_id, timestamp)
);

-- Convert to hypertable
SELECT create_hypertable('route_optimization', 'timestamp', 'vehicle_id', 4);

-- Create indexes for route optimization
CREATE INDEX IF NOT EXISTS idx_route_optimization_vehicle_time ON route_optimization (vehicle_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_route_optimization_driver_time ON route_optimization (driver_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_route_optimization_trip_id ON route_optimization (trip_id, timestamp DESC);

-- Create continuous aggregates for performance
-- Daily GPS summary
CREATE MATERIALIZED VIEW IF NOT EXISTS gps_daily_summary
WITH (timescaledb.continuous) AS
SELECT 
    vehicle_id,
    driver_id,
    time_bucket('1 day', timestamp) AS day,
    COUNT(*) AS total_points,
    AVG(speed) AS avg_speed,
    MAX(speed) AS max_speed,
    MIN(speed) AS min_speed,
    ST_Collect(location) AS route_geometry,
    SUM(fuel_consumed) AS total_fuel_consumed,
    AVG(fuel_level) AS avg_fuel_level
FROM gps_tracks
GROUP BY vehicle_id, driver_id, day;

-- Weekly driver performance summary
CREATE MATERIALIZED VIEW IF NOT EXISTS driver_weekly_performance
WITH (timescaledb.continuous) AS
SELECT 
    driver_id,
    time_bucket('1 week', timestamp) AS week,
    COUNT(*) AS total_events,
    COUNT(*) FILTER (WHERE event_type = 'speeding') AS speeding_events,
    COUNT(*) FILTER (WHERE event_type = 'harsh_braking') AS harsh_braking_events,
    COUNT(*) FILTER (WHERE event_type = 'rapid_acceleration') AS rapid_acceleration_events,
    COUNT(*) FILTER (WHERE severity = 'high' OR severity = 'critical') AS critical_events,
    AVG(speed_at_event) AS avg_speed,
    MAX(speed_at_event) AS max_speed
FROM behavior_events
GROUP BY driver_id, week;

-- Monthly fuel consumption summary
CREATE MATERIALIZED VIEW IF NOT EXISTS fuel_monthly_summary
WITH (timescaledb.continuous) AS
SELECT 
    vehicle_id,
    driver_id,
    time_bucket('1 month', timestamp) AS month,
    SUM(fuel_consumed) AS total_fuel_consumed,
    SUM(fuel_cost) AS total_fuel_cost,
    AVG(price_per_liter) AS avg_price_per_liter,
    COUNT(*) AS refuel_events,
    AVG(fuel_efficiency) AS avg_fuel_efficiency
FROM fuel_consumption
GROUP BY vehicle_id, driver_id, month;

-- Create data retention policy (keep 2 years of data)
SELECT add_retention_policy('gps_tracks', INTERVAL '2 years');
SELECT add_retention_policy('fuel_consumption', INTERVAL '2 years');
SELECT add_retention_policy('engine_diagnostics', INTERVAL '1 year');
SELECT add_retention_policy('behavior_events', INTERVAL '2 years');
SELECT add_retention_policy('route_optimization', INTERVAL '1 year');

-- Create compression policy for older data
SELECT add_compression_policy('gps_tracks', INTERVAL '30 days');
SELECT add_compression_policy('fuel_consumption', INTERVAL '30 days');
SELECT add_compression_policy('engine_diagnostics', INTERVAL '30 days');
SELECT add_compression_policy('behavior_events', INTERVAL '30 days');
SELECT add_compression_policy('route_optimization', INTERVAL '30 days');

-- Grant permissions to fleettracker user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO fleettracker;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO fleettracker;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO fleettracker;
GRANT ALL PRIVILEGES ON ALL MATERIALIZED VIEWS IN SCHEMA public TO fleettracker;

-- Success message
DO $$
BEGIN
    RAISE NOTICE 'FleetTracker Pro TimescaleDB initialized successfully!';
    RAISE NOTICE 'Time-series tables created with compression and retention policies';
    RAISE NOTICE 'Continuous aggregates created for performance optimization';
END
$$;
