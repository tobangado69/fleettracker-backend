package validators

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
)

// ValidationMiddleware provides request validation middleware
type ValidationMiddleware struct {
	sanitizer *Sanitizer
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{
		sanitizer: NewSanitizer(),
	}
}

// SanitizeQueryParams sanitizes all query parameters
func (vm *ValidationMiddleware) SanitizeQueryParams() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get all query parameters
		queryParams := c.Request.URL.Query()

		// Sanitize each parameter
		for key, values := range queryParams {
			for i, value := range values {
				queryParams[key][i] = vm.sanitizer.SanitizeInput(value)
			}
		}

		// Update request
		c.Request.URL.RawQuery = queryParams.Encode()

		c.Next()
	}
}

// ValidateRequestSize limits request body size
func ValidateRequestSize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// ValidateContentType validates Content-Type header
func ValidateContentType(allowedTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip validation for GET requests
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		
		for _, allowed := range allowedTypes {
			if strings.Contains(contentType, allowed) {
				c.Next()
				return
			}
		}

		middleware.AbortWithBadRequest(c, fmt.Sprintf("Invalid Content-Type: must be one of %v", allowedTypes))
	}
}

// ValidateVehicleRequest validates vehicle creation/update request
func ValidateVehicleRequest(c *gin.Context) {
	var req struct {
		LicensePlate string `json:"license_plate"`
		VIN          string `json:"vin"`
		Make         string `json:"make"`
		Model        string `json:"model"`
		Year         int    `json:"year"`
		FuelType     string `json:"fuel_type"`
		Capacity     int    `json:"capacity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Validate license plate
	if req.LicensePlate != "" {
		if err := ValidatePlateNumber(req.LicensePlate); err != nil {
			middleware.AbortWithBadRequest(c, "Invalid license plate: "+err.Error())
			return
		}
		// Normalize format
		req.LicensePlate = FormatPlateNumber(req.LicensePlate)
	}

	// Validate VIN
	if req.VIN != "" {
		if err := ValidateVIN(req.VIN); err != nil {
			middleware.AbortWithBadRequest(c, "Invalid VIN: "+err.Error())
			return
		}
	}

	// Validate year
	if req.Year != 0 {
		if err := ValidateVehicleYear(req.Year); err != nil {
			middleware.AbortWithBadRequest(c, err.Error())
			return
		}
	}

	// Validate fuel type
	if req.FuelType != "" {
		if err := ValidateFuelType(req.FuelType); err != nil {
			middleware.AbortWithBadRequest(c, err.Error())
			return
		}
	}

	// Validate capacity
	if req.Capacity != 0 {
		if err := ValidateVehicleCapacity(req.Capacity); err != nil {
			middleware.AbortWithBadRequest(c, err.Error())
			return
		}
	}

	// Store validated request back to context
	c.Set("validated_request", req)
	c.Next()
}

// ValidateDriverRequest validates driver creation/update request
func ValidateDriverRequest(c *gin.Context) {
	var req struct {
		Name      string    `json:"name"`
		NIK       string    `json:"nik"`
		SIMNumber string    `json:"sim_number"`
		SIMType   string    `json:"sim_type"`
		Phone     string    `json:"phone"`
		Email     string    `json:"email"`
		BirthDate time.Time `json:"birth_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Validate NIK
	if req.NIK != "" {
		nik := NormalizeNIK(req.NIK)
		if err := ValidateNIK(nik); err != nil {
			middleware.AbortWithBadRequest(c, "Invalid NIK: "+err.Error())
			return
		}
		req.NIK = nik
	}

	// Validate SIM
	if req.SIMNumber != "" {
		if err := ValidateSIM(req.SIMNumber); err != nil {
			middleware.AbortWithBadRequest(c, "Invalid SIM: "+err.Error())
			return
		}
	}

	// Validate SIM type
	if req.SIMType != "" {
		if err := ValidateSIMType(req.SIMType); err != nil {
			middleware.AbortWithBadRequest(c, err.Error())
			return
		}
	}

	// Validate phone
	if req.Phone != "" {
		phone := CleanPhoneNumber(req.Phone)
		if err := ValidatePhoneNumber(phone); err != nil {
			middleware.AbortWithBadRequest(c, "Invalid phone: "+err.Error())
			return
		}
		req.Phone = FormatPhoneNumber(phone)
	}

	// Validate email
	if req.Email != "" {
		if err := ValidateEmail(req.Email); err != nil {
			middleware.AbortWithBadRequest(c, err.Error())
			return
		}
	}

	// Validate driver age
	if !req.BirthDate.IsZero() {
		if err := ValidateDriverAge(req.BirthDate); err != nil {
			middleware.AbortWithBadRequest(c, err.Error())
			return
		}
	}

	c.Set("validated_request", req)
	c.Next()
}

// ValidateGPSTrackRequest validates GPS track data
func ValidateGPSTrackRequest(c *gin.Context) {
	var req struct {
		VehicleID string  `json:"vehicle_id"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Speed     float64 `json:"speed"`
		Heading   float64 `json:"heading"`
		Accuracy  float64 `json:"accuracy"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Validate coordinates
	if err := ValidateCoordinates(req.Latitude, req.Longitude); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Sanitize coordinates
	lat, lng, err := SanitizeCoordinates(req.Latitude, req.Longitude)
	if err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}
	req.Latitude = lat
	req.Longitude = lng

	// Validate speed
	if req.Speed != 0 {
		if err := ValidateSpeed(req.Speed); err != nil {
			middleware.AbortWithBadRequest(c, err.Error())
			return
		}
	}

	// Validate heading
	if req.Heading != 0 {
		if err := ValidateHeading(req.Heading); err != nil {
			middleware.AbortWithBadRequest(c, err.Error())
			return
		}
	}

	// Validate accuracy
	if req.Accuracy != 0 {
		if err := ValidateAccuracy(req.Accuracy); err != nil {
			middleware.AbortWithBadRequest(c, err.Error())
			return
		}
	}

	c.Set("validated_request", req)
	c.Next()
}

// ValidatePaymentRequest validates payment request
func ValidatePaymentRequest(c *gin.Context) {
	var req struct {
		InvoiceID     string  `json:"invoice_id"`
		Amount        float64 `json:"amount"`
		PaymentMethod string  `json:"payment_method"`
		Currency      string  `json:"currency"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Validate amount
	if err := ValidateAmount(req.Amount); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Validate payment method
	if req.PaymentMethod != "" {
		if err := ValidatePaymentMethod(req.PaymentMethod); err != nil {
			middleware.AbortWithBadRequest(c, err.Error())
			return
		}
	}

	// Validate currency
	currency := req.Currency
	if currency == "" {
		currency = "IDR" // Default to Indonesian Rupiah
	}
	if err := ValidateCurrency(currency); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}
	req.Currency = currency

	c.Set("validated_request", req)
	c.Next()
}

// ValidatePaginationParams validates common pagination parameters
func ValidatePaginationParams() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get limit
		limitStr := c.DefaultQuery("limit", "20")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 1000 {
			middleware.AbortWithBadRequest(c, "Invalid limit: must be between 1 and 1000")
			return
		}

		// Get offset
		offsetStr := c.DefaultQuery("offset", "0")
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			middleware.AbortWithBadRequest(c, "Invalid offset: must be non-negative")
			return
		}

		// Store validated values
		c.Set("validated_limit", limit)
		c.Set("validated_offset", offset)

		c.Next()
	}
}

// ValidateDateRangeParams validates date range query parameters
func ValidateDateRangeParams(maxRange time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		startDateStr := c.Query("start_date")
		endDateStr := c.Query("end_date")

		if startDateStr == "" || endDateStr == "" {
			middleware.AbortWithBadRequest(c, "start_date and end_date are required")
			return
		}

		// Parse dates
		startDate, err := ValidateDate(startDateStr)
		if err != nil {
			middleware.AbortWithBadRequest(c, "Invalid start_date: "+err.Error())
			return
		}

		endDate, err := ValidateDate(endDateStr)
		if err != nil {
			middleware.AbortWithBadRequest(c, "Invalid end_date: "+err.Error())
			return
		}

		// Validate range
		if err := ValidateTimeRange(startDate, endDate, maxRange); err != nil {
			middleware.AbortWithBadRequest(c, err.Error())
			return
		}

		// Store validated dates
		c.Set("validated_start_date", startDate)
		c.Set("validated_end_date", endDate)

		c.Next()
	}
}

