package validators

import (
	"fmt"
	"strings"
	"time"
)

// Business rule validation errors
var (
	ErrInvalidVehicleCapacity = fmt.Errorf("vehicle capacity must be between 1 and 100")
	ErrInvalidFuelAmount      = fmt.Errorf("fuel amount must be positive and less than 1000 liters")
	ErrInvalidFuelCost        = fmt.Errorf("fuel cost must be positive")
	ErrInvalidSpeed           = fmt.Errorf("speed must be between 0 and 300 km/h")
	ErrInvalidDistance        = fmt.Errorf("distance must be positive")
	ErrInvalidOdometer        = fmt.Errorf("odometer reading must be positive")
	ErrInvalidYear            = fmt.Errorf("year must be between 1900 and current year + 1")
	ErrInvalidAmount          = fmt.Errorf("amount must be positive")
	ErrInvalidPercentage      = fmt.Errorf("percentage must be between 0 and 100")
)

// ValidateVehicleCapacity validates vehicle passenger capacity
func ValidateVehicleCapacity(capacity int) error {
	if capacity < 1 || capacity > 100 {
		return ErrInvalidVehicleCapacity
	}
	return nil
}

// ValidateFuelAmount validates fuel amount in liters
func ValidateFuelAmount(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("fuel amount must be positive")
	}
	if amount > 1000 {
		return ErrInvalidFuelAmount
	}
	return nil
}

// ValidateFuelCost validates fuel cost in Indonesian Rupiah
func ValidateFuelCost(cost float64) error {
	if cost <= 0 {
		return ErrInvalidFuelCost
	}
	if cost > 100000000 { // 100 million IDR (sanity check)
		return fmt.Errorf("fuel cost exceeds reasonable limit")
	}
	return nil
}

// ValidateFuelEfficiency validates fuel efficiency (km/liter)
func ValidateFuelEfficiency(efficiency float64) error {
	if efficiency < 0 {
		return fmt.Errorf("fuel efficiency cannot be negative")
	}
	if efficiency > 50 { // Sanity check (very high efficiency)
		return fmt.Errorf("fuel efficiency exceeds reasonable limit (50 km/L)")
	}
	return nil
}

// ValidateSpeed validates speed in km/h
func ValidateSpeed(speed float64) error {
	if speed < 0 {
		return fmt.Errorf("speed cannot be negative")
	}
	if speed > 300 {
		return ErrInvalidSpeed
	}
	return nil
}

// ValidateSpeedLimit validates speed limit based on Indonesian regulations
func ValidateSpeedLimit(speedLimit float64, roadType string) error {
	limits := map[string]float64{
		"highway":      100, // Jalan tol
		"national":     80,  // Jalan nasional
		"urban":        50,  // Jalan dalam kota
		"residential":  30,  // Jalan perumahan
		"school_zone":  20,  // Zona sekolah
	}

	maxLimit, exists := limits[roadType]
	if !exists {
		maxLimit = 100 // Default to highway
	}

	if speedLimit < 0 || speedLimit > maxLimit {
		return fmt.Errorf("speed limit for %s must be between 0 and %.0f km/h", roadType, maxLimit)
	}

	return nil
}

// ValidateDistance validates distance in kilometers
func ValidateDistance(distance float64) error {
	if distance < 0 {
		return ErrInvalidDistance
	}
	if distance > 10000 { // Sanity check (10,000 km is across Indonesia multiple times)
		return fmt.Errorf("distance exceeds reasonable limit")
	}
	return nil
}

// ValidateOdometerReading validates odometer reading
func ValidateOdometerReading(reading float64) error {
	if reading < 0 {
		return ErrInvalidOdometer
	}
	if reading > 10000000 { // 10 million km (sanity check)
		return fmt.Errorf("odometer reading exceeds reasonable limit")
	}
	return nil
}

// ValidateOdometerIncrement validates odometer increment
func ValidateOdometerIncrement(oldReading, newReading float64) error {
	if newReading < oldReading {
		return fmt.Errorf("new odometer reading cannot be less than previous reading")
	}

	increment := newReading - oldReading
	if increment > 1000 { // More than 1000 km in one update
		return fmt.Errorf("odometer increment too large (%.2f km) - possible error", increment)
	}

	return nil
}

// ValidateVehicleYear validates vehicle manufacturing year
func ValidateVehicleYear(year int) error {
	currentYear := time.Now().Year()
	if year < 1900 || year > currentYear+1 {
		return ErrInvalidYear
	}
	return nil
}

// ValidateAmount validates monetary amount in IDR
func ValidateAmount(amount float64) error {
	if amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}
	if amount > 10000000000 { // 10 billion IDR (sanity check)
		return fmt.Errorf("amount exceeds reasonable limit")
	}
	return nil
}

// ValidatePaymentAmount validates payment amount
func ValidatePaymentAmount(amount float64, minAmount float64) error {
	if amount < minAmount {
		return fmt.Errorf("payment amount must be at least %.2f", minAmount)
	}
	return ValidateAmount(amount)
}

// ValidatePercentage validates percentage (0-100)
func ValidatePercentage(percentage float64) error {
	if percentage < 0 || percentage > 100 {
		return ErrInvalidPercentage
	}
	return nil
}

// ValidateScore validates score (0-100)
func ValidateScore(score float64) error {
	return ValidatePercentage(score)
}

// ValidateDuration validates duration
func ValidateDuration(duration time.Duration, minDuration, maxDuration time.Duration) error {
	if duration < minDuration {
		return fmt.Errorf("duration must be at least %s", minDuration)
	}
	if duration > maxDuration {
		return fmt.Errorf("duration cannot exceed %s", maxDuration)
	}
	return nil
}

// ValidateTripDuration validates trip duration (prevent unrealistic trips)
func ValidateTripDuration(duration time.Duration) error {
	if duration < 0 {
		return fmt.Errorf("trip duration cannot be negative")
	}
	if duration > 24*time.Hour {
		return fmt.Errorf("trip duration exceeds 24 hours - possible error")
	}
	return nil
}

// ValidateDateRange validates date range
func ValidateDateRange(startDate, endDate time.Time) error {
	if endDate.Before(startDate) {
		return fmt.Errorf("end date must be after start date")
	}

	duration := endDate.Sub(startDate)
	if duration > 365*24*time.Hour {
		return fmt.Errorf("date range cannot exceed 1 year")
	}

	return nil
}

// ValidateFutureDate validates date is in the future
func ValidateFutureDate(date time.Time, fieldName string) error {
	if date.Before(time.Now()) {
		return fmt.Errorf("%s must be in the future", fieldName)
	}
	return nil
}

// ValidatePastDate validates date is in the past
func ValidatePastDate(date time.Time, fieldName string) error {
	if date.After(time.Now()) {
		return fmt.Errorf("%s cannot be in the future", fieldName)
	}
	return nil
}

// ValidateCoordinates validates GPS coordinates for Indonesian region
func ValidateCoordinates(latitude, longitude float64) error {
	// Basic coordinate validation
	if latitude < -90 || latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if longitude < -180 || longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	// Check if exactly 0,0 (likely invalid/default value)
	if latitude == 0 && longitude == 0 {
		return fmt.Errorf("invalid coordinates: (0, 0)")
	}

	return nil
}

// ValidateAccuracy validates GPS accuracy in meters
func ValidateAccuracy(accuracy float64) error {
	if accuracy < 0 {
		return fmt.Errorf("accuracy cannot be negative")
	}
	if accuracy > 1000 {
		return fmt.Errorf("accuracy too low (> 1000m) - GPS signal poor")
	}
	return nil
}

// ValidateHeading validates compass heading (0-360 degrees)
func ValidateHeading(heading float64) error {
	if heading < 0 || heading > 360 {
		return fmt.Errorf("heading must be between 0 and 360 degrees")
	}
	return nil
}

// ValidateMaintenanceCost validates maintenance cost
func ValidateMaintenanceCost(cost float64) error {
	if cost < 0 {
		return fmt.Errorf("maintenance cost cannot be negative")
	}
	if cost > 100000000 { // 100 million IDR
		return fmt.Errorf("maintenance cost exceeds reasonable limit")
	}
	return nil
}

// ValidatePageLimit validates pagination limit
func ValidatePageLimit(limit int) error {
	if limit < 1 {
		return fmt.Errorf("limit must be at least 1")
	}
	if limit > 1000 {
		return fmt.Errorf("limit cannot exceed 1000")
	}
	return nil
}

// ValidatePageOffset validates pagination offset
func ValidatePageOffset(offset int) error {
	if offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	return nil
}

// ValidateTimeRange validates time range for queries
func ValidateTimeRange(start, end time.Time, maxRange time.Duration) error {
	if end.Before(start) {
		return fmt.Errorf("end time must be after start time")
	}

	duration := end.Sub(start)
	if duration > maxRange {
		return fmt.Errorf("time range exceeds maximum allowed (%s)", maxRange)
	}

	return nil
}

// ValidateDriverAge validates driver age (must be 17+ in Indonesia)
func ValidateDriverAge(birthDate time.Time) error {
	age := time.Now().Year() - birthDate.Year()

	// Adjust if birthday hasn't occurred this year
	if time.Now().YearDay() < birthDate.YearDay() {
		age--
	}

	if age < 17 {
		return fmt.Errorf("driver must be at least 17 years old")
	}

	if age > 100 {
		return fmt.Errorf("invalid birth date: age exceeds 100 years")
	}

	return nil
}

// ValidateMaintenanceInterval validates maintenance interval in kilometers
func ValidateMaintenanceInterval(interval float64) error {
	if interval < 1000 {
		return fmt.Errorf("maintenance interval must be at least 1000 km")
	}
	if interval > 50000 {
		return fmt.Errorf("maintenance interval cannot exceed 50,000 km")
	}
	return nil
}

// ValidateGeofenceRadius validates geofence radius in meters
func ValidateGeofenceRadius(radius float64) error {
	if radius < 10 {
		return fmt.Errorf("geofence radius must be at least 10 meters")
	}
	if radius > 100000 {
		return fmt.Errorf("geofence radius cannot exceed 100 km")
	}
	return nil
}

// ValidateDriverRating validates driver rating (1-5 stars)
func ValidateDriverRating(rating float64) error {
	if rating < 1 || rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}
	return nil
}

// ValidatePriority validates priority value
func ValidatePriority(priority int, minPriority, maxPriority int) error {
	if priority < minPriority || priority > maxPriority {
		return fmt.Errorf("priority must be between %d and %d", minPriority, maxPriority)
	}
	return nil
}

// ValidateRetryCount validates retry count
func ValidateRetryCount(retryCount, maxRetries int) error {
	if retryCount < 0 {
		return fmt.Errorf("retry count cannot be negative")
	}
	if retryCount > maxRetries {
		return fmt.Errorf("retry count cannot exceed max retries (%d)", maxRetries)
	}
	return nil
}

// ValidateSubscriptionTier validates subscription tier
func ValidateSubscriptionTier(tier string) error {
	validTiers := map[string]bool{
		"basic":      true,
		"professional": true,
		"enterprise": true,
		"trial":      true,
	}

	tier = strings.ToLower(strings.TrimSpace(tier))
	if !validTiers[tier] {
		return fmt.Errorf("invalid subscription tier: must be basic, professional, or enterprise")
	}

	return nil
}

// ValidateRole validates user role
func ValidateRole(role string) error {
	validRoles := map[string]bool{
		"admin":    true,
		"manager":  true,
		"operator": true,
		"viewer":   true,
	}

	role = strings.ToLower(strings.TrimSpace(role))
	if !validRoles[role] {
		return fmt.Errorf("invalid role: must be admin, manager, operator, or viewer")
	}

	return nil
}

// ValidateStatus validates generic status field
func ValidateStatus(status string, validStatuses []string) error {
	status = strings.ToLower(strings.TrimSpace(status))

	for _, validStatus := range validStatuses {
		if status == strings.ToLower(validStatus) {
			return nil
		}
	}

	return fmt.Errorf("invalid status: must be one of %v", validStatuses)
}

// ValidateVehicleStatus validates vehicle status
func ValidateVehicleStatus(status string) error {
	return ValidateStatus(status, []string{"available", "in_use", "maintenance", "inactive"})
}

// ValidateDriverStatus validates driver status
func ValidateDriverStatus(status string) error {
	return ValidateStatus(status, []string{"available", "on_trip", "off_duty", "inactive"})
}

// ValidateTripStatus validates trip status
func ValidateTripStatus(status string) error {
	return ValidateStatus(status, []string{"planned", "active", "in_progress", "completed", "cancelled"})
}

// ValidatePaymentStatus validates payment status
func ValidatePaymentStatus(status string) error {
	return ValidateStatus(status, []string{"pending", "processing", "completed", "failed", "refunded"})
}

// ValidateInvoiceStatus validates invoice status
func ValidateInvoiceStatus(status string) error {
	return ValidateStatus(status, []string{"draft", "sent", "unpaid", "partial", "paid", "overdue", "cancelled"})
}

// ValidateMaintenanceType validates maintenance type
func ValidateMaintenanceType(maintenanceType string) error {
	validTypes := []string{
		"regular_service", "oil_change", "tire_rotation", "brake_check",
		"engine_repair", "transmission_repair", "electrical", "bodywork",
		"inspection", "emergency_repair", "preventive",
	}
	return ValidateStatus(maintenanceType, validTypes)
}

// ValidateEventSeverity validates event severity
func ValidateEventSeverity(severity string) error {
	return ValidateStatus(severity, []string{"low", "medium", "high", "critical"})
}

// ValidateLanguage validates language code
func ValidateLanguage(language string) error {
	validLanguages := map[string]bool{
		"id": true, // Indonesian
		"en": true, // English
	}

	language = strings.ToLower(strings.TrimSpace(language))
	if !validLanguages[language] {
		return fmt.Errorf("invalid language: must be 'id' or 'en'")
	}

	return nil
}

// ValidateTimezone validates timezone
func ValidateTimezone(timezone string) error {
	// Load timezone to verify it's valid
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone: %s", timezone)
	}

	// Recommend Indonesian timezones
	indonesianTimezones := map[string]bool{
		"Asia/Jakarta":   true, // WIB (UTC+7)
		"Asia/Makassar":  true, // WITA (UTC+8)
		"Asia/Jayapura":  true, // WIT (UTC+9)
		"Asia/Pontianak": true, // WIB
	}

	if !indonesianTimezones[timezone] {
		// Still valid, but might want to log a warning
	}

	return nil
}

// ValidateCurrency validates currency code
func ValidateCurrency(currency string) error {
	currency = strings.ToUpper(strings.TrimSpace(currency))

	if currency != "IDR" && currency != "USD" && currency != "EUR" {
		return fmt.Errorf("currency must be IDR, USD, or EUR")
	}

	return nil
}

// ValidatePaymentMethod validates payment method
func ValidatePaymentMethod(method string) error {
	validMethods := []string{
		"bank_transfer", "credit_card", "debit_card", "e_wallet",
		"gopay", "ovo", "dana", "linkaja", "qris",
		"cash", "virtual_account",
	}
	return ValidateStatus(method, validMethods)
}

// ValidateFileSize validates file size in bytes
func ValidateFileSize(size int64, maxSize int64) error {
	if size <= 0 {
		return fmt.Errorf("file size must be positive")
	}
	if size > maxSize {
		return fmt.Errorf("file size (%d bytes) exceeds limit (%d bytes)", size, maxSize)
	}
	return nil
}

// ValidateFileType validates file MIME type
func ValidateFileType(mimeType string, allowedTypes []string) error {
	for _, allowed := range allowedTypes {
		if mimeType == allowed {
			return nil
		}
	}
	return fmt.Errorf("file type %s not allowed (allowed: %v)", mimeType, allowedTypes)
}

// ValidateImageFile validates image file
func ValidateImageFile(mimeType string) error {
	allowedTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}
	return ValidateFileType(mimeType, allowedTypes)
}

// ValidateDocumentFile validates document file
func ValidateDocumentFile(mimeType string) error {
	allowedTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/plain",
	}
	return ValidateFileType(mimeType, allowedTypes)
}

// ValidateExportFormat validates export format
func ValidateExportFormat(format string) error {
	return ValidateStatus(format, []string{"json", "csv", "pdf", "xlsx"})
}

// ValidateReportType validates report type
func ValidateReportType(reportType string) error {
	return ValidateStatus(reportType, []string{
		"fleet_summary", "driver_performance", "fuel_consumption",
		"maintenance_report", "compliance_report", "financial_report",
	})
}

// ValidateAnalyticsPeriod validates analytics period
func ValidateAnalyticsPeriod(period string) error {
	return ValidateStatus(period, []string{"daily", "weekly", "monthly", "quarterly", "yearly", "custom"})
}

// ValidateNotificationType validates notification type
func ValidateNotificationType(notificationType string) error {
	return ValidateStatus(notificationType, []string{
		"maintenance_due", "insurance_expiring", "sim_expiring",
		"geofence_violation", "speed_violation", "low_fuel",
		"trip_completed", "invoice_due", "payment_received",
	})
}

// ValidateGeofenceType validates geofence type
func ValidateGeofenceType(geofenceType string) error {
	return ValidateStatus(geofenceType, []string{"circular", "polygon", "route"})
}

// ValidateAlertType validates alert type
func ValidateAlertType(alertType string) error {
	return ValidateStatus(alertType, []string{
		"info", "warning", "danger", "critical", "success",
	})
}


