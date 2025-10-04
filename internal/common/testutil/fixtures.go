package testutil

import (
	"time"

	"github.com/google/uuid"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// NewTestCompany creates a test company with default values
func NewTestCompany() *models.Company {
	return &models.Company{
		ID:               uuid.New().String(),
		Name:             "Test Company",
		Email:            "test@company.com",
		Phone:            "+62 21 1234567",
		NPWP:             "01.234.567.8-901.000",
		City:             "Jakarta",
		Province:         "DKI Jakarta",
		Country:          "Indonesia",
		CompanyType:      "PT",
		FleetSize:        10,
		MaxVehicles:      100,
		IsActive:         true,
		SubscriptionTier: "basic",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// NewTestUser creates a test user with default values
func NewTestUser(companyID string) *models.User {
	return &models.User{
		ID:          uuid.New().String(),
		CompanyID:   companyID,
		Email:       "test@user.com",
		Username:    "testuser",
		Password:    "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5lW5h8TQz5yPW", // password123
		FirstName:   "Test",
		LastName:    "User",
		Phone:       "+62 811 1234567",
		Role:        "admin",
		Status:      "active",
		IsActive:    true,
		IsVerified:  true,
		Language:    "id",
		Timezone:    "Asia/Jakarta",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewTestVehicle creates a test vehicle with default values
func NewTestVehicle(companyID string) *models.Vehicle {
	return &models.Vehicle{
		ID:            uuid.New().String(),
		CompanyID:     companyID,
		LicensePlate:  "B 1234 ABC",
		Make:          "Toyota",
		Model:         "Avanza",
		Year:          2023,
		Type:          "van",
		FuelType:      "gasoline",
		TankCapacity:  45.0,
		PurchasePrice: 250000000.0, // IDR
		Status:        "active",
		IsGPSEnabled:  true,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// NewTestDriver creates a test driver with default values
func NewTestDriver(companyID string) *models.Driver {
	licenseExpiry := time.Now().AddDate(2, 0, 0) // 2 years from now
	
	return &models.Driver{
		ID:                uuid.New().String(),
		CompanyID:         companyID,
		FirstName:         "Test",
		LastName:          "Driver",
		Email:             "driver@test.com",
		Phone:             "+62 812 3456789",
		NIK:               "3174012345678901", // 16 digits
		SIMNumber:         "SIM1234567890",
		SIMType:           "B1",
		SIMExpiry:         &licenseExpiry,
		Status:            "active",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

// NewTestGPSTrack creates a test GPS track with default values
func NewTestGPSTrack(vehicleID string) *models.GPSTrack {
	timestamp := time.Now()
	
	return &models.GPSTrack{
		ID:        uuid.New().String(),
		VehicleID: vehicleID,
		Latitude:  -6.2088,    // Jakarta
		Longitude: 106.8456,   // Jakarta
		Speed:     60.0,
		Heading:   180.0,
		Accuracy:  5.0,
		Timestamp: timestamp,
		CreatedAt: time.Now(),
	}
}

// NewTestTrip creates a test trip with default values
func NewTestTrip(companyID, vehicleID, driverID string) *models.Trip {
	startTime := time.Now().Add(-2 * time.Hour)
	
	return &models.Trip{
		ID:             uuid.New().String(),
		CompanyID:      companyID,
		VehicleID:      vehicleID,
		DriverID:       &driverID,
		Name:           "Test Trip",
		StartTime:      &startTime,
		Status:         "in_progress",
		StartLocation:  "Jakarta",
		StartLatitude:  -6.2088,
		StartLongitude: 106.8456,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// NewTestInvoice creates a test invoice with default values
func NewTestInvoice(companyID string) *models.Invoice {
	now := time.Now()
	dueDate := now.AddDate(0, 0, 30) // 30 days from now
	billingStart := now.AddDate(0, -1, 0) // 1 month ago
	
	return &models.Invoice{
		ID:                 uuid.New().String(),
		CompanyID:          companyID,
		InvoiceNumber:      "INV-" + uuid.New().String()[:8],
		InvoiceDate:        now,
		DueDate:            dueDate,
		BillingPeriodStart: billingStart,
		BillingPeriodEnd:   now,
		Subtotal:           1000000.0, // IDR
		TaxAmount:          110000.0,  // 11% PPN
		TotalAmount:        1110000.0,
		Currency:           "IDR",
		Status:             "draft",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}


// Helper function to create pointer to string
func PtrString(s string) *string {
	return &s
}

// Helper function to create pointer to time
func PtrTime(t time.Time) *time.Time {
	return &t
}

