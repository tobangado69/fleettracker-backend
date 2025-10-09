# Real-Time Features Documentation

## Overview

The FleetTracker Pro backend now includes comprehensive real-time features powered by WebSocket technology, providing live updates for fleet tracking, analytics, and alerts. This document outlines the implementation and usage of these real-time capabilities.

## Architecture

### Core Components

1. **Enhanced WebSocket Hub** (`internal/common/realtime/websocket_hub.go`)
2. **Analytics Broadcaster** (`internal/common/realtime/analytics_broadcaster.go`)
3. **Alert System** (`internal/common/realtime/alert_system.go`)
4. **Integrated Tracking Service** (Updated `internal/tracking/service.go`)

## WebSocket Hub Features

### Enhanced Connection Management

```go
type WebSocketHub struct {
    clients         map[*Client]bool
    register        chan *Client
    unregister      chan *Client
    broadcast       chan []byte
    companyChannels map[string]chan []byte
    redis           *redis.Client
    mutex           sync.RWMutex
    config          *WebSocketConfig
}
```

**Key Features:**
- **Company-Scoped Broadcasting**: Messages can be sent to specific companies
- **User-Specific Broadcasting**: Messages can be sent to specific users
- **Cross-Instance Communication**: Redis pub/sub for multi-instance deployments
- **Connection Health Monitoring**: Ping/pong heartbeat mechanism
- **Graceful Connection Handling**: Automatic cleanup of disconnected clients

### WebSocket Configuration

```go
type WebSocketConfig struct {
    ReadBufferSize  int           // 1024 bytes
    WriteBufferSize int           // 1024 bytes
    PingPeriod      time.Duration // 54 seconds
    PongWait        time.Duration // 60 seconds
    WriteWait       time.Duration // 10 seconds
    MaxMessageSize  int64         // 512 bytes
}
```

### Client Management

```go
type Client struct {
    ID        string
    CompanyID string
    UserID    string
    Conn      *websocket.Conn
    Send      chan []byte
    Hub       *WebSocketHub
}
```

## Real-Time Analytics Broadcasting

### Fleet Dashboard Updates

```go
type FleetDashboardUpdate struct {
    Type            string    `json:"type"`
    CompanyID       string    `json:"company_id"`
    ActiveVehicles  int       `json:"active_vehicles"`
    TotalTrips      int       `json:"total_trips"`
    DistanceTraveled float64  `json:"distance_traveled"`
    FuelConsumed    float64   `json:"fuel_consumed"`
    UtilizationRate float64   `json:"utilization_rate"`
    CostPerKm       float64   `json:"cost_per_km"`
    Timestamp       time.Time `json:"timestamp"`
}
```

**Features:**
- **Real-time KPI Updates**: Live fleet performance metrics
- **Automatic Broadcasting**: Periodic dashboard updates (configurable interval)
- **Company-Scoped Updates**: Each company receives only their data

### Vehicle Location Updates

```go
type VehicleLocationUpdate struct {
    Type        string    `json:"type"`
    CompanyID   string    `json:"company_id"`
    VehicleID   string    `json:"vehicle_id"`
    DriverID    string    `json:"driver_id"`
    Latitude    float64   `json:"latitude"`
    Longitude   float64   `json:"longitude"`
    Speed       float64   `json:"speed"`
    Heading     float64   `json:"heading"`
    Timestamp   time.Time `json:"timestamp"`
}
```

**Features:**
- **Live GPS Tracking**: Real-time vehicle location updates
- **Speed and Heading**: Complete vehicle movement data
- **Driver Association**: Links location to specific drivers

### Trip Updates

```go
type TripUpdate struct {
    Type        string     `json:"type"`
    CompanyID   string     `json:"company_id"`
    TripID      string     `json:"trip_id"`
    VehicleID   string     `json:"vehicle_id"`
    DriverID    string     `json:"driver_id"`
    Status      string     `json:"status"` // "started", "completed", "cancelled"
    StartTime   *time.Time `json:"start_time,omitempty"`
    EndTime     *time.Time `json:"end_time,omitempty"`
    Distance    float64    `json:"distance"`
    Duration    float64    `json:"duration"` // in minutes
    Timestamp   time.Time  `json:"timestamp"`
}
```

**Features:**
- **Trip Lifecycle Tracking**: Start, progress, and completion updates
- **Real-time Metrics**: Distance and duration calculations
- **Status Notifications**: Immediate trip status changes

## Real-Time Alert System

### Alert Types

```go
const (
    AlertTypeSpeedViolation     = "speed_violation"
    AlertTypeGeofenceViolation  = "geofence_violation"
    AlertTypeMaintenance        = "maintenance"
    AlertTypeFuelTheft          = "fuel_theft"
    AlertTypeDriverEvent        = "driver_event"
    AlertTypeVehicleOffline     = "vehicle_offline"
    AlertTypeSystemError        = "system_error"
    AlertTypePaymentReceived    = "payment_received"
    AlertTypeInvoiceGenerated   = "invoice_generated"
)
```

### Alert Severities

```go
const (
    AlertSeverityLow      = "low"
    AlertSeverityMedium   = "medium"
    AlertSeverityHigh     = "high"
    AlertSeverityCritical = "critical"
)
```

### Alert Structure

```go
type Alert struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"`
    CompanyID   string                 `json:"company_id"`
    UserID      string                 `json:"user_id,omitempty"`
    VehicleID   string                 `json:"vehicle_id,omitempty"`
    DriverID    string                 `json:"driver_id,omitempty"`
    Severity    string                 `json:"severity"`
    Title       string                 `json:"title"`
    Message     string                 `json:"message"`
    Data        map[string]interface{} `json:"data,omitempty"`
    Timestamp   time.Time              `json:"timestamp"`
    Read        bool                   `json:"read"`
    ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}
```

### Alert Management

**Create Alert:**
```go
func (as *AlertSystem) CreateAlert(ctx context.Context, alert *Alert) error
```

**Get Company Alerts:**
```go
func (as *AlertSystem) GetCompanyAlerts(ctx context.Context, companyID string, limit int64) ([]*Alert, error)
```

**Mark Alert as Read:**
```go
func (as *AlertSystem) MarkAlertAsRead(ctx context.Context, companyID, alertID string) error
```

**Delete Alert:**
```go
func (as *AlertSystem) DeleteAlert(ctx context.Context, companyID, alertID string) error
```

## WebSocket Message Format

### Standard Message Structure

```go
type WebSocketMessage struct {
    Type      string      `json:"type"`
    Data      interface{} `json:"data"`
    Timestamp time.Time   `json:"timestamp"`
    CompanyID string      `json:"company_id,omitempty"`
    UserID    string      `json:"user_id,omitempty"`
}
```

### Message Types

1. **connection_established** - Welcome message on connection
2. **fleet_dashboard_update** - Fleet dashboard data updates
3. **vehicle_location_update** - Real-time vehicle location
4. **driver_event_update** - Driver behavior events
5. **geofence_violation_update** - Geofence violations
6. **trip_update** - Trip status changes
7. **maintenance_alert_update** - Maintenance alerts
8. **alert** - General alert notifications
9. **alert_read** - Alert marked as read
10. **alert_deleted** - Alert deleted

## Integration with Tracking Service

### Enhanced GPS Processing

The tracking service now includes real-time broadcasting for:

1. **GPS Data Processing**: Real-time location updates
2. **Driver Behavior Analysis**: Speed violations and behavior events
3. **Trip Management**: Start and end notifications
4. **Geofence Monitoring**: Violation alerts

### Real-Time Broadcasting Integration

```go
// Broadcast real-time location update
go func() {
    if err := s.analyticsBroadcaster.BroadcastVehicleLocationUpdate(ctx, gpsTrack); err != nil {
        fmt.Printf("Failed to broadcast vehicle location update %s: %v\n", gpsTrack.VehicleID, err)
    }
}()

// Create real-time alert for speed violations
go func() {
    if err := s.alertSystem.CreateSpeedViolationAlert(ctx, vehicle.CompanyID, gpsTrack.VehicleID, *gpsTrack.DriverID, gpsTrack.Speed, 80, location); err != nil {
        fmt.Printf("Failed to create speed violation alert: %v\n", err)
    }
}()
```

## WebSocket Connection

### Connection URL

```
ws://localhost:8080/ws/tracking?company_id={companyID}&user_id={userID}
```

### Connection Parameters

- **company_id** (required): Company identifier
- **user_id** (optional): User identifier for user-specific messages

### Connection Handling

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/tracking?company_id=123&user_id=456');

ws.onopen = function(event) {
    console.log('Connected to FleetTracker Pro');
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('Received:', message);
    
    switch(message.type) {
        case 'vehicle_location_update':
            updateVehicleLocation(message.data);
            break;
        case 'alert':
            showAlert(message.data);
            break;
        case 'fleet_dashboard_update':
            updateDashboard(message.data);
            break;
    }
};

ws.onclose = function(event) {
    console.log('Disconnected from FleetTracker Pro');
};
```

## Performance Features

### Cross-Instance Communication

- **Redis Pub/Sub**: Enables real-time updates across multiple server instances
- **Channel-based Broadcasting**: Efficient message distribution
- **Automatic Failover**: Graceful handling of Redis connection issues

### Connection Optimization

- **Heartbeat Mechanism**: Ping/pong to maintain connection health
- **Automatic Reconnection**: Client-side reconnection logic
- **Message Queuing**: Buffered message sending for reliability
- **Connection Limits**: Configurable maximum connections per company

### Memory Management

- **Automatic Cleanup**: Disconnected clients are automatically removed
- **Alert Expiration**: Alerts automatically expire after 24 hours
- **Message Size Limits**: Configurable maximum message sizes
- **Connection Pooling**: Efficient connection management

## Security Features

### Origin Checking

```go
CheckOrigin: func(r *http.Request) bool {
    // Implement proper origin checking in production
    return true
}
```

### Company Isolation

- **Company-Scoped Messages**: Users only receive messages for their company
- **User-Specific Broadcasting**: Optional user-level message targeting
- **Data Isolation**: Complete separation of company data

### Authentication Integration

- **JWT Token Validation**: Optional token-based authentication
- **Session Management**: Integration with existing auth system
- **Permission-Based Access**: Role-based message filtering

## Monitoring and Diagnostics

### Connection Metrics

```go
// Get total connected clients
func (h *WebSocketHub) GetConnectedClients() int

// Get company-specific client count
func (h *WebSocketHub) GetCompanyClients(companyID string) int
```

### Alert Management

```go
// Get company alerts with pagination
alerts, err := alertSystem.GetCompanyAlerts(ctx, companyID, 50)

// Mark alert as read
err := alertSystem.MarkAlertAsRead(ctx, companyID, alertID)

// Delete alert
err := alertSystem.DeleteAlert(ctx, companyID, alertID)
```

### Health Checks

- **Connection Health**: Ping/pong heartbeat monitoring
- **Redis Health**: Pub/sub connection monitoring
- **Message Delivery**: Success/failure tracking
- **Alert Cleanup**: Automatic expired alert removal

## Usage Examples

### Frontend Integration

```javascript
class FleetTrackerWebSocket {
    constructor(companyId, userId) {
        this.companyId = companyId;
        this.userId = userId;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
    }
    
    connect() {
        const url = `ws://localhost:8080/ws/tracking?company_id=${this.companyId}&user_id=${this.userId}`;
        this.ws = new WebSocket(url);
        
        this.ws.onopen = () => {
            console.log('Connected to FleetTracker Pro');
            this.reconnectAttempts = 0;
        };
        
        this.ws.onmessage = (event) => {
            this.handleMessage(JSON.parse(event.data));
        };
        
        this.ws.onclose = () => {
            this.handleReconnect();
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }
    
    handleMessage(message) {
        switch(message.type) {
            case 'vehicle_location_update':
                this.updateVehicleLocation(message.data);
                break;
            case 'alert':
                this.showAlert(message.data);
                break;
            case 'fleet_dashboard_update':
                this.updateDashboard(message.data);
                break;
        }
    }
    
    handleReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            setTimeout(() => this.connect(), 1000 * this.reconnectAttempts);
        }
    }
}
```

### Real-Time Dashboard Updates

```javascript
function updateDashboard(data) {
    document.getElementById('active-vehicles').textContent = data.active_vehicles;
    document.getElementById('total-trips').textContent = data.total_trips;
    document.getElementById('distance-traveled').textContent = data.distance_traveled.toFixed(2);
    document.getElementById('fuel-consumed').textContent = data.fuel_consumed.toFixed(2);
    document.getElementById('utilization-rate').textContent = data.utilization_rate.toFixed(1) + '%';
    document.getElementById('cost-per-km').textContent = 'IDR ' + data.cost_per_km.toFixed(0);
}
```

### Alert Notifications

```javascript
function showAlert(alert) {
    const notification = document.createElement('div');
    notification.className = `alert alert-${alert.severity}`;
    notification.innerHTML = `
        <h4>${alert.title}</h4>
        <p>${alert.message}</p>
        <button onclick="markAlertAsRead('${alert.id}')">Mark as Read</button>
    `;
    
    document.getElementById('alerts-container').appendChild(notification);
    
    // Auto-remove after 10 seconds
    setTimeout(() => {
        notification.remove();
    }, 10000);
}
```

## Configuration

### Environment Variables

```bash
# Redis configuration for pub/sub
REDIS_URL=redis://localhost:6379

# WebSocket configuration
WS_READ_BUFFER_SIZE=1024
WS_WRITE_BUFFER_SIZE=1024
WS_PING_PERIOD=54s
WS_PONG_WAIT=60s
WS_WRITE_WAIT=10s
WS_MAX_MESSAGE_SIZE=512

# Alert configuration
ALERT_CLEANUP_INTERVAL=1h
ALERT_EXPIRY_DURATION=24h
```

### Service Initialization

```go
// Create WebSocket hub
hub := realtime.NewWebSocketHub(redis, realtime.DefaultWebSocketConfig())

// Create analytics broadcaster
analyticsBroadcaster := realtime.NewAnalyticsBroadcaster(hub, redis, db, repoManager)

// Create alert system
alertSystem := realtime.NewAlertSystem(hub, redis)

// Start periodic dashboard updates
go analyticsBroadcaster.StartPeriodicDashboardUpdates(ctx, 30*time.Second)

// Start alert cleanup
go alertSystem.StartAlertCleanup(ctx, 1*time.Hour)
```

## Best Practices

### Client-Side

1. **Implement Reconnection Logic**: Handle connection drops gracefully
2. **Message Queuing**: Queue messages during disconnection
3. **Error Handling**: Handle WebSocket errors appropriately
4. **Resource Cleanup**: Close connections when not needed

### Server-Side

1. **Graceful Shutdown**: Handle server shutdown properly
2. **Connection Limits**: Implement per-company connection limits
3. **Message Validation**: Validate incoming WebSocket messages
4. **Monitoring**: Monitor connection health and performance

### Security

1. **Origin Validation**: Implement proper origin checking
2. **Authentication**: Integrate with existing auth system
3. **Rate Limiting**: Implement message rate limiting
4. **Data Validation**: Validate all incoming data

## Troubleshooting

### Common Issues

1. **Connection Drops**: Check network stability and Redis connectivity
2. **Message Loss**: Verify Redis pub/sub configuration
3. **High Memory Usage**: Monitor connection count and cleanup
4. **Performance Issues**: Check message frequency and size

### Debugging

```go
// Enable WebSocket debugging
hub.config.Debug = true

// Monitor connection count
fmt.Printf("Connected clients: %d\n", hub.GetConnectedClients())

// Check Redis connectivity
err := redis.Ping(ctx).Err()
if err != nil {
    log.Printf("Redis connection error: %v", err)
}
```

## Future Enhancements

### Planned Features

1. **Message Compression**: Implement message compression for large payloads
2. **Selective Subscriptions**: Allow clients to subscribe to specific event types
3. **Message Persistence**: Store important messages for offline clients
4. **Advanced Analytics**: Real-time performance metrics and insights
5. **Mobile Push Notifications**: Integration with mobile push services
6. **Voice Alerts**: Audio notifications for critical alerts
7. **Custom Dashboards**: User-configurable real-time dashboards

### Performance Optimizations

1. **Message Batching**: Batch multiple updates into single messages
2. **Delta Updates**: Send only changed data instead of full updates
3. **Connection Pooling**: Optimize connection management
4. **Caching Integration**: Leverage existing cache infrastructure
5. **Load Balancing**: Distribute WebSocket connections across instances

The FleetTracker Pro real-time features provide a comprehensive solution for live fleet monitoring, analytics, and alerting, enabling fleet managers to make informed decisions based on real-time data and immediate notifications.
