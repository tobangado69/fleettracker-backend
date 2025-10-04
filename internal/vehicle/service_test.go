package vehicle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/testutil"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

func TestService_CreateVehicle(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test company
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	purchaseDate := time.Now().AddDate(-1, 0, 0) // 1 year ago
	inspectionDate := time.Now().AddDate(0, -1, 0) // 1 month ago

	tests := []struct {
		name    string
		request CreateVehicleRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid vehicle creation",
			request: CreateVehicleRequest{
				Make:                  "Toyota",
				Model:                 "Avanza",
				Year:                  2023,
				LicensePlate:          "B 1234 ABC",
				VIN:                   "MHFZ12345678901234",
				Color:                 "Silver",
				FuelType:              "gasoline",
				CurrentOdometer:       50000,
				PurchaseDate:          &purchaseDate,
				STNKNumber:            "STNK123456789",
				BPKBNumber:            "BPKB123456789",
				InsurancePolicyNumber: "INS-2023-12345",
				LastInspectionDate:    &inspectionDate,
			},
			wantErr: false,
		},
		{
			name: "invalid license plate format",
			request: CreateVehicleRequest{
				Make:                  "Toyota",
				Model:                 "Avanza",
				Year:                  2023,
				LicensePlate:          "INVALID",
				VIN:                   "MHFZ12345678901234",
				Color:                 "Silver",
				FuelType:              "gasoline",
				STNKNumber:            "STNK123456789",
				BPKBNumber:            "BPKB123456789",
				InsurancePolicyNumber: "INS-2023-12345",
			},
			wantErr: true,
			errMsg:  "license plate",
		},
		{
			name: "invalid year",
			request: CreateVehicleRequest{
				Make:                  "Toyota",
				Model:                 "Avanza",
				Year:                  1800, // Too old
				LicensePlate:          "B 1234 ABC",
				VIN:                   "MHFZ12345678901234",
				Color:                 "Silver",
				FuelType:              "gasoline",
				STNKNumber:            "STNK123456789",
				BPKBNumber:            "BPKB123456789",
				InsurancePolicyNumber: "INS-2023-12345",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vehicle, err := service.CreateVehicle(company.ID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, vehicle)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, vehicle)
				testutil.AssertValidUUID(t, vehicle.ID)
				assert.Equal(t, company.ID, vehicle.CompanyID)
				assert.Equal(t, tt.request.Make, vehicle.Make)
				assert.Equal(t, tt.request.Model, vehicle.Model)
				assert.Equal(t, tt.request.Year, vehicle.Year)
				testutil.AssertValidLicensePlate(t, vehicle.LicensePlate)
				assert.True(t, vehicle.IsActive)
			}
		})
	}
}

func TestService_GetVehicle(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	vehicle := testutil.NewTestVehicle(company.ID)
	require.NoError(t, db.Create(vehicle).Error)

	t.Run("get existing vehicle", func(t *testing.T) {
		result, err := service.GetVehicle(company.ID, vehicle.ID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, vehicle.ID, result.ID)
		assert.Equal(t, vehicle.LicensePlate, result.LicensePlate)
	})

	t.Run("get non-existent vehicle", func(t *testing.T) {
		result, err := service.GetVehicle(company.ID, "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestService_UpdateVehicle(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	vehicle := testutil.NewTestVehicle(company.ID)
	require.NoError(t, db.Create(vehicle).Error)

	t.Run("update vehicle make and model", func(t *testing.T) {
		newMake := "Honda"
		newModel := "CR-V"

		updated, err := service.UpdateVehicle(company.ID, vehicle.ID, UpdateVehicleRequest{
			Make:  &newMake,
			Model: &newModel,
		})

		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, newMake, updated.Make)
		assert.Equal(t, newModel, updated.Model)
	})

	t.Run("update vehicle status", func(t *testing.T) {
		newStatus := "maintenance"

		updated, err := service.UpdateVehicle(company.ID, vehicle.ID, UpdateVehicleRequest{
			Status: &newStatus,
		})

		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, newStatus, updated.Status)
	})
}

func TestService_DeleteVehicle(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	vehicle := testutil.NewTestVehicle(company.ID)
	require.NoError(t, db.Create(vehicle).Error)

	t.Run("soft delete vehicle", func(t *testing.T) {
		err := service.DeleteVehicle(company.ID, vehicle.ID)

		assert.NoError(t, err)

		// Verify vehicle is soft deleted
		var deletedVehicle models.Vehicle
		err = db.Unscoped().Where("id = ?", vehicle.ID).First(&deletedVehicle).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedVehicle.DeletedAt)
	})
}

func TestService_ListVehicles(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	// Create multiple vehicles
	for i := 0; i < 5; i++ {
		vehicle := testutil.NewTestVehicle(company.ID)
		require.NoError(t, db.Create(vehicle).Error)
	}

	t.Run("list all vehicles", func(t *testing.T) {
		filters := VehicleFilters{
			Page:      1,
			Limit:     10,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		vehicles, total, err := service.ListVehicles(company.ID, filters)

		assert.NoError(t, err)
		assert.NotEmpty(t, vehicles)
		assert.GreaterOrEqual(t, total, int64(5))
		assert.LessOrEqual(t, len(vehicles), 10)
	})

	t.Run("filter by make", func(t *testing.T) {
		make := "Toyota"
		filters := VehicleFilters{
			Make:      &make,
			Page:      1,
			Limit:     10,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		vehicles, _, err := service.ListVehicles(company.ID, filters)

		assert.NoError(t, err)
		for _, v := range vehicles {
			assert.Equal(t, "Toyota", v.Make)
		}
	})
}

func TestService_UpdateVehicleStatus(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	vehicle := testutil.NewTestVehicle(company.ID)
	require.NoError(t, db.Create(vehicle).Error)

	tests := []struct {
		name   string
		status VehicleStatus
		reason string
	}{
		{
			name:   "set to maintenance",
			status: StatusMaintenance,
			reason: "Scheduled maintenance",
		},
		{
			name:   "set to active",
			status: StatusActive,
			reason: "Maintenance completed",
		},
		{
			name:   "set to inactive",
			status: StatusInactive,
			reason: "Temporarily out of service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateVehicleStatus(company.ID, vehicle.ID, tt.status, tt.reason)

			assert.NoError(t, err)

			// Verify status was updated
			updated, err := service.GetVehicle(company.ID, vehicle.ID)
			assert.NoError(t, err)
			assert.Equal(t, string(tt.status), updated.Status)
		})
	}
}

func TestService_AssignDriver(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	vehicle := testutil.NewTestVehicle(company.ID)
	require.NoError(t, db.Create(vehicle).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	t.Run("assign driver to vehicle", func(t *testing.T) {
		err := service.AssignDriver(company.ID, vehicle.ID, driver.ID)

		assert.NoError(t, err)

		// Verify driver was assigned
		updated, err := service.GetVehicle(company.ID, vehicle.ID)
		assert.NoError(t, err)
		assert.NotNil(t, updated.DriverID)
		assert.Equal(t, driver.ID, *updated.DriverID)
	})

	t.Run("unassign driver from vehicle", func(t *testing.T) {
		err := service.UnassignDriver(company.ID, vehicle.ID)

		assert.NoError(t, err)

		// Verify driver was unassigned
		updated, err := service.GetVehicle(company.ID, vehicle.ID)
		assert.NoError(t, err)
		assert.Nil(t, updated.DriverID)
	})
}

func TestService_GetVehicleDriver(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	vehicle := testutil.NewTestVehicle(company.ID)
	require.NoError(t, db.Create(vehicle).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	// Assign driver
	err := service.AssignDriver(company.ID, vehicle.ID, driver.ID)
	require.NoError(t, err)

	t.Run("get assigned driver", func(t *testing.T) {
		result, err := service.GetVehicleDriver(company.ID, vehicle.ID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, driver.ID, result.ID)
		assert.Equal(t, driver.FirstName, result.FirstName)
	})
}

func TestService_UpdateInspectionDate(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	vehicle := testutil.NewTestVehicle(company.ID)
	require.NoError(t, db.Create(vehicle).Error)

	t.Run("update inspection date", func(t *testing.T) {
		inspectionDate := time.Now()

		err := service.UpdateInspectionDate(company.ID, vehicle.ID, inspectionDate)

		assert.NoError(t, err)

		// Verify vehicle was updated
		_, err = service.GetVehicle(company.ID, vehicle.ID)
		assert.NoError(t, err)
	})
}

func TestService_ValidateIndonesianCompliance(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	tests := []struct {
		name         string
		stnkNumber   string
		bpkbNumber   string
		licensePlate string
		wantErr      bool
	}{
		{
			name:         "valid Indonesian documents",
			stnkNumber:   "STNK123456789",
			bpkbNumber:   "BPKB123456789",
			licensePlate: "B 1234 ABC",
			wantErr:      false,
		},
		{
			name:         "invalid STNK number - too short",
			stnkNumber:   "SHORT",
			bpkbNumber:   "BPKB123456789",
			licensePlate: "B 1234 ABC",
			wantErr:      true,
		},
		{
			name:         "invalid license plate format",
			stnkNumber:   "STNK123456789",
			bpkbNumber:   "BPKB123456789",
			licensePlate: "INVALID",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateIndonesianCompliance(tt.stnkNumber, tt.bpkbNumber, tt.licensePlate)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ValidateIndonesianLicensePlate(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db)

	tests := []struct {
		name  string
		plate string
		valid bool
	}{
		{
			name:  "valid Jakarta plate - B 1234 ABC",
			plate: "B 1234 ABC",
			valid: true,
		},
		{
			name:  "valid Surabaya plate - L 5678 DEF",
			plate: "L 5678 DEF",
			valid: true,
		},
		{
			name:  "valid compact format - B1234ABC",
			plate: "B1234ABC",
			valid: true,
		},
		{
			name:  "invalid format - no numbers",
			plate: "B ABCD EFG",
			valid: false,
		},
		{
			name:  "invalid format - too short",
			plate: "B 12 A",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateIndonesianLicensePlate(tt.plate)

			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

