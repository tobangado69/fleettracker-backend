package validators

import (
	"fmt"
	"strings"
)

// Validator provides comprehensive validation functionality
type Validator struct {
	sanitizer *Sanitizer
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		sanitizer: NewSanitizer(),
	}
}

// ValidationError represents a validation error with field information
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// Error implements error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", ve.Field, ve.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}

	messages := make([]string, len(ve))
	for i, err := range ve {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

// AddError adds a validation error
func (ve *ValidationErrors) AddError(field, message string) {
	*ve = append(*ve, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors returns true if there are validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// ValidateVehicleData validates complete vehicle data
func (v *Validator) ValidateVehicleData(data map[string]interface{}) error {
	errors := ValidationErrors{}

	// Validate license plate
	if licensePlate, ok := data["license_plate"].(string); ok && licensePlate != "" {
		if err := ValidatePlateNumber(licensePlate); err != nil {
			errors.AddError("license_plate", err.Error())
		}
	}

	// Validate VIN
	if vin, ok := data["vin"].(string); ok && vin != "" {
		if err := ValidateVIN(vin); err != nil {
			errors.AddError("vin", err.Error())
		}
	}

	// Validate year
	if year, ok := data["year"].(float64); ok {
		if err := ValidateVehicleYear(int(year)); err != nil {
			errors.AddError("year", err.Error())
		}
	}

	// Validate fuel type
	if fuelType, ok := data["fuel_type"].(string); ok && fuelType != "" {
		if err := ValidateFuelType(fuelType); err != nil {
			errors.AddError("fuel_type", err.Error())
		}
	}

	// Validate capacity
	if capacity, ok := data["capacity"].(float64); ok {
		if err := ValidateVehicleCapacity(int(capacity)); err != nil {
			errors.AddError("capacity", err.Error())
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// ValidateDriverData validates complete driver data
func (v *Validator) ValidateDriverData(data map[string]interface{}) error {
	errors := ValidationErrors{}

	// Validate NIK
	if nik, ok := data["nik"].(string); ok && nik != "" {
		nikClean := NormalizeNIK(nik)
		if err := ValidateNIK(nikClean); err != nil {
			errors.AddError("nik", err.Error())
		}
	}

	// Validate SIM
	if sim, ok := data["sim_number"].(string); ok && sim != "" {
		if err := ValidateSIM(sim); err != nil {
			errors.AddError("sim_number", err.Error())
		}
	}

	// Validate phone
	if phone, ok := data["phone"].(string); ok && phone != "" {
		phoneClean := CleanPhoneNumber(phone)
		if err := ValidatePhoneNumber(phoneClean); err != nil {
			errors.AddError("phone", err.Error())
		}
	}

	// Validate email
	if email, ok := data["email"].(string); ok && email != "" {
		if err := ValidateEmail(email); err != nil {
			errors.AddError("email", err.Error())
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// ValidateGPSData validates GPS track data
func (v *Validator) ValidateGPSData(latitude, longitude, speed, heading, accuracy float64) error {
	errors := ValidationErrors{}

	// Validate coordinates
	if err := ValidateCoordinates(latitude, longitude); err != nil {
		errors.AddError("coordinates", err.Error())
	}

	// Validate speed
	if speed != 0 {
		if err := ValidateSpeed(speed); err != nil {
			errors.AddError("speed", err.Error())
		}
	}

	// Validate heading
	if heading != 0 {
		if err := ValidateHeading(heading); err != nil {
			errors.AddError("heading", err.Error())
		}
	}

	// Validate accuracy
	if accuracy != 0 {
		if err := ValidateAccuracy(accuracy); err != nil {
			errors.AddError("accuracy", err.Error())
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// ValidatePaymentData validates payment data
func (v *Validator) ValidatePaymentData(amount float64, method, currency string) error {
	errors := ValidationErrors{}

	// Validate amount
	if err := ValidateAmount(amount); err != nil {
		errors.AddError("amount", err.Error())
	}

	// Validate payment method
	if method != "" {
		if err := ValidatePaymentMethod(method); err != nil {
			errors.AddError("payment_method", err.Error())
		}
	}

	// Validate currency
	if currency != "" {
		if err := ValidateCurrency(currency); err != nil {
			errors.AddError("currency", err.Error())
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// SanitizeAndValidate sanitizes input then validates it
func (v *Validator) SanitizeAndValidate(input string, validator func(string) error) (string, error) {
	// Sanitize first
	sanitized := v.sanitizer.SanitizeInput(input)

	// Then validate
	if err := validator(sanitized); err != nil {
		return "", err
	}

	return sanitized, nil
}

