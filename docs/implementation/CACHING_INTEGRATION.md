# Redis Caching Integration

## Overview

The FleetTracker Pro backend now includes comprehensive Redis caching support across all services. This document explains how the caching system works and how to use it.

## Architecture

### Service Structure
All services now follow this pattern:
```go
type Service struct {
    db    *gorm.DB
    redis *redis.Client
    cache *CacheService
}
```

### Cache Service Pattern
Each service includes a dedicated `CacheService`:
```go
type CacheService struct {
    redis *redis.Client
}
```

## Configuration

### Environment Variables
```bash
# Redis Configuration
REDIS_URL=redis://localhost:6379

# Cache Configuration
CACHE_TTL=5m
CACHE_MAX_SIZE=1000
```

### Service Initialization
```go
// Initialize Redis client
redisClient, err := database.ConnectRedis(cfg.RedisURL)
if err != nil {
    log.Fatal("Failed to connect to Redis:", err)
}

// Initialize services with Redis support
authService := auth.NewService(db, redisClient, cfg.JWTSecret)
vehicleService := vehicle.NewService(db, redisClient)
driverService := driver.NewService(db, redisClient)
paymentService := payment.NewService(db, redisClient, cfg, repoManager)
```

## Caching Implementation Examples

### Vehicle Service Caching

#### Individual Vehicle Cache Operations
```go
// Get from cache
cachedVehicle, err := s.cache.GetVehicleFromCache(ctx, vehicleID)

// Set in cache with expiration
err := s.cache.SetVehicleInCache(ctx, vehicle, 30*time.Minute)

// Invalidate cache
err := s.cache.InvalidateVehicleCache(ctx, vehicleID)
```

#### Vehicle List Cache Operations
```go
// Get vehicle list from cache
cachedVehicles, cachedTotal, err := s.cache.GetVehicleListFromCache(ctx, companyID, filters)

// Set vehicle list in cache
err := s.cache.SetVehicleListInCache(ctx, companyID, filters, vehicles, total, 15*time.Minute)

// Invalidate all vehicle lists for a company
err := s.cache.InvalidateVehicleListCache(ctx, companyID)
```

### Driver Service Caching

#### Individual Driver Cache Operations
```go
// Get from cache
cachedDriver, err := s.cache.GetDriverFromCache(ctx, driverID)

// Set in cache with expiration
err := s.cache.SetDriverInCache(ctx, driver, 30*time.Minute)

// Invalidate cache
err := s.cache.InvalidateDriverCache(ctx, driverID)
```

#### Driver List Cache Operations
```go
// Get driver list from cache
cachedDrivers, cachedTotal, err := s.cache.GetDriverListFromCache(ctx, companyID, filters)

// Set driver list in cache
err := s.cache.SetDriverListInCache(ctx, companyID, filters, drivers, total, 15*time.Minute)

// Invalidate all driver lists for a company
err := s.cache.InvalidateDriverListCache(ctx, companyID)
```

### Auth Service Caching

#### User Cache Operations
```go
// Get user from cache
cachedUser, err := s.cache.GetUserFromCache(ctx, userID)

// Set user in cache
err := s.cache.SetUserInCache(ctx, user, 30*time.Minute)

// Invalidate user cache
err := s.cache.InvalidateUserCache(ctx, userID)
```

#### Session Cache Operations
```go
// Set session in cache
err := s.cache.SetSessionInCache(ctx, sessionID, userID, 15*time.Minute)

// Get session from cache
userID, err := s.cache.GetSessionFromCache(ctx, sessionID)

// Invalidate session cache
err := s.cache.InvalidateSessionCache(ctx, sessionID)
```

### Payment Service Caching

#### Individual Invoice Cache Operations
```go
// Get invoice from cache by ID
cachedInvoice, err := s.cache.GetInvoiceFromCache(ctx, invoiceID)

// Get invoice from cache by invoice number
cachedInvoice, err := s.cache.GetInvoiceByNumberFromCache(ctx, invoiceNumber)

// Set invoice in cache (both ID and number)
err := s.cache.SetInvoiceInCache(ctx, invoice, 30*time.Minute)
err := s.cache.SetInvoiceByNumberInCache(ctx, invoice, 30*time.Minute)

// Invalidate invoice cache (both ID and number)
err := s.cache.InvalidateInvoiceCache(ctx, invoiceID)
err := s.cache.InvalidateInvoiceByNumberCache(ctx, invoiceNumber)
```

#### Individual Payment Cache Operations
```go
// Get payment from cache by ID
cachedPayment, err := s.cache.GetPaymentFromCache(ctx, paymentID)

// Get payment from cache by reference number
cachedPayment, err := s.cache.GetPaymentByReferenceFromCache(ctx, referenceNumber)

// Set payment in cache (both ID and reference)
err := s.cache.SetPaymentInCache(ctx, payment, 30*time.Minute)
err := s.cache.SetPaymentByReferenceInCache(ctx, payment, 30*time.Minute)

// Invalidate payment cache (both ID and reference)
err := s.cache.InvalidatePaymentCache(ctx, paymentID)
err := s.cache.InvalidatePaymentByReferenceCache(ctx, referenceNumber)
```

#### Invoice List Cache Operations
```go
// Get invoice list from cache
cachedInvoices, err := s.cache.GetInvoiceListFromCache(ctx, companyID, status, limit, offset)

// Set invoice list in cache
err := s.cache.SetInvoiceListInCache(ctx, companyID, status, limit, offset, invoices, 10*time.Minute)

// Invalidate all invoice lists for a company
err := s.cache.InvalidateInvoiceListCache(ctx, companyID)
```

#### Payment Instructions Cache Operations
```go
// Get payment instructions from cache
cachedInstructions, err := s.cache.GetPaymentInstructionsFromCache(ctx, invoiceID)

// Set payment instructions in cache
err := s.cache.SetPaymentInstructionsInCache(ctx, invoiceID, instructions, 1*time.Hour)

// Invalidate payment instructions cache
err := s.cache.InvalidatePaymentInstructionsCache(ctx, invoiceID)
```

#### Bulk Cache Operations
```go
// Bulk cache multiple invoices (both ID and number keys)
err := s.cache.BulkSetInvoicesInCache(ctx, invoices, 30*time.Minute)

// Bulk cache multiple payments (both ID and reference keys)
err := s.cache.BulkSetPaymentsInCache(ctx, payments, 30*time.Minute)
```

### Tracking Service Caching

#### Current Location Cache Operations
```go
// Get current vehicle location from cache
cachedLocation, err := s.cache.GetCurrentLocationFromCache(ctx, vehicleID)

// Set current vehicle location in cache
err := s.cache.SetCurrentLocationInCache(ctx, gpsTrack, 5*time.Minute)

// Cache is automatically invalidated when new GPS data arrives
```

#### Location History Cache Operations
```go
// Get location history from cache
cachedTracks, cachedTotal, err := s.cache.GetLocationHistoryFromCache(ctx, vehicleID, filters)

// Set location history in cache
err := s.cache.SetLocationHistoryInCache(ctx, vehicleID, filters, tracks, total, 15*time.Minute)

// Invalidate location history cache for a vehicle
err := s.cache.InvalidateLocationHistoryCache(ctx, vehicleID)
```

#### Trip Cache Operations
```go
// Get trip from cache
cachedTrip, err := s.cache.GetTripFromCache(ctx, tripID)

// Set trip in cache
err := s.cache.SetTripInCache(ctx, trip, 1*time.Hour)

// Invalidate trip cache
err := s.cache.InvalidateTripCache(ctx, tripID)
```

#### Geofence Cache Operations
```go
// Get individual geofence from cache
cachedGeofence, err := s.cache.GetGeofenceFromCache(ctx, geofenceID)

// Set individual geofence in cache
err := s.cache.SetGeofenceInCache(ctx, geofence, 1*time.Hour)

// Get company geofences from cache
cachedGeofences, err := s.cache.GetGeofencesByCompanyFromCache(ctx, companyID)

// Set company geofences in cache
err := s.cache.SetGeofencesByCompanyInCache(ctx, companyID, geofences, 30*time.Minute)

// Invalidate geofence caches
err := s.cache.InvalidateGeofenceCache(ctx, geofenceID)
err := s.cache.InvalidateGeofencesByCompanyCache(ctx, companyID)
```

### Analytics Service Caching

#### Fleet Dashboard Cache Operations
```go
// Get fleet dashboard from cache
cachedDashboard, err := s.cache.GetFleetDashboardFromCache(ctx, companyID)

// Set fleet dashboard in cache
err := s.cache.SetFleetDashboardInCache(ctx, companyID, dashboard, 10*time.Minute)

// Dashboard includes: active vehicles, trips, distance, fuel, utilization rate, etc.
```

#### Fuel Analytics Cache Operations
```go
// Get fuel analytics from cache
cachedAnalytics, err := s.cache.GetFuelAnalyticsFromCache(ctx, companyID, startDate, endDate)

// Set fuel analytics in cache
err := s.cache.SetFuelAnalyticsInCache(ctx, companyID, startDate, endDate, analytics, 30*time.Minute)

// Fuel analytics include: consumption, efficiency, costs, trends, theft alerts
```

#### Driver Performance Cache Operations
```go
// Get driver performance from cache
cachedPerformance, err := s.cache.GetDriverPerformanceFromCache(ctx, companyID, driverID, period)

// Set driver performance in cache
err := s.cache.SetDriverPerformanceInCache(ctx, companyID, driverID, period, performance, 20*time.Minute)

// Driver performance includes: behavior metrics, score, recommendations, trends
```

#### Compliance Report Cache Operations
```go
// Get compliance report from cache
cachedReport, err := s.cache.GetComplianceReportFromCache(ctx, companyID, period)

// Set compliance report in cache
err := s.cache.SetComplianceReportInCache(ctx, companyID, period, report, 1*time.Hour)

// Compliance reports include: driver hours, vehicle inspections, tax reports
```

#### Analytics Cache Invalidation
```go
// Invalidate all analytics cache for a company
err := s.cache.InvalidateAnalyticsCache(ctx, companyID)

// This invalidates:
// - Fleet dashboard cache
// - All fuel analytics cache
// - All driver performance cache
// - All compliance report cache
```

#### Enhanced Individual Item Caching Methods
```go
// GetInvoice - Cached individual invoice retrieval by ID
func (s *Service) GetInvoice(ctx context.Context, invoiceID string) (*models.Invoice, error) {
    // Try cache first, fallback to database, then cache result
    cachedInvoice, err := s.cache.GetInvoiceFromCache(ctx, invoiceID)
    if err == nil && cachedInvoice != nil {
        return cachedInvoice, nil
    }
    
    // Database lookup with automatic caching
    invoice, err := s.repoManager.GetInvoices().GetByID(ctx, invoiceID)
    if err != nil {
        return nil, err
    }
    
    // Cache by both ID and invoice number
    s.cache.SetInvoiceInCache(ctx, invoice, 30*time.Minute)
    s.cache.SetInvoiceByNumberInCache(ctx, invoice, 30*time.Minute)
    
    return invoice, nil
}

// GetInvoiceByNumber - Cached invoice lookup by invoice number
func (s *Service) GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error) {
    // Cache-first strategy with dual caching
    cachedInvoice, err := s.cache.GetInvoiceByNumberFromCache(ctx, invoiceNumber)
    if err == nil && cachedInvoice != nil {
        return cachedInvoice, nil
    }
    
    // Database lookup with automatic dual caching
    var invoice models.Invoice
    if err := s.db.Where("invoice_number = ?", invoiceNumber).First(&invoice).Error; err != nil {
        return nil, err
    }
    
    // Cache by both ID and invoice number for future lookups
    s.cache.SetInvoiceInCache(ctx, &invoice, 30*time.Minute)
    s.cache.SetInvoiceByNumberInCache(ctx, &invoice, 30*time.Minute)
    
    return &invoice, nil
}

// GetPayment - Cached individual payment retrieval by ID
func (s *Service) GetPayment(ctx context.Context, paymentID string) (*models.Payment, error) {
    // Cache-first strategy
    cachedPayment, err := s.cache.GetPaymentFromCache(ctx, paymentID)
    if err == nil && cachedPayment != nil {
        return cachedPayment, nil
    }
    
    // Database lookup with automatic caching
    payment, err := s.repoManager.PaymentRepository().GetByID(ctx, paymentID)
    if err != nil {
        return nil, err
    }
    
    // Cache by both ID and reference number
    s.cache.SetPaymentInCache(ctx, payment, 30*time.Minute)
    if payment.ReferenceNumber != "" {
        s.cache.SetPaymentByReferenceInCache(ctx, payment, 30*time.Minute)
    }
    
    return payment, nil
}

// GetPaymentByReference - Cached payment lookup by reference number
func (s *Service) GetPaymentByReference(ctx context.Context, referenceNumber string) (*models.Payment, error) {
    // Cache-first strategy
    cachedPayment, err := s.cache.GetPaymentByReferenceFromCache(ctx, referenceNumber)
    if err == nil && cachedPayment != nil {
        return cachedPayment, nil
    }
    
    // Database lookup with automatic dual caching
    var payment models.Payment
    if err := s.db.Where("reference_number = ?", referenceNumber).First(&payment).Error; err != nil {
        return nil, err
    }
    
    // Cache by both ID and reference number
    s.cache.SetPaymentInCache(ctx, &payment, 30*time.Minute)
    s.cache.SetPaymentByReferenceInCache(ctx, &payment, 30*time.Minute)
    
    return &payment, nil
}
```

#### Cached GetVehicle Method
```go
func (s *Service) GetVehicle(companyID, vehicleID string) (*models.Vehicle, error) {
    ctx := context.Background()
    
    // Try cache first
    cachedVehicle, err := s.cache.GetVehicleFromCache(ctx, vehicleID)
    if err != nil {
        // Log cache error but continue
        fmt.Printf("Cache error for vehicle %s: %v\n", vehicleID, err)
    }
    
    if cachedVehicle != nil && cachedVehicle.CompanyID == companyID {
        return cachedVehicle, nil
    }
    
    // Get from database
    var vehicle models.Vehicle
    if err := s.db.Preload("Driver").Where("company_id = ? AND id = ?", companyID, vehicleID).First(&vehicle).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, apperrors.NewNotFoundError("Vehicle")
        }
        return nil, apperrors.NewInternalError("Failed to fetch vehicle").WithInternal(err)
    }

    // Cache the result
    if err := s.cache.SetVehicleInCache(ctx, &vehicle, 30*time.Minute); err != nil {
        fmt.Printf("Failed to cache vehicle %s: %v\n", vehicleID, err)
    }

    return &vehicle, nil
}
```

#### Cache Invalidation on Updates
```go
func (s *Service) UpdateVehicle(companyID, vehicleID string, req UpdateVehicleRequest) (*models.Vehicle, error) {
    // ... update logic ...
    
    // Save changes
    if err := s.db.Save(vehicle).Error; err != nil {
        return nil, apperrors.NewInternalError("Failed to update vehicle").WithInternal(err)
    }

    // Invalidate cache after update
    ctx := context.Background()
    if err := s.cache.InvalidateVehicleCache(ctx, vehicleID); err != nil {
        fmt.Printf("Failed to invalidate vehicle cache %s: %v\n", vehicleID, err)
    }

    return vehicle, nil
}
```

## Cache Key Patterns

### Vehicle Cache Keys
- `vehicle:{vehicleID}` - Individual vehicle data
- `vehicle:list:{companyID}:{filterHash}` - Vehicle lists with filters
- `vehicle:stats:{companyID}` - Vehicle statistics (future implementation)

### Driver Cache Keys
- `driver:{driverID}` - Individual driver data
- `driver:list:{companyID}:{filterHash}` - Driver lists with filters
- `driver:stats:{companyID}` - Driver statistics (future implementation)

### Auth Cache Keys
- `user:{userID}` - Individual user data
- `session:{sessionID}` - User session data
- `token:{tokenHash}` - JWT token validation (future implementation)

### Payment Cache Keys
- `payment:{paymentID}` - Individual payment data by ID
- `payment:reference:{referenceNumber}` - Individual payment data by reference number
- `invoice:{invoiceID}` - Individual invoice data by ID
- `invoice:number:{invoiceNumber}` - Individual invoice data by invoice number
- `invoice:list:{companyID}:{status}:{limit}:{offset}` - Invoice lists with pagination
- `payment_instructions:{invoiceID}` - Payment instructions for invoices
- `payment:list:{companyID}:{filterHash}` - Payment lists (future implementation)
- `payment:stats:{companyID}` - Payment statistics (future implementation)

### Tracking Cache Keys
- `vehicle:location:{vehicleID}` - Current vehicle location (real-time)
- `location:history:{vehicleID}:{filterHash}` - Location history with filters
- `trip:{tripID}` - Individual trip data
- `geofence:{geofenceID}` - Individual geofence data
- `geofences:company:{companyID}` - Company geofences list

### Analytics Cache Keys
- `analytics:dashboard:{companyID}` - Fleet dashboard statistics
- `analytics:fuel:{companyID}:{startDate}:{endDate}` - Fuel consumption analytics
- `analytics:driver:{companyID}:{driverID}:{period}` - Driver performance analytics
- `analytics:compliance:{companyID}:{period}` - Compliance reports

## Best Practices

### 1. Cache-First Strategy
Always try cache first, then fall back to database:
```go
// Try cache first
cached, err := s.cache.GetFromCache(ctx, key)
if err == nil && cached != nil {
    return cached, nil
}

// Fall back to database
data, err := s.db.Find(...)
if err != nil {
    return nil, err
}

// Cache the result
s.cache.SetInCache(ctx, key, data, expiration)
return data, nil
```

### 2. Graceful Cache Failures
Never let cache failures break your application:
```go
if err := s.cache.SetInCache(ctx, key, data, expiration); err != nil {
    // Log error but don't fail the request
    log.Printf("Cache set failed: %v", err)
}
```

### 3. Cache Invalidation
Always invalidate cache when data changes:
```go
// After update/delete operations
if err := s.cache.InvalidateCache(ctx, key); err != nil {
    log.Printf("Cache invalidation failed: %v", err)
}
```

### 4. Appropriate TTL Values
- **Frequently accessed data**: 30 minutes - 1 hour
- **Rarely changing data**: 1-24 hours
- **Real-time data**: 1-5 minutes
- **User sessions**: 15-30 minutes
- **Payment instructions**: 1 hour (rarely change)
- **Invoice lists**: 10 minutes (shorter due to payment activity)
- **Payment data**: 30 minutes (sensitive financial data)
- **Current vehicle locations**: 5 minutes (real-time GPS data)
- **Location history**: 15 minutes (historical data)
- **Active trips**: 1 hour (frequently accessed)
- **Completed trips**: 2 hours (longer TTL for completed data)
- **Geofences**: 1 hour (individual), 30 minutes (company lists)
- **Fleet dashboard**: 10 minutes (frequently changing data)
- **Fuel analytics**: 30 minutes (moderate change frequency)
- **Driver performance**: 20 minutes (moderate change frequency)
- **Compliance reports**: 1 hour (rarely changing data)

## Performance Benefits

### Expected Improvements
- **Database load reduction**: 70-85% fewer database queries
- **Response time improvement**: 60-95% faster for cached data
- **List query performance**: 80-95% faster for paginated results
- **Session validation**: 90-99% faster authentication checks
- **Real-time tracking**: 90-99% faster current location retrieval
- **Location history**: 80-95% faster historical data access
- **Trip management**: 85-98% faster trip data retrieval
- **Geofence checking**: 70-90% faster geofence violation detection
- **Fleet dashboard**: 85-95% faster dashboard statistics loading
- **Fuel analytics**: 80-90% faster fuel consumption reports
- **Driver performance**: 75-85% faster driver analytics generation
- **Compliance reports**: 90-98% faster regulatory report generation
- **Scalability**: Better handling of concurrent requests
- **Cost reduction**: Lower database resource usage
- **User experience**: Near-instant response times for frequently accessed data

### Monitoring
Monitor cache hit rates and performance:
```go
// Example metrics to track
- Cache hit rate
- Cache miss rate
- Average response time
- Database query count
```

## Future Enhancements

### Planned Features
1. **List Caching**: Cache paginated results
2. **Statistics Caching**: Cache dashboard statistics
3. **Session Caching**: Cache user sessions
4. **Geospatial Caching**: Cache location-based queries
5. **Cache Warming**: Pre-populate frequently accessed data

### Advanced Patterns
1. **Write-Through Caching**: Update cache on database writes
2. **Cache-Aside Pattern**: Application manages cache
3. **Distributed Caching**: Multi-instance cache coordination
4. **Cache Compression**: Compress large cached objects

## Troubleshooting

### Common Issues

#### Redis Connection Errors
```bash
# Check Redis is running
redis-cli ping

# Check Redis logs
tail -f /var/log/redis/redis-server.log
```

#### Cache Miss Issues
- Verify cache keys are consistent
- Check TTL values are appropriate
- Ensure cache invalidation is working

#### Memory Issues
- Monitor Redis memory usage
- Implement cache eviction policies
- Use appropriate data structures

## Development Guidelines

### Adding Caching to New Services
1. Add Redis client to service struct
2. Create CacheService for the domain
3. Implement cache operations (Get, Set, Invalidate)
4. Update service methods to use cache-first pattern
5. Add cache invalidation on data mutations

### Testing with Caching
```go
// Mock Redis for unit tests
mockRedis := &MockRedisClient{}

// Test cache hit scenarios
// Test cache miss scenarios
// Test cache invalidation
```

This caching integration provides a solid foundation for high-performance fleet tracking operations while maintaining data consistency and reliability.
