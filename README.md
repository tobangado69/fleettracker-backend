# 🚛 FleetTracker Pro - Backend

**Indonesian Fleet Management SaaS - Go Backend API**

[![Go Version](https://img.shields.io/badge/Go-1.24.0-blue.svg)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-18-blue.svg)](https://postgresql.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-Framework-green.svg)](https://gin-gonic.com/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 📋 Overview

FleetTracker Pro Backend is a high-performance Go API service designed for Indonesian fleet management companies. It provides real-time mobile GPS tracking, driver behavior monitoring, and comprehensive fleet analytics with Indonesian market compliance.

### 🎯 Key Features

- **📱 Mobile GPS Tracking**: Smartphone-based GPS tracking (no dedicated hardware needed)
- **🇮🇩 Indonesian Compliance**: NPWP, SIUP, NIK, SIM validation and Indonesian Rupiah (IDR) support
- **🔐 JWT Authentication**: Secure authentication with role-based access control
- **💳 Payment Integration**: QRIS, Indonesian banks (BCA, Mandiri, BNI, BRI), and e-wallets
- **📊 Real-time Analytics**: Driver behavior monitoring and fuel consumption tracking
- **🐳 Docker Ready**: Complete containerized development environment

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Mobile Apps   │    │   Backend API   │    │   PostgreSQL    │
│                 │    │                 │    │                 │
│ GPS Tracking    │◄──►│ Go + Gin        │◄──►│ Mobile GPS      │
│ Driver App      │    │ JWT Auth        │    │ Data Storage    │
│ Fleet Manager   │    │ WebSocket       │    │ Indonesian      │
│                 │    │ REST APIs       │    │ Compliance      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 🛠️ Technology Stack

- **Backend**: Go 1.24.0 with Gin Framework
- **Database**: PostgreSQL 18 (optimized for mobile GPS data)
- **Authentication**: JWT with Better Auth compatibility
- **Real-time**: WebSocket for live GPS updates
- **Caching**: Redis for session management
- **Documentation**: Swagger/OpenAPI 3.0

## 🚀 Quick Start

### Prerequisites

- Go 1.24.0 or higher
- PostgreSQL 18 or higher
- Redis (for caching)
- Docker & Docker Compose (optional)

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/tobangado69/fleettracker-backend.git
   cd fleettracker-backend
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Set up environment variables**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Start development environment**:
   ```bash
   # Using Docker Compose (recommended)
   docker-compose up -d postgres redis
   
   # Or start manually
   make docker-up
   ```

5. **Run database migrations**:
   ```bash
   make migrate-up
   ```

6. **Start the server**:
   ```bash
   make run
   # Or: go run cmd/server/main.go
   ```

The API will be available at `http://localhost:8080`

## 📁 Project Structure

```
backend/
├── cmd/
│   └── server/           # Application entry point
│       └── main.go
├── internal/             # Private application code
│   ├── auth/            # Authentication service
│   │   ├── handler.go   # HTTP handlers
│   │   └── service.go   # Business logic
│   ├── vehicle/         # Vehicle management
│   ├── driver/          # Driver management
│   ├── tracking/        # Mobile GPS tracking
│   ├── payment/         # Indonesian payment integration
│   └── common/          # Shared utilities
│       ├── config/      # Configuration management
│       ├── database/    # Database connections
│       └── middleware/  # HTTP middleware
├── pkg/
│   └── models/          # GORM models
│       ├── company.go   # Company entity
│       ├── user.go      # User entity
│       ├── vehicle.go   # Vehicle entity
│       ├── driver.go    # Driver entity
│       ├── tracking.go  # GPS tracking entities
│       └── payment.go   # Payment entities
├── docs/                # API documentation
├── migrations/          # Database migrations
├── docker-compose.yml   # Development environment
├── Dockerfile          # Container configuration
├── Makefile           # Development commands
└── README.md          # This file
```

## 🔧 Development Commands

```bash
# Start development environment
make docker-up

# Stop development environment
make docker-down

# Run the server
make run

# Build for production
make build

# Run tests
make test

# View logs
make logs

# Check health
make health

# Generate Swagger docs
make swagger

# Database migrations
make migrate-up      # Apply migrations
make migrate-down    # Rollback migrations
make migrate-create  # Create new migration
```

## 🌐 API Endpoints

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `GET /api/v1/auth/profile` - Get user profile

### Vehicle Management
- `GET /api/v1/vehicles` - List vehicles
- `POST /api/v1/vehicles` - Create vehicle
- `GET /api/v1/vehicles/:id` - Get vehicle details
- `PUT /api/v1/vehicles/:id` - Update vehicle
- `DELETE /api/v1/vehicles/:id` - Delete vehicle

### Driver Management
- `GET /api/v1/drivers` - List drivers
- `POST /api/v1/drivers` - Create driver
- `GET /api/v1/drivers/:id` - Get driver details
- `PUT /api/v1/drivers/:id` - Update driver
- `GET /api/v1/drivers/:id/performance` - Driver performance

### Mobile GPS Tracking
- `POST /api/v1/tracking/gps` - Submit GPS data
- `GET /api/v1/tracking/vehicles/:id/history` - GPS history
- `GET /api/v1/tracking/vehicles/:id/current` - Current location
- `WebSocket /ws/tracking` - Real-time GPS updates

### Payment Integration
- `POST /api/v1/payments/qris` - Create QRIS payment
- `POST /api/v1/payments/bank` - Bank transfer
- `POST /api/v1/payments/ewallet` - E-wallet payment
- `GET /api/v1/payments/:id/status` - Payment status

## 📱 Mobile GPS Integration

### GPS Data Format
```json
{
  "vehicle_id": "uuid",
  "driver_id": "uuid",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "speed": 45.5,
  "heading": 180,
  "altitude": 10.0,
  "accuracy": 5.0,
  "battery_level": 85,
  "timestamp": "2025-01-04T10:30:00Z"
}
```

### WebSocket Events
```json
{
  "type": "gps_update",
  "vehicle_id": "uuid",
  "data": {
    "latitude": -6.2088,
    "longitude": 106.8456,
    "speed": 45.5,
    "timestamp": "2025-01-04T10:30:00Z"
  }
}
```

## 🇮🇩 Indonesian Market Features

### Compliance Fields
- **NPWP**: Indonesian Tax ID validation
- **SIUP**: Indonesian Business License
- **NIK**: Indonesian ID Number (16-digit validation)
- **SIM**: Indonesian Driver's License format validation

### Payment Integration
- **QRIS**: Indonesian standardized QR payment
- **Bank Transfers**: BCA, Mandiri, BNI, BRI
- **E-Wallets**: GoPay, OVO, DANA, ShopeePay
- **Currency**: Indonesian Rupiah (IDR) support

### Data Residency
- All data stored within Indonesia
- Indonesian cloud provider deployment
- Compliance with Indonesian data protection laws

## 🔐 Security Features

- JWT-based authentication
- Role-based access control (RBAC)
- Rate limiting (100 requests/minute)
- Input validation and sanitization
- SQL injection prevention
- XSS protection
- CORS configuration for Indonesian domains

## 📊 Database Schema

### Core Entities
- **Companies**: Multi-tenant company management
- **Users**: Authentication and user profiles
- **Vehicles**: Fleet vehicle information
- **Drivers**: Driver profiles and performance
- **GPSTracks**: Mobile GPS tracking data
- **Trips**: Journey tracking and analytics
- **Payments**: Indonesian payment processing

### Indonesian Compliance
- NPWP validation and storage
- SIUP business license management
- Indonesian driver's license (SIM) tracking
- Indonesian vehicle registration (STNK, BPKB)
- Indonesian Rupiah currency support

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -cover ./...

# Run specific package tests
go test ./internal/auth/...

# Integration tests
go test -tags=integration ./...
```

## 🐳 Docker Deployment

### Development
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f backend

# Stop services
docker-compose down
```

### Production
```bash
# Build production image
docker build -t fleettracker-backend .

# Run with environment variables
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://..." \
  -e JWT_SECRET="your-secret" \
  fleettracker-backend
```

## 📈 Performance

- **API Response Time**: <200ms (95th percentile)
- **GPS Data Processing**: <30 seconds
- **Concurrent Users**: 1000+ simultaneous connections
- **Database Queries**: Optimized with proper indexing
- **Memory Usage**: <512MB typical

## 🔧 Configuration

### Environment Variables

```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/fleettracker?sslmode=disable
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRY=24h

# Server
PORT=8080
FRONTEND_URL=http://localhost:5173

# Indonesian Payment APIs
QRIS_API_URL=https://api.qris.id
QRIS_API_KEY=your-qris-api-key

# External APIs
GOOGLE_MAPS_API_KEY=your-google-maps-key
WHATSAPP_API_URL=https://api.whatsapp.com
```

## 📚 API Documentation

Once the server is running, visit:
- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **OpenAPI Spec**: `http://localhost:8080/swagger/doc.json`

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Write comprehensive tests for new features
- Update documentation for API changes
- Ensure Indonesian compliance features are maintained
- Test with mobile GPS data scenarios

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [FleetTracker Pro Docs](https://github.com/tobangado69/fleettracker-docs)
- **Issues**: [GitHub Issues](https://github.com/tobangado69/fleettracker-backend/issues)
- **Discussions**: [GitHub Discussions](https://github.com/tobangado69/fleettracker-pro/discussions)

## 🎯 Roadmap

### Phase 1: Core Backend ✅
- [x] Go backend infrastructure
- [x] Mobile GPS tracking
- [x] Indonesian compliance
- [x] JWT authentication

### Phase 2: Business Logic 🚧
- [ ] Vehicle management APIs
- [ ] Driver management APIs
- [ ] Mobile GPS data ingestion
- [ ] Payment integration (QRIS)

### Phase 3: Advanced Features 📋
- [ ] Real-time analytics
- [ ] Driver behavior monitoring
- [ ] Fuel consumption tracking
- [ ] Route optimization

---

**Built with ❤️ for Indonesian Fleet Management Companies**

*FleetTracker Pro - Making fleet management simple, efficient, and compliant with Indonesian regulations.*
