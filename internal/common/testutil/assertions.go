package testutil

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertValidUUID checks if a string is a valid UUID
func AssertValidUUID(t *testing.T, id string, msgAndArgs ...interface{}) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return assert.Regexp(t, uuidRegex, id, msgAndArgs...)
}

// AssertValidEmail checks if a string is a valid email
func AssertValidEmail(t *testing.T, email string, msgAndArgs ...interface{}) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return assert.Regexp(t, emailRegex, email, msgAndArgs...)
}

// AssertValidNIK checks if a string is a valid Indonesian NIK (16 digits)
func AssertValidNIK(t *testing.T, nik string, msgAndArgs ...interface{}) bool {
	if !assert.Len(t, nik, 16, msgAndArgs...) {
		return false
	}
	nikRegex := regexp.MustCompile(`^\d{16}$`)
	return assert.Regexp(t, nikRegex, nik, msgAndArgs...)
}

// AssertValidNPWP checks if a string is a valid Indonesian NPWP format
func AssertValidNPWP(t *testing.T, npwp string, msgAndArgs ...interface{}) bool {
	// Format: XX.XXX.XXX.X-XXX.XXX
	npwpRegex := regexp.MustCompile(`^\d{2}\.\d{3}\.\d{3}\.\d-\d{3}\.\d{3}$`)
	return assert.Regexp(t, npwpRegex, npwp, msgAndArgs...)
}

// AssertValidIndonesianPhone checks if a string is a valid Indonesian phone number
func AssertValidIndonesianPhone(t *testing.T, phone string, msgAndArgs ...interface{}) bool {
	// Format: +62 XXX XXXX XXXX or similar
	phoneRegex := regexp.MustCompile(`^\+62\s?\d{2,3}\s?\d{4}\s?\d{4,5}$`)
	return assert.Regexp(t, phoneRegex, phone, msgAndArgs...)
}

// AssertValidLicensePlate checks if a string is a valid Indonesian license plate
func AssertValidLicensePlate(t *testing.T, plate string, msgAndArgs ...interface{}) bool {
	// Format: B 1234 ABC (can have 1-2 letters, 1-4 numbers, 1-3 letters)
	plateRegex := regexp.MustCompile(`^[A-Z]{1,2}\s?\d{1,4}\s?[A-Z]{1,3}$`)
	return assert.Regexp(t, plateRegex, plate, msgAndArgs...)
}

// AssertValidCurrency checks if amount is in valid IDR format (positive number)
func AssertValidCurrency(t *testing.T, amount float64, msgAndArgs ...interface{}) bool {
	return assert.Greater(t, amount, 0.0, msgAndArgs...)
}

// AssertValidPPN11 checks if tax amount is correctly calculated as 11% of base amount
func AssertValidPPN11(t *testing.T, baseAmount, taxAmount float64, msgAndArgs ...interface{}) bool {
	expectedTax := baseAmount * 0.11
	// Allow small floating point differences
	return assert.InDelta(t, expectedTax, taxAmount, 0.01, msgAndArgs...)
}

// AssertValidSIMType checks if SIM type is valid (A, B1, B2, C)
func AssertValidSIMType(t *testing.T, simType string, msgAndArgs ...interface{}) bool {
	validTypes := []string{"A", "B1", "B2", "C"}
	return assert.Contains(t, validTypes, simType, msgAndArgs...)
}

