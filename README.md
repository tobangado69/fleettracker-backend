# FleetTracker Pro - Backend

**Indonesian Fleet Management SaaS Platform**

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- Docker & Docker Compose
- golang-migrate CLI

### Installation

1. **Install golang-migrate:**
   ```bash
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```

2. **Start services:**
   ```bash
   make docker-up
   ```

3. **Run migrations:**
   ```bash
   make migrate-up
   ```
   
   *(Note: Seed data requires model fixes - migrations are ready to use!)*

4. **Start backend:**
   ```bash
   make run
   ```

5. **Open Swagger UI:**
   ```
   http://localhost:8080/swagger/index.html
   ```

### Test Credentials
```
Email:    admin@logistikjkt.co.id
Password: password123
```

---

## 📋 Available Commands

### Docker
```bash
make docker-up              # Start all services
make docker-down            # Stop services
make docker-restart         # Restart services
make docker-status          # Check health
make docker-logs            # View all logs
make docker-logs-backend    # Backend logs only
make docker-shell-postgres  # Connect to PostgreSQL
```

### Database
```bash
make migrate-up             # Apply migrations
make migrate-down           # Rollback migration
make migrate-version        # Current version
make seed                   # Populate test data
make db-reset               # Drop + migrate + seed ⚠️
make db-status              # Show database info
```

### Development
```bash
make build                  # Build binary
make run                    # Start server
make swagger                # Generate API docs
make help                   # Show all commands
```

---

## 🗄️ Database Migrations

### Create New Migration
```bash
make migrate-create NAME=add_feature
```

This creates:
- `migrations/XXX_add_feature.up.sql`
- `migrations/XXX_add_feature.down.sql`

### Migration Best Practices
- Always include both `.up.sql` and `.down.sql`
- Use `IF EXISTS` / `IF NOT EXISTS` for idempotency
- Test rollback before committing
- Never edit applied migrations

### Troubleshooting

**"Dirty database"**
```bash
make migrate-force VERSION=1
make migrate-up
```

**Connection refused**
```bash
make docker-up
# Wait 30 seconds
make migrate-up
```

---

## 🌱 Seed Data

### What's Included
- 2 Indonesian companies (Jakarta & Surabaya)
- 5 users (1 admin, 2 managers, 2 operators)
- 10 vehicles with Indonesian license plates
- 5 drivers with valid SIM
- 100+ GPS tracking points (real routes)
- 20 completed trips with fuel data

### Seed Commands
```bash
make seed               # All data
make seed-companies     # Companies only
make seed-users         # Users only
```

### User Accounts
| Email | Password | Role | Company |
|-------|----------|------|---------|
| admin@logistikjkt.co.id | password123 | admin | Jakarta |
| manager.jakarta@logistikjkt.co.id | password123 | manager | Jakarta |
| operator.jakarta@logistikjkt.co.id | password123 | operator | Jakarta |
| manager.surabaya@transportsby.co.id | password123 | manager | Surabaya |
| operator.surabaya@transportsby.co.id | password123 | operator | Surabaya |

---

## 🐳 Docker Services

| Service | Port | Credentials |
|---------|------|-------------|
| Backend API | 8080 | - |
| PostgreSQL | 5432 | fleettracker / password123 |
| TimescaleDB | 5433 | fleettracker / password123 |
| Redis | 6379 | - |
| pgAdmin | 5050 | admin@fleettracker.id / admin123 |
| Redis Commander | 8081 | - |

### Service Health
```bash
make docker-status
```

### View Logs
```bash
make docker-logs              # All services
make docker-logs-backend      # Backend only
make docker-logs-postgres     # PostgreSQL only
```

### Connect to Database
```bash
make docker-shell-postgres
# Inside PostgreSQL shell:
\dt                          # List tables
SELECT * FROM companies;     # Query data
```

---

## 📚 API Documentation

### Swagger UI
```
http://localhost:8080/swagger/index.html
```

### Regenerate Docs
```bash
make swagger
```

### Key Endpoints
- **Auth:** `POST /api/v1/auth/login`
- **Vehicles:** `GET /api/v1/vehicles`
- **Drivers:** `GET /api/v1/drivers`
- **GPS Tracking:** `GET /api/v1/tracking/vehicles/:id/history`
- **Trips:** `GET /api/v1/trips`
- **Analytics:** `GET /api/v1/analytics/dashboard`

---

## 🔧 Troubleshooting

### Swagger Not Loading
```bash
make swagger                 # Regenerate docs
make docker-rebuild-backend  # Rebuild container
make docker-restart          # Restart services
```

### TimescaleDB Won't Start
```bash
make docker-down
docker volume rm backend_timescale_data
make docker-up
```

### Migration Fails
```bash
make docker-logs-postgres    # Check logs
make migrate-version         # Check current version
make migrate-force VERSION=0 # Reset if needed
make migrate-up              # Try again
```

### Seed Data Errors
```bash
make db-reset               # Fresh start (deletes all data!)
```

---

## 📁 Project Structure

```
backend/
├── cmd/
│   ├── server/main.go      # Main application entry
│   └── seed/main.go        # Database seeder CLI
├── internal/
│   ├── analytics/          # Analytics & reporting
│   ├── auth/               # Authentication
│   ├── driver/             # Driver management
│   ├── payment/            # Payment processing
│   ├── tracking/           # GPS tracking
│   ├── vehicle/            # Vehicle management
│   └── common/
│       ├── config/         # Configuration
│       ├── database/       # Database connection
│       ├── middleware/     # HTTP middleware
│       └── repository/     # Data repositories
├── pkg/
│   └── models/             # Data models
├── migrations/             # SQL migrations
│   ├── 001_initial_schema.up.sql
│   └── 001_initial_schema.down.sql
├── seeds/                  # Seed data generators
│   ├── seed.go
│   ├── companies.go
│   ├── users.go
│   ├── vehicles.go
│   ├── drivers.go
│   ├── gps_tracks.go
│   └── trips.go
├── docs/                   # Swagger documentation
├── Makefile                # Build & run commands
└── docker-compose.yml      # Docker services
```

---

## 🏗️ Development Workflow

### Daily Development
```bash
make docker-up              # Start services
make run                    # Start backend
# Develop features...
make swagger                # Update API docs
```

### Adding New Features
```bash
# 1. Create migration
make migrate-create NAME=add_alerts

# 2. Edit migration files
vim migrations/XXX_add_alerts.up.sql
vim migrations/XXX_add_alerts.down.sql

# 3. Apply migration
make migrate-up

# 4. Add seed data (optional)
vim seeds/alerts.go

# 5. Test
make db-reset
make run
```

### Testing Changes
```bash
make db-reset               # Fresh database
make run                    # Start backend
# Test via Swagger UI
```

---

## 🇮🇩 Indonesian Compliance

All seed data uses authentic Indonesian formats:

- **NPWP:** `XX.XXX.XXX.X-XXX.XXX` (Tax ID)
- **NIK:** 16-digit National ID
- **SIM:** Driver's License format
- **License Plates:** B (Jakarta), L (Surabaya)
- **Real GPS Routes:**
  - Jakarta: Monas → Blok M
  - Surabaya: Tugu Pahlawan → Delta Plaza

---

## 🔐 Environment Variables

Create `.env` file (optional, defaults provided):

```env
DATABASE_URL=postgres://fleettracker:password123@localhost:5432/fleettracker?sslmode=disable
TIMESERIES_URL=postgres://fleettracker:password123@localhost:5433/fleettracker_timeseries?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-super-secret-jwt-key-for-development
PORT=8080
ENVIRONMENT=development
DEBUG=true
```

---

## 📊 Database Schema

18 tables with Indonesian compliance fields:

- **Core:** companies, users, sessions, audit_logs
- **Fleet:** vehicles, maintenance_logs, fuel_logs
- **Drivers:** drivers, driver_events, performance_logs
- **Tracking:** gps_tracks, trips, geofences
- **Billing:** subscriptions, payments, invoices
- **History:** vehicle_history, password_reset_tokens

---

## 🧪 Testing

### API Testing
1. Start backend: `make run`
2. Open Swagger: http://localhost:8080/swagger/index.html
3. Login with test credentials
4. Test endpoints with real data

### Manual Testing
```bash
# Test login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@logistikjkt.co.id","password":"password123"}'

# Get vehicles
curl http://localhost:8080/api/v1/vehicles \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 🚨 Common Issues

### Port Already in Use
```bash
# Change port in docker-compose.yml or stop conflicting service
lsof -ti:8080 | xargs kill -9
```

### Out of Memory
```bash
# Increase Docker memory limit
# Docker Desktop > Settings > Resources > Memory
```

### Slow Performance
```bash
# Check Docker resources
docker stats

# Restart services
make docker-restart
```

---

## 📖 Additional Resources

- **Swagger API Docs:** http://localhost:8080/swagger/index.html
- **pgAdmin:** http://localhost:5050
- **Redis Commander:** http://localhost:8081
- **Migrations:** `/migrations/` directory
- **Seeds:** `/seeds/` directory

---

## 🎯 Next Steps

1. ✅ Install golang-migrate
2. ✅ Start Docker services
3. ✅ Run migrations
4. ✅ Seed database
5. ✅ Start backend
6. ✅ Test via Swagger UI
7. 🚀 Start building features!

---

## 💡 Tips

- Use `make help` to see all available commands
- Check `make db-status` to verify database state
- Run `make docker-logs` if something fails
- Use `make db-reset` for a fresh start (⚠️ deletes all data!)
- Swagger UI automatically updates when you run `make swagger`

---

## 📝 License

Private - Indonesian Fleet Management SaaS Platform

---

**Need Help?** Run `make help` or check the Makefile for all available commands.
