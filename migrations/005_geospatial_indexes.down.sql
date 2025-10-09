-- Rollback Geospatial Indexes

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_gps_location_geo ON gps_tracks;
DROP FUNCTION IF EXISTS update_gps_location_geo();

-- Drop geography column
ALTER TABLE gps_tracks DROP COLUMN IF EXISTS location_geo;

-- Drop geospatial indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_location_geo;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_location_gist;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_geography;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_recent_location;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_vehicle_location;
DROP INDEX CONCURRENTLY IF EXISTS idx_geofences_boundary_gist;
DROP INDEX CONCURRENTLY IF EXISTS idx_geofences_active_boundary;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_bbox;
DROP INDEX CONCURRENTLY IF EXISTS idx_gps_tracks_indonesia;

