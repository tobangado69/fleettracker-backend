-- FleetTracker Pro TimescaleDB Initialization
-- TimescaleDB for GPS tracking time-series data
-- Indonesian Fleet Management SaaS Application

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- Enable PostGIS extension for location data
CREATE EXTENSION IF NOT EXISTS postgis CASCADE;

-- GPS tracking data (optimized for time-series)
CREATE TABLE IF NOT EXISTS gps_tracks (
    id BIGSERIAL,
    vehicle_id UUID NOT NULL,
    driver_id UUID,
    latitude DECIMAL(10,7) NOT NULL,
    longitude DECIMAL(10,7) NOT NULL,
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
    
    PRIMARY KEY (vehicle_id, timestamp)
);

-- Convert to TimescaleDB hypertable for better time-series performance
SELECT create_hypertable('gps_tracks', 'timestamp', 
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_gps_tracks_vehicle_time ON gps_tracks (vehicle_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_gps_tracks_speed ON gps_tracks (speed) WHERE speed > 80; -- Speed violations
CREATE INDEX IF NOT EXISTS idx_gps_tracks_driver_time ON gps_tracks (driver_id, timestamp DESC) WHERE driver_id IS NOT NULL;

-- Driver behavior events table
CREATE TABLE IF NOT EXISTS behavior_events (
    id BIGSERIAL,
    vehicle_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL, -- 'harsh_braking', 'speeding', 'rapid_acceleration'
    severity VARCHAR(20) DEFAULT 'medium', -- 'low', 'medium', 'high', 'critical'
    latitude DECIMAL(10,7),
    longitude DECIMAL(10,7),
    speed_at_event DECIMAL(5,2), -- km/h
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    PRIMARY KEY (vehicle_id, timestamp)
);

-- Convert to hypertable
SELECT create_hypertable('behavior_events', 'timestamp',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_behavior_events_vehicle_time ON behavior_events (vehicle_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_behavior_events_driver_time ON behavior_events (driver_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_behavior_events_type ON behavior_events (event_type, timestamp DESC);

-- Create data retention policy (keep 2 years of data)
SELECT add_retention_policy('gps_tracks', INTERVAL '2 years', if_not_exists => TRUE);
SELECT add_retention_policy('behavior_events', INTERVAL '2 years', if_not_exists => TRUE);

-- Create compression policy for older data (compress data older than 7 days)
ALTER TABLE gps_tracks SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'vehicle_id'
);

ALTER TABLE behavior_events SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'vehicle_id,driver_id'
);

SELECT add_compression_policy('gps_tracks', INTERVAL '7 days', if_not_exists => TRUE);
SELECT add_compression_policy('behavior_events', INTERVAL '7 days', if_not_exists => TRUE);

-- Success message
DO $$
BEGIN
    RAISE NOTICE '✅ FleetTracker Pro TimescaleDB initialized successfully!';
    RAISE NOTICE '✅ GPS tracking and behavior events tables created';
    RAISE NOTICE '✅ Compression and retention policies configured';
END
$$;
