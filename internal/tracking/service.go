package tracking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	apperrors "github.com/tobangado69/fleettracker-pro/backend/pkg/errors"
	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/internal/common/realtime"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

var ctx = context.Background()

// Service handles mobile GPS tracking operations
type Service struct {
	db                    *gorm.DB
	redis                 *redis.Client
	websocketHub          *realtime.WebSocketHub
	localWebSocketHub     *WebSocketHub
	cache                 *CacheService
	analyticsBroadcaster  *realtime.AnalyticsBroadcaster
	alertSystem           *realtime.AlertSystem
}

// CacheService provides caching functionality for tracking operations
type CacheService struct {
	redis *redis.Client
}

// NewCacheService creates a new cache service
func NewCacheService(redis *redis.Client) *CacheService {
	return &CacheService{redis: redis}
}

// GetCurrentLocationFromCache retrieves current vehicle location from cache
func (cs *CacheService) GetCurrentLocationFromCache(ctx context.Context, vehicleID string) (*models.GPSTrack, error) {
	key := fmt.Sprintf("vehicle:location:%s", vehicleID)
	
	result := cs.redis.HGetAll(ctx, key)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get current location from cache: %w", result.Err())
	}
	
	data := result.Val()
	if len(data) == 0 {
		return nil, nil // Cache miss
	}
	
	// Parse cached data
	lat, _ := strconv.ParseFloat(data["latitude"], 64)
	lng, _ := strconv.ParseFloat(data["longitude"], 64)
	speed, _ := strconv.ParseFloat(data["speed"], 64)
	heading, _ := strconv.ParseFloat(data["heading"], 64)
	timestamp, _ := strconv.ParseInt(data["timestamp"], 10, 64)
	
	return &models.GPSTrack{
		VehicleID: vehicleID,
		Latitude:  lat,
		Longitude: lng,
		Speed:     speed,
		Heading:   heading,
		Timestamp: time.Unix(timestamp, 0),
	}, nil
}

// SetCurrentLocationInCache stores current vehicle location in cache
func (cs *CacheService) SetCurrentLocationInCache(ctx context.Context, gpsTrack *models.GPSTrack, expiration time.Duration) error {
	key := fmt.Sprintf("vehicle:location:%s", gpsTrack.VehicleID)
	
	location := map[string]interface{}{
		"latitude":  gpsTrack.Latitude,
		"longitude": gpsTrack.Longitude,
		"speed":     gpsTrack.Speed,
		"heading":   gpsTrack.Heading,
		"timestamp": gpsTrack.Timestamp.Unix(),
	}
	
	if err := cs.redis.HMSet(ctx, key, location).Err(); err != nil {
		return fmt.Errorf("failed to set current location in cache: %w", err)
	}
	
	if err := cs.redis.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set current location cache expiration: %w", err)
	}
	
	return nil
}

// GetLocationHistoryFromCache retrieves location history from cache
func (cs *CacheService) GetLocationHistoryFromCache(ctx context.Context, vehicleID string, filters GPSFilters) ([]models.GPSTrack, int64, error) {
	cacheKey := cs.generateLocationHistoryCacheKey(vehicleID, filters)
	
	var result struct {
		Tracks []models.GPSTrack `json:"tracks"`
		Total  int64            `json:"total"`
	}
	
	data, err := cs.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, 0, nil // Cache miss
		}
		return nil, 0, fmt.Errorf("failed to get location history from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal location history from cache: %w", err)
	}
	
	return result.Tracks, result.Total, nil
}

// SetLocationHistoryInCache stores location history in cache
func (cs *CacheService) SetLocationHistoryInCache(ctx context.Context, vehicleID string, filters GPSFilters, tracks []models.GPSTrack, total int64, expiration time.Duration) error {
	cacheKey := cs.generateLocationHistoryCacheKey(vehicleID, filters)
	
	result := struct {
		Tracks []models.GPSTrack `json:"tracks"`
		Total  int64            `json:"total"`
	}{
		Tracks: tracks,
		Total:  total,
	}
	
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal location history for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, cacheKey, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set location history in cache: %w", err)
	}
	
	return nil
}

// GetTripFromCache retrieves a trip from cache
func (cs *CacheService) GetTripFromCache(ctx context.Context, tripID string) (*models.Trip, error) {
	key := fmt.Sprintf("trip:%s", tripID)
	
	var trip models.Trip
	data, err := cs.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get trip from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &trip); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trip from cache: %w", err)
	}
	
	return &trip, nil
}

// SetTripInCache stores a trip in cache
func (cs *CacheService) SetTripInCache(ctx context.Context, trip *models.Trip, expiration time.Duration) error {
	key := fmt.Sprintf("trip:%s", trip.ID)
	
	data, err := json.Marshal(trip)
	if err != nil {
		return fmt.Errorf("failed to marshal trip for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set trip in cache: %w", err)
	}
	
	return nil
}

// GetGeofenceFromCache retrieves a geofence from cache
func (cs *CacheService) GetGeofenceFromCache(ctx context.Context, geofenceID string) (*models.Geofence, error) {
	key := fmt.Sprintf("geofence:%s", geofenceID)
	
	var geofence models.Geofence
	data, err := cs.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get geofence from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &geofence); err != nil {
		return nil, fmt.Errorf("failed to unmarshal geofence from cache: %w", err)
	}
	
	return &geofence, nil
}

// SetGeofenceInCache stores a geofence in cache
func (cs *CacheService) SetGeofenceInCache(ctx context.Context, geofence *models.Geofence, expiration time.Duration) error {
	key := fmt.Sprintf("geofence:%s", geofence.ID)
	
	data, err := json.Marshal(geofence)
	if err != nil {
		return fmt.Errorf("failed to marshal geofence for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set geofence in cache: %w", err)
	}
	
	return nil
}

// GetGeofencesByCompanyFromCache retrieves geofences for a company from cache
func (cs *CacheService) GetGeofencesByCompanyFromCache(ctx context.Context, companyID string) ([]models.Geofence, error) {
	key := fmt.Sprintf("geofences:company:%s", companyID)
	
	var geofences []models.Geofence
	data, err := cs.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get geofences from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &geofences); err != nil {
		return nil, fmt.Errorf("failed to unmarshal geofences from cache: %w", err)
	}
	
	return geofences, nil
}

// SetGeofencesByCompanyInCache stores geofences for a company in cache
func (cs *CacheService) SetGeofencesByCompanyInCache(ctx context.Context, companyID string, geofences []models.Geofence, expiration time.Duration) error {
	key := fmt.Sprintf("geofences:company:%s", companyID)
	
	data, err := json.Marshal(geofences)
	if err != nil {
		return fmt.Errorf("failed to marshal geofences for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set geofences in cache: %w", err)
	}
	
	return nil
}

// InvalidateLocationHistoryCache removes location history cache for a vehicle
func (cs *CacheService) InvalidateLocationHistoryCache(ctx context.Context, vehicleID string) error {
	pattern := fmt.Sprintf("location:history:%s:*", vehicleID)
	
	keys, err := cs.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get location history cache keys: %w", err)
	}
	
	if len(keys) > 0 {
		if err := cs.redis.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to invalidate location history cache: %w", err)
		}
	}
	
	return nil
}

// InvalidateTripCache removes a trip from cache
func (cs *CacheService) InvalidateTripCache(ctx context.Context, tripID string) error {
	key := fmt.Sprintf("trip:%s", tripID)
	
	if err := cs.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate trip cache: %w", err)
	}
	
	return nil
}

// InvalidateGeofenceCache removes a geofence from cache
func (cs *CacheService) InvalidateGeofenceCache(ctx context.Context, geofenceID string) error {
	key := fmt.Sprintf("geofence:%s", geofenceID)
	
	if err := cs.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate geofence cache: %w", err)
	}
	
	return nil
}

// InvalidateGeofencesByCompanyCache removes geofences cache for a company
func (cs *CacheService) InvalidateGeofencesByCompanyCache(ctx context.Context, companyID string) error {
	key := fmt.Sprintf("geofences:company:%s", companyID)
	
	if err := cs.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate geofences cache: %w", err)
	}
	
	return nil
}

// generateLocationHistoryCacheKey creates a cache key for location history queries
func (cs *CacheService) generateLocationHistoryCacheKey(vehicleID string, filters GPSFilters) string {
	// Create a hash of the filters for the cache key
	filterHash := fmt.Sprintf("%v", filters)
	hash := fmt.Sprintf("%x", filterHash)
	return fmt.Sprintf("location:history:%s:%s", vehicleID, hash)
}

// WebSocketHub manages WebSocket connections
type WebSocketHub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Register registers a new WebSocket connection
func (h *WebSocketHub) Register(conn *websocket.Conn) {
	h.register <- conn
}

// Unregister unregisters a WebSocket connection
func (h *WebSocketHub) Unregister(conn *websocket.Conn) {
	h.unregister <- conn
}

// Broadcast sends a message to all connected clients
func (h *WebSocketHub) Broadcast(message []byte) {
	h.broadcast <- message
}

// GetClientCount returns the number of connected clients
func (h *WebSocketHub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mutex.Lock()
			h.clients[conn] = true
			h.mutex.Unlock()

		case conn := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for conn := range h.clients {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					conn.Close()
					delete(h.clients, conn)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// GPSDataRequest represents GPS data from mobile device
type GPSDataRequest struct {
	VehicleID     string    `json:"vehicle_id" validate:"required"`
	DriverID      string    `json:"driver_id" validate:"required"`
	Latitude      float64   `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude     float64   `json:"longitude" validate:"required,min=-180,max=180"`
	Altitude      float64   `json:"altitude"`
	Speed         float64   `json:"speed" validate:"min=0,max=200"` // km/h
	Heading       float64   `json:"heading" validate:"min=0,max=360"`
	Accuracy      float64   `json:"accuracy" validate:"min=0,max=100"` // meters
	Timestamp     time.Time `json:"timestamp" validate:"required"`
	BatteryLevel  float64   `json:"battery_level" validate:"min=0,max=100"`
	NetworkType   string    `json:"network_type"` // 4G, 5G, WiFi
	IsOfflineSync bool      `json:"is_offline_sync"`
}

// GPSFilters represents filters for GPS data queries
type GPSFilters struct {
	VehicleID   *string    `json:"vehicle_id" form:"vehicle_id"`
	DriverID    *string    `json:"driver_id" form:"driver_id"`
	StartTime   *time.Time `json:"start_time" form:"start_time"`
	EndTime     *time.Time `json:"end_time" form:"end_time"`
	MinAccuracy *float64   `json:"min_accuracy" form:"min_accuracy"`
	MaxSpeed    *float64   `json:"max_speed" form:"max_speed"`
	
	// Pagination
	Page      int    `json:"page" form:"page" validate:"min=1"`
	Limit     int    `json:"limit" form:"limit" validate:"min=1,max=1000"`
	SortBy    string `json:"sort_by" form:"sort_by" validate:"oneof=timestamp latitude longitude speed"`
	SortOrder string `json:"sort_order" form:"sort_order" validate:"oneof=asc desc"`
}

// DriverEventRequest represents a driver behavior event
type DriverEventRequest struct {
	VehicleID   string    `json:"vehicle_id" validate:"required"`
	DriverID    string    `json:"driver_id" validate:"required"`
	EventType   string    `json:"event_type" validate:"required,oneof=speed_violation harsh_braking rapid_acceleration sharp_cornering idle_time driving_hours_violation"`
	Severity    string    `json:"severity" validate:"required,oneof=low medium high critical"`
	Latitude    float64   `json:"latitude" validate:"required"`
	Longitude   float64   `json:"longitude" validate:"required"`
	Timestamp   time.Time `json:"timestamp" validate:"required"`
	Speed       float64   `json:"speed"`
	Details     string    `json:"details"`
	Value       float64   `json:"value"` // Speed, acceleration, etc.
}

// TripRequest represents trip start/end data
type TripRequest struct {
	VehicleID     string    `json:"vehicle_id" validate:"required"`
	DriverID      string    `json:"driver_id" validate:"required"`
	Action        string    `json:"action" validate:"required,oneof=start end"`
	StartLocation *Location `json:"start_location,omitempty"`
	EndLocation   *Location `json:"end_location,omitempty"`
	Timestamp     time.Time `json:"timestamp" validate:"required"`
	OdometerStart *float64  `json:"odometer_start,omitempty"`
	OdometerEnd   *float64  `json:"odometer_end,omitempty"`
}

// Location represents a GPS location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
}

// GeofenceRequest represents geofence data
type GeofenceRequest struct {
	CompanyID      string  `json:"company_id" validate:"required"`
	Name           string  `json:"name" validate:"required"`
	Type           string  `json:"type" validate:"required,oneof=zone pickup delivery restricted"`
	CenterLat      float64 `json:"center_lat" validate:"required,min=-90,max=90"`
	CenterLng      float64 `json:"center_lng" validate:"required,min=-180,max=180"`
	Radius         float64 `json:"radius" validate:"required,min=10,max=10000"` // meters
	AlertOnEntry   bool    `json:"alert_on_entry"`
	AlertOnExit    bool    `json:"alert_on_exit"`
	IsActive       bool    `json:"is_active"`
	Description    string  `json:"description"`
}

// NewService creates a new tracking service
func NewService(db *gorm.DB, redis *redis.Client) *Service {
	// Create enhanced WebSocket hub
	hub := realtime.NewWebSocketHub(redis, realtime.DefaultWebSocketConfig())
	
	// Create local WebSocket hub for additional functionality
	localHub := NewWebSocketHub()
	
	// Create analytics broadcaster
	analyticsBroadcaster := realtime.NewAnalyticsBroadcaster(hub, redis, db, nil)
	
	// Create alert system
	alertSystem := realtime.NewAlertSystem(hub, redis)

	service := &Service{
		db:                   db,
		redis:                redis,
		websocketHub:         hub,
		localWebSocketHub:    localHub,
		cache:                NewCacheService(redis),
		analyticsBroadcaster: analyticsBroadcaster,
		alertSystem:          alertSystem,
	}
	
	// Start the local WebSocket hub in a goroutine
	go service.localWebSocketHub.Run()
	
	return service
}

// ProcessGPSData processes incoming GPS data from mobile devices
func (s *Service) ProcessGPSData(req GPSDataRequest) (*models.GPSTrack, error) {
	// Validate GPS coordinates
	if err := s.validateGPSCoordinates(req.Latitude, req.Longitude, req.Accuracy); err != nil {
		return nil, err
	}

	// Check if vehicle exists and is active
	var vehicle models.Vehicle
	if err := s.db.Where("id = ? AND is_active = ?", req.VehicleID, true).First(&vehicle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("vehicle")
		}
		return nil, apperrors.Wrap(err, "failed to validate vehicle")
	}

	// Check if driver exists and is active
	var driver models.Driver
	if err := s.db.Where("id = ? AND is_active = ?", req.DriverID, true).First(&driver).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("driver")
		}
		return nil, apperrors.Wrap(err, "failed to validate driver")
	}

	// Validate driver is assigned to vehicle
	if driver.VehicleID == nil || *driver.VehicleID != req.VehicleID {
		return nil, apperrors.NewBadRequestError("driver not assigned to this vehicle")
	}

	// Create GPS track
	gpsTrack := &models.GPSTrack{
		VehicleID:   req.VehicleID,
		DriverID:    &req.DriverID,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Altitude:    req.Altitude,
		Speed:       req.Speed,
		Heading:     req.Heading,
		Accuracy:    req.Accuracy,
		Timestamp:   req.Timestamp,
		ProcessedAt: time.Now(),
	}

	// Save to database
	if err := s.db.Create(gpsTrack).Error; err != nil {
		return nil, fmt.Errorf("failed to save GPS track: %w", err)
	}

	// Update vehicle's last known location
	if err := s.updateVehicleLocation(req.VehicleID, req.Latitude, req.Longitude, req.Speed, req.Timestamp); err != nil {
		// Log error but don't fail the GPS processing
		fmt.Printf("Failed to update vehicle location: %v\n", err)
	}

	// Process driver behavior events
	go s.processDriverBehavior(gpsTrack)

	// Broadcast real-time location update
	go func() {
		if err := s.analyticsBroadcaster.BroadcastVehicleLocationUpdate(ctx, gpsTrack); err != nil {
			fmt.Printf("Failed to broadcast vehicle location update %s: %v\n", gpsTrack.VehicleID, err)
		}
	}()
	
	// Broadcast to local WebSocket clients
	go s.broadcastGPSUpdate(gpsTrack)

	// Cache current location using new cache service
	go func() {
		if err := s.cache.SetCurrentLocationInCache(ctx, gpsTrack, 5*time.Minute); err != nil {
			fmt.Printf("Failed to cache current location %s: %v\n", gpsTrack.VehicleID, err)
		}
		
		// Invalidate location history cache for this vehicle
		if err := s.cache.InvalidateLocationHistoryCache(ctx, gpsTrack.VehicleID); err != nil {
			fmt.Printf("Failed to invalidate location history cache %s: %v\n", gpsTrack.VehicleID, err)
		}
	}()

	return gpsTrack, nil
}

// validateGPSCoordinates validates GPS coordinates and accuracy
func (s *Service) validateGPSCoordinates(lat, lng, accuracy float64) error {
	// Validate latitude
	if lat < -90 || lat > 90 {
		return apperrors.NewValidationError("invalid latitude: must be between -90 and 90 degrees")
	}

	// Validate longitude
	if lng < -180 || lng > 180 {
		return apperrors.NewValidationError("invalid longitude: must be between -180 and 180 degrees")
	}

	// Validate accuracy (filter out inaccurate readings)
	if accuracy > 50 {
		return apperrors.NewValidationError("GPS accuracy too low: accuracy must be less than 50 meters")
	}

	// Check for impossible coordinates (middle of ocean, etc.)
	if s.isImpossibleLocation(lat, lng) {
		return apperrors.NewValidationError("GPS coordinates appear to be invalid")
	}

	return nil
}

// isImpossibleLocation checks for obviously invalid coordinates
func (s *Service) isImpossibleLocation(lat, lng float64) bool {
	// Check for coordinates in the middle of oceans or impossible locations
	// This is a simple check - in production, you might use a more sophisticated approach
	
	// Middle of Pacific Ocean
	if lat == 0 && lng == 0 {
		return true
	}
	
	// North Pole (impossible for vehicles)
	if lat > 89 {
		return true
	}
	
	// South Pole (impossible for vehicles)
	if lat < -89 {
		return true
	}
	
	return false
}

// updateVehicleLocation updates the vehicle's last known location
func (s *Service) updateVehicleLocation(vehicleID string, lat, lng, _ float64, timestamp time.Time) error {
	return s.db.Model(&models.Vehicle{}).Where("id = ?", vehicleID).Updates(map[string]interface{}{
		"last_latitude":   lat,
		"last_longitude":  lng,
		"last_updated_at": timestamp,
	}).Error
}

// processDriverBehavior analyzes GPS data for driver behavior events
func (s *Service) processDriverBehavior(gpsTrack *models.GPSTrack) {
	// Get recent GPS tracks for this driver to analyze behavior
	var recentTracks []models.GPSTrack
	if err := s.db.Where("driver_id = ? AND timestamp > ?", 
		gpsTrack.DriverID, time.Now().Add(-5*time.Minute)).Order("timestamp DESC").Limit(10).Find(&recentTracks).Error; err != nil {
		return
	}

	// Check for speed violations (Indonesian speed limits)
	if gpsTrack.Speed > 80 { // 80 km/h is typical urban speed limit in Indonesia
		// Create driver event
		event := models.DriverEvent{
			DriverID:    *gpsTrack.DriverID,
			VehicleID:   gpsTrack.VehicleID,
			EventType:   "speed_violation",
			Severity:    s.getSpeedViolationSeverity(gpsTrack.Speed),
			Latitude:    gpsTrack.Latitude,
			Longitude:   gpsTrack.Longitude,
			Speed:       gpsTrack.Speed,
			Description: fmt.Sprintf("Speed violation: %.1f km/h", gpsTrack.Speed),
		}
		
		if err := s.createDriverEvent(event); err != nil {
			// Log error but don't fail the GPS tracking
			fmt.Printf("Failed to create speed violation event: %v\n", err)
		}
		
		// Create real-time alert
		go func() {
			// Get vehicle to get company ID
			var vehicle models.Vehicle
			if err := s.db.Where("id = ?", gpsTrack.VehicleID).First(&vehicle).Error; err == nil {
				if err := s.alertSystem.CreateSpeedViolationAlert(ctx, vehicle.CompanyID, gpsTrack.VehicleID, *gpsTrack.DriverID, gpsTrack.Speed, 80, fmt.Sprintf("%.6f,%.6f", gpsTrack.Latitude, gpsTrack.Longitude)); err != nil {
					fmt.Printf("Failed to create speed violation alert: %v\n", err)
				}
			}
		}()
	}

	// Check for harsh braking
	if len(recentTracks) >= 2 {
		prevTrack := recentTracks[1]
		deceleration := (prevTrack.Speed - gpsTrack.Speed) / float64(gpsTrack.Timestamp.Sub(prevTrack.Timestamp).Seconds())
		if deceleration > 3.5 { // m/s² threshold for harsh braking
			if err := s.createDriverEvent(models.DriverEvent{
				DriverID:    *gpsTrack.DriverID,
				VehicleID:   gpsTrack.VehicleID,
				EventType:   "harsh_braking",
				Severity:    s.getBrakingSeverity(deceleration),
				Latitude:    gpsTrack.Latitude,
				Longitude:   gpsTrack.Longitude,
				Speed:       gpsTrack.Speed,
				Description: fmt.Sprintf("Harsh braking: %.2f m/s²", deceleration),
			}); err != nil {
				// Log error but don't fail the GPS tracking
				fmt.Printf("Failed to create harsh braking event: %v\n", err)
			}
		}
	}

	// Check for rapid acceleration
	if len(recentTracks) >= 2 {
		prevTrack := recentTracks[1]
		acceleration := (gpsTrack.Speed - prevTrack.Speed) / float64(gpsTrack.Timestamp.Sub(prevTrack.Timestamp).Seconds())
		if acceleration > 2.5 { // m/s² threshold for rapid acceleration
			if err := s.createDriverEvent(models.DriverEvent{
				DriverID:    *gpsTrack.DriverID,
				VehicleID:   gpsTrack.VehicleID,
				EventType:   "rapid_acceleration",
				Severity:    s.getAccelerationSeverity(acceleration),
				Latitude:    gpsTrack.Latitude,
				Longitude:   gpsTrack.Longitude,
				Speed:       gpsTrack.Speed,
				Description: fmt.Sprintf("Rapid acceleration: %.2f m/s²", acceleration),
			}); err != nil {
				// Log error but don't fail the GPS tracking
				fmt.Printf("Failed to create rapid acceleration event: %v\n", err)
			}
		}
	}
}

// createDriverEvent creates a driver behavior event
func (s *Service) createDriverEvent(event models.DriverEvent) error {
	if err := s.db.Create(&event).Error; err != nil {
		return fmt.Errorf("failed to create driver event: %w", err)
	}
	
	// Broadcast event to WebSocket clients
	go s.broadcastDriverEvent(&event)
	
	return nil
}

// getSpeedViolationSeverity returns severity based on speed
func (s *Service) getSpeedViolationSeverity(speed float64) string {
	if speed > 120 {
		return "critical"
	} else if speed > 100 {
		return "high"
	} else if speed > 90 {
		return "medium"
	}
	return "low"
}

// getBrakingSeverity returns severity based on deceleration
func (s *Service) getBrakingSeverity(deceleration float64) string {
	if deceleration > 5.0 {
		return "critical"
	} else if deceleration > 4.0 {
		return "high"
	}
	return "medium"
}

// getAccelerationSeverity returns severity based on acceleration
func (s *Service) getAccelerationSeverity(acceleration float64) string {
	if acceleration > 3.5 {
		return "critical"
	} else if acceleration > 3.0 {
		return "high"
	}
	return "medium"
}

// broadcastGPSUpdate broadcasts GPS update to WebSocket clients
func (s *Service) broadcastGPSUpdate(gpsTrack *models.GPSTrack) {
	// Create GPS update message
	message := map[string]interface{}{
		"type":      "gps_update",
		"vehicle_id": gpsTrack.VehicleID,
		"latitude":  gpsTrack.Latitude,
		"longitude": gpsTrack.Longitude,
		"speed":     gpsTrack.Speed,
		"heading":   gpsTrack.Heading,
		"timestamp": gpsTrack.Timestamp,
	}
	
	// Marshal to JSON
	data, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Failed to marshal GPS update message: %v\n", err)
		return
	}
	
	// Broadcast to local WebSocket clients
	s.localWebSocketHub.Broadcast(data)
}

// broadcastDriverEvent broadcasts driver event to WebSocket clients
func (s *Service) broadcastDriverEvent(event *models.DriverEvent) {
	// Create driver event message
	message := map[string]interface{}{
		"type":        "driver_event",
		"driver_id":   event.DriverID,
		"vehicle_id":  event.VehicleID,
		"event_type":  event.EventType,
		"severity":    event.Severity,
		"latitude":    event.Latitude,
		"longitude":   event.Longitude,
		"speed":       event.Speed,
		"description": event.Description,
		"timestamp":   event.CreatedAt,
	}
	
	// Marshal to JSON
	data, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Failed to marshal driver event message: %v\n", err)
		return
	}
	
	// Broadcast to local WebSocket clients
	s.localWebSocketHub.Broadcast(data)
}


// GetCurrentLocation gets the current location of a vehicle
func (s *Service) GetCurrentLocation(vehicleID string) (*models.GPSTrack, error) {
	// Try to get from cache first
	cached, err := s.cache.GetCurrentLocationFromCache(ctx, vehicleID)
	if err != nil {
		// Log cache error but continue with database lookup
		fmt.Printf("Cache error for current location %s: %v\n", vehicleID, err)
	}
	
	if cached != nil {
		return cached, nil
	}

	// Get from database
	var gpsTrack models.GPSTrack
	if err := s.db.Where("vehicle_id = ?", vehicleID).Order("timestamp DESC").First(&gpsTrack).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("GPS data for vehicle")
		}
		return nil, apperrors.Wrap(err, "failed to get current location")
	}

	// Cache the result for 5 minutes (real-time data)
	if err := s.cache.SetCurrentLocationInCache(ctx, &gpsTrack, 5*time.Minute); err != nil {
		// Log cache error but don't fail the request
		fmt.Printf("Failed to cache current location %s: %v\n", vehicleID, err)
	}

	return &gpsTrack, nil
}


// GetLocationHistory gets historical GPS data for a vehicle
func (s *Service) GetLocationHistory(vehicleID string, filters GPSFilters) ([]models.GPSTrack, int64, error) {
	// Try to get from cache first
	cachedTracks, cachedTotal, err := s.cache.GetLocationHistoryFromCache(ctx, vehicleID, filters)
	if err != nil {
		// Log cache error but continue with database lookup
		fmt.Printf("Cache error for location history %s: %v\n", vehicleID, err)
	}
	
	if cachedTracks != nil {
		return cachedTracks, cachedTotal, nil
	}

	var gpsTracks []models.GPSTrack
	var total int64

	// Build query
	query := s.db.Model(&models.GPSTrack{}).Where("vehicle_id = ?", vehicleID)

	// Apply filters
	if filters.DriverID != nil {
		query = query.Where("driver_id = ?", *filters.DriverID)
	}
	if filters.StartTime != nil {
		query = query.Where("timestamp >= ?", *filters.StartTime)
	}
	if filters.EndTime != nil {
		query = query.Where("timestamp <= ?", *filters.EndTime)
	}
	if filters.MinAccuracy != nil {
		query = query.Where("accuracy <= ?", *filters.MinAccuracy)
	}
	if filters.MaxSpeed != nil {
		query = query.Where("speed <= ?", *filters.MaxSpeed)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to count GPS tracks")
	}

	// Apply sorting
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "timestamp"
	}
	sortOrder := strings.ToUpper(filters.SortOrder)
	if sortOrder == "" {
		sortOrder = "DESC"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 || limit > 1000 {
		limit = 100
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// Execute query
	if err := query.Find(&gpsTracks).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to get location history")
	}

	// Cache the result for 15 minutes (historical data)
	if err := s.cache.SetLocationHistoryInCache(ctx, vehicleID, filters, gpsTracks, total, 15*time.Minute); err != nil {
		// Log cache error but don't fail the request
		fmt.Printf("Failed to cache location history %s: %v\n", vehicleID, err)
	}

	return gpsTracks, total, nil
}

// ProcessDriverEvent processes a driver behavior event
func (s *Service) ProcessDriverEvent(req DriverEventRequest) (*models.DriverEvent, error) {
	// Validate vehicle and driver
	var vehicle models.Vehicle
	if err := s.db.Where("id = ? AND is_active = ?", req.VehicleID, true).First(&vehicle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("vehicle")
		}
		return nil, apperrors.Wrap(err, "failed to validate vehicle")
	}

	var driver models.Driver
	if err := s.db.Where("id = ? AND is_active = ?", req.DriverID, true).First(&driver).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("driver")
		}
		return nil, apperrors.Wrap(err, "failed to validate driver")
	}

	// Create driver event
	event := &models.DriverEvent{
		DriverID:    req.DriverID,
		VehicleID:   req.VehicleID,
		EventType:   req.EventType,
		Severity:    req.Severity,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Speed:       req.Speed,
		Description: req.Details,
	}

	// Save to database
	if err := s.db.Create(event).Error; err != nil {
		return nil, fmt.Errorf("failed to create driver event: %w", err)
	}

	// Broadcast to WebSocket clients
	go s.broadcastDriverEvent(event)

	// Update driver performance scores based on event
	go s.updateDriverPerformanceFromEvent(event)

	return event, nil
}

// updateDriverPerformanceFromEvent updates driver performance based on events
func (s *Service) updateDriverPerformanceFromEvent(event *models.DriverEvent) {
	// Get current driver performance scores
	var driver models.Driver
	if err := s.db.Where("id = ?", event.DriverID).First(&driver).Error; err != nil {
		return
	}

	// Calculate performance impact based on event type and severity
	impact := s.calculatePerformanceImpact(event.EventType, event.Severity)
	
	// Update performance scores
	newSafetyScore := math.Max(0, driver.SafetyScore-impact)
	driver.SafetyScore = newSafetyScore
	driver.OverallScore = (driver.PerformanceScore + driver.SafetyScore + driver.EfficiencyScore) / 3

	// Save updated scores
	s.db.Model(&driver).Updates(map[string]interface{}{
		"safety_score":   driver.SafetyScore,
		"overall_score":  driver.OverallScore,
	})
}

// calculatePerformanceImpact calculates performance impact based on event
func (s *Service) calculatePerformanceImpact(eventType, severity string) float64 {
	baseImpact := map[string]float64{
		"speed_violation":           5.0,
		"harsh_braking":            3.0,
		"rapid_acceleration":       3.0,
		"sharp_cornering":          2.0,
		"idle_time":                1.0,
		"driving_hours_violation":  10.0,
	}

	severityMultiplier := map[string]float64{
		"low":      0.5,
		"medium":   1.0,
		"high":     1.5,
		"critical": 2.0,
	}

	impact := baseImpact[eventType]
	if impact == 0 {
		impact = 1.0
	}

	multiplier := severityMultiplier[severity]
	if multiplier == 0 {
		multiplier = 1.0
	}

	return impact * multiplier
}

// HandleWebSocket handles WebSocket connections for real-time tracking
func (s *Service) HandleWebSocket(c *gin.Context) {
	s.websocketHub.HandleWebSocket(c)
}

// HandleLocalWebSocket handles WebSocket connections using the local hub
func (s *Service) HandleLocalWebSocket(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}
	
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("Failed to upgrade WebSocket connection: %v\n", err)
		return
	}
	
	// Register connection with local hub
	s.localWebSocketHub.Register(conn)
	
	// Send welcome message
	welcomeMsg := map[string]interface{}{
		"type":    "welcome",
		"message": "Connected to tracking WebSocket",
		"client_count": s.localWebSocketHub.GetClientCount(),
	}
	
	data, err := json.Marshal(welcomeMsg)
	if err == nil {
		conn.WriteMessage(websocket.TextMessage, data)
	}
	
	// Handle incoming messages and cleanup on disconnect
	defer func() {
		s.localWebSocketHub.Unregister(conn)
		conn.Close()
	}()
	
	// Keep connection alive and handle incoming messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}
	}
}

// GetWebSocketClientCount returns the number of connected WebSocket clients
func (s *Service) GetWebSocketClientCount() int {
	return s.localWebSocketHub.GetClientCount()
}


// StartTrip starts a new trip
func (s *Service) StartTrip(req TripRequest) (*models.Trip, error) {
	// Validate vehicle and driver
	var vehicle models.Vehicle
	if err := s.db.Where("id = ? AND is_active = ?", req.VehicleID, true).First(&vehicle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("vehicle")
		}
		return nil, apperrors.Wrap(err, "failed to validate vehicle")
	}

	var driver models.Driver
	if err := s.db.Where("id = ? AND is_active = ?", req.DriverID, true).First(&driver).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("driver")
		}
		return nil, apperrors.Wrap(err, "failed to validate driver")
	}

	// Create trip
	trip := &models.Trip{
		CompanyID:       vehicle.CompanyID,
		VehicleID:       req.VehicleID,
		DriverID:        &req.DriverID,
		StartTime:       &req.Timestamp,
		StartLatitude:   req.StartLocation.Latitude,
		StartLongitude:  req.StartLocation.Longitude,
		StartLocation:   req.StartLocation.Address,
		StartFuelLevel:  *req.OdometerStart, // Using odometer as fuel level for now
		Status:          "active",
	}

	// Save to database
	if err := s.db.Create(trip).Error; err != nil {
		return nil, fmt.Errorf("failed to create trip: %w", err)
	}

	// Cache the trip for 1 hour (active trips are frequently accessed)
	if err := s.cache.SetTripInCache(ctx, trip, 1*time.Hour); err != nil {
		fmt.Printf("Failed to cache trip %s: %v\n", trip.ID, err)
	}

	// Broadcast real-time trip update
	go func() {
		if err := s.analyticsBroadcaster.BroadcastTripUpdate(ctx, trip); err != nil {
			fmt.Printf("Failed to broadcast trip update %s: %v\n", trip.ID, err)
		}
	}()

	return trip, nil
}

// EndTrip ends an active trip
func (s *Service) EndTrip(req TripRequest) (*models.Trip, error) {
	// Find active trip
	var trip models.Trip
	if err := s.db.Where("vehicle_id = ? AND driver_id = ? AND status = ?",
		req.VehicleID, req.DriverID, "active").First(&trip).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("active trip")
		}
		return nil, apperrors.Wrap(err, "failed to find active trip")
	}

	// Update trip end data
	trip.EndTime = &req.Timestamp
	trip.EndLatitude = req.EndLocation.Latitude
	trip.EndLongitude = req.EndLocation.Longitude
	trip.EndLocation = req.EndLocation.Address
	trip.EndFuelLevel = *req.OdometerEnd // Using odometer as fuel level for now
	trip.Status = "completed"

	// Calculate trip metrics
	if err := s.calculateTripMetrics(&trip); err != nil {
		return nil, fmt.Errorf("failed to calculate trip metrics: %w", err)
	}

	// Save updated trip
	if err := s.db.Save(&trip).Error; err != nil {
		return nil, fmt.Errorf("failed to update trip: %w", err)
	}

	// Update trip cache with completed trip data
	if err := s.cache.SetTripInCache(ctx, &trip, 2*time.Hour); err != nil {
		fmt.Printf("Failed to cache completed trip %s: %v\n", trip.ID, err)
	}

	// Broadcast real-time trip update
	go func() {
		if err := s.analyticsBroadcaster.BroadcastTripUpdate(ctx, &trip); err != nil {
			fmt.Printf("Failed to broadcast trip update %s: %v\n", trip.ID, err)
		}
	}()

	return &trip, nil
}

// calculateTripMetrics calculates trip distance, duration, and other metrics
func (s *Service) calculateTripMetrics(trip *models.Trip) error {
	// Calculate duration
	if trip.EndTime != nil && trip.StartTime != nil {
		trip.TotalDuration = int(trip.EndTime.Sub(*trip.StartTime).Seconds())
	}

	// Calculate distance using GPS tracks
	var gpsTracks []models.GPSTrack
	if err := s.db.Where("vehicle_id = ? AND driver_id = ? AND timestamp BETWEEN ? AND ?",
		trip.VehicleID, *trip.DriverID, trip.StartTime, trip.EndTime).Order("timestamp ASC").Find(&gpsTracks).Error; err != nil {
		return err
	}

	// Calculate total distance
	var totalDistance float64
	var maxSpeed float64
	var totalSpeed float64
	var speedCount int

	for i := 1; i < len(gpsTracks); i++ {
		prev := gpsTracks[i-1]
		curr := gpsTracks[i]
		
		// Calculate distance between two points (Haversine formula)
		distance := s.calculateDistance(prev.Latitude, prev.Longitude, curr.Latitude, curr.Longitude)
		totalDistance += distance
		
		// Track max speed
		if curr.Speed > maxSpeed {
			maxSpeed = curr.Speed
		}
		
		// Calculate average speed
		if curr.Speed > 0 {
			totalSpeed += curr.Speed
			speedCount++
		}
	}

	trip.TotalDistance = totalDistance
	trip.MaxSpeed = maxSpeed
	if speedCount > 0 {
		trip.AverageSpeed = totalSpeed / float64(speedCount)
	}

	return nil
}

// calculateDistance calculates distance between two GPS points using Haversine formula
func (s *Service) calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371e3 // Earth's radius in meters
	
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// CreateGeofence creates a new geofence
func (s *Service) CreateGeofence(req GeofenceRequest) (*models.Geofence, error) {
	// Validate company exists
	var company models.Company
	if err := s.db.Where("id = ? AND is_active = ?", req.CompanyID, true).First(&company).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("company")
		}
		return nil, apperrors.Wrap(err, "failed to validate company")
	}

	// Create geofence
	geofence := &models.Geofence{
		CompanyID:       req.CompanyID,
		Name:            req.Name,
		Type:            req.Type,
		CenterLatitude:  req.CenterLat,
		CenterLongitude: req.CenterLng,
		Radius:          req.Radius,
		AlertOnEnter:    req.AlertOnEntry,
		AlertOnExit:     req.AlertOnExit,
		IsActive:        req.IsActive,
		Description:     req.Description,
	}

	// Save to database
	if err := s.db.Create(geofence).Error; err != nil {
		return nil, fmt.Errorf("failed to create geofence: %w", err)
	}

	// Cache the geofence for 1 hour
	if err := s.cache.SetGeofenceInCache(ctx, geofence, 1*time.Hour); err != nil {
		fmt.Printf("Failed to cache geofence %s: %v\n", geofence.ID, err)
	}
	
	// Invalidate company geofences cache
	if err := s.cache.InvalidateGeofencesByCompanyCache(ctx, req.CompanyID); err != nil {
		fmt.Printf("Failed to invalidate company geofences cache %s: %v\n", req.CompanyID, err)
	}

	return geofence, nil
}

// CheckGeofenceViolations checks if a GPS point violates any geofences
func (s *Service) CheckGeofenceViolations(vehicleID string, lat, lng float64) ([]models.Geofence, error) {
	// Get vehicle's company ID
	var vehicle models.Vehicle
	if err := s.db.Where("id = ?", vehicleID).First(&vehicle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("vehicle")
		}
		return nil, apperrors.Wrap(err, "failed to get vehicle")
	}

	// Try to get geofences from cache first
	geofences, err := s.cache.GetGeofencesByCompanyFromCache(ctx, vehicle.CompanyID)
	if err != nil {
		// Log cache error but continue with database lookup
		fmt.Printf("Cache error for geofences %s: %v\n", vehicle.CompanyID, err)
	}
	
	if geofences == nil {
		// Get active geofences for the company from database
		if err := s.db.Where("company_id = ? AND is_active = ?", vehicle.CompanyID, true).Find(&geofences).Error; err != nil {
			return nil, apperrors.Wrap(err, "failed to get geofences")
		}
		
		// Cache the geofences for 30 minutes
		if err := s.cache.SetGeofencesByCompanyInCache(ctx, vehicle.CompanyID, geofences, 30*time.Minute); err != nil {
			fmt.Printf("Failed to cache geofences %s: %v\n", vehicle.CompanyID, err)
		}
	}

	// Check each geofence
	var violations []models.Geofence
	for _, geofence := range geofences {
		distance := s.calculateDistance(lat, lng, geofence.CenterLatitude, geofence.CenterLongitude)
		if distance <= geofence.Radius {
			violations = append(violations, geofence)
		}
	}

	return violations, nil
}
