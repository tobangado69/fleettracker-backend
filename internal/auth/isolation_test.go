package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/testutil"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// TestCompanyIsolation_VehicleAccess tests that owners cannot access other company's vehicles
func TestCompanyIsolation_VehicleAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create two companies
	companyA := &models.Company{
		Name:  "Company A",
		Email: "companya@test.com",
	}
	require.NoError(t, db.Create(companyA).Error)

	companyB := &models.Company{
		Name:  "Company B",
		Email: "companyb@test.com",
	}
	require.NoError(t, db.Create(companyB).Error)

	// Create vehicle for Company A
	vehicleA := &models.Vehicle{
		CompanyID:    companyA.ID,
		LicensePlate: "B1234ABC",
		Make:         "Toyota",
		Model:        "Avanza",
		Year:         2023,
		Status:       "active",
	}
	require.NoError(t, db.Create(vehicleA).Error)

	// Create vehicle for Company B
	vehicleB := &models.Vehicle{
		CompanyID:    companyB.ID,
		LicensePlate: "B5678DEF",
		Make:         "Honda",
		Model:        "Brio",
		Year:         2023,
		Status:       "active",
	}
	require.NoError(t, db.Create(vehicleB).Error)

	t.Run("Owner cannot access other company's vehicle", func(t *testing.T) {
		// Try to access Company B's vehicle with Company A's ID
		var vehicle models.Vehicle
		err := db.WithContext(ctx).
			Where("id = ? AND company_id = ?", vehicleB.ID, companyA.ID).
			First(&vehicle).Error

		// Should not find the vehicle
		assert.Error(t, err)
		assert.Equal(t, "record not found", err.Error())
	})

	t.Run("Owner can access own company's vehicle", func(t *testing.T) {
		// Access Company A's vehicle with Company A's ID
		var vehicle models.Vehicle
		err := db.WithContext(ctx).
			Where("id = ? AND company_id = ?", vehicleA.ID, companyA.ID).
			First(&vehicle).Error

		// Should find the vehicle
		assert.NoError(t, err)
		assert.Equal(t, vehicleA.ID, vehicle.ID)
		assert.Equal(t, companyA.ID, vehicle.CompanyID)
	})

	t.Run("Super-admin can access any company's vehicle", func(t *testing.T) {
		// Access Company B's vehicle without company filter (super-admin)
		var vehicle models.Vehicle
		err := db.WithContext(ctx).
			Where("id = ?", vehicleB.ID).
			First(&vehicle).Error

		// Should find the vehicle
		assert.NoError(t, err)
		assert.Equal(t, vehicleB.ID, vehicle.ID)
		assert.Equal(t, companyB.ID, vehicle.CompanyID)
	})
}

// TestCompanyIsolation_DriverAccess tests that owners cannot access other company's drivers
func TestCompanyIsolation_DriverAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create two companies
	companyA := &models.Company{
		Name:  "Company A",
		Email: "companya@test.com",
	}
	require.NoError(t, db.Create(companyA).Error)

	companyB := &models.Company{
		Name:  "Company B",
		Email: "companyb@test.com",
	}
	require.NoError(t, db.Create(companyB).Error)

	// Create driver for Company A
	driverA := &models.Driver{
		CompanyID:   companyA.ID,
		FirstName:   "Driver",
		LastName:    "A",
		NIK:         "1234567890123456",
		SIMNumber:   "SIM-A-001",
		SIMType:     "A",
		Phone:       "+628123456789",
		Status:   "active",
		IsActive: true,
	}
	require.NoError(t, db.Create(driverA).Error)

	// Create driver for Company B
	driverB := &models.Driver{
		CompanyID:   companyB.ID,
		FirstName:   "Driver",
		LastName:    "B",
		NIK:         "6543210987654321",
		SIMNumber:   "SIM-B-001",
		SIMType:     "A",
		Phone:       "+628987654321",
		Status:   "active",
		IsActive: true,
	}
	require.NoError(t, db.Create(driverB).Error)

	t.Run("Owner cannot access other company's driver", func(t *testing.T) {
		// Try to access Company B's driver with Company A's ID
		var driver models.Driver
		err := db.WithContext(ctx).
			Where("id = ? AND company_id = ?", driverB.ID, companyA.ID).
			First(&driver).Error

		// Should not find the driver
		assert.Error(t, err)
		assert.Equal(t, "record not found", err.Error())
	})

	t.Run("Owner can access own company's driver", func(t *testing.T) {
		// Access Company A's driver with Company A's ID
		var driver models.Driver
		err := db.WithContext(ctx).
			Where("id = ? AND company_id = ?", driverA.ID, companyA.ID).
			First(&driver).Error

		// Should find the driver
		assert.NoError(t, err)
		assert.Equal(t, driverA.ID, driver.ID)
		assert.Equal(t, companyA.ID, driver.CompanyID)
	})

	t.Run("Super-admin can access any company's driver", func(t *testing.T) {
		// Access Company B's driver without company filter (super-admin)
		var driver models.Driver
		err := db.WithContext(ctx).
			Where("id = ?", driverB.ID).
			First(&driver).Error

		// Should find the driver
		assert.NoError(t, err)
		assert.Equal(t, driverB.ID, driver.ID)
		assert.Equal(t, companyB.ID, driver.CompanyID)
	})
}

// TestCompanyIsolation_UserAccess tests that owners cannot access other company's users
func TestCompanyIsolation_UserAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create two companies
	companyA := &models.Company{
		Name:  "Company A",
		Email: "companya@test.com",
	}
	require.NoError(t, db.Create(companyA).Error)

	companyB := &models.Company{
		Name:  "Company B",
		Email: "companyb@test.com",
	}
	require.NoError(t, db.Create(companyB).Error)

	// Create user for Company A
	userA := &models.User{
		CompanyID: companyA.ID,
		Email:     "usera@companya.com",
		Username:  "usera",
		FirstName: "User",
		LastName:  "A",
		Role:      "owner",
		IsActive:  true,
		Status:    "active",
	}
	require.NoError(t, db.Create(userA).Error)

	// Create user for Company B
	userB := &models.User{
		CompanyID: companyB.ID,
		Email:     "userb@companyb.com",
		Username:  "userb",
		FirstName: "User",
		LastName:  "B",
		Role:      "owner",
		IsActive:  true,
		Status:    "active",
	}
	require.NoError(t, db.Create(userB).Error)

	t.Run("Owner cannot see other company's users", func(t *testing.T) {
		// Try to access Company B's users with Company A's ID
		var users []models.User
		err := db.WithContext(ctx).
			Where("company_id = ?", companyA.ID).
			Find(&users).Error

		require.NoError(t, err)
		assert.Equal(t, 1, len(users))
		assert.Equal(t, userA.ID, users[0].ID)

		// Should not include Company B's users
		for _, user := range users {
			assert.NotEqual(t, userB.ID, user.ID)
			assert.Equal(t, companyA.ID, user.CompanyID)
		}
	})

	t.Run("Super-admin can see all users", func(t *testing.T) {
		// Access all users without company filter (super-admin)
		var users []models.User
		err := db.WithContext(ctx).
			Find(&users).Error

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 2)

		// Should include users from both companies
		var foundA, foundB bool
		for _, user := range users {
			if user.ID == userA.ID {
				foundA = true
			}
			if user.ID == userB.ID {
				foundB = true
			}
		}
		assert.True(t, foundA, "Should find Company A's user")
		assert.True(t, foundB, "Should find Company B's user")
	})
}

// TestCompanyIsolation_ListQueries tests that list queries properly filter by company
func TestCompanyIsolation_ListQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create two companies
	companyA := &models.Company{
		Name:  "Company A",
		Email: "companya@test.com",
	}
	require.NoError(t, db.Create(companyA).Error)

	companyB := &models.Company{
		Name:  "Company B",
		Email: "companyb@test.com",
	}
	require.NoError(t, db.Create(companyB).Error)

	// Create 3 vehicles for Company A
	for i := 1; i <= 3; i++ {
		vehicle := &models.Vehicle{
			CompanyID:    companyA.ID,
			LicensePlate: "B" + string(rune('0'+i)) + "000A",
			Make:         "Toyota",
			Model:        "Avanza",
			Year:         2023,
			Status:       "active",
		}
		require.NoError(t, db.Create(vehicle).Error)
	}

	// Create 2 vehicles for Company B
	for i := 1; i <= 2; i++ {
		vehicle := &models.Vehicle{
			CompanyID:    companyB.ID,
			LicensePlate: "D" + string(rune('0'+i)) + "000B",
			Make:         "Honda",
			Model:        "Brio",
			Year:         2023,
			Status:       "active",
		}
		require.NoError(t, db.Create(vehicle).Error)
	}

	t.Run("List query returns only company's vehicles", func(t *testing.T) {
		// Query Company A's vehicles
		var vehicles []models.Vehicle
		err := db.WithContext(ctx).
			Where("company_id = ?", companyA.ID).
			Find(&vehicles).Error

		require.NoError(t, err)
		assert.Equal(t, 3, len(vehicles))

		// All vehicles should belong to Company A
		for _, vehicle := range vehicles {
			assert.Equal(t, companyA.ID, vehicle.CompanyID)
		}
	})

	t.Run("Count query returns only company's count", func(t *testing.T) {
		// Count Company B's vehicles
		var count int64
		err := db.Model(&models.Vehicle{}).
			Where("company_id = ?", companyB.ID).
			Count(&count).Error

		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})
}

