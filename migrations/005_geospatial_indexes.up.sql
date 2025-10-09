-- Geospatial Indexes for GPS Data Optimization
-- Created: 2025-10-08
-- Description: Add PostGIS geospatial indexes for location-based queries

-- Ensure PostGIS extension is enabled
CREATE EXTENSION IF NOT EXISTS postgis;

-- =============================================================================
-- GEOSPATIAL POINT INDEXES
-- =============================================================================

-- GPS tracks location index (GIST for spatial queries)
-- Enables: distance queries, nearest neighbor, within radius
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_location_gist 
ON gps_tracks USING GIST(ST_MakePoint(longitude, latitude));

-- GPS tracks with geography type (for accurate distance calculations)
-- Better for long-distance calculations across globe
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_geography 
ON gps_tracks USING GIST(
    ST_Transform(
        ST_SetSRID(ST_MakePoint(longitude, latitude), 4326),
        4326
    )::geography
);

-- Recent GPS tracks with location (last 7 days - most active data)
CREATE INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_recent_location 
ON gps_tracks USING GIST(ST_MakePoint(longitude, latitude)) 
WHERE timestamp >= NOW() - INTERVAL '7 days';

-- GPS tracks by vehicle with spatial index
-- Composite: vehicle_id + location for vehicle-specific spatial queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_vehicle_location 
ON gps_tracks USING GIST(vehicle_id, ST_MakePoint(longitude, latitude));

-- =============================================================================
-- GEOFENCE BOUNDARY INDEXES
-- =============================================================================

-- Geofence boundaries (polygon/circle geospatial index)
-- Assumes geofences table has a 'boundary' geometry column
-- If using JSONB for coordinates, this needs to be added separately
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_boundary_gist 
ON geofences USING GIST(
    ST_GeomFromGeoJSON(coordinates::text)
) WHERE coordinates IS NOT NULL;

-- Active geofences with spatial index
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_active_boundary 
ON geofences USING GIST(
    ST_GeomFromGeoJSON(coordinates::text)
) WHERE is_active = true AND coordinates IS NOT NULL;

-- =============================================================================
-- SPATIAL QUERY OPTIMIZATIONS
-- =============================================================================

-- Bounding box queries (faster than distance calculations)
-- For queries like: "vehicles within visible map area"
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_bbox 
ON gps_tracks(vehicle_id, timestamp DESC, 
    latitude, longitude
) WHERE latitude BETWEEN -90 AND 90 
  AND longitude BETWEEN -180 AND 180;

-- GPS tracks within Indonesia bounds (optimize for Indonesian fleet)
-- Indonesia: approximately 6°N to 11°S, 95°E to 141°E
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_indonesia 
ON gps_tracks USING GIST(ST_MakePoint(longitude, latitude)) 
WHERE latitude BETWEEN -11 AND 6 
  AND longitude BETWEEN 95 AND 141;

-- =============================================================================
-- DISTANCE CALCULATION OPTIMIZATIONS
-- =============================================================================

-- Add computed geography column for efficient distance queries
-- This is optional but significantly speeds up distance-based queries
ALTER TABLE gps_tracks 
ADD COLUMN IF NOT EXISTS location_geo geography(Point, 4326);

-- Populate the geography column
UPDATE gps_tracks 
SET location_geo = ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography 
WHERE location_geo IS NULL AND latitude IS NOT NULL AND longitude IS NOT NULL;

-- Create index on geography column
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gps_tracks_location_geo 
ON gps_tracks USING GIST(location_geo);

-- Trigger to auto-update location_geo on insert/update
CREATE OR REPLACE FUNCTION update_gps_location_geo()
RETURNS TRIGGER AS $$
BEGIN
    NEW.location_geo = ST_SetSRID(ST_MakePoint(NEW.longitude, NEW.latitude), 4326)::geography;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_gps_location_geo ON gps_tracks;
CREATE TRIGGER trigger_update_gps_location_geo
    BEFORE INSERT OR UPDATE OF latitude, longitude ON gps_tracks
    FOR EACH ROW
    EXECUTE FUNCTION update_gps_location_geo();

-- =============================================================================
-- SPATIAL CLUSTERING (for better locality)
-- =============================================================================

-- Cluster GPS tracks by vehicle and timestamp for sequential reads
-- This physically reorders data on disk for better performance
CLUSTER gps_tracks USING idx_gps_tracks_vehicle_time_speed;

-- =============================================================================
-- VACUUM AND ANALYZE
-- =============================================================================

-- Update statistics after adding indexes
ANALYZE gps_tracks;
ANALYZE geofences;
ANALYZE vehicles;
ANALYZE trips;

