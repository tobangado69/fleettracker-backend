package tracking

import (
	"context"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// TrackingRepository defines the interface for GPS tracking data operations
type TrackingRepository interface {
	// GPS Track operations
	CreateGPSTrack(ctx context.Context, track *models.GPSTrack) error
	GetCurrentLocation(ctx context.Context, vehicleID string) (*models.GPSTrack, error)
	GetLocationHistory(ctx context.Context, vehicleID string, startTime, endTime time.Time, limit int) ([]*models.GPSTrack, error)
	GetSpeedViolations(ctx context.Context, vehicleID string, minSpeed float64, startTime, endTime time.Time) ([]*models.GPSTrack, error)
	GetRouteDistance(ctx context.Context, vehicleID string, startTime, endTime time.Time) (float64, error)
	
	// Trip operations
	CreateTrip(ctx context.Context, trip *models.Trip) error
	FindTripByID(ctx context.Context, id string) (*models.Trip, error)
	FindTripsByVehicle(ctx context.Context, vehicleID string, limit int) ([]*models.Trip, error)
	FindTripsByDriver(ctx context.Context, driverID string, limit int) ([]*models.Trip, error)
	FindTripsByCompany(ctx context.Context, companyID string, startTime, endTime time.Time, limit int) ([]*models.Trip, error)
	UpdateTrip(ctx context.Context, trip *models.Trip) error
	GetTripStats(ctx context.Context, tripID string) (map[string]interface{}, error)
	
	// Driver event operations
	CreateDriverEvent(ctx context.Context, event *models.DriverEvent) error
	GetDriverEvents(ctx context.Context, driverID string, startTime, endTime time.Time) ([]*models.DriverEvent, error)
	GetVehicleEvents(ctx context.Context, vehicleID string, startTime, endTime time.Time) ([]*models.DriverEvent, error)
	
	// Geofence operations
	CreateGeofence(ctx context.Context, geofence *models.Geofence) error
	FindGeofenceByID(ctx context.Context, id string) (*models.Geofence, error)
	FindGeofencesByCompany(ctx context.Context, companyID string) ([]*models.Geofence, error)
	UpdateGeofence(ctx context.Context, geofence *models.Geofence) error
	DeleteGeofence(ctx context.Context, id string) error
	
	// Statistics
	GetVehicleTrackingStats(ctx context.Context, vehicleID string, startTime, endTime time.Time) (map[string]interface{}, error)
	GetCompanyTrackingStats(ctx context.Context, companyID string) (map[string]interface{}, error)
}

// trackingRepository implements TrackingRepository interface
type trackingRepository struct {
	db              *gorm.DB
	optimizedQueries *OptimizedQueryService
}

// NewTrackingRepository creates a new tracking repository
func NewTrackingRepository(db *gorm.DB) TrackingRepository {
	return &trackingRepository{
		db:              db,
		optimizedQueries: NewOptimizedQueryService(db),
	}
}

// CreateGPSTrack creates a new GPS track
func (r *trackingRepository) CreateGPSTrack(ctx context.Context, track *models.GPSTrack) error {
	return r.db.WithContext(ctx).Create(track).Error
}

// GetCurrentLocation gets the current location
func (r *trackingRepository) GetCurrentLocation(ctx context.Context, vehicleID string) (*models.GPSTrack, error) {
	return r.optimizedQueries.GetCurrentLocationOptimized(ctx, vehicleID)
}

// GetLocationHistory gets location history
func (r *trackingRepository) GetLocationHistory(ctx context.Context, vehicleID string, startTime, endTime time.Time, limit int) ([]*models.GPSTrack, error) {
	return r.optimizedQueries.GetLocationHistoryOptimized(ctx, vehicleID, startTime, endTime, limit)
}

// GetSpeedViolations gets speed violations
func (r *trackingRepository) GetSpeedViolations(ctx context.Context, vehicleID string, minSpeed float64, startTime, endTime time.Time) ([]*models.GPSTrack, error) {
	return r.optimizedQueries.GetSpeedViolationsOptimized(ctx, vehicleID, minSpeed, startTime, endTime)
}

// GetRouteDistance gets route distance
func (r *trackingRepository) GetRouteDistance(ctx context.Context, vehicleID string, startTime, endTime time.Time) (float64, error) {
	return r.optimizedQueries.GetRouteDistanceOptimized(ctx, vehicleID, startTime, endTime)
}

// CreateTrip creates a new trip
func (r *trackingRepository) CreateTrip(ctx context.Context, trip *models.Trip) error {
	return r.db.WithContext(ctx).Create(trip).Error
}

// FindTripByID finds a trip by ID
func (r *trackingRepository) FindTripByID(ctx context.Context, id string) (*models.Trip, error) {
	var trip models.Trip
	if err := r.db.WithContext(ctx).
		Preload("Vehicle").
		Preload("Driver").
		First(&trip, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &trip, nil
}

// FindTripsByVehicle finds trips by vehicle
func (r *trackingRepository) FindTripsByVehicle(ctx context.Context, vehicleID string, limit int) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).
		Where("vehicle_id = ?", vehicleID).
		Order("start_time DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Preload("Vehicle").Preload("Driver").Find(&trips).Error; err != nil {
		return nil, err
	}
	return trips, nil
}

// FindTripsByDriver finds trips by driver
func (r *trackingRepository) FindTripsByDriver(ctx context.Context, driverID string, limit int) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).
		Where("driver_id = ?", driverID).
		Order("start_time DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Preload("Vehicle").Preload("Driver").Find(&trips).Error; err != nil {
		return nil, err
	}
	return trips, nil
}

// FindTripsByCompany finds trips by company
func (r *trackingRepository) FindTripsByCompany(ctx context.Context, companyID string, startTime, endTime time.Time, limit int) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).
		Where("company_id = ? AND start_time BETWEEN ? AND ?", companyID, startTime, endTime).
		Order("start_time DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Preload("Vehicle").Preload("Driver").Find(&trips).Error; err != nil {
		return nil, err
	}
	return trips, nil
}

// UpdateTrip updates a trip
func (r *trackingRepository) UpdateTrip(ctx context.Context, trip *models.Trip) error {
	return r.db.WithContext(ctx).Save(trip).Error
}

// GetTripStats gets trip statistics
func (r *trackingRepository) GetTripStats(ctx context.Context, tripID string) (map[string]interface{}, error) {
	var trip models.Trip
	if err := r.db.WithContext(ctx).First(&trip, "id = ?", tripID).Error; err != nil {
		return nil, err
	}
	
	duration := 0.0
	if trip.EndTime != nil && trip.StartTime != nil {
		duration = trip.EndTime.Sub(*trip.StartTime).Minutes()
	}
	
	return map[string]interface{}{
		"total_distance": trip.TotalDistance,
		"duration":       duration,
	}, nil
}

// CreateDriverEvent creates a new driver event
func (r *trackingRepository) CreateDriverEvent(ctx context.Context, event *models.DriverEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// GetDriverEvents gets driver events
func (r *trackingRepository) GetDriverEvents(ctx context.Context, driverID string, startTime, endTime time.Time) ([]*models.DriverEvent, error) {
	var events []*models.DriverEvent
	if err := r.db.WithContext(ctx).
		Where("driver_id = ? AND timestamp BETWEEN ? AND ?", driverID, startTime, endTime).
		Order("timestamp DESC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetVehicleEvents gets vehicle events
func (r *trackingRepository) GetVehicleEvents(ctx context.Context, vehicleID string, startTime, endTime time.Time) ([]*models.DriverEvent, error) {
	var events []*models.DriverEvent
	if err := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND timestamp BETWEEN ? AND ?", vehicleID, startTime, endTime).
		Order("timestamp DESC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// CreateGeofence creates a new geofence
func (r *trackingRepository) CreateGeofence(ctx context.Context, geofence *models.Geofence) error {
	return r.db.WithContext(ctx).Create(geofence).Error
}

// FindGeofenceByID finds a geofence by ID
func (r *trackingRepository) FindGeofenceByID(ctx context.Context, id string) (*models.Geofence, error) {
	var geofence models.Geofence
	if err := r.db.WithContext(ctx).First(&geofence, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &geofence, nil
}

// FindGeofencesByCompany finds geofences by company
func (r *trackingRepository) FindGeofencesByCompany(ctx context.Context, companyID string) ([]*models.Geofence, error) {
	var geofences []*models.Geofence
	if err := r.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ?", companyID, true).
		Order("created_at DESC").
		Find(&geofences).Error; err != nil {
		return nil, err
	}
	return geofences, nil
}

// UpdateGeofence updates a geofence
func (r *trackingRepository) UpdateGeofence(ctx context.Context, geofence *models.Geofence) error {
	return r.db.WithContext(ctx).Save(geofence).Error
}

// DeleteGeofence soft deletes a geofence
func (r *trackingRepository) DeleteGeofence(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Geofence{}, "id = ?", id).Error
}

// GetVehicleTrackingStats gets vehicle tracking statistics
func (r *trackingRepository) GetVehicleTrackingStats(ctx context.Context, vehicleID string, startTime, endTime time.Time) (map[string]interface{}, error) {
	var stats struct {
		TotalTracks int64   `gorm:"column:total_tracks"`
		AvgSpeed    float64 `gorm:"column:avg_speed"`
	}
	
	query := `
		SELECT 
			COUNT(*) as total_tracks,
			COALESCE(AVG(speed), 0) as avg_speed
		FROM gps_tracks
		WHERE vehicle_id = ? AND timestamp BETWEEN ? AND ?
	`
	
	if err := r.db.WithContext(ctx).Raw(query, vehicleID, startTime, endTime).Scan(&stats).Error; err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"total_tracks": stats.TotalTracks,
		"avg_speed":    stats.AvgSpeed,
	}, nil
}

// GetCompanyTrackingStats gets company tracking statistics
func (r *trackingRepository) GetCompanyTrackingStats(ctx context.Context, companyID string) (map[string]interface{}, error) {
	var stats struct {
		TotalTracks   int64 `gorm:"column:total_tracks"`
		ActiveVehicles int64 `gorm:"column:active_vehicles"`
	}
	
	query := `
		SELECT 
			COUNT(DISTINCT gt.vehicle_id) as active_vehicles,
			COUNT(*) as total_tracks
		FROM gps_tracks gt
		INNER JOIN vehicles v ON v.id = gt.vehicle_id
		WHERE v.company_id = ?
			AND gt.timestamp >= NOW() - INTERVAL '24 hours'
	`
	
	if err := r.db.WithContext(ctx).Raw(query, companyID).Scan(&stats).Error; err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"active_vehicles": stats.ActiveVehicles,
		"total_tracks":    stats.TotalTracks,
	}, nil
}

