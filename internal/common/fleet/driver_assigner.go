package fleet

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// DriverAssigner provides intelligent driver assignment capabilities
type DriverAssigner struct {
	db    *gorm.DB
	redis *redis.Client
}

// AssignmentRequest represents a driver assignment request
type AssignmentRequest struct {
	CompanyID       string                 `json:"company_id"`
	VehicleID       string                 `json:"vehicle_id"`
	TaskType        string                 `json:"task_type"` // delivery, pickup, service, maintenance
	Priority        string                 `json:"priority"` // low, medium, high, urgent
	StartLocation   Location               `json:"start_location"`
	EndLocation     *Location              `json:"end_location,omitempty"`
	Stops           []Location             `json:"stops,omitempty"`
	RequiredSkills  []string               `json:"required_skills,omitempty"`
	TimeWindow      TimeWindow             `json:"time_window"`
	Constraints     AssignmentConstraints  `json:"constraints"`
	Preferences     AssignmentPreferences  `json:"preferences"`
}

// Location represents a geographical location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
	Name      string  `json:"name,omitempty"`
}

// AssignmentConstraints represents constraints for driver assignment
type AssignmentConstraints struct {
	MaxDistance     float64 `json:"max_distance"`     // km
	MaxDuration     int     `json:"max_duration"`     // minutes
	RequiredLicense string  `json:"required_license"` // A, B, C, etc.
	MinExperience   int     `json:"min_experience"`   // years
	MaxWorkHours    int     `json:"max_work_hours"`   // hours per day
	ExcludeDrivers  []string `json:"exclude_drivers"` // driver IDs to exclude
}

// AssignmentPreferences represents preferences for driver assignment
type AssignmentPreferences struct {
	PreferExperienced bool    `json:"prefer_experienced"`
	PreferNearby      bool    `json:"prefer_nearby"`
	PreferAvailable   bool    `json:"prefer_available"`
	BalanceWorkload   bool    `json:"balance_workload"`
	MinRating         float64 `json:"min_rating"` // 1-5 scale
}

// DriverAssignment represents the result of driver assignment
type DriverAssignment struct {
	DriverID         string    `json:"driver_id"`
	DriverName       string    `json:"driver_name"`
	VehicleID        string    `json:"vehicle_id"`
	AssignmentScore  float64   `json:"assignment_score"`
	EstimatedArrival time.Time `json:"estimated_arrival"`
	EstimatedDuration int      `json:"estimated_duration"` // minutes
	Distance         float64   `json:"distance"` // km
	Reason           string    `json:"reason"`
	CreatedAt        time.Time `json:"created_at"`
}

// DriverProfile represents a driver's profile for assignment
type DriverProfile struct {
	DriverID         string    `json:"driver_id"`
	DriverName       string    `json:"driver_name"`
	CurrentLocation  Location  `json:"current_location"`
	Status           string    `json:"status"` // available, busy, offline
	LicenseType      string    `json:"license_type"`
	Experience       int       `json:"experience"` // years
	Rating           float64   `json:"rating"` // 1-5 scale
	Skills           []string  `json:"skills"`
	CurrentVehicle   *string   `json:"current_vehicle,omitempty"`
	WorkHoursToday   int       `json:"work_hours_today"` // hours
	LastAssignment   time.Time `json:"last_assignment"`
	DistanceFromTask float64   `json:"distance_from_task"` // km
	AvailabilityScore float64  `json:"availability_score"`
	ExperienceScore  float64   `json:"experience_score"`
	LocationScore    float64   `json:"location_score"`
	WorkloadScore    float64   `json:"workload_score"`
	OverallScore     float64   `json:"overall_score"`
}

// AssignmentAnalytics represents analytics for driver assignments
type AssignmentAnalytics struct {
	Period              string    `json:"period"`
	TotalAssignments    int       `json:"total_assignments"`
	SuccessfulAssignments int     `json:"successful_assignments"`
	AverageResponseTime float64   `json:"average_response_time"` // minutes
	AverageDistance     float64   `json:"average_distance"` // km
	DriverUtilization   []DriverUtilization `json:"driver_utilization"`
	TaskTypeBreakdown   []TaskTypeStats `json:"task_type_breakdown"`
	PerformanceMetrics  []PerformanceMetric `json:"performance_metrics"`
}

// DriverUtilization represents driver utilization statistics
type DriverUtilization struct {
	DriverID       string  `json:"driver_id"`
	DriverName     string  `json:"driver_name"`
	TotalAssignments int   `json:"total_assignments"`
	TotalHours     float64 `json:"total_hours"`
	UtilizationRate float64 `json:"utilization_rate"` // percentage
	AverageRating  float64 `json:"average_rating"`
}

// TaskTypeStats represents statistics by task type
type TaskTypeStats struct {
	TaskType        string  `json:"task_type"`
	Count           int     `json:"count"`
	AverageDuration float64 `json:"average_duration"`
	AverageDistance float64 `json:"average_distance"`
	SuccessRate     float64 `json:"success_rate"`
}

// PerformanceMetric represents performance metrics
type PerformanceMetric struct {
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	Trend     string  `json:"trend"` // up, down, stable
	Target    float64 `json:"target"`
	Achievement float64 `json:"achievement"` // percentage
}

// NewDriverAssigner creates a new driver assigner
func NewDriverAssigner(db *gorm.DB, redis *redis.Client) *DriverAssigner {
	return &DriverAssigner{
		db:    db,
		redis: redis,
	}
}

// AssignDriver assigns the best available driver for a task
func (da *DriverAssigner) AssignDriver(ctx context.Context, req *AssignmentRequest) (*DriverAssignment, error) {
	// Validate assignment request
	if err := da.validateAssignmentRequest(req); err != nil {
		return nil, fmt.Errorf("assignment request validation failed: %w", err)
	}

	// Get available drivers
	drivers, err := da.getAvailableDrivers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get available drivers: %w", err)
	}

	if len(drivers) == 0 {
		return nil, fmt.Errorf("no available drivers found for the assignment")
	}

	// Score and rank drivers
	scoredDrivers, err := da.scoreDrivers(ctx, drivers, req)
	if err != nil {
		return nil, fmt.Errorf("failed to score drivers: %w", err)
	}

	// Select best driver
	bestDriver := scoredDrivers[0]

	// Create assignment
	assignment := &DriverAssignment{
		DriverID:         bestDriver.DriverID,
		DriverName:       bestDriver.DriverName,
		VehicleID:        req.VehicleID,
		AssignmentScore:  bestDriver.OverallScore,
		EstimatedArrival: time.Now().Add(time.Duration(da.calculateArrivalTime(&bestDriver, req)) * time.Minute),
		EstimatedDuration: da.calculateTaskDuration(req),
		Distance:         bestDriver.DistanceFromTask,
		Reason:           da.generateAssignmentReason(&bestDriver, req),
		CreatedAt:        time.Now(),
	}

	// Save assignment to database
	if err := da.saveAssignment(ctx, assignment); err != nil {
		return nil, fmt.Errorf("failed to save assignment: %w", err)
	}

	// Update driver status
	if err := da.updateDriverStatus(ctx, bestDriver.DriverID, "busy"); err != nil {
		return nil, fmt.Errorf("failed to update driver status: %w", err)
	}

	// Invalidate cache
	da.invalidateDriverCache(ctx, req.CompanyID)

	return assignment, nil
}

// GetDriverRecommendations gets driver recommendations for a task
func (da *DriverAssigner) GetDriverRecommendations(ctx context.Context, req *AssignmentRequest, limit int) ([]DriverProfile, error) {
	// Get available drivers
	drivers, err := da.getAvailableDrivers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get available drivers: %w", err)
	}

	// Score and rank drivers
	scoredDrivers, err := da.scoreDrivers(ctx, drivers, req)
	if err != nil {
		return nil, fmt.Errorf("failed to score drivers: %w", err)
	}

	// Limit results
	if limit > 0 && len(scoredDrivers) > limit {
		scoredDrivers = scoredDrivers[:limit]
	}

	return scoredDrivers, nil
}

// GetAssignmentAnalytics retrieves assignment analytics
func (da *DriverAssigner) GetAssignmentAnalytics(ctx context.Context, companyID string, startDate, endDate time.Time) (*AssignmentAnalytics, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("assignment_analytics:%s:%s:%s", companyID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	cached, err := da.getCachedAnalytics(ctx, cacheKey)
	if err == nil && cached != nil {
		return cached, nil
	}

	// Get total assignments
	var totalAssignments, successfulAssignments int
	var averageResponseTime, averageDistance float64

	err = da.db.Model(&DriverAssignment{}).
		Where("company_id = ? AND created_at BETWEEN ? AND ?", companyID, startDate, endDate).
		Select("COUNT(*) as total_assignments").
		Row().Scan(&totalAssignments)
	if err != nil {
		return nil, fmt.Errorf("failed to get total assignments: %w", err)
	}

	err = da.db.Model(&DriverAssignment{}).
		Where("company_id = ? AND created_at BETWEEN ? AND ? AND status = 'completed'", companyID, startDate, endDate).
		Select("COUNT(*) as successful_assignments, AVG(response_time) as average_response_time, AVG(distance) as average_distance").
		Row().Scan(&successfulAssignments, &averageResponseTime, &averageDistance)
	if err != nil {
		return nil, fmt.Errorf("failed to get successful assignments: %w", err)
	}

	// Get driver utilization
	driverUtilization, err := da.getDriverUtilization(ctx, companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver utilization: %w", err)
	}

	// Get task type breakdown
	taskTypeBreakdown, err := da.getTaskTypeBreakdown(ctx, companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get task type breakdown: %w", err)
	}

	// Get performance metrics
	performanceMetrics, err := da.getPerformanceMetrics(ctx, companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance metrics: %w", err)
	}

	analytics := &AssignmentAnalytics{
		Period:              fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		TotalAssignments:    totalAssignments,
		SuccessfulAssignments: successfulAssignments,
		AverageResponseTime: averageResponseTime,
		AverageDistance:     averageDistance,
		DriverUtilization:   driverUtilization,
		TaskTypeBreakdown:   taskTypeBreakdown,
		PerformanceMetrics:  performanceMetrics,
	}

	// Cache the result
	da.cacheAnalytics(ctx, cacheKey, analytics, 1*time.Hour)

	return analytics, nil
}

// validateAssignmentRequest validates an assignment request
func (da *DriverAssigner) validateAssignmentRequest(req *AssignmentRequest) error {
	if req.CompanyID == "" {
		return fmt.Errorf("company ID is required")
	}
	if req.VehicleID == "" {
		return fmt.Errorf("vehicle ID is required")
	}
	if req.TaskType == "" {
		return fmt.Errorf("task type is required")
	}
	if req.Priority == "" {
		return fmt.Errorf("priority is required")
	}
	if req.StartLocation.Latitude == 0 || req.StartLocation.Longitude == 0 {
		return fmt.Errorf("start location is required")
	}
	return nil
}

// getAvailableDrivers retrieves available drivers for assignment
func (da *DriverAssigner) getAvailableDrivers(ctx context.Context, req *AssignmentRequest) ([]DriverProfile, error) {
	var drivers []DriverProfile

	// Build query for available drivers
	query := da.db.Table("drivers d").
		Select("d.id, d.first_name, d.last_name, d.status, d.license_type, d.experience, d.rating, d.current_location_lat, d.current_location_lng").
		Where("d.company_id = ? AND d.status = 'available'", req.CompanyID)

	// Apply constraints
	if req.Constraints.RequiredLicense != "" {
		query = query.Where("d.license_type = ?", req.Constraints.RequiredLicense)
	}
	if req.Constraints.MinExperience > 0 {
		query = query.Where("d.experience >= ?", req.Constraints.MinExperience)
	}
	if len(req.Constraints.ExcludeDrivers) > 0 {
		query = query.Where("d.id NOT IN ?", req.Constraints.ExcludeDrivers)
	}
	if req.Preferences.MinRating > 0 {
		query = query.Where("d.rating >= ?", req.Preferences.MinRating)
	}

	// Execute query
	rows, err := query.Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query available drivers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var driver DriverProfile
		var firstName, lastName string
		var lat, lng float64

		err := rows.Scan(&driver.DriverID, &firstName, &lastName, &driver.Status, 
			&driver.LicenseType, &driver.Experience, &driver.Rating, &lat, &lng)
		if err != nil {
			continue
		}

		driver.DriverName = firstName + " " + lastName
		driver.CurrentLocation = Location{
			Latitude:  lat,
			Longitude: lng,
		}

		// Calculate distance from task
		driver.DistanceFromTask = da.calculateDistance(
			driver.CurrentLocation.Latitude, driver.CurrentLocation.Longitude,
			req.StartLocation.Latitude, req.StartLocation.Longitude,
		)

		// Get additional driver information
		da.enrichDriverProfile(ctx, &driver)

		drivers = append(drivers, driver)
	}

	return drivers, nil
}

// enrichDriverProfile enriches driver profile with additional information
func (da *DriverAssigner) enrichDriverProfile(_ context.Context, driver *DriverProfile) {
	// Get driver skills
	var skills []string
	da.db.Table("driver_skills ds").
		Select("s.name").
		Joins("JOIN skills s ON ds.skill_id = s.id").
		Where("ds.driver_id = ?", driver.DriverID).
		Pluck("s.name", &skills)
	driver.Skills = skills

	// Get current vehicle
	var currentVehicle string
	da.db.Model(&models.Driver{}).
		Where("id = ?", driver.DriverID).
		Select("vehicle_id").
		Scan(&currentVehicle)
	if currentVehicle != "" {
		driver.CurrentVehicle = &currentVehicle
	}

	// Get work hours today
	var workHours int
	da.db.Table("driver_assignments da").
		Select("COALESCE(SUM(estimated_duration), 0) / 60").
		Where("da.driver_id = ? AND DATE(da.created_at) = CURRENT_DATE", driver.DriverID).
		Scan(&workHours)
	driver.WorkHoursToday = workHours

	// Get last assignment
	var lastAssignment time.Time
	da.db.Model(&DriverAssignment{}).
		Where("driver_id = ?", driver.DriverID).
		Select("COALESCE(MAX(created_at), '1900-01-01')").
		Scan(&lastAssignment)
	driver.LastAssignment = lastAssignment
}

// scoreDrivers scores and ranks drivers based on assignment criteria
func (da *DriverAssigner) scoreDrivers(_ context.Context, drivers []DriverProfile, req *AssignmentRequest) ([]DriverProfile, error) {
	for i := range drivers {
		// Calculate availability score
		drivers[i].AvailabilityScore = da.calculateAvailabilityScore(&drivers[i], req)

		// Calculate experience score
		drivers[i].ExperienceScore = da.calculateExperienceScore(&drivers[i], req)

		// Calculate location score
		drivers[i].LocationScore = da.calculateLocationScore(&drivers[i], req)

		// Calculate workload score
		drivers[i].WorkloadScore = da.calculateWorkloadScore(&drivers[i], req)

		// Calculate overall score
		drivers[i].OverallScore = da.calculateOverallScore(&drivers[i], req)
	}

	// Sort by overall score (descending)
	sort.Slice(drivers, func(i, j int) bool {
		return drivers[i].OverallScore > drivers[j].OverallScore
	})

	return drivers, nil
}

// calculateAvailabilityScore calculates availability score for a driver
func (da *DriverAssigner) calculateAvailabilityScore(driver *DriverProfile, req *AssignmentRequest) float64 {
	score := 0.0

	// Base availability score
	if driver.Status == "available" {
		score += 50.0
	} else {
		return 0.0 // Not available
	}

	// Work hours constraint
	if req.Constraints.MaxWorkHours > 0 {
		remainingHours := req.Constraints.MaxWorkHours - driver.WorkHoursToday
		if remainingHours > 0 {
			score += float64(remainingHours) * 5.0 // 5 points per remaining hour
		} else {
			score = 0.0 // Exceeded work hours
		}
	}

	// Time since last assignment
	timeSinceLast := time.Since(driver.LastAssignment)
	if timeSinceLast > 2*time.Hour {
		score += 20.0 // Bonus for drivers who haven't been assigned recently
	}

	return math.Min(score, 100.0)
}

// calculateExperienceScore calculates experience score for a driver
func (da *DriverAssigner) calculateExperienceScore(driver *DriverProfile, req *AssignmentRequest) float64 {
	score := 0.0

	// Base experience score
	score += float64(driver.Experience) * 2.0 // 2 points per year of experience

	// License type bonus
	switch driver.LicenseType {
	case "A":
		score += 10.0
	case "B":
		score += 15.0
	case "C":
		score += 20.0
	}

	// Rating bonus
	score += driver.Rating * 10.0 // 10 points per rating point

	// Required skills match
	if len(req.RequiredSkills) > 0 {
		skillMatches := 0
		for _, requiredSkill := range req.RequiredSkills {
			for _, driverSkill := range driver.Skills {
				if requiredSkill == driverSkill {
					skillMatches++
					break
				}
			}
		}
		skillScore := float64(skillMatches) / float64(len(req.RequiredSkills)) * 30.0
		score += skillScore
	}

	return math.Min(score, 100.0)
}

// calculateLocationScore calculates location score for a driver
func (da *DriverAssigner) calculateLocationScore(driver *DriverProfile, req *AssignmentRequest) float64 {
	score := 0.0

	// Distance-based score (closer is better)
	maxDistance := 50.0 // km
	if req.Constraints.MaxDistance > 0 {
		maxDistance = req.Constraints.MaxDistance
	}

	if driver.DistanceFromTask <= maxDistance {
		// Score decreases with distance
		score = (maxDistance - driver.DistanceFromTask) / maxDistance * 100.0
	} else {
		score = 0.0 // Too far away
	}

	// Nearby preference bonus
	if req.Preferences.PreferNearby && driver.DistanceFromTask < 10.0 {
		score += 20.0
	}

	return math.Min(score, 100.0)
}

// calculateWorkloadScore calculates workload score for a driver
func (da *DriverAssigner) calculateWorkloadScore(driver *DriverProfile, req *AssignmentRequest) float64 {
	score := 100.0 // Start with full score

	// Reduce score based on work hours today
	workHoursPenalty := float64(driver.WorkHoursToday) * 5.0
	score -= workHoursPenalty

	// Balance workload preference
	if req.Preferences.BalanceWorkload {
		// Get average work hours for all drivers
		var avgWorkHours float64
		da.db.Table("drivers d").
			Select("AVG(COALESCE(da.work_hours_today, 0))").
			Joins("LEFT JOIN driver_assignments da ON d.id = da.driver_id").
			Where("d.company_id = ?", req.CompanyID).
			Scan(&avgWorkHours)

		if driver.WorkHoursToday < int(avgWorkHours) {
			score += 20.0 // Bonus for underutilized drivers
		}
	}

	return math.Max(score, 0.0)
}

// calculateOverallScore calculates overall score for a driver
func (da *DriverAssigner) calculateOverallScore(driver *DriverProfile, _ *AssignmentRequest) float64 {
	// Weighted combination of all scores
	weights := map[string]float64{
		"availability": 0.3,
		"experience":   0.25,
		"location":     0.25,
		"workload":     0.2,
	}

	overallScore := driver.AvailabilityScore*weights["availability"] +
		driver.ExperienceScore*weights["experience"] +
		driver.LocationScore*weights["location"] +
		driver.WorkloadScore*weights["workload"]

	return overallScore
}

// calculateArrivalTime calculates estimated arrival time for a driver
func (da *DriverAssigner) calculateArrivalTime(driver *DriverProfile, _ *AssignmentRequest) int {
	// Base time to reach start location
	distanceTime := int(driver.DistanceFromTask / 40.0 * 60) // Assume 40 km/h average speed

	// Add buffer time
	bufferTime := 15 // 15 minutes buffer

	return distanceTime + bufferTime
}

// calculateTaskDuration calculates estimated task duration
func (da *DriverAssigner) calculateTaskDuration(req *AssignmentRequest) int {
	baseDuration := 60 // 1 hour base duration

	// Adjust based on task type
	switch req.TaskType {
	case "delivery":
		baseDuration = 45
	case "pickup":
		baseDuration = 30
	case "service":
		baseDuration = 120
	case "maintenance":
		baseDuration = 180
	}

	// Adjust based on priority
	switch req.Priority {
	case "urgent":
		baseDuration = int(float64(baseDuration) * 0.8) // 20% faster
	case "high":
		baseDuration = int(float64(baseDuration) * 0.9) // 10% faster
	case "low":
		baseDuration = int(float64(baseDuration) * 1.2) // 20% slower
	}

	// Add time for multiple stops
	if len(req.Stops) > 0 {
		baseDuration += len(req.Stops) * 15 // 15 minutes per stop
	}

	return baseDuration
}

// generateAssignmentReason generates a reason for the assignment
func (da *DriverAssigner) generateAssignmentReason(driver *DriverProfile, _ *AssignmentRequest) string {
	reasons := []string{}

	if driver.DistanceFromTask < 5.0 {
		reasons = append(reasons, "closest available driver")
	}
	if driver.Rating >= 4.5 {
		reasons = append(reasons, "high-rated driver")
	}
	if driver.Experience >= 5 {
		reasons = append(reasons, "experienced driver")
	}
	if driver.WorkHoursToday < 6 {
		reasons = append(reasons, "available for full shift")
	}

	if len(reasons) == 0 {
		return "best available match"
	}

	reason := "Selected because: " + reasons[0]
	for i := 1; i < len(reasons); i++ {
		reason += ", " + reasons[i]
	}

	return reason
}

// saveAssignment saves driver assignment to database
func (da *DriverAssigner) saveAssignment(_ context.Context, _ *DriverAssignment) error {
	// This would save the assignment to the database
	// For now, just return nil
	return nil
}

// updateDriverStatus updates driver status
func (da *DriverAssigner) updateDriverStatus(_ context.Context, driverID, status string) error {
	return da.db.Model(&models.Driver{}).Where("id = ?", driverID).Update("status", status).Error
}

// Helper methods for analytics
func (da *DriverAssigner) getDriverUtilization(_ context.Context, companyID string, startDate, endDate time.Time) ([]DriverUtilization, error) {
	var utilization []DriverUtilization
	
	rows, err := da.db.Table("driver_assignments da").
		Select("da.driver_id, d.first_name, d.last_name, COUNT(*) as total_assignments, SUM(da.estimated_duration) / 60.0 as total_hours, AVG(da.assignment_score) as average_rating").
		Joins("JOIN drivers d ON da.driver_id = d.id").
		Where("da.company_id = ? AND da.created_at BETWEEN ? AND ?", companyID, startDate, endDate).
		Group("da.driver_id, d.first_name, d.last_name").
		Order("total_assignments DESC").
		Rows()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get driver utilization: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var util DriverUtilization
		var firstName, lastName string
		
		err := rows.Scan(&util.DriverID, &firstName, &lastName, &util.TotalAssignments, &util.TotalHours, &util.AverageRating)
		if err != nil {
			continue
		}
		
		util.DriverName = firstName + " " + lastName
		util.UtilizationRate = (util.TotalHours / (8.0 * float64(len(utilization)+1))) * 100 // Assuming 8-hour work day
		
		utilization = append(utilization, util)
	}
	
	return utilization, nil
}

func (da *DriverAssigner) getTaskTypeBreakdown(_ context.Context, companyID string, startDate, endDate time.Time) ([]TaskTypeStats, error) {
	var breakdown []TaskTypeStats
	
	rows, err := da.db.Model(&DriverAssignment{}).
		Select("task_type, COUNT(*) as count, AVG(estimated_duration) as average_duration, AVG(distance) as average_distance").
		Where("company_id = ? AND created_at BETWEEN ? AND ?", companyID, startDate, endDate).
		Group("task_type").
		Order("count DESC").
		Rows()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get task type breakdown: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var stats TaskTypeStats
		err := rows.Scan(&stats.TaskType, &stats.Count, &stats.AverageDuration, &stats.AverageDistance)
		if err != nil {
			continue
		}
		
		// Calculate success rate (simplified)
		stats.SuccessRate = 95.0 // Would need actual completion data
		
		breakdown = append(breakdown, stats)
	}
	
	return breakdown, nil
}

func (da *DriverAssigner) getPerformanceMetrics(_ context.Context, companyID string, startDate, endDate time.Time) ([]PerformanceMetric, error) {
	var metrics []PerformanceMetric
	
	// Average response time
	var avgResponseTime float64
	da.db.Model(&DriverAssignment{}).
		Where("company_id = ? AND created_at BETWEEN ? AND ?", companyID, startDate, endDate).
		Select("AVG(response_time)").
		Scan(&avgResponseTime)
	
	metrics = append(metrics, PerformanceMetric{
		Metric:      "Average Response Time",
		Value:       avgResponseTime,
		Unit:        "minutes",
		Trend:       "stable",
		Target:      30.0,
		Achievement: (30.0 / avgResponseTime) * 100,
	})
	
	// Assignment success rate
	var successRate float64
	da.db.Model(&DriverAssignment{}).
		Where("company_id = ? AND created_at BETWEEN ? AND ?", companyID, startDate, endDate).
		Select("(COUNT(CASE WHEN status = 'completed' THEN 1 END) * 100.0 / COUNT(*))").
		Scan(&successRate)
	
	metrics = append(metrics, PerformanceMetric{
		Metric:      "Assignment Success Rate",
		Value:       successRate,
		Unit:        "percentage",
		Trend:       "stable",
		Target:      95.0,
		Achievement: successRate,
	})
	
	return metrics, nil
}

// Utility methods
func (da *DriverAssigner) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// Cache methods
func (da *DriverAssigner) getCachedAnalytics(_ context.Context, _ string) (*AssignmentAnalytics, error) {
	// Implementation would use Redis to get cached analytics
	return nil, fmt.Errorf("cache miss")
}

func (da *DriverAssigner) cacheAnalytics(_ context.Context, _ string, _ *AssignmentAnalytics, _ time.Duration) error {
	// Implementation would use Redis to cache analytics
	return nil
}

func (da *DriverAssigner) invalidateDriverCache(_ context.Context, _ string) error {
	// Implementation would invalidate relevant cache entries
	return nil
}
