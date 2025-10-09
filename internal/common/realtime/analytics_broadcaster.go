package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/internal/common/repository"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// AnalyticsBroadcaster provides real-time analytics broadcasting
type AnalyticsBroadcaster struct {
	hub         *WebSocketHub
	redis       *redis.Client
	db          *gorm.DB
	repoManager *repository.RepositoryManager
}

// NewAnalyticsBroadcaster creates a new analytics broadcaster
func NewAnalyticsBroadcaster(hub *WebSocketHub, redis *redis.Client, db *gorm.DB, repoManager *repository.RepositoryManager) *AnalyticsBroadcaster {
	return &AnalyticsBroadcaster{
		hub:         hub,
		redis:       redis,
		db:          db,
		repoManager: repoManager,
	}
}

// FleetDashboardUpdate represents a real-time fleet dashboard update
type FleetDashboardUpdate struct {
	Type           string    `json:"type"`
	CompanyID      string    `json:"company_id"`
	ActiveVehicles int       `json:"active_vehicles"`
	TotalTrips     int       `json:"total_trips"`
	DistanceTraveled float64 `json:"distance_traveled"`
	FuelConsumed   float64   `json:"fuel_consumed"`
	UtilizationRate float64  `json:"utilization_rate"`
	CostPerKm      float64   `json:"cost_per_km"`
	Timestamp      time.Time `json:"timestamp"`
}

// VehicleLocationUpdate represents a real-time vehicle location update
type VehicleLocationUpdate struct {
	Type        string    `json:"type"`
	CompanyID   string    `json:"company_id"`
	VehicleID   string    `json:"vehicle_id"`
	DriverID    string    `json:"driver_id"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Speed       float64   `json:"speed"`
	Heading     float64   `json:"heading"`
	Timestamp   time.Time `json:"timestamp"`
}

// DriverEventUpdate represents a real-time driver event update
type DriverEventUpdate struct {
	Type        string    `json:"type"`
	CompanyID   string    `json:"company_id"`
	DriverID    string    `json:"driver_id"`
	VehicleID   string    `json:"vehicle_id"`
	EventType   string    `json:"event_type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// GeofenceViolationUpdate represents a real-time geofence violation update
type GeofenceViolationUpdate struct {
	Type        string    `json:"type"`
	CompanyID   string    `json:"company_id"`
	VehicleID   string    `json:"vehicle_id"`
	DriverID    string    `json:"driver_id"`
	GeofenceID  string    `json:"geofence_id"`
	GeofenceName string   `json:"geofence_name"`
	ViolationType string  `json:"violation_type"` // "enter" or "exit"
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Timestamp   time.Time `json:"timestamp"`
}

// TripUpdate represents a real-time trip update
type TripUpdate struct {
	Type        string    `json:"type"`
	CompanyID   string    `json:"company_id"`
	TripID      string    `json:"trip_id"`
	VehicleID   string    `json:"vehicle_id"`
	DriverID    string    `json:"driver_id"`
	Status      string    `json:"status"` // "started", "completed", "cancelled"
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Distance    float64   `json:"distance"`
	Duration    float64   `json:"duration"` // in minutes
	Timestamp   time.Time `json:"timestamp"`
}

// MaintenanceAlertUpdate represents a real-time maintenance alert update
type MaintenanceAlertUpdate struct {
	Type        string    `json:"type"`
	CompanyID   string    `json:"company_id"`
	VehicleID   string    `json:"vehicle_id"`
	AlertType   string    `json:"alert_type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// BroadcastFleetDashboardUpdate broadcasts a fleet dashboard update
func (ab *AnalyticsBroadcaster) BroadcastFleetDashboardUpdate(ctx context.Context, companyID string) error {
	// Get current fleet dashboard data
	update, err := ab.generateFleetDashboardUpdate(ctx, companyID)
	if err != nil {
		return fmt.Errorf("failed to generate fleet dashboard update: %w", err)
	}
	
	// Broadcast to company clients
	message := WebSocketMessage{
		Type:      "fleet_dashboard_update",
		Data:      update,
		Timestamp: time.Now(),
		CompanyID: companyID,
	}
	
	ab.hub.BroadcastToCompany(companyID, message)
	
	// Also publish to Redis for cross-instance communication
	return ab.publishToRedis("fleet_dashboard_update", message)
}

// BroadcastVehicleLocationUpdate broadcasts a vehicle location update
func (ab *AnalyticsBroadcaster) BroadcastVehicleLocationUpdate(ctx context.Context, gpsTrack *models.GPSTrack) error {
	// Get vehicle and driver information
	vehicle, err := ab.repoManager.GetVehicles().GetByID(ctx, gpsTrack.VehicleID)
	if err != nil {
		return fmt.Errorf("failed to get vehicle: %w", err)
	}
	
	update := VehicleLocationUpdate{
		Type:        "vehicle_location_update",
		CompanyID:   vehicle.CompanyID,
		VehicleID:   gpsTrack.VehicleID,
		DriverID:    *gpsTrack.DriverID,
		Latitude:    gpsTrack.Latitude,
		Longitude:   gpsTrack.Longitude,
		Speed:       gpsTrack.Speed,
		Heading:     gpsTrack.Heading,
		Timestamp:   gpsTrack.Timestamp,
	}
	
	// Broadcast to company clients
	message := WebSocketMessage{
		Type:      "vehicle_location_update",
		Data:      update,
		Timestamp: time.Now(),
		CompanyID: vehicle.CompanyID,
	}
	
	ab.hub.BroadcastToCompany(vehicle.CompanyID, message)
	
	// Publish to Redis for cross-instance communication
	return ab.publishToRedis("vehicle_location_update", message)
}

// BroadcastDriverEventUpdate broadcasts a driver event update
func (ab *AnalyticsBroadcaster) BroadcastDriverEventUpdate(ctx context.Context, event *models.DriverEvent) error {
	// Get vehicle information
	vehicle, err := ab.repoManager.GetVehicles().GetByID(ctx, event.VehicleID)
	if err != nil {
		return fmt.Errorf("failed to get vehicle: %w", err)
	}
	
	update := DriverEventUpdate{
		Type:        "driver_event_update",
		CompanyID:   vehicle.CompanyID,
		DriverID:    event.DriverID,
		VehicleID:   event.VehicleID,
		EventType:   event.EventType,
		Severity:    event.Severity,
		Description: event.Description,
		Timestamp:   event.CreatedAt,
	}
	
	// Broadcast to company clients
	message := WebSocketMessage{
		Type:      "driver_event_update",
		Data:      update,
		Timestamp: time.Now(),
		CompanyID: vehicle.CompanyID,
	}
	
	ab.hub.BroadcastToCompany(vehicle.CompanyID, message)
	
	// Publish to Redis for cross-instance communication
	return ab.publishToRedis("driver_event_update", message)
}

// BroadcastGeofenceViolationUpdate broadcasts a geofence violation update
func (ab *AnalyticsBroadcaster) BroadcastGeofenceViolationUpdate(ctx context.Context, vehicleID, driverID, geofenceID, geofenceName, violationType string, lat, lng float64) error {
	// Get vehicle information
	vehicle, err := ab.repoManager.GetVehicles().GetByID(ctx, vehicleID)
	if err != nil {
		return fmt.Errorf("failed to get vehicle: %w", err)
	}
	
	update := GeofenceViolationUpdate{
		Type:          "geofence_violation_update",
		CompanyID:     vehicle.CompanyID,
		VehicleID:     vehicleID,
		DriverID:      driverID,
		GeofenceID:    geofenceID,
		GeofenceName:  geofenceName,
		ViolationType: violationType,
		Latitude:      lat,
		Longitude:     lng,
		Timestamp:     time.Now(),
	}
	
	// Broadcast to company clients
	message := WebSocketMessage{
		Type:      "geofence_violation_update",
		Data:      update,
		Timestamp: time.Now(),
		CompanyID: vehicle.CompanyID,
	}
	
	ab.hub.BroadcastToCompany(vehicle.CompanyID, message)
	
	// Publish to Redis for cross-instance communication
	return ab.publishToRedis("geofence_violation_update", message)
}

// BroadcastTripUpdate broadcasts a trip update
func (ab *AnalyticsBroadcaster) BroadcastTripUpdate(ctx context.Context, trip *models.Trip) error {
	update := TripUpdate{
		Type:      "trip_update",
		CompanyID: trip.CompanyID,
		TripID:    trip.ID,
		VehicleID: trip.VehicleID,
		DriverID:  *trip.DriverID,
		Status:    trip.Status,
		StartTime: trip.StartTime,
		EndTime:   trip.EndTime,
		Distance:  trip.TotalDistance,
		Duration:  float64(trip.TotalDuration),
		Timestamp: time.Now(),
	}
	
	// Broadcast to company clients
	message := WebSocketMessage{
		Type:      "trip_update",
		Data:      update,
		Timestamp: time.Now(),
		CompanyID: trip.CompanyID,
	}
	
	ab.hub.BroadcastToCompany(trip.CompanyID, message)
	
	// Publish to Redis for cross-instance communication
	return ab.publishToRedis("trip_update", message)
}

// BroadcastMaintenanceAlertUpdate broadcasts a maintenance alert update
func (ab *AnalyticsBroadcaster) BroadcastMaintenanceAlertUpdate(ctx context.Context, companyID, vehicleID, alertType, severity, description string) error {
	update := MaintenanceAlertUpdate{
		Type:        "maintenance_alert_update",
		CompanyID:   companyID,
		VehicleID:   vehicleID,
		AlertType:   alertType,
		Severity:    severity,
		Description: description,
		Timestamp:   time.Now(),
	}
	
	// Broadcast to company clients
	message := WebSocketMessage{
		Type:      "maintenance_alert_update",
		Data:      update,
		Timestamp: time.Now(),
		CompanyID: companyID,
	}
	
	ab.hub.BroadcastToCompany(companyID, message)
	
	// Publish to Redis for cross-instance communication
	return ab.publishToRedis("maintenance_alert_update", message)
}

// generateFleetDashboardUpdate generates a fleet dashboard update
func (ab *AnalyticsBroadcaster) generateFleetDashboardUpdate(ctx context.Context, companyID string) (*FleetDashboardUpdate, error) {
	// Get active vehicles
	vehicleFilters := repository.FilterOptions{
		CompanyID: companyID,
		Where: map[string]interface{}{
			"status": "active",
		},
	}
	vehicles, err := ab.repoManager.GetVehicles().List(ctx, vehicleFilters, repository.Pagination{Page: 1, PageSize: 1000})
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicles: %w", err)
	}
	
	// Get recent trips
	tripFilters := repository.FilterOptions{
		CompanyID: companyID,
	}
	trips, err := ab.repoManager.GetTrips().List(ctx, tripFilters, repository.Pagination{Page: 1, PageSize: 1000})
	if err != nil {
		return nil, fmt.Errorf("failed to get trips: %w", err)
	}
	
	// Calculate metrics
	activeVehicles := len(vehicles)
	totalTrips := len(trips)
	
	distanceTraveled := 0.0
	fuelConsumed := 0.0
	
	for _, trip := range trips {
		if trip.TotalDistance > 0 {
			distanceTraveled += trip.TotalDistance
		}
		if trip.FuelConsumed > 0 {
			fuelConsumed += trip.FuelConsumed
		}
	}
	
	// Calculate utilization rate
	utilizationRate := 0.0
	if activeVehicles > 0 {
		utilizationRate = float64(totalTrips) / float64(activeVehicles) * 100
	}
	
	// Calculate cost per km (simplified)
	costPerKm := 0.0
	if distanceTraveled > 0 {
		costPerKm = (fuelConsumed * 15000) / distanceTraveled // IDR per km
	}
	
	return &FleetDashboardUpdate{
		Type:            "fleet_dashboard_update",
		CompanyID:       companyID,
		ActiveVehicles:  activeVehicles,
		TotalTrips:      totalTrips,
		DistanceTraveled: distanceTraveled,
		FuelConsumed:    fuelConsumed,
		UtilizationRate: utilizationRate,
		CostPerKm:       costPerKm,
		Timestamp:       time.Now(),
	}, nil
}

// publishToRedis publishes a message to Redis for cross-instance communication
func (ab *AnalyticsBroadcaster) publishToRedis(_ string, message WebSocketMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	return ab.redis.Publish(context.Background(), "fleet_tracker:websocket", data).Err()
}

// StartPeriodicDashboardUpdates starts periodic dashboard updates
func (ab *AnalyticsBroadcaster) StartPeriodicDashboardUpdates(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get all companies and broadcast dashboard updates
			var companies []models.Company
			if err := ab.db.Where("is_active = ?", true).Find(&companies).Error; err != nil {
				fmt.Printf("Failed to get companies for dashboard updates: %v\n", err)
				continue
			}
			
			for _, company := range companies {
				if err := ab.BroadcastFleetDashboardUpdate(ctx, company.ID); err != nil {
					fmt.Printf("Failed to broadcast dashboard update for company %s: %v\n", company.ID, err)
				}
			}
		}
	}
}
