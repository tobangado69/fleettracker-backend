package driver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/testutil"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

func TestService_CreateDriver(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test company
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	dob := time.Now().AddDate(-30, 0, 0) // 30 years old
	hireDate := time.Now().AddDate(-1, 0, 0) // Hired 1 year ago
	simExpiry := time.Now().AddDate(2, 0, 0) // SIM expires in 2 years

	tests := []struct {
		name    string
		request CreateDriverRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid driver creation",
			request: CreateDriverRequest{
				FirstName:   "Budi",
				LastName:    "Santoso",
				PhoneNumber: "+6281234567890",
				Email:       "budi@test.com",
				Address:     "Jl. Sudirman No. 123",
				City:        "Jakarta",
				Province:    "DKI Jakarta",
				DateOfBirth: dob,
				HireDate:    hireDate,
				NIK:         "3174012345678901", // 16 digit NIK
				SIMNumber:   "SIM1234567890",
				SIMExpiry:   &simExpiry,
			},
			wantErr: false,
		},
		{
			name: "invalid NIK - not 16 digits",
			request: CreateDriverRequest{
				FirstName:   "Budi",
				LastName:    "Santoso",
				PhoneNumber: "+6281234567890",
				Email:       "budi@test.com",
				Address:     "Jl. Sudirman No. 123",
				City:        "Jakarta",
				Province:    "DKI Jakarta",
				DateOfBirth: dob,
				HireDate:    hireDate,
				NIK:         "123456", // Invalid NIK
				SIMNumber:   "SIM1234567890",
			},
			wantErr: true,
			errMsg:  "NIK",
		},
		{
			name: "invalid email format",
			request: CreateDriverRequest{
				FirstName:   "Budi",
				LastName:    "Santoso",
				PhoneNumber: "+6281234567890",
				Email:       "invalid-email",
				Address:     "Jl. Sudirman No. 123",
				City:        "Jakarta",
				Province:    "DKI Jakarta",
				DateOfBirth: dob,
				HireDate:    hireDate,
				NIK:         "3174012345678901",
				SIMNumber:   "SIM1234567890",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, err := service.CreateDriver(company.ID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, driver)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, driver)
				testutil.AssertValidUUID(t, driver.ID)
				assert.Equal(t, company.ID, driver.CompanyID)
				assert.Equal(t, tt.request.FirstName, driver.FirstName)
				assert.Equal(t, tt.request.LastName, driver.LastName)
				testutil.AssertValidNIK(t, driver.NIK)
			}
		})
	}
}

func TestService_GetDriver(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	t.Run("get existing driver", func(t *testing.T) {
		result, err := service.GetDriver(company.ID, driver.ID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, driver.ID, result.ID)
		assert.Equal(t, driver.FirstName, result.FirstName)
		testutil.AssertValidNIK(t, result.NIK)
	})

	t.Run("get non-existent driver", func(t *testing.T) {
		result, err := service.GetDriver(company.ID, "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestService_UpdateDriver(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	t.Run("update driver name", func(t *testing.T) {
		newFirstName := "Ahmad"
		newLastName := "Wijaya"

		updated, err := service.UpdateDriver(company.ID, driver.ID, UpdateDriverRequest{
			FirstName: &newFirstName,
			LastName:  &newLastName,
		})

		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, newFirstName, updated.FirstName)
		assert.Equal(t, newLastName, updated.LastName)
	})

	t.Run("update driver phone", func(t *testing.T) {
		newPhone := "+6285555555555"

		updated, err := service.UpdateDriver(company.ID, driver.ID, UpdateDriverRequest{
			PhoneNumber: &newPhone,
		})

		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, newPhone, updated.Phone)
	})
}

func TestService_DeleteDriver(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	t.Run("soft delete driver", func(t *testing.T) {
		err := service.DeleteDriver(company.ID, driver.ID)

		assert.NoError(t, err)

		// Verify driver is soft deleted
		var deletedDriver models.Driver
		err = db.Unscoped().Where("id = ?", driver.ID).First(&deletedDriver).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedDriver.DeletedAt)
	})
}

func TestService_ListDrivers(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	// Create multiple drivers
	for i := 0; i < 5; i++ {
		driver := testutil.NewTestDriver(company.ID)
		require.NoError(t, db.Create(driver).Error)
	}

	t.Run("list all drivers", func(t *testing.T) {
		filters := DriverFilters{
			Page:      1,
			Limit:     10,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		drivers, total, err := service.ListDrivers(company.ID, filters)

		assert.NoError(t, err)
		assert.NotEmpty(t, drivers)
		assert.GreaterOrEqual(t, total, int64(5))
		assert.LessOrEqual(t, len(drivers), 10)

		// Verify all have valid NIK
		for _, d := range drivers {
			testutil.AssertValidNIK(t, d.NIK)
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		status := "active"
		filters := DriverFilters{
			Status:    &status,
			Page:      1,
			Limit:     10,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		drivers, _, err := service.ListDrivers(company.ID, filters)

		assert.NoError(t, err)
		for _, d := range drivers {
			assert.Equal(t, "active", d.Status)
		}
	})
}

func TestService_UpdateDriverStatus(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	tests := []struct {
		name   string
		status DriverStatus
		reason string
	}{
		{
			name:   "set to busy",
			status: StatusBusy,
			reason: "On delivery",
		},
		{
			name:   "set to available",
			status: StatusAvailable,
			reason: "Delivery completed",
		},
		{
			name:   "set to suspended",
			status: StatusSuspended,
			reason: "Policy violation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateDriverStatus(company.ID, driver.ID, tt.status, tt.reason)

			assert.NoError(t, err)

			// Verify status was updated
			updated, err := service.GetDriver(company.ID, driver.ID)
			assert.NoError(t, err)
			assert.Equal(t, string(tt.status), updated.Status)
		})
	}
}

func TestService_UpdateDriverPerformance(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	tests := []struct {
		name       string
		performance float64
		safety     float64
		efficiency float64
		expectErr  bool
	}{
		{
			name:       "excellent performance",
			performance: 95.0,
			safety:     98.0,
			efficiency: 92.0,
			expectErr:  false,
		},
		{
			name:       "average performance",
			performance: 75.0,
			safety:     80.0,
			efficiency: 70.0,
			expectErr:  false,
		},
		{
			name:       "invalid performance - out of range",
			performance: 150.0, // Invalid
			safety:     98.0,
			efficiency: 92.0,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateDriverPerformance(company.ID, driver.ID, tt.performance, tt.safety, tt.efficiency)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetDriverPerformance(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	// Update performance
	err := service.UpdateDriverPerformance(company.ID, driver.ID, 90.0, 95.0, 88.0)
	require.NoError(t, err)

	t.Run("get driver performance", func(t *testing.T) {
		result, err := service.GetDriverPerformance(company.ID, driver.ID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, driver.ID, result.ID)
	})
}

func TestService_AssignVehicle(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	vehicle := testutil.NewTestVehicle(company.ID)
	require.NoError(t, db.Create(vehicle).Error)

	t.Run("assign vehicle to driver", func(t *testing.T) {
		err := service.AssignVehicle(company.ID, driver.ID, vehicle.ID)

		assert.NoError(t, err)

		// Verify vehicle was assigned
		updated, err := service.GetDriver(company.ID, driver.ID)
		assert.NoError(t, err)
		assert.NotNil(t, updated.VehicleID)
		assert.Equal(t, vehicle.ID, *updated.VehicleID)
	})

	t.Run("unassign vehicle from driver", func(t *testing.T) {
		err := service.UnassignVehicle(company.ID, driver.ID)

		assert.NoError(t, err)

		// Verify vehicle was unassigned
		updated, err := service.GetDriver(company.ID, driver.ID)
		assert.NoError(t, err)
		assert.Nil(t, updated.VehicleID)
	})
}

func TestService_GetDriverVehicle(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	vehicle := testutil.NewTestVehicle(company.ID)
	require.NoError(t, db.Create(vehicle).Error)

	// Assign vehicle
	err := service.AssignVehicle(company.ID, driver.ID, vehicle.ID)
	require.NoError(t, err)

	t.Run("get assigned vehicle", func(t *testing.T) {
		result, err := service.GetDriverVehicle(company.ID, driver.ID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, vehicle.ID, result.ID)
		assert.Equal(t, vehicle.LicensePlate, result.LicensePlate)
	})
}

func TestService_UpdateMedicalCheckup(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	t.Run("update medical checkup date", func(t *testing.T) {
		checkupDate := time.Now()

		err := service.UpdateMedicalCheckup(company.ID, driver.ID, checkupDate)

		assert.NoError(t, err)

		// Verify driver was updated
		_, err = service.GetDriver(company.ID, driver.ID)
		assert.NoError(t, err)
	})
}

func TestService_UpdateTrainingStatus(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	t.Run("mark training as completed", func(t *testing.T) {
		expiryDate := time.Now().AddDate(1, 0, 0) // 1 year from now

		err := service.UpdateTrainingStatus(company.ID, driver.ID, true, &expiryDate)

		assert.NoError(t, err)

		// Verify training status was updated
		updated, err := service.GetDriver(company.ID, driver.ID)
		assert.NoError(t, err)
		assert.NotNil(t, updated)
	})
}

func TestService_ValidateIndonesianCompliance(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	dob := time.Now().AddDate(-30, 0, 0) // 30 years old

	tests := []struct {
		name        string
		nik         string
		simNumber   string
		dateOfBirth time.Time
		wantErr     bool
	}{
		{
			name:        "valid Indonesian documents",
			nik:         "3174012345678901", // 16 digits
			simNumber:   "SIM1234567890",
			dateOfBirth: dob,
			wantErr:     false,
		},
		{
			name:        "invalid NIK - not 16 digits",
			nik:         "123456", // Too short
			simNumber:   "SIM1234567890",
			dateOfBirth: dob,
			wantErr:     true,
		},
		{
			name:        "invalid SIM number - too short",
			nik:         "3174012345678901",
			simNumber:   "SHORT",
			dateOfBirth: dob,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateIndonesianCompliance(tt.nik, tt.simNumber, tt.dateOfBirth)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ValidateNIK(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	tests := []struct {
		name  string
		nik   string
		valid bool
	}{
		{
			name:  "valid NIK - Jakarta",
			nik:   "3174012345678901",
			valid: true,
		},
		{
			name:  "valid NIK - Surabaya",
			nik:   "3578012345678901",
			valid: true,
		},
		{
			name:  "invalid NIK - too short",
			nik:   "123456",
			valid: false,
		},
		{
			name:  "invalid NIK - contains letters",
			nik:   "3174AB1234567890",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateNIK(tt.nik)

			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestService_ValidateSIMNumber(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	tests := []struct {
		name      string
		simNumber string
		valid     bool
	}{
		{
			name:      "valid SIM number",
			simNumber: "SIM1234567890",
			valid:     true,
		},
		{
			name:      "valid SIM number - alternative format",
			simNumber: "1234567890ABCD",
			valid:     true,
		},
		{
			name:      "invalid SIM number - too short",
			simNumber: "SHORT",
			valid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateSIMNumber(tt.simNumber)

			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

