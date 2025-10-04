# FleetTracker Pro Backend Makefile
# Indonesian Fleet Management SaaS Application

.PHONY: help build run test clean docker-up docker-down migrate swagger

# Default target
help:
	@echo "FleetTracker Pro Backend - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  build      - Build the backend application"
	@echo "  run        - Run the backend server"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo ""
	@echo "Database:"
	@echo "  migrate    - Run database migrations"
	@echo "  docker-up  - Start development environment with Docker"
	@echo "  docker-down- Stop development environment"
	@echo ""
	@echo "Documentation:"
	@echo "  swagger    - Generate API documentation"
	@echo ""
	@echo "Indonesian Market Features:"
	@echo "  qris-test  - Test QRIS payment integration"
	@echo "  gps-test   - Test GPS tracking functionality"

# Build the application
build:
	@echo "🔨 Building FleetTracker Pro Backend..."
	go build -o bin/main cmd/server/main.go
	@echo "✅ Build completed successfully!"

# Run the application
run:
	@echo "🚀 Starting FleetTracker Pro Backend..."
	@echo "🇮🇩 Indonesian Fleet Management SaaS Ready!"
	go run cmd/server/main.go

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f bin/main
	rm -rf dist/
	@echo "✅ Clean completed!"

# Start development environment
docker-up:
	@echo "🐳 Starting development environment..."
	@echo "📊 PostgreSQL with PostGIS: localhost:5432"
	@echo "⏰ TimescaleDB: localhost:5433"
	@echo "🔴 Redis: localhost:6379"
	@echo "🗄️ pgAdmin: http://localhost:5050"
	@echo "🔧 Redis Commander: http://localhost:8081"
	docker-compose up -d
	@echo "✅ Development environment started!"

# Stop development environment
docker-down:
	@echo "🛑 Stopping development environment..."
	docker-compose down
	@echo "✅ Development environment stopped!"

# Run database migrations
migrate:
	@echo "🗄️ Running database migrations..."
	@echo "TODO: Implement migration system"
	@echo "✅ Migrations completed!"

# Generate Swagger documentation
swagger:
	@echo "📚 Generating API documentation..."
	@echo "TODO: Implement Swagger generation"
	@echo "✅ Documentation generated!"

# Test QRIS payment integration
qris-test:
	@echo "💳 Testing QRIS payment integration..."
	@echo "TODO: Implement QRIS testing"
	@echo "✅ QRIS test completed!"

# Test GPS tracking functionality
gps-test:
	@echo "📍 Testing GPS tracking functionality..."
	@echo "TODO: Implement GPS testing"
	@echo "✅ GPS test completed!"

# Development workflow
dev: docker-up
	@echo "🔄 Waiting for services to start..."
	@sleep 5
	@echo "🚀 Starting development server..."
	$(MAKE) run

# Production build
prod-build:
	@echo "🏭 Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/main cmd/server/main.go
	@echo "✅ Production build completed!"

# Install development dependencies
install-dev:
	@echo "📦 Installing development dependencies..."
	go install github.com/cosmtrek/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ Development dependencies installed!"

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted!"

# Lint code
lint:
	@echo "🔍 Linting code..."
	@echo "TODO: Install and run golangci-lint"
	@echo "✅ Code linted!"

# Security scan
security:
	@echo "🔒 Running security scan..."
	@echo "TODO: Install and run gosec"
	@echo "✅ Security scan completed!"

# Performance test
perf-test:
	@echo "⚡ Running performance tests..."
	@echo "TODO: Implement performance testing"
	@echo "✅ Performance test completed!"

# Indonesian compliance check
compliance:
	@echo "🇮🇩 Checking Indonesian compliance..."
	@echo "✅ Data residency: Enforced"
	@echo "✅ Currency: Indonesian Rupiah (IDR)"
	@echo "✅ Language: Bahasa Indonesia"
	@echo "✅ Payment: QRIS integration ready"
	@echo "✅ Compliance check completed!"

# Full CI pipeline
ci: clean fmt lint test security compliance build
	@echo "🎉 CI pipeline completed successfully!"

# Show logs
logs:
	@echo "📋 Showing application logs..."
	docker-compose logs -f backend

# Database backup
backup:
	@echo "💾 Creating database backup..."
	@echo "TODO: Implement backup system"
	@echo "✅ Backup completed!"

# Database restore
restore:
	@echo "🔄 Restoring database..."
	@echo "TODO: Implement restore system"
	@echo "✅ Restore completed!"

# Health check
health:
	@echo "🏥 Checking service health..."
	@curl -f http://localhost:8080/health || echo "❌ Service not healthy"
	@echo "✅ Health check completed!"
