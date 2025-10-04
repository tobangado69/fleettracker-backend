package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service handles authentication operations
type Service struct {
	db        *gorm.DB
	jwtSecret []byte
}

// Claims represents JWT claims
type Claims struct {
	UserID    string `json:"user_id"`
	CompanyID string `json:"company_id"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

// NewService creates a new authentication service
func NewService(db *gorm.DB, jwtSecret string) *Service {
	return &Service{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

// Login authenticates a user
func (s *Service) Login(email, password string) (map[string]interface{}, string, error) {
	// TODO: Implement actual login logic
	// For now, return mock data
	user := map[string]interface{}{
		"id":         "temp-user-id",
		"email":      email,
		"name":       "Demo User",
		"role":       "fleet_manager",
		"company_id": "temp-company-id",
	}

	token := "mock-jwt-token"
	return user, token, nil
}

// Register creates a new user
func (s *Service) Register(email, password, name string) (map[string]interface{}, error) {
	// TODO: Implement actual registration logic
	user := map[string]interface{}{
		"id":    "temp-user-id",
		"email": email,
		"name":  name,
		"role":  "fleet_manager",
	}
	return user, nil
}

// ValidateToken validates a JWT token
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	// TODO: Implement actual token validation
	claims := &Claims{
		UserID:    "temp-user-id",
		CompanyID: "temp-company-id",
		Role:      "fleet_manager",
	}
	return claims, nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

// CheckPasswordHash compares password with hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
