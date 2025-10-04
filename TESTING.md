# 🧪 FleetTracker Pro - Testing Guide

Complete guide for running and understanding the test suite for FleetTracker Pro backend.

## 📋 Table of Contents

- [Overview](#overview)
- [Test Infrastructure](#test-infrastructure)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Testing Philosophy](#testing-philosophy)
- [CI/CD Integration](#cicd-integration)

## 🎯 Overview

FleetTracker Pro uses **real database integration tests** (no mocking) to ensure code quality and Indonesian compliance.

### Test Statistics

| Metric | Value |
|--------|-------|
| **Total Test Files** | 6 files |
| **Test Infrastructure** | 766 lines |
| **Service Tests** | 3,400+ lines |
| **Integration Tests** | 400+ lines |
| **Total Test Code** | **4,566 lines** |
| **Test Suites** | 70+ suites |
| **Test Cases** | 150+ cases |
| **Coverage Target** | 80%+ per service |
| **Mocks Used** | **0 (ZERO)** ✅ |

### Services Tested

- ✅ **Auth Service** (13 test cases + integration tests)
- ✅ **GPS Tracking Service** (35+ test cases)
- ✅ **Payment Service** (30+ test cases with Indonesian tax)
- ✅ **Vehicle Service** (40+ test cases with STNK/BPKB)
- ✅ **Driver Service** (50+ test cases with NIK/SIM)

## 🏗️ Test Infrastructure

### Test Utilities Package

Located in `internal/common/testutil/`:

- **`database.go`** - Test database setup and cleanup
- **`fixtures.go`** - Test data generators for all models
- **`assertions.go`** - Indonesian-specific validators

### Key Features

1. **Real PostgreSQL Database**
   - Uses actual postgres connection
   - Auto-migration of all models
   - Automatic cleanup between tests

2. **Indonesian Compliance Testing**
   - NIK validation (16-digit Indonesian ID)
   - NPWP validation (tax ID format)
   - SIM validation (driver's license)
   - STNK/BPKB validation (vehicle documents)
   - PPN 11% tax calculation
   - Indonesian phone numbers
   - License plate formats

3. **Test Fixtures**
   - Pre-built test data for all models
   - Realistic Indonesian data
   - UUID generation
   - Proper relationships

## 🚀 Running Tests

### Prerequisites

Ensure PostgreSQL is running:

```bash
# Via Docker Compose
docker-compose up -d postgres

# Check if postgres is ready
docker exec fleettracker-postgres pg_isready
```

### Run All Tests

```bash
# From backend directory
cd backend

# Run all unit tests
go test -v ./internal/...

# Run with coverage
go test -v -cover ./internal/...

# Run specific service tests
go test -v ./internal/auth/...
go test -v ./internal/tracking/...
go test -v ./internal/payment/...
go test -v ./internal/vehicle/...
go test -v ./internal/driver/...
```

### Run Integration Tests

```bash
# Run handler integration tests
go test -v ./internal/auth/handler_test.go

# Run all integration tests
go test -v ./internal/*/handler_test.go
```

### Using Test Scripts

#### Basic Test Script

```bash
./run-tests.sh
```

#### Comprehensive Coverage Report

```bash
# Run comprehensive test suite with coverage
./test-coverage.sh

# View HTML coverage report
open coverage/coverage.html  # macOS
xdg-open coverage/coverage.html  # Linux
start coverage/coverage.html  # Windows
```

## 📊 Test Coverage

### Generate Coverage Report

```bash
# Generate coverage for all packages
go test -coverprofile=coverage.out ./internal/...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Check total coverage
go tool cover -func=coverage.out | grep total
```

### Coverage by Service

Target coverage: **80%+**

```bash
# Auth Service
go test -cover ./internal/auth/...

# GPS Tracking Service
go test -cover ./internal/tracking/...

# Payment Service (Indonesian Tax)
go test -cover ./internal/payment/...

# Vehicle Service
go test -cover ./internal/vehicle/...

# Driver Service
go test -cover ./internal/driver/...
```

### Coverage Thresholds

The CI/CD pipeline enforces a minimum **75% coverage** threshold.

## 🧪 Testing Philosophy

### Why Real Database Tests?

We use **real database integration tests** instead of mocks:

**Benefits:**
- ✅ Tests actual database behavior
- ✅ Catches real SQL/GORM issues
- ✅ No mock maintenance overhead
- ✅ Tests data integrity constraints
- ✅ Validates foreign key relationships
- ✅ Real transaction behavior
- ✅ Indonesian compliance validation

**Trade-offs:**
- ⚠️ Slightly slower than unit tests with mocks
- ⚠️ Requires running PostgreSQL
- ⚠️ Database state management needed

### Test Structure

All tests follow the **AAA pattern**:

```go
func TestService_Example(t *testing.T) {
    // Arrange - Setup
    db, cleanup := testutil.SetupTestDB(t)
    defer cleanup()
    
    service := NewService(db)
    company := testutil.NewTestCompany()
    require.NoError(t, db.Create(company).Error)
    
    // Act - Execute
    result, err := service.DoSomething(company.ID)
    
    // Assert - Verify
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, expected, result.Field)
}
```

### Indonesian Compliance Testing

All tests validate Indonesian-specific requirements:

```go
// Validate NIK (16-digit Indonesian ID)
testutil.AssertValidNIK(t, driver.NIK)

// Validate NPWP (tax ID)
testutil.AssertValidNPWP(t, company.NPWP)

// Validate PPN 11% (Indonesian VAT)
testutil.AssertValidPPN11(t, invoice.Subtotal, invoice.TaxAmount)

// Validate SIM (driver's license)
testutil.AssertValidSIMType(t, driver.SIMType)

// Validate license plate
testutil.AssertValidLicensePlate(t, vehicle.LicensePlate)
```

## 🔄 CI/CD Integration

### GitHub Actions Workflow

Located in `.github/workflows/test.yml`

The CI/CD pipeline automatically:

1. **Sets up test environment**
   - PostgreSQL with PostGIS
   - Go 1.24
   - Dependencies

2. **Runs database migrations**
   - Applies all migrations
   - Prepares test database

3. **Executes test suite**
   - Unit tests for all services
   - Integration tests for handlers
   - Race condition detection (`-race`)
   - Coverage reporting

4. **Generates coverage reports**
   - Per-service coverage
   - Combined coverage report
   - Uploads to Codecov

5. **Enforces quality gates**
   - Minimum 75% coverage
   - All tests must pass
   - Linting must pass
   - Build must succeed

### Running Locally Like CI

```bash
# Simulate CI environment
docker-compose up -d postgres

# Wait for postgres to be ready
sleep 5

# Run migrations
export DATABASE_URL="postgres://fleettracker:password123@localhost:5432/fleettracker?sslmode=disable"
migrate -path migrations -database "$DATABASE_URL" up

# Run tests with race detection
go test -v -race -coverprofile=coverage.out ./internal/...

# Check coverage threshold
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE < 75" | bc -l) )); then
    echo "❌ Coverage below threshold"
    exit 1
fi
```

## 🐛 Troubleshooting

### Database Connection Issues

**Problem:** Tests fail with "connection refused"

**Solution:**
```bash
# Check if postgres is running
docker ps | grep postgres

# Restart postgres
docker-compose restart postgres

# Check logs
docker logs fleettracker-postgres
```

### Permission Issues

**Problem:** "no pg_hba.conf entry for host"

**Solution:**
```bash
# Use Docker network or host.docker.internal
export DATABASE_URL="postgres://fleettracker:password123@host.docker.internal:5432/fleettracker?sslmode=disable"
```

### Coverage Not Generated

**Problem:** Coverage files not created

**Solution:**
```bash
# Ensure coverage directory exists
mkdir -p coverage

# Run with explicit output
go test -coverprofile=coverage/coverage.out ./internal/...
```

## 📚 Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [GORM Testing Guide](https://gorm.io/docs/testing.html)
- [Indonesian Tax Regulations (PPN)](https://www.pajak.go.id/)

## 🤝 Contributing

When adding new features:

1. ✅ Write tests first (TDD)
2. ✅ Use real database (no mocks)
3. ✅ Validate Indonesian compliance
4. ✅ Maintain 80%+ coverage
5. ✅ Run full test suite before PR
6. ✅ Update test documentation

## 📝 Test Examples

### Example: Service Test

```go
func TestService_CreateVehicle(t *testing.T) {
    db, cleanup := testutil.SetupTestDB(t)
    defer cleanup()
    
    service := NewService(db)
    company := testutil.NewTestCompany()
    require.NoError(t, db.Create(company).Error)
    
    vehicle, err := service.CreateVehicle(company.ID, CreateVehicleRequest{
        Make:         "Toyota",
        Model:        "Avanza",
        Year:         2023,
        LicensePlate: "B 1234 ABC",
        STNKNumber:   "STNK123456789",
        BPKBNumber:   "BPKB123456789",
    })
    
    assert.NoError(t, err)
    assert.NotNil(t, vehicle)
    testutil.AssertValidLicensePlate(t, vehicle.LicensePlate)
}
```

### Example: Integration Test

```go
func TestHandler_Login(t *testing.T) {
    db, cleanup := testutil.SetupTestDB(t)
    defer cleanup()
    
    service := NewService(db, "test-secret")
    handler := NewHandler(service)
    router := setupTestRouter()
    router.POST("/auth/login", handler.Login)
    
    payload := map[string]interface{}{
        "email":    "test@example.com",
        "password": "SecurePass123!",
    }
    
    jsonData, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

---

**Happy Testing! 🎉**

For questions or issues, please open a GitHub issue or contact the development team.

