# FleetTracker Pro API Documentation

**Version**: 1.0.0  
**Base URL**: `http://localhost:8080/api/v1`  
**Interactive Docs**: http://localhost:8080/swagger/index.html

---

## 🚀 Quick Start

### 1. Register First User (Company Owner)
**⚠️ IMPORTANT**: Public registration is restricted to the first user only. All subsequent users must be created by administrators.

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "owner@company.com",
    "username": "owner",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+6281234567890",
    "company_name": "PT Fleet Indonesia"
  }'
```

**Note**: If you're not the first user, contact your company administrator to create your account.

### 2. Login & Get Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@company.com",
    "password": "SecurePass123"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "refresh_token_here",
    "expires_at": "2025-01-09T10:00:00Z"
  }
}
```

### 3. Use the Token
```bash
curl http://localhost:8080/api/v1/vehicles \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## 🔐 Authentication

All API requests (except `/auth/register` and `/auth/login`) require authentication.

### Headers
```http
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json
```

### Token Expiry
- Access token: 24 hours
- Refresh token: 7 days

### Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "YOUR_REFRESH_TOKEN"}'
```

---

## 📊 Response Format

### Success Response
```json
{
  "success": true,
  "data": { ... }
}
```

### Error Response
```json
{
  "success": false,
  "error": "Error type",
  "message": "Detailed error message"
}
```

### Paginated Response
```json
{
  "success": true,
  "data": [ ... ],
  "meta": {
    "total": 100,
    "page": 1,
    "limit": 10,
    "total_pages": 10,
    "has_next": true,
    "has_previous": false
  }
}
```

---

## 🚗 Core Resources

### [User Management](#-user-management) ⭐ NEW
- **7+ endpoints** for admin-controlled user creation
- Role-based access control (5 levels)
- Company isolation enforcement
- Privilege escalation prevention
- **Roles**: super-admin, owner, admin, operator, driver

### [Vehicles](resources/vehicles.md)
- **10+ endpoints** for vehicle fleet management
- CRUD operations with Indonesian compliance (STNK, BPKB)
- Vehicle status tracking
- Maintenance history
- Assignment management

### [Drivers](resources/drivers.md)
- **9+ endpoints** for driver management
- SIM validation (Indonesian driver's license)
- Performance tracking
- Compliance monitoring
- Vehicle assignment

### [GPS Tracking](resources/tracking.md)
- **8+ endpoints** for real-time tracking
- Location history
- Trip management
- Geofencing
- WebSocket support for live updates

### [Payments](resources/payments.md)
- **12+ endpoints** for payment processing
- Invoice generation
- Payment confirmation
- Subscription management
- Indonesian tax calculation

### [Analytics](resources/analytics.md)
- **15+ endpoints** for reporting
- Dashboard data
- Fuel consumption
- Driver performance
- Fleet utilization
- Custom reports

---

## 🌐 Common Headers

### Request Headers
```http
Authorization: Bearer YOUR_TOKEN       # Required for protected endpoints
Content-Type: application/json         # For POST/PUT requests
Accept-Encoding: gzip                  # Get compressed responses (60-80% smaller)
```

### Response Headers
```http
Content-Encoding: gzip                 # Response is compressed
X-API-Version: 1.0.0                   # API version
X-Service-Name: FleetTracker Pro API   # Service identifier
X-RateLimit-Limit: 100                 # Requests allowed per window
X-RateLimit-Remaining: 95              # Requests remaining
X-RateLimit-Reset: 1704711600          # Unix timestamp when limit resets
```

---

## ⚡ Rate Limiting

### Limits
- **General endpoints**: 100 requests/minute
- **Login endpoint**: 10 requests/minute
- **Analytics endpoints**: 50 requests/minute

### Headers
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1704711600
```

### Rate Limited Response (429)
```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Please try again later.",
  "retry_after": 60,
  "limit": 100,
  "remaining": 0,
  "reset": 1704711600
}
```

---

## 📖 Guides

### [Quick Start Guide](guides/quick-start.md)
Get up and running in 5 minutes

### [Authentication Flow](guides/authentication.md)
Complete authentication guide with examples

### [Pagination](guides/pagination.md)
How to paginate large result sets

### [Error Handling](guides/errors.md)
All error codes and how to handle them

### [Rate Limiting](guides/rate-limiting.md)
Understanding and working with rate limits

### [Indonesian Compliance](guides/indonesian-compliance.md)
NIK, NPWP, SIM, STNK validation and requirements

---

## 💻 Code Examples

### JavaScript/TypeScript
```javascript
class FleetTrackerAPI {
  constructor(baseURL, token) {
    this.baseURL = baseURL;
    this.token = token;
  }

  async request(endpoint, options = {}) {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json',
        'Accept-Encoding': 'gzip',
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'API request failed');
    }

    return response.json();
  }

  // Vehicle methods
  async getVehicles(page = 1, limit = 10) {
    return this.request(`/vehicles?page=${page}&limit=${limit}`);
  }

  async createVehicle(data) {
    return this.request('/vehicles', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }
}

// Usage
const api = new FleetTrackerAPI('http://localhost:8080/api/v1', 'YOUR_TOKEN');
const vehicles = await api.getVehicles();
```

### Python
```python
import requests

class FleetTrackerAPI:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.token = token
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json',
            'Accept-Encoding': 'gzip',
        })

    def get_vehicles(self, page=1, limit=10):
        response = self.session.get(
            f'{self.base_url}/vehicles',
            params={'page': page, 'limit': limit}
        )
        response.raise_for_status()
        return response.json()

    def create_vehicle(self, data):
        response = self.session.post(
            f'{self.base_url}/vehicles',
            json=data
        )
        response.raise_for_status()
        return response.json()

# Usage
api = FleetTrackerAPI('http://localhost:8080/api/v1', 'YOUR_TOKEN')
vehicles = api.get_vehicles()
```

---

## 🇮🇩 Indonesian Features

### NIK Validation (National ID)
```
Format: 16 digits
Example: 3174012801950001
Pattern: PPDDMMYYXXXXNNNN
  PP   = Province code (31 = DKI Jakarta)
  DD   = District code  
  MMYY = Birth date (month +40 for females)
  XXXX = Serial number
  NNNN = Registration number
```

### NPWP Validation (Tax ID)
```
Format: 15 digits
Example: 012345678901000
Can be formatted: 01.234.567.8-901.000
```

### SIM Validation (Driver's License)
```
Format: 12 digits
Example: 123456789012
Types: A, B, C, D (motorcycle to bus)
```

### License Plate Format
```
Format: X 1234 ABC (or XX 1234 ABC)
Examples:
  B 1234 ABC  (Jakarta)
  D 5678 XYZ  (Bandung)
  L 9012 DEF  (Surabaya)
```

---

## 🔍 Search & Filter

### Common Query Parameters
```
?page=1               # Page number (default: 1)
?limit=10             # Results per page (default: 10, max: 100)
?search=keyword       # Search by keyword
?sort=created_at      # Sort field
?order=desc           # Sort order (asc/desc)
?status=active        # Filter by status
?company_id=uuid      # Filter by company (multi-tenant)
```

### Example
```bash
curl "http://localhost:8080/api/v1/vehicles?page=1&limit=20&status=active&sort=created_at&order=desc" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 🏥 Health & Monitoring

### Health Checks
```bash
# Basic health
GET /health

# Kubernetes readiness
GET /health/ready

# Kubernetes liveness  
GET /health/live

# Detailed status
GET /health/detailed
```

### Metrics
```bash
# Prometheus format
GET /metrics

# JSON format
GET /metrics/json
```

---

## 🔐 Authentication & Sessions

### GET /auth/sessions
Get all active sessions for the current user.

**Authentication**: Required  
**Authorization**: Authenticated users

**Example Request:**
```bash
curl http://localhost:8080/api/v1/auth/sessions \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Example Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)...",
      "ip_address": "192.168.1.100",
      "is_active": true,
      "expires_at": "2025-10-15T10:00:00Z",
      "created_at": "2025-10-08T10:00:00Z",
      "is_current": true
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "user_agent": "FleetTracker Mobile App/1.0 (Android)",
      "ip_address": "192.168.1.50",
      "is_active": true,
      "expires_at": "2025-10-15T08:30:00Z",
      "created_at": "2025-10-08T08:30:00Z",
      "is_current": false
    }
  ]
}
```

**Use Cases:**
- View all devices/locations logged in
- Identify suspicious logins
- Monitor session security
- Prepare to revoke specific sessions

---

### DELETE /auth/sessions/:id
Revoke a specific session (logout from specific device).

**Authentication**: Required  
**Authorization**: Authenticated users (own sessions only)

**Example Request:**
```bash
curl -X DELETE http://localhost:8080/api/v1/auth/sessions/550e8400-e29b-41d4-a716-446655440002 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Example Response:**
```json
{
  "success": true,
  "message": "Session revoked successfully"
}
```

**Error Responses:**

```json
// Not Found - Session doesn't exist or doesn't belong to user
{
  "success": false,
  "error": "NOT_FOUND",
  "message": "Session not found"
}

// Unauthorized
{
  "success": false,
  "error": "UNAUTHORIZED",
  "message": "User ID not found"
}
```

**Security Notes (Session Isolation):**
- ✅ **User Isolation**: Users can ONLY view/revoke their OWN sessions
  - Query filters: `WHERE user_id = ?` (GetActiveSessions)
  - Query filters: `WHERE id = ? AND user_id = ?` (RevokeSession)
  - User A CANNOT see User B's sessions
  - User A CANNOT revoke User B's sessions
- ✅ Current session is marked with `is_current: true` (avoid self-logout)
- ✅ Revoked sessions are immediately invalidated (database + Redis cache)
- ✅ Session cannot be reused after revocation

**Use Cases:**
- Logout from specific device remotely
- Revoke suspicious session
- Security: logged in on public computer, revoke that session
- Lost phone: revoke mobile app session

---

## 👥 User Management

### Role Hierarchy & Multi-Tenant Isolation

FleetTracker Pro uses a strict 5-level role hierarchy with **100% company isolation**:

```
┌─────────────────────────────────────────────────────────────┐
│                      SUPER-ADMIN                            │
│  Platform-level | Can access ALL companies                  │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┴───────────────────┐
        │                                       │
┌───────────────────┐                  ┌───────────────────┐
│   COMPANY A       │                  │   COMPANY B       │
│ (100% Isolated)   │                  │ (100% Isolated)   │
├───────────────────┤                  ├───────────────────┤
│  OWNER ────────── │                  │ ────────── OWNER  │
│  ADMIN ────────── │                  │ ────────── ADMIN  │
│  OPERATOR ─────── │                  │ ─────── OPERATOR  │
│  DRIVER ───────── │                  │ ───────── DRIVER  │
│                   │                  │                   │
│  ❌ Cannot see     │                  │  ❌ Cannot see     │
│     Company B     │                  │     Company A     │
└───────────────────┘                  └───────────────────┘
```

| Role | Description | Can Create | Company Access | Cross-Company |
|------|-------------|------------|----------------|---------------|
| **super-admin** | System administrator | All roles | All companies | ✅ YES |
| **owner** | Company owner | admin, operator, driver | Own company | ❌ NO |
| **admin** | Client administrator | operator, driver | Own company | ❌ NO |
| **operator** | Regular user | None | Own company | ❌ NO |
| **driver** | Mobile app user | None | Own company | ❌ NO |

### Security Rules
- ✅ **Registration restricted** to first user (company owner)
- ✅ **All other users** must be created by admins
- ✅ **Company isolation** enforced for ALL roles (owner/admin/operator/driver)
- ✅ **ONLY super-admin** can access cross-company data
- ✅ **Privilege escalation** prevented
- ✅ **Role hierarchy** strictly enforced

### Multi-Tenant Isolation Guarantee
**Critical**: Users from Company A **CANNOT** see Company B's data:
- ❌ Owner A cannot see Company B
- ❌ Admin A cannot see Company B
- ❌ Operator A cannot see Company B
- ❌ Driver A cannot see Company B
- ✅ ONLY super-admin can access all companies (for platform support)

---

### GET /users/allowed-roles
Get roles that current user can assign.

**Authentication**: Required  
**Authorization**: Authenticated users

**Example Request:**
```bash
curl http://localhost:8080/api/v1/users/allowed-roles \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Example Response:**
```json
{
  "success": true,
  "data": {
    "allowed_roles": ["admin", "operator", "driver"],
    "descriptions": {
      "admin": "Client administrator, can manage operators and drivers",
      "operator": "Regular user, can access company resources",
      "driver": "Mobile app user, can track trips and update location"
    }
  }
}
```

---

### POST /users
Create a new user (admin-only, role hierarchy enforced).

**Authentication**: Required  
**Authorization**: super-admin, owner, or admin

**Cross-Company Creation (Super-Admin Only):**
- ✅ Super-admin can specify `company_id` to create users in ANY company
- ❌ Owner/Admin CANNOT specify `company_id` (automatically uses their company)

**Request Body:**
```json
{
  "email": "newuser@company.com",
  "username": "newuser",
  "password": "SecurePass123!",
  "first_name": "Jane",
  "last_name": "Doe",
  "phone": "+6281234567890",
  "role": "operator",
  "company_id": "company-b-uuid"  // ONLY for super-admin (cross-company creation)
}
```

**Examples:**

**Super-Admin creating admin for Company A:**
```json
{
  "email": "admin@companya.com",
  "username": "admin_a",
  "password": "SecurePass123!",
  "first_name": "Admin",
  "last_name": "A",
  "role": "admin",
  "company_id": "company-a-uuid"  // Super-admin specifies target company
}
```

**Super-Admin creating driver for Company B:**
```json
{
  "email": "driver@companyb.com",
  "username": "driver_b",
  "password": "SecurePass123!",
  "first_name": "Driver",
  "last_name": "B",
  "role": "driver",
  "company_id": "company-b-uuid"  // Super-admin specifies different company
}
```

**Owner creating operator (company_id ignored):**
```json
{
  "email": "operator@company.com",
  "username": "operator1",
  "password": "SecurePass123!",
  "first_name": "Jane",
  "last_name": "Smith",
  "role": "operator"
  // company_id automatically set to owner's company
}
```

**Example Request:**
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "operator@company.com",
    "username": "operator1",
    "password": "SecurePass123!",
    "first_name": "Jane",
    "last_name": "Smith",
    "phone": "+6281234567890",
    "role": "operator"
  }'
```

**Example Response:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "operator@company.com",
    "username": "operator1",
    "first_name": "Jane",
    "last_name": "Smith",
    "phone": "+6281234567890",
    "role": "operator",
    "company_id": "company-uuid",
    "is_active": true,
    "is_verified": false,
    "created_at": "2025-01-08T10:00:00Z"
  },
  "message": "User created successfully"
}
```

**Error Responses:**

```json
// Forbidden - Insufficient permissions
{
  "success": false,
  "error": "FORBIDDEN",
  "message": "role admin cannot create users with role owner"
}

// Forbidden - Privilege escalation attempt
{
  "success": false,
  "error": "FORBIDDEN",
  "message": "role admin cannot assign role owner (privilege escalation prevented)"
}

// Conflict - Email already exists
{
  "success": false,
  "error": "CONFLICT",
  "message": "email operator@company.com already exists"
}
```

---

### GET /users
List all users in the company (admin-only, paginated).

**Authentication**: Required  
**Authorization**: super-admin, owner, or admin

**Query Parameters:**
- `page` (optional, default: 1) - Page number
- `limit` (optional, default: 10, max: 100) - Results per page

**Example Request:**
```bash
curl "http://localhost:8080/api/v1/users?page=1&limit=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Example Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "user-uuid-1",
      "email": "owner@company.com",
      "username": "owner",
      "first_name": "John",
      "last_name": "Doe",
      "phone": "+6281234567890",
      "role": "owner",
      "company_id": "company-uuid",
      "is_active": true,
      "is_verified": true,
      "last_login_at": "2025-01-08T09:30:00Z",
      "created_at": "2025-01-01T10:00:00Z"
    },
    {
      "id": "user-uuid-2",
      "email": "admin@company.com",
      "username": "admin1",
      "first_name": "Jane",
      "last_name": "Smith",
      "role": "admin",
      "company_id": "company-uuid",
      "is_active": true,
      "is_verified": true,
      "created_at": "2025-01-02T10:00:00Z"
    }
  ],
  "meta": {
    "total": 15,
    "page": 1,
    "limit": 20,
    "total_pages": 1
  }
}
```

---

### GET /users/:id
Get user details by ID (admin-only).

**Authentication**: Required  
**Authorization**: super-admin, owner, or admin

**Example Request:**
```bash
curl http://localhost:8080/api/v1/users/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Example Response:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "operator@company.com",
    "username": "operator1",
    "first_name": "Jane",
    "last_name": "Smith",
    "phone": "+6281234567890",
    "role": "operator",
    "company_id": "company-uuid",
    "is_active": true,
    "is_verified": true,
    "last_login_at": "2025-01-08T08:00:00Z",
    "created_at": "2025-01-05T10:00:00Z"
  }
}
```

**Error Responses:**

```json
// Not Found
{
  "success": false,
  "error": "NOT_FOUND",
  "message": "user 550e8400-e29b-41d4-a716-446655440000 not found"
}

// Forbidden - Different company
{
  "success": false,
  "error": "FORBIDDEN",
  "message": "You can only access data from your own company"
}
```

---

### PUT /users/:id
Update user details (admin-only).

**Authentication**: Required  
**Authorization**: super-admin, owner, or admin

**Request Body:**
```json
{
  "first_name": "Updated Name",
  "last_name": "Updated Last",
  "phone": "+6281234567891",
  "email": "newemail@company.com"
}
```

**Example Request:**
```bash
curl -X PUT http://localhost:8080/api/v1/users/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Jane Updated",
    "phone": "+6281234567891"
  }'
```

**Example Response:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "operator@company.com",
    "username": "operator1",
    "first_name": "Jane Updated",
    "last_name": "Smith",
    "phone": "+6281234567891",
    "role": "operator",
    "company_id": "company-uuid",
    "is_active": true,
    "is_verified": true,
    "created_at": "2025-01-05T10:00:00Z"
  },
  "message": "User updated successfully"
}
```

---

### DELETE /users/:id
Deactivate a user (owner/super-admin only).

**Authentication**: Required  
**Authorization**: super-admin or owner only

**Example Request:**
```bash
curl -X DELETE http://localhost:8080/api/v1/users/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Example Response:**
```json
{
  "success": true,
  "message": "User deactivated successfully"
}
```

**Error Responses:**

```json
// Forbidden - Insufficient permissions
{
  "success": false,
  "error": "FORBIDDEN",
  "message": "Only super-admin or owner can deactivate users"
}

// Bad Request - Self-deactivation
{
  "success": false,
  "error": "BAD_REQUEST",
  "message": "Cannot deactivate your own account"
}
```

**Note:** Deactivation also invalidates all active sessions for the user.

---

### PUT /users/:id/role
Change a user's role (admin-only, role hierarchy enforced).

**Authentication**: Required  
**Authorization**: super-admin, owner, or admin

**Request Body:**
```json
{
  "new_role": "admin"
}
```

**Example Request:**
```bash
curl -X PUT http://localhost:8080/api/v1/users/550e8400-e29b-41d4-a716-446655440000/role \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "new_role": "admin"
  }'
```

**Example Response:**
```json
{
  "success": true,
  "message": "User role changed successfully"
}
```

**Error Responses:**

```json
// Forbidden - Privilege escalation prevented
{
  "success": false,
  "error": "FORBIDDEN",
  "message": "role admin cannot assign role owner (privilege escalation prevented)"
}

// Forbidden - Cannot create super-admin
{
  "success": false,
  "error": "FORBIDDEN",
  "message": "owner cannot assign super-admin or owner roles"
}

// Validation Error
{
  "success": false,
  "error": "VALIDATION_ERROR",
  "message": "invalid role: unknown-role"
}
```

---

## 📚 Additional Resources

- **Interactive API Docs**: http://localhost:8080/swagger/index.html
- **Postman Collection**: [Download](../examples/FleetTracker-Pro.postman_collection.json)
- **Architecture Guide**: [ARCHITECTURE.md](../guides/ARCHITECTURE.md)
- **Deployment Guide**: [README.md#deployment](../../README.md#deployment)

---

## 🆘 Support

- **Issues**: GitHub Issues
- **Email**: api-support@fleettracker.id  
- **Documentation**: This guide
- **Interactive Testing**: Swagger UI

---

**Last Updated**: October 2025 | **Version**: 1.0.0

