package fleet

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// FuelManager provides comprehensive fuel management capabilities
type FuelManager struct {
	db    *gorm.DB
	redis *redis.Client
}

// FuelRecord represents a fuel consumption record
type FuelRecord struct {
	ID              string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	VehicleID       string    `json:"vehicle_id" gorm:"type:uuid;not null;index"`
	DriverID        string    `json:"driver_id" gorm:"type:uuid;index"`
	CompanyID       string    `json:"company_id" gorm:"type:uuid;not null;index"`
	
	// Fuel Data
	FuelType        string    `json:"fuel_type" gorm:"type:varchar(20);not null"` // gasoline, diesel, electric
	Quantity        float64   `json:"quantity" gorm:"not null"` // liters or kWh
	UnitPrice       float64   `json:"unit_price" gorm:"not null"` // IDR per liter/kWh
	TotalCost       float64   `json:"total_cost" gorm:"not null"` // IDR
	StationName     string    `json:"station_name" gorm:"type:varchar(255)"`
	StationLocation string    `json:"station_location" gorm:"type:varchar(500)"`
	
	// Odometer and Performance
	OdometerReading float64   `json:"odometer_reading" gorm:"not null"` // km
	PreviousReading float64   `json:"previous_reading" gorm:"not null"` // km
	DistanceTraveled float64  `json:"distance_traveled" gorm:"not null"` // km
	FuelEfficiency  float64   `json:"fuel_efficiency" gorm:"not null"` // km/liter or km/kWh
	
	// Trip Information
	TripID          *string   `json:"trip_id" gorm:"type:uuid;index"`
	RouteID         *string   `json:"route_id" gorm:"type:uuid;index"`
	
	// Environmental Data
	CO2Emission     float64   `json:"co2_emission" gorm:"not null"` // kg CO2
	CarbonFootprint float64   `json:"carbon_footprint" gorm:"not null"` // kg CO2/km
	
	// Metadata
	RecordedAt      time.Time `json:"recorded_at" gorm:"not null"`
	CreatedAt       time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"not null"`
}

// FuelAnalytics represents fuel consumption analytics
type FuelAnalytics struct {
	Period          string    `json:"period"`
	TotalFuel       float64   `json:"total_fuel"` // liters or kWh
	TotalCost       float64   `json:"total_cost"` // IDR
	TotalDistance   float64   `json:"total_distance"` // km
	AverageEfficiency float64 `json:"average_efficiency"` // km/liter or km/kWh
	AverageCost     float64   `json:"average_cost"` // IDR per km
	CO2Emission     float64   `json:"co2_emission"` // kg CO2
	FuelTrend       []FuelTrendPoint `json:"fuel_trend"`
	TopConsumers    []VehicleFuelStats `json:"top_consumers"`
	EfficiencyRanking []VehicleEfficiency `json:"efficiency_ranking"`
}

// FuelTrendPoint represents a point in fuel consumption trend
type FuelTrendPoint struct {
	Date        time.Time `json:"date"`
	FuelUsed    float64   `json:"fuel_used"`
	Distance    float64   `json:"distance"`
	Efficiency  float64   `json:"efficiency"`
	Cost        float64   `json:"cost"`
}

// VehicleFuelStats represents fuel statistics for a vehicle
type VehicleFuelStats struct {
	VehicleID       string  `json:"vehicle_id"`
	LicensePlate    string  `json:"license_plate"`
	Make            string  `json:"make"`
	Model           string  `json:"model"`
	TotalFuel       float64 `json:"total_fuel"`
	TotalCost       float64 `json:"total_cost"`
	TotalDistance   float64 `json:"total_distance"`
	AverageEfficiency float64 `json:"average_efficiency"`
	FuelCostPerKm   float64 `json:"fuel_cost_per_km"`
}

// VehicleEfficiency represents vehicle efficiency ranking
type VehicleEfficiency struct {
	VehicleID       string  `json:"vehicle_id"`
	LicensePlate    string  `json:"license_plate"`
	Make            string  `json:"make"`
	Model           string  `json:"model"`
	Efficiency      float64 `json:"efficiency"` // km/liter or km/kWh
	Rank            int     `json:"rank"`
	Improvement     float64 `json:"improvement"` // % improvement from previous period
}

// FuelAlert represents a fuel-related alert
type FuelAlert struct {
	ID          string    `json:"id"`
	VehicleID   string    `json:"vehicle_id"`
	AlertType   string    `json:"alert_type"` // low_efficiency, high_consumption, fuel_theft, maintenance
	Severity    string    `json:"severity"` // low, medium, high, critical
	Message     string    `json:"message"`
	Threshold   float64   `json:"threshold"`
	ActualValue float64   `json:"actual_value"`
	CreatedAt   time.Time `json:"created_at"`
	IsResolved  bool      `json:"is_resolved"`
}

// NewFuelManager creates a new fuel manager
func NewFuelManager(db *gorm.DB, redis *redis.Client) *FuelManager {
	return &FuelManager{
		db:    db,
		redis: redis,
	}
}

// RecordFuelConsumption records a fuel consumption event
func (fm *FuelManager) RecordFuelConsumption(ctx context.Context, record *FuelRecord) error {
	// Validate fuel record
	if err := fm.validateFuelRecord(record); err != nil {
		return fmt.Errorf("fuel record validation failed: %w", err)
	}

	// Calculate derived fields
	fm.calculateDerivedFields(record)

	// Save to database
	if err := fm.db.Create(record).Error; err != nil {
		return fmt.Errorf("failed to save fuel record: %w", err)
	}

	// Update vehicle fuel statistics
	if err := fm.updateVehicleFuelStats(ctx, record); err != nil {
		return fmt.Errorf("failed to update vehicle fuel stats: %w", err)
	}

	// Check for fuel alerts
	go fm.checkFuelAlerts(ctx, record)

	// Invalidate cache
	fm.invalidateFuelCache(ctx, record.CompanyID, record.VehicleID)

	return nil
}

// GetFuelAnalytics retrieves fuel consumption analytics
func (fm *FuelManager) GetFuelAnalytics(ctx context.Context, companyID string, period string, startDate, endDate time.Time) (*FuelAnalytics, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("fuel_analytics:%s:%s:%s:%s", companyID, period, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	cached, err := fm.getCachedAnalytics(ctx, cacheKey)
	if err == nil && cached != nil {
		return cached, nil
	}

	// Build query
	query := fm.db.Model(&FuelRecord{}).Where("company_id = ? AND recorded_at BETWEEN ? AND ?", companyID, startDate, endDate)

	// Get total fuel consumption
	var totalFuel, totalCost, totalDistance, totalCO2 float64
	var count int64

	err = query.Select("SUM(quantity) as total_fuel, SUM(total_cost) as total_cost, SUM(distance_traveled) as total_distance, SUM(co2_emission) as total_co2, COUNT(*) as count").
		Row().Scan(&totalFuel, &totalCost, &totalDistance, &totalCO2, &count)
	if err != nil {
		return nil, fmt.Errorf("failed to get fuel totals: %w", err)
	}

	// Calculate average efficiency
	var averageEfficiency float64
	if totalFuel > 0 {
		averageEfficiency = totalDistance / totalFuel
	}

	// Calculate average cost per km
	var averageCost float64
	if totalDistance > 0 {
		averageCost = totalCost / totalDistance
	}

	// Get fuel trend
	trend, err := fm.getFuelTrend(ctx, companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get fuel trend: %w", err)
	}

	// Get top consumers
	topConsumers, err := fm.getTopFuelConsumers(ctx, companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get top consumers: %w", err)
	}

	// Get efficiency ranking
	efficiencyRanking, err := fm.getEfficiencyRanking(ctx, companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get efficiency ranking: %w", err)
	}

	analytics := &FuelAnalytics{
		Period:            period,
		TotalFuel:         totalFuel,
		TotalCost:         totalCost,
		TotalDistance:     totalDistance,
		AverageEfficiency: averageEfficiency,
		AverageCost:       averageCost,
		CO2Emission:       totalCO2,
		FuelTrend:         trend,
		TopConsumers:      topConsumers,
		EfficiencyRanking: efficiencyRanking,
	}

	// Cache the result
	fm.cacheAnalytics(ctx, cacheKey, analytics, 30*time.Minute)

	return analytics, nil
}

// GetVehicleFuelHistory retrieves fuel history for a specific vehicle
func (fm *FuelManager) GetVehicleFuelHistory(ctx context.Context, vehicleID string, limit int) ([]FuelRecord, error) {
	var records []FuelRecord
	
	query := fm.db.Where("vehicle_id = ?", vehicleID).
		Order("recorded_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle fuel history: %w", err)
	}
	
	return records, nil
}

// GetFuelAlerts retrieves active fuel alerts
func (fm *FuelManager) GetFuelAlerts(ctx context.Context, companyID string, severity string) ([]FuelAlert, error) {
	var alerts []FuelAlert
	
	query := fm.db.Where("company_id = ? AND is_resolved = false", companyID)
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	
	err := query.Order("created_at DESC").Find(&alerts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get fuel alerts: %w", err)
	}
	
	return alerts, nil
}

// PredictFuelConsumption predicts fuel consumption for a route
func (fm *FuelManager) PredictFuelConsumption(ctx context.Context, vehicleID string, distance float64, routeType string) (float64, float64, error) {
	// Get vehicle's average efficiency
	var avgEfficiency float64
	err := fm.db.Model(&FuelRecord{}).
		Where("vehicle_id = ? AND recorded_at >= ?", vehicleID, time.Now().AddDate(0, 0, -30)).
		Select("AVG(fuel_efficiency)").
		Row().Scan(&avgEfficiency)
	
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get vehicle efficiency: %w", err)
	}
	
	// Adjust efficiency based on route type
	routeMultiplier := fm.getRouteTypeMultiplier(routeType)
	adjustedEfficiency := avgEfficiency * routeMultiplier
	
	// Calculate predicted fuel consumption
	predictedFuel := distance / adjustedEfficiency
	
	// Get current fuel price
	fuelPrice := fm.getCurrentFuelPrice()
	predictedCost := predictedFuel * fuelPrice
	
	return predictedFuel, predictedCost, nil
}

// validateFuelRecord validates a fuel record
func (fm *FuelManager) validateFuelRecord(record *FuelRecord) error {
	if record.VehicleID == "" {
		return fmt.Errorf("vehicle ID is required")
	}
	if record.CompanyID == "" {
		return fmt.Errorf("company ID is required")
	}
	if record.Quantity <= 0 {
		return fmt.Errorf("fuel quantity must be positive")
	}
	if record.UnitPrice <= 0 {
		return fmt.Errorf("unit price must be positive")
	}
	if record.OdometerReading <= 0 {
		return fmt.Errorf("odometer reading must be positive")
	}
	if record.PreviousReading < 0 {
		return fmt.Errorf("previous reading cannot be negative")
	}
	if record.OdometerReading < record.PreviousReading {
		return fmt.Errorf("odometer reading cannot be less than previous reading")
	}
	return nil
}

// calculateDerivedFields calculates derived fields for a fuel record
func (fm *FuelManager) calculateDerivedFields(record *FuelRecord) {
	// Calculate total cost
	record.TotalCost = record.Quantity * record.UnitPrice
	
	// Calculate distance traveled
	record.DistanceTraveled = record.OdometerReading - record.PreviousReading
	
	// Calculate fuel efficiency
	if record.Quantity > 0 {
		record.FuelEfficiency = record.DistanceTraveled / record.Quantity
	}
	
	// Calculate CO2 emission (approximate values)
	co2PerLiter := fm.getCO2EmissionFactor(record.FuelType)
	record.CO2Emission = record.Quantity * co2PerLiter
	
	// Calculate carbon footprint
	if record.DistanceTraveled > 0 {
		record.CarbonFootprint = record.CO2Emission / record.DistanceTraveled
	}
	
	// Set recorded time if not set
	if record.RecordedAt.IsZero() {
		record.RecordedAt = time.Now()
	}
}

// updateVehicleFuelStats updates vehicle fuel statistics
func (fm *FuelManager) updateVehicleFuelStats(_ context.Context, _ *FuelRecord) error {
	// This would update aggregated fuel statistics for the vehicle
	// For now, we'll just return nil as the implementation would depend on
	// the specific requirements for vehicle fuel statistics
	return nil
}

// checkFuelAlerts checks for fuel-related alerts
func (fm *FuelManager) checkFuelAlerts(ctx context.Context, record *FuelRecord) {
	// Check for low efficiency alert
	if record.FuelEfficiency < 8.0 { // Less than 8 km/liter
		alert := &FuelAlert{
			VehicleID:   record.VehicleID,
			AlertType:   "low_efficiency",
			Severity:    "medium",
			Message:     fmt.Sprintf("Low fuel efficiency detected: %.2f km/liter", record.FuelEfficiency),
			Threshold:   8.0,
			ActualValue: record.FuelEfficiency,
			CreatedAt:   time.Now(),
			IsResolved:  false,
		}
		fm.createFuelAlert(ctx, alert)
	}
	
	// Check for high consumption alert
	if record.Quantity > 50.0 { // More than 50 liters
		alert := &FuelAlert{
			VehicleID:   record.VehicleID,
			AlertType:   "high_consumption",
			Severity:    "high",
			Message:     fmt.Sprintf("High fuel consumption detected: %.2f liters", record.Quantity),
			Threshold:   50.0,
			ActualValue: record.Quantity,
			CreatedAt:   time.Now(),
			IsResolved:  false,
		}
		fm.createFuelAlert(ctx, alert)
	}
}

// createFuelAlert creates a fuel alert
func (fm *FuelManager) createFuelAlert(_ context.Context, alert *FuelAlert) error {
	alert.ID = fmt.Sprintf("alert_%d", time.Now().UnixNano())
	return fm.db.Create(alert).Error
}

// getFuelTrend retrieves fuel consumption trend
func (fm *FuelManager) getFuelTrend(_ context.Context, companyID string, startDate, endDate time.Time) ([]FuelTrendPoint, error) {
	var trend []FuelTrendPoint
	
	// Group by date and calculate daily totals
	rows, err := fm.db.Model(&FuelRecord{}).
		Select("DATE(recorded_at) as date, SUM(quantity) as fuel_used, SUM(distance_traveled) as distance, AVG(fuel_efficiency) as efficiency, SUM(total_cost) as cost").
		Where("company_id = ? AND recorded_at BETWEEN ? AND ?", companyID, startDate, endDate).
		Group("DATE(recorded_at)").
		Order("date ASC").
		Rows()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get fuel trend: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var point FuelTrendPoint
		var dateStr string
		
		err := rows.Scan(&dateStr, &point.FuelUsed, &point.Distance, &point.Efficiency, &point.Cost)
		if err != nil {
			continue
		}
		
		point.Date, _ = time.Parse("2006-01-02", dateStr)
		trend = append(trend, point)
	}
	
	return trend, nil
}

// getTopFuelConsumers retrieves top fuel consuming vehicles
func (fm *FuelManager) getTopFuelConsumers(_ context.Context, companyID string, startDate, endDate time.Time) ([]VehicleFuelStats, error) {
	var consumers []VehicleFuelStats
	
	rows, err := fm.db.Table("fuel_records fr").
		Select("fr.vehicle_id, v.license_plate, v.make, v.model, SUM(fr.quantity) as total_fuel, SUM(fr.total_cost) as total_cost, SUM(fr.distance_traveled) as total_distance, AVG(fr.fuel_efficiency) as average_efficiency").
		Joins("JOIN vehicles v ON fr.vehicle_id = v.id").
		Where("fr.company_id = ? AND fr.recorded_at BETWEEN ? AND ?", companyID, startDate, endDate).
		Group("fr.vehicle_id, v.license_plate, v.make, v.model").
		Order("total_fuel DESC").
		Limit(10).
		Rows()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get top consumers: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var consumer VehicleFuelStats
		err := rows.Scan(&consumer.VehicleID, &consumer.LicensePlate, &consumer.Make, &consumer.Model, 
			&consumer.TotalFuel, &consumer.TotalCost, &consumer.TotalDistance, &consumer.AverageEfficiency)
		if err != nil {
			continue
		}
		
		if consumer.TotalDistance > 0 {
			consumer.FuelCostPerKm = consumer.TotalCost / consumer.TotalDistance
		}
		
		consumers = append(consumers, consumer)
	}
	
	return consumers, nil
}

// getEfficiencyRanking retrieves vehicle efficiency ranking
func (fm *FuelManager) getEfficiencyRanking(_ context.Context, companyID string, startDate, endDate time.Time) ([]VehicleEfficiency, error) {
	var ranking []VehicleEfficiency
	
	rows, err := fm.db.Table("fuel_records fr").
		Select("fr.vehicle_id, v.license_plate, v.make, v.model, AVG(fr.fuel_efficiency) as efficiency").
		Joins("JOIN vehicles v ON fr.vehicle_id = v.id").
		Where("fr.company_id = ? AND fr.recorded_at BETWEEN ? AND ?", companyID, startDate, endDate).
		Group("fr.vehicle_id, v.license_plate, v.make, v.model").
		Having("COUNT(*) >= 3"). // At least 3 fuel records
		Order("efficiency DESC").
		Rows()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get efficiency ranking: %w", err)
	}
	defer rows.Close()
	
	rank := 1
	for rows.Next() {
		var efficiency VehicleEfficiency
		err := rows.Scan(&efficiency.VehicleID, &efficiency.LicensePlate, &efficiency.Make, &efficiency.Model, &efficiency.Efficiency)
		if err != nil {
			continue
		}
		
		efficiency.Rank = rank
		// Calculate improvement from previous period (simplified)
		efficiency.Improvement = 0.0 // Would need historical data to calculate
		
		ranking = append(ranking, efficiency)
		rank++
	}
	
	return ranking, nil
}

// Helper methods
func (fm *FuelManager) getRouteTypeMultiplier(routeType string) float64 {
	switch routeType {
	case "highway":
		return 1.2 // Better efficiency on highways
	case "city":
		return 0.8 // Lower efficiency in city traffic
	case "mixed":
		return 1.0 // Average efficiency
	default:
		return 1.0
	}
}

func (fm *FuelManager) getCurrentFuelPrice() float64 {
	// This would typically fetch from an external API or database
	// For now, return a fixed price
	return 15000.0 // IDR per liter
}

func (fm *FuelManager) getCO2EmissionFactor(fuelType string) float64 {
	switch fuelType {
	case "gasoline":
		return 2.31 // kg CO2 per liter
	case "diesel":
		return 2.68 // kg CO2 per liter
	case "electric":
		return 0.0 // No direct CO2 emission
	default:
		return 2.31 // Default to gasoline
	}
}

// Cache methods
func (fm *FuelManager) getCachedAnalytics(_ context.Context, _ string) (*FuelAnalytics, error) {
	// Implementation would use Redis to get cached analytics
	return nil, fmt.Errorf("cache miss")
}

func (fm *FuelManager) cacheAnalytics(_ context.Context, _ string, _ *FuelAnalytics, _ time.Duration) error {
	// Implementation would use Redis to cache analytics
	return nil
}

func (fm *FuelManager) invalidateFuelCache(_ context.Context, _, _ string) error {
	// Implementation would invalidate relevant cache entries
	return nil
}
