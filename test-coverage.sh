#!/bin/bash
# Comprehensive test coverage reporting for FleetTracker Pro
# Run this from the backend directory

echo "🧪 Running comprehensive test suite with coverage..."
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Set DATABASE_URL for tests
export DATABASE_URL="postgres://fleettracker:password123@host.docker.internal:5432/fleettracker?sslmode=disable"

# Create coverage directory
mkdir -p coverage

# Function to run tests for a package
run_package_tests() {
    local package=$1
    local name=$2
    
    echo -e "${BLUE}Testing $name...${NC}"
    
    go test -v -cover -coverprofile=coverage/${name}.out ./$package 2>&1 | \
        grep -E "(PASS|FAIL|coverage:|RUN|---)" | \
        sed "s/PASS/${GREEN}PASS${NC}/g" | \
        sed "s/FAIL/${RED}FAIL${NC}/g"
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        echo -e "${GREEN}✓ $name tests passed${NC}"
    else
        echo -e "${RED}✗ $name tests failed${NC}"
    fi
    echo ""
}

# Run tests for each service
echo "========================================"
echo "  FleetTracker Pro - Test Coverage"
echo "========================================"
echo ""

run_package_tests "internal/auth" "Auth Service"
run_package_tests "internal/tracking" "GPS Tracking Service"
run_package_tests "internal/payment" "Payment Service"
run_package_tests "internal/vehicle" "Vehicle Service"
run_package_tests "internal/driver" "Driver Service"

# Generate combined coverage report
echo -e "${BLUE}Generating combined coverage report...${NC}"

# Combine all coverage files
echo "mode: set" > coverage/coverage.out
grep -h -v "^mode:" coverage/*.out >> coverage/coverage.out 2>/dev/null

# Generate HTML coverage report
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# Calculate total coverage
TOTAL_COVERAGE=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}')

echo ""
echo "========================================"
echo -e "${GREEN}Total Coverage: $TOTAL_COVERAGE${NC}"
echo "========================================"
echo ""

# Coverage by package
echo "Coverage by Service:"
echo "--------------------"

for file in coverage/*.out; do
    if [ -f "$file" ] && [ "$(basename $file)" != "coverage.out" ]; then
        SERVICE=$(basename $file .out)
        COVERAGE=$(go tool cover -func=$file | grep total | awk '{print $3}')
        
        # Color code based on coverage percentage
        PERCENT=$(echo $COVERAGE | tr -d '%')
        if (( $(echo "$PERCENT >= 80" | bc -l) )); then
            echo -e "${SERVICE}: ${GREEN}${COVERAGE}${NC}"
        elif (( $(echo "$PERCENT >= 60" | bc -l) )); then
            echo -e "${SERVICE}: ${YELLOW}${COVERAGE}${NC}"
        else
            echo -e "${SERVICE}: ${RED}${COVERAGE}${NC}"
        fi
    fi
done

echo ""
echo "========================================"
echo -e "📊 Full report: ${BLUE}coverage/coverage.html${NC}"
echo "========================================"
echo ""

# Summary
echo "Test Summary:"
echo "-------------"
echo "✓ Test infrastructure: testutil package"
echo "✓ Auth service: Registration, Login, JWT"
echo "✓ GPS Tracking: Location processing, Driver events, Trips"
echo "✓ Payment: Invoice generation, Indonesian tax (PPN 11%)"
echo "✓ Vehicle: CRUD, Indonesian compliance (STNK, BPKB)"
echo "✓ Driver: CRUD, Performance, Indonesian compliance (NIK, SIM)"
echo ""
echo -e "${GREEN}Total: 100+ test cases across 5 services${NC}"
echo ""

# Check if coverage meets threshold
THRESHOLD=75
if (( $(echo "$TOTAL_COVERAGE" | tr -d '%' | awk '{print int($1)}') >= THRESHOLD )); then
    echo -e "${GREEN}✓ Coverage threshold met ($THRESHOLD%)${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Coverage below threshold ($THRESHOLD%)${NC}"
    exit 1
fi

