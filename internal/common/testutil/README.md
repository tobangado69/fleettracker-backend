# 🧪 Test Utilities Package

Comprehensive testing infrastructure for FleetTracker Pro backend with Indonesian-specific test data and assertions.

## 📦 Package Contents

### 1. `database.go` - Test Database Management
```go
// Setup test database with auto-migration and cleanup
db, cleanup := testutil.SetupTestDB(t)
defer cleanup()

// Clear database between tests
testutil.ClearDatabase(db)
```

**Features:**
- Automatic Postgres connection with test database
- Auto-migration of all 18 models
- Automatic cleanup before and after tests
- Silent logger mode for clean test output

### 2. `fixtures.go` - Test Data Generators
Pre-built test data factories for all Indonesian fleet management models:

```go
// Create test company
company := testutil.NewTestCompany()

// Create test user
user := testutil.NewTestUser(company.ID)

// Create test vehicle
vehicle := testutil.NewTestVehicle(company.ID)

// Create test driver
driver := testutil.NewTestDriver(company.ID)

// Create test GPS track
gpsTrack := testutil.NewTestGPSTrack(vehicle.ID)

// Create test trip
trip := testutil.NewTestTrip(company.ID, vehicle.ID, driver.ID)

// Create test invoice
invoice := testutil.NewTestInvoice(company.ID)
```

**All fixtures include:**
- ✅ Valid Indonesian data (NPWP, NIK, SIM numbers)
- ✅ Realistic Jakarta/Indonesia coordinates
- ✅ Proper UUID generation
- ✅ Correct field types and values
- ✅ Indonesian language defaults

### 3. `assertions.go` - Indonesian Validators

Custom assertions for Indonesian data formats:

```go
// Validate UUID format
testutil.AssertValidUUID(t, id)

// Validate email format
testutil.AssertValidEmail(t, "user@example.com")

// Validate Indonesian NIK (16 digits)
testutil.AssertValidNIK(t, "3174012345678901")

// Validate Indonesian NPWP format (XX.XXX.XXX.X-XXX.XXX)
testutil.AssertValidNPWP(t, "01.234.567.8-901.000")

// Validate Indonesian phone number (+62 XXX XXXX XXXX)
testutil.AssertValidIndonesianPhone(t, "+62 811 1234 5678")

// Validate Indonesian license plate (B 1234 ABC)
testutil.AssertValidLicensePlate(t, "B 1234 ABC")

// Validate Indonesian currency (IDR, positive)
testutil.AssertValidCurrency(t, 1000000.0)

// Validate Indonesian PPN 11% tax calculation
testutil.AssertValidPPN11(t, baseAmount, taxAmount)

// Validate SIM type (A, B1, B2, C)
testutil.AssertValidSIMType(t, "B1")
```

## 🚀 Quick Start

### 1. Basic Test Setup
```go
package myservice

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/tobangado69/fleettracker-pro/backend/internal/common/testutil"
)

func TestMyService(t *testing.T) {
    // Setup test database
    db, cleanup := testutil.SetupTestDB(t)
    defer cleanup()
    
    // Create test data
    company := testutil.NewTestCompany()
    require.NoError(t, db.Create(company).Error)
    
    // Run your test
    service := NewService(db)
    result, err := service.DoSomething(company.ID)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### 2. Test with Multiple Fixtures
```go
func TestComplexScenario(t *testing.T) {
    db, cleanup := testutil.SetupTestDB(t)
    defer cleanup()
    
    // Create complete test scenario
    company := testutil.NewTestCompany()
    require.NoError(t, db.Create(company).Error)
    
    user := testutil.NewTestUser(company.ID)
    require.NoError(t, db.Create(user).Error)
    
    vehicle := testutil.NewTestVehicle(company.ID)
    require.NoError(t, db.Create(vehicle).Error)
    
    driver := testutil.NewTestDriver(company.ID)
    require.NoError(t, db.Create(driver).Error)
    
    // Create trip with all relationships
    trip := testutil.NewTestTrip(company.ID, vehicle.ID, driver.ID)
    require.NoError(t, db.Create(trip).Error)
    
    // Test your service
    service := NewTripService(db)
    result, err := service.GetTripDetails(trip.ID)
    
    assert.NoError(t, err)
    assert.Equal(t, trip.ID, result.ID)
}
```

### 3. Test with Indonesian Validation
```go
func TestIndonesianCompliance(t *testing.T) {
    db, cleanup := testutil.SetupTestDB(t)
    defer cleanup()
    
    company := testutil.NewTestCompany()
    require.NoError(t, db.Create(company).Error)
    
    // Validate Indonesian formats
    testutil.AssertValidNPWP(t, company.NPWP)
    testutil.AssertValidIndonesianPhone(t, company.Phone)
    
    driver := testutil.NewTestDriver(company.ID)
    require.NoError(t, db.Create(driver).Error)
    
    testutil.AssertValidNIK(t, driver.NIK)
    testutil.AssertValidSIMType(t, driver.SIMType)
    
    vehicle := testutil.NewTestVehicle(company.ID)
    require.NoError(t, db.Create(vehicle).Error)
    
    testutil.AssertValidLicensePlate(t, vehicle.LicensePlate)
}
```

### 4. Test Payment Calculations
```go
func TestPaymentWithTax(t *testing.T) {
    db, cleanup := testutil.SetupTestDB(t)
    defer cleanup()
    
    company := testutil.NewTestCompany()
    require.NoError(t, db.Create(company).Error)
    
    invoice := testutil.NewTestInvoice(company.ID)
    require.NoError(t, db.Create(invoice).Error)
    
    // Validate Indonesian tax calculation (PPN 11%)
    testutil.AssertValidPPN11(t, invoice.Subtotal, invoice.TaxAmount)
    testutil.AssertValidCurrency(t, invoice.TotalAmount)
    
    assert.Equal(t, "IDR", invoice.Currency)
    assert.Equal(t, invoice.Subtotal + invoice.TaxAmount, invoice.TotalAmount)
}
```

## 📊 Test Data Examples

### Company Fixture
```json
{
  "name": "Test Company",
  "email": "test@company.com",
  "npwp": "01.234.567.8-901.000",
  "city": "Jakarta",
  "province": "DKI Jakarta",
  "country": "Indonesia",
  "company_type": "PT"
}
```

### Driver Fixture
```json
{
  "first_name": "Test",
  "last_name": "Driver",
  "nik": "3174012345678901",
  "sim_number": "SIM1234567890",
  "sim_type": "B1",
  "phone": "+62 812 3456789"
}
```

### Vehicle Fixture
```json
{
  "license_plate": "B 1234 ABC",
  "make": "Toyota",
  "model": "Avanza",
  "year": 2023,
  "type": "van",
  "fuel_type": "gasoline"
}
```

## 🔧 Configuration

### Database Connection
By default, tests use:
```
postgres://fleettracker:password123@host.docker.internal:5432/fleettracker?sslmode=disable
```

**Note**: Tests require either:
1. Docker network access to postgres container
2. Proper pg_hba.conf configuration for host connections
3. Running tests inside Docker container

### Environment Variables
```bash
# Optional: Override test database URL
export TEST_DATABASE_URL="postgres://user:pass@localhost:5432/testdb?sslmode=disable"

# Optional: Use main database URL
export DATABASE_URL="postgres://user:pass@localhost:5432/fleettracker?sslmode=disable"
```

## 📝 Best Practices

### 1. Always Use Cleanup
```go
db, cleanup := testutil.SetupTestDB(t)
defer cleanup() // Ensures database is cleared after test
```

### 2. Create Required Data First
```go
// Create parent entities before children
company := testutil.NewTestCompany()
db.Create(company)

// Then create child entities
user := testutil.NewTestUser(company.ID)
db.Create(user)
```

### 3. Use require for Setup, assert for Tests
```go
// require.NoError stops test on setup failure
require.NoError(t, db.Create(company).Error)

// assert.NoError continues test to show all failures
assert.NoError(t, err)
assert.NotNil(t, result)
```

### 4. Test Indonesian Compliance
```go
// Always validate Indonesian-specific formats
testutil.AssertValidNPWP(t, company.NPWP)
testutil.AssertValidNIK(t, driver.NIK)
testutil.AssertValidPPN11(t, invoice.Subtotal, invoice.TaxAmount)
```

## 🎯 Coverage Goals

Target test coverage by service:
- **Auth Service**: 85%+ ✅ (Completed)
- **GPS Tracking**: 85%+ (Next)
- **Payment Service**: 85%+ (Next)
- **Vehicle Service**: 80%+
- **Driver Service**: 80%+
- **Analytics Service**: 75%+

## 🐛 Troubleshooting

### Issue: "Binary was compiled with 'CGO_ENABLED=0'"
**Solution**: Use Postgres instead of SQLite (already configured)

### Issue: "password authentication failed"
**Solution**: Use host.docker.internal or run tests in Docker network

### Issue: "no pg_hba.conf entry for host"
**Solution**: Configure pg_hba.conf to trust connections from host, or run tests inside Docker

### Issue: "relation does not exist"
**Solution**: Auto-migration runs automatically, but ensure database URL is correct

## 📚 Dependencies

```go
require (
    github.com/stretchr/testify v1.11.1  // Assertions and testing utilities
    github.com/DATA-DOG/go-sqlmock v1.5.2  // SQL mocking (for unit tests)
    gorm.io/gorm v1.30.0                 // ORM
    gorm.io/driver/postgres              // Postgres driver
)
```

## 🤝 Contributing

When adding new fixtures:
1. Follow Indonesian data format standards
2. Use realistic test data (Jakarta coordinates, valid NPWP, etc.)
3. Add corresponding custom assertions
4. Update this README with examples

## 📄 License

Part of FleetTracker Pro - Indonesian Fleet Management SaaS Platform

