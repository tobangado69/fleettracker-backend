package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a system user (admin, manager, operator)
type User struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID string    `json:"company_id" gorm:"type:uuid;not null;index"`
	Email     string    `json:"email" gorm:"type:varchar(255);unique;not null"`
	Username  string    `json:"username" gorm:"type:varchar(100);unique;not null"`
	Password  string    `json:"-" gorm:"type:varchar(255);not null"` // Hidden from JSON
	
	// Personal Information
	FirstName   string    `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName    string    `json:"last_name" gorm:"type:varchar(100);not null"`
	Phone       string    `json:"phone" gorm:"type:varchar(20)"`
	Avatar      string    `json:"avatar" gorm:"type:varchar(500)"`
	
	// Indonesian Fields
	NIK         string    `json:"nik" gorm:"type:varchar(16);unique"`           // Indonesian ID
	Address     string    `json:"address" gorm:"type:text"`
	City        string    `json:"city" gorm:"type:varchar(100)"`
	Province    string    `json:"province" gorm:"type:varchar(100)"`
	PostalCode  string    `json:"postal_code" gorm:"type:varchar(10)"`
	
	// Role and Permissions
	Role        string    `json:"role" gorm:"type:varchar(50);not null;default:'operator'"` // admin, manager, operator
	Permissions JSON      `json:"permissions" gorm:"type:jsonb"`                            // Custom permissions
	
	// Account Status
	Status      string    `json:"status" gorm:"type:varchar(20);default:'active'"` // active, inactive, suspended
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	IsVerified  bool      `json:"is_verified" gorm:"default:false"`
	LastLoginAt *time.Time `json:"last_login_at"`
	
	// Authentication Security
	FailedLoginAttempts int       `json:"failed_login_attempts" gorm:"default:0"`
	LockedUntil         *time.Time `json:"locked_until"`
	EmailVerifiedAt     *time.Time `json:"email_verified_at"`
	EmailVerificationToken string  `json:"-" gorm:"type:varchar(255)"` // Hidden from JSON
	
	// Language and Preferences
	Language    string    `json:"language" gorm:"type:varchar(10);default:'id'"` // id, en
	Timezone    string    `json:"timezone" gorm:"type:varchar(50);default:'Asia/Jakarta'"`
	Preferences JSON      `json:"preferences" gorm:"type:jsonb"`                 // User preferences
	
	// Security
	TwoFactorEnabled bool      `json:"two_factor_enabled" gorm:"default:false"`
	TwoFactorSecret  string    `json:"-" gorm:"type:varchar(255)"` // Hidden from JSON
	PasswordChangedAt time.Time `json:"password_changed_at"`
	
	// Timestamps
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Company             Company               `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	Sessions            []Session             `json:"sessions,omitempty" gorm:"foreignKey:UserID"`
	AuditLogs           []AuditLog            `json:"audit_logs,omitempty" gorm:"foreignKey:UserID"`
	PasswordResetTokens []PasswordResetToken  `json:"password_reset_tokens,omitempty" gorm:"foreignKey:UserID"`
}

// Session represents user login sessions
type Session struct {
	ID           string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Token        string    `json:"-" gorm:"type:varchar(500);unique;not null"` // Hidden from JSON
	RefreshToken string    `json:"-" gorm:"type:varchar(500)"`                 // Hidden from JSON
	UserAgent    string    `json:"user_agent" gorm:"type:text"`
	IPAddress    string    `json:"ip_address" gorm:"type:varchar(45)"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// AuditLog represents user activity logs
type AuditLog struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Action    string    `json:"action" gorm:"type:varchar(100);not null"`
	Resource  string    `json:"resource" gorm:"type:varchar(100)"`
	ResourceID string   `json:"resource_id" gorm:"type:varchar(100)"`
	Details   JSON      `json:"details" gorm:"type:jsonb"`
	IPAddress string    `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent string    `json:"user_agent" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// PasswordResetToken represents password reset tokens
type PasswordResetToken struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Token     string    `json:"-" gorm:"type:varchar(255);unique;not null"` // Hidden from JSON
	ExpiresAt time.Time `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// TableName specifies the table name for the Session model
func (Session) TableName() string {
	return "sessions"
}

// TableName specifies the table name for the AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}

// TableName specifies the table name for the PasswordResetToken model
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// BeforeCreate hook to hash password and set default values
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	
	// Set default values
	if u.Language == "" {
		u.Language = "id"
	}
	if u.Timezone == "" {
		u.Timezone = "Asia/Jakarta"
	}
	if u.Role == "" {
		u.Role = "operator"
	}
	if u.Status == "" {
		u.Status = "active"
	}
	
	u.PasswordChangedAt = time.Now()
	
	return nil
}

// BeforeUpdate hook to hash password if changed
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("Password") {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
		u.PasswordChangedAt = time.Now()
	}
	return nil
}

// CheckPassword verifies the provided password against the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// IsValidNIK validates Indonesian NIK format (16 digits)
func (u *User) IsValidNIK() bool {
	if len(u.NIK) != 16 {
		return false
	}
	// Additional NIK validation logic can be added here
	return true
}

// HasPermission checks if user has a specific permission
func (u *User) HasPermission(permission string) bool {
	if u.Permissions == nil {
		return false
	}
	value, exists := u.Permissions[permission]
	if !exists {
		return false
	}
	
	if boolValue, ok := value.(bool); ok {
		return boolValue
	}
	return false
}

// SetPermission sets a specific permission for the user
func (u *User) SetPermission(permission string, value bool) {
	if u.Permissions == nil {
		u.Permissions = make(JSON)
	}
	u.Permissions[permission] = value
}

// IsAdmin checks if user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// IsManager checks if user has manager role
func (u *User) IsManager() bool {
	return u.Role == "manager"
}

// IsOperator checks if user has operator role
func (u *User) IsOperator() bool {
	return u.Role == "operator"
}

// CanAccessCompany checks if user can access a specific company
func (u *User) CanAccessCompany(companyID string) bool {
	return u.CompanyID == companyID
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

// IsAccountLocked checks if the account is currently locked
func (u *User) IsAccountLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// LockAccount locks the account for the specified duration
func (u *User) LockAccount(duration time.Duration) {
	lockUntil := time.Now().Add(duration)
	u.LockedUntil = &lockUntil
}

// UnlockAccount unlocks the account and resets failed attempts
func (u *User) UnlockAccount() {
	u.LockedUntil = nil
	u.FailedLoginAttempts = 0
}

// IncrementFailedAttempts increments failed login attempts
func (u *User) IncrementFailedAttempts() {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= 5 {
		// Lock account for 30 minutes after 5 failed attempts
		u.LockAccount(30 * time.Minute)
	}
}

// ResetFailedAttempts resets failed login attempts
func (u *User) ResetFailedAttempts() {
	u.FailedLoginAttempts = 0
}

// VerifyEmail marks the email as verified
func (u *User) VerifyEmail() {
	now := time.Now()
	u.EmailVerifiedAt = &now
	u.IsVerified = true
	u.EmailVerificationToken = ""
}

// IsEmailVerified checks if the email is verified
func (u *User) IsEmailVerified() bool {
	return u.IsVerified && u.EmailVerifiedAt != nil
}
