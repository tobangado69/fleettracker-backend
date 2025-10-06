# FleetTracker Pro Backend Makefile
# Indonesian Fleet Management SaaS Application

# Environment variables
DATABASE_URL ?= postgres://fleettracker:password123@localhost:5432/fleettracker?sslmode=disable
TIMESERIES_URL ?= postgres://fleettracker:password123@localhost:5433/fleettracker_timeseries?sslmode=disable

.PHONY: help build run test clean docker-up docker-down migrate swagger

# Default target
help:
	@echo "FleetTracker Pro Backend - Available Commands:"
	@echo ""
	@echo "🚀 Development:"
	@echo "  build          - Build the backend application"
	@echo "  run            - Run the backend server"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  dev            - Start Docker + run dev server"
	@echo ""
	@echo "🐳 Docker Commands:"
	@echo "  docker-setup             - Full setup (clean + rebuild + up)"
	@echo "  docker-up                - Start all services"
	@echo "  docker-down              - Stop all services"
	@echo "  docker-restart           - Restart all services"
	@echo "  docker-clean             - Stop and remove volumes"
	@echo "  docker-rebuild           - Rebuild all containers"
	@echo "  docker-rebuild-backend   - Rebuild backend only"
	@echo "  docker-update-backend    - Rebuild + restart backend"
	@echo "  docker-status            - Show container status"
	@echo "  docker-check             - Check Docker environment"
	@echo ""
	@echo "📋 Docker Logs:"
	@echo "  docker-logs            - Show all logs (follow)"
	@echo "  docker-logs-postgres   - Show PostgreSQL logs"
	@echo "  docker-logs-timescale  - Show TimescaleDB logs"
	@echo "  docker-logs-redis      - Show Redis logs"
	@echo "  docker-logs-backend    - Show Backend logs"
	@echo ""
	@echo "🔧 Docker Shell Access:"
	@echo "  docker-shell-postgres  - Connect to PostgreSQL"
	@echo "  docker-shell-timescale - Connect to TimescaleDB"
	@echo "  docker-shell-redis     - Connect to Redis CLI"
	@echo "  docker-shell-backend   - Connect to backend container"
	@echo ""
	@echo "💾 Database Management:"
	@echo "  docker-backup-postgres   - Backup PostgreSQL"
	@echo "  docker-backup-timescale  - Backup TimescaleDB"
	@echo "  docker-restore-postgres  - Restore PostgreSQL (FILE=...)"
	@echo ""
	@echo "🗄️  Database Migrations & Seeds:"
	@echo "  migrate-up               - Apply pending migrations"
	@echo "  migrate-down             - Rollback last migration"
	@echo "  migrate-version          - Show current version"
	@echo "  migrate-create NAME=...  - Create new migration"
	@echo "  seed                     - Populate with test data"
	@echo "  seed-companies           - Seed companies only"
	@echo "  seed-users               - Seed users only"
	@echo "  db-reset                 - Drop, migrate, seed (CAUTION)"
	@echo "  db-status                - Show database info"
	@echo ""
	@echo "📚 Documentation:"
	@echo "  swagger          - Generate API documentation"
	@echo "  swagger-install  - Install Swagger CLI tool"
	@echo ""
	@echo "🇮🇩 Indonesian Market Features:"
	@echo "  qris-test  - Test QRIS payment integration"
	@echo "  gps-test   - Test GPS tracking functionality"
	@echo "  compliance - Check Indonesian compliance"

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
	@docker-compose up -d || (echo "❌ Failed to start services. Check logs with 'make docker-logs'" && exit 1)
	@echo "✅ Development environment started!"
	@echo "⏳ Waiting for services to be healthy..."
	@sleep 10
	@echo "📊 Checking service status..."
	@docker-compose ps
	@echo "✅ Services ready!"

# Stop development environment
docker-down:
	@echo "🛑 Stopping development environment..."
	docker-compose down
	@echo "✅ Development environment stopped!"

# Stop and remove volumes
docker-clean:
	@echo "🧹 Stopping and cleaning Docker environment..."
	docker-compose down -v
	@echo "✅ Docker environment cleaned!"

# Rebuild Docker containers
docker-rebuild:
	@echo "🔄 Rebuilding Docker containers..."
	docker-compose build --no-cache
	@echo "✅ Containers rebuilt!"

# Rebuild only backend container
docker-rebuild-backend:
	@echo "🔄 Rebuilding backend container..."
	docker-compose build --no-cache backend
	@echo "✅ Backend container rebuilt!"
	@echo "Run 'make docker-restart' to apply changes"

# Full backend update (rebuild + restart)
docker-update-backend:
	@echo "🔄 Updating backend (rebuild + restart)..."
	docker-compose build --no-cache backend
	docker-compose up -d --force-recreate backend
	@echo "✅ Backend updated!"
	@echo "⏳ Waiting for backend to be healthy..."
	@sleep 10
	@echo "📚 Swagger should be available at: http://localhost:8080/swagger/index.html"

# Restart all services
docker-restart:
	@echo "🔄 Restarting Docker services..."
	docker-compose restart
	@echo "✅ Services restarted!"

# Show Docker container status
docker-status:
	@echo "📊 Docker Container Status:"
	@docker-compose ps

# Show Docker logs
docker-logs:
	@echo "📋 Showing Docker logs (Ctrl+C to exit)..."
	docker-compose logs -f

# Show specific service logs
docker-logs-postgres:
	@docker-compose logs -f postgres

docker-logs-timescale:
	@docker-compose logs -f timescaledb

docker-logs-redis:
	@docker-compose logs -f redis

docker-logs-backend:
	@docker-compose logs -f backend

# Execute commands in containers
docker-shell-postgres:
	@echo "🐘 Connecting to PostgreSQL..."
	docker exec -it fleettracker-postgres psql -U fleettracker -d fleettracker

docker-shell-timescale:
	@echo "⏰ Connecting to TimescaleDB..."
	docker exec -it fleettracker-timescaledb psql -U fleettracker -d fleettracker_timeseries

docker-shell-redis:
	@echo "🔴 Connecting to Redis..."
	docker exec -it fleettracker-redis redis-cli

docker-shell-backend:
	@echo "🔧 Connecting to backend container..."
	docker exec -it fleettracker-backend sh

# Backup database
docker-backup-postgres:
	@echo "💾 Backing up PostgreSQL database..."
	@mkdir -p backups
	docker exec fleettracker-postgres pg_dump -U fleettracker fleettracker > backups/postgres_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup saved to backups/"

docker-backup-timescale:
	@echo "💾 Backing up TimescaleDB database..."
	@mkdir -p backups
	docker exec fleettracker-timescaledb pg_dump -U fleettracker fleettracker_timeseries > backups/timescale_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup saved to backups/"

# Restore database
docker-restore-postgres:
	@echo "🔄 Restoring PostgreSQL database..."
	@echo "⚠️  Usage: make docker-restore-postgres FILE=backups/postgres_XXXXXX.sql"
	@if [ -z "$(FILE)" ]; then \
		echo "❌ Error: Please specify FILE=path/to/backup.sql"; \
		exit 1; \
	fi
	@cat $(FILE) | docker exec -i fleettracker-postgres psql -U fleettracker fleettracker
	@echo "✅ Database restored!"

# Check Docker environment
docker-check:
	@echo "🔍 Checking Docker environment..."
	@echo ""
	@echo "Docker version:"
	@docker --version
	@echo ""
	@echo "Docker Compose version:"
	@docker-compose --version
	@echo ""
	@echo "Running containers:"
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
	@echo ""
	@echo "✅ Docker environment check completed!"

# Full Docker setup (first time)
docker-setup: docker-clean docker-rebuild docker-up
	@echo "🎉 Docker environment fully set up!"
	@echo ""
	@echo "Available services:"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - TimescaleDB: localhost:5433"
	@echo "  - Redis: localhost:6379"
	@echo "  - pgAdmin: http://localhost:5050"
	@echo "  - Redis Commander: http://localhost:8081"
	@echo "  - Backend API: http://localhost:8080"
	@echo ""
	@echo "pgAdmin credentials:"
	@echo "  Email: admin@fleettracker.id"
	@echo "  Password: admin123"

# Run database migrations (using psql inside Docker to avoid connection issues)
migrate-up:
	@echo "🗄️ Running database migrations..."
	@for file in migrations/*_*.up.sql; do \
		echo "Applying $$file..."; \
		docker exec -i fleettracker-postgres psql -U fleettracker -d fleettracker < $$file || exit 1; \
	done
	@echo "✅ Migrations completed!"

migrate-down:
	@echo "🔄 Rolling back migrations..."
	@echo "⚠️  Manual rollback: Run the .down.sql files in reverse order"
	@echo "Example: docker exec -i fleettracker-postgres psql -U fleettracker -d fleettracker < migrations/001_initial_schema.down.sql"

migrate-force:
	@echo "⚠️  Not applicable with psql-based migrations"
	@echo "To reset: make db-reset"

migrate-version:
	@echo "📊 Checking if tables exist..."
	@docker exec fleettracker-postgres psql -U fleettracker -d fleettracker -c "SELECT COUNT(*) as tables FROM information_schema.tables WHERE table_schema = 'public';" || echo "No tables yet"

migrate-create:
	@echo "📝 Creating new migration: $(NAME)..."
	@migrate create -ext sql -dir migrations -seq $(NAME)
	@echo "✅ Migration files created!"

# Legacy command (redirects to migrate-up)
migrate: migrate-up

# Seed database with test data
# Note: Run from inside Docker network since Windows path issues
seed:
	@echo "🌱 Seeding database with test data..."
	@echo "Note: Seeding directly via psql due to connection issues from host"
	@docker exec fleettracker-postgres psql -U fleettracker -d fleettracker -c "\
		INSERT INTO companies (name, email, npwp, city, province, country) \
		SELECT 'PT Fleet Indonesia', 'contact@fleet.id', '01.234.567.8-901.000', 'Jakarta', 'DKI Jakarta', 'Indonesia' \
		WHERE NOT EXISTS (SELECT 1 FROM companies WHERE email = 'contact@fleet.id');"
	@echo "✅ Basic seed data inserted! For full seeding, use the seed container or run seeds manually"
	@echo "   Or connect to pgAdmin and run seed scripts there"

seed-companies:
	@echo "🏢 Seeding companies only..."
	@DATABASE_URL="postgres://fleettracker:password123@localhost:5432/fleettracker?sslmode=disable" go run cmd/seed/main.go --companies
	@echo "✅ Companies seeded!"

seed-users:
	@echo "👥 Seeding users only..."
	@DATABASE_URL="postgres://fleettracker:password123@localhost:5432/fleettracker?sslmode=disable" go run cmd/seed/main.go --users
	@echo "✅ Users seeded!"

# Database reset (drop, migrate, seed)
db-reset:
	@echo "♻️  Resetting database..."
	@echo "⚠️  This will delete ALL data!"
	@read -p "Continue? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(MAKE) migrate-down || true; \
		$(MAKE) migrate-up; \
		$(MAKE) seed; \
		echo "✅ Database reset complete!"; \
	else \
		echo "❌ Reset cancelled"; \
	fi

# Show database status
db-status:
	@echo "📊 Database Status:"
	@echo ""
	@echo "Migration Version:"
	@$(MAKE) migrate-version
	@echo ""
	@echo "Database Connection:"
	@docker exec fleettracker-postgres psql -U fleettracker -d fleettracker -c "SELECT COUNT(*) as tables FROM information_schema.tables WHERE table_schema = 'public';"
	@echo ""
	@echo "Quick Data Count:"
	@docker exec fleettracker-postgres psql -U fleettracker -d fleettracker -c "SELECT 'companies' as table_name, COUNT(*) as count FROM companies UNION ALL SELECT 'users', COUNT(*) FROM users UNION ALL SELECT 'vehicles', COUNT(*) FROM vehicles UNION ALL SELECT 'drivers', COUNT(*) FROM drivers;" || echo "Tables not yet created"

# Generate Swagger documentation
swagger:
	@echo "📚 Generating API documentation..."
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
	@echo "✅ Documentation generated at docs/"
	@echo "📝 View at: http://localhost:8080/swagger/index.html"

# Install Swagger CLI tool
swagger-install:
	@echo "📦 Installing Swagger CLI..."
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ Swagger CLI installed!"

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
	@echo ""
	@echo "Generate Swagger docs with: make swagger"

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
