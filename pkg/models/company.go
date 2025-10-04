package models

import (
	"time"

	"gorm.io/gorm"
)

// Company represents a fleet management company
type Company struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Email       string    `json:"email" gorm:"type:varchar(255);unique;not null"`
	Phone       string    `json:"phone" gorm:"type:varchar(20)"`
	Address     string    `json:"address" gorm:"type:text"`
	City        string    `json:"city" gorm:"type:varchar(100)"`
	Province    string    `json:"province" gorm:"type:varchar(100)"`
	PostalCode  string    `json:"postal_code" gorm:"type:varchar(10)"`
	Country     string    `json:"country" gorm:"type:varchar(100);default:'Indonesia'"`

	// Indonesian Compliance Fields
	NPWP        string    `json:"npwp" gorm:"type:varchar(20);unique"`           // Tax ID
	SIUP        string    `json:"siup" gorm:"type:varchar(50)"`                  // Business License
	SKT         string    `json:"skt" gorm:"type:varchar(50)"`                   // Tax Certificate
	PKP         bool      `json:"pkp" gorm:"default:false"`                      // VAT Registered
	CompanyType string    `json:"company_type" gorm:"type:varchar(50)"`         // PT, CV, UD, etc.

	// Business Information
	Industry        string    `json:"industry" gorm:"type:varchar(100)"`
	FleetSize       int       `json:"fleet_size" gorm:"default:0"`
	MaxVehicles     int       `json:"max_vehicles" gorm:"default:100"`
	SubscriptionTier string   `json:"subscription_tier" gorm:"type:varchar(50);default:'basic'"`
	
	// Status and Settings
	Status      string    `json:"status" gorm:"type:varchar(20);default:'active'"` // active, suspended, inactive
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	Settings    JSON      `json:"settings" gorm:"type:jsonb"`                      // Company-specific settings
	
	// Timestamps
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Users      []User      `json:"users,omitempty" gorm:"foreignKey:CompanyID"`
	Vehicles   []Vehicle   `json:"vehicles,omitempty" gorm:"foreignKey:CompanyID"`
	Drivers    []Driver    `json:"drivers,omitempty" gorm:"foreignKey:CompanyID"`
	Trips      []Trip      `json:"trips,omitempty" gorm:"foreignKey:CompanyID"`
	Subscriptions []Subscription `json:"subscriptions,omitempty" gorm:"foreignKey:CompanyID"`
	Payments   []Payment   `json:"payments,omitempty" gorm:"foreignKey:CompanyID"`
}

// JSON represents a JSON field type for GORM
type JSON map[string]interface{}

// TableName specifies the table name for the Company model
func (Company) TableName() string {
	return "companies"
}

// BeforeCreate hook to set default values
func (c *Company) BeforeCreate(tx *gorm.DB) error {
	if c.Country == "" {
		c.Country = "Indonesia"
	}
	if c.Status == "" {
		c.Status = "active"
	}
	if c.SubscriptionTier == "" {
		c.SubscriptionTier = "basic"
	}
	return nil
}

// IsValidNPWP validates Indonesian NPWP format (15 digits)
func (c *Company) IsValidNPWP() bool {
	if len(c.NPWP) != 15 {
		return false
	}
	// Additional NPWP validation logic can be added here
	return true
}

// IsValidSIUP validates Indonesian SIUP format
func (c *Company) IsValidSIUP() bool {
	if len(c.SIUP) < 10 {
		return false
	}
	// Additional SIUP validation logic can be added here
	return true
}

// GetDisplayName returns a formatted company name for display
func (c *Company) GetDisplayName() string {
	if c.CompanyType != "" {
		return c.CompanyType + " " + c.Name
	}
	return c.Name
}

// GetFullAddress returns a formatted address string
func (c *Company) GetFullAddress() string {
	address := c.Address
	if c.City != "" {
		address += ", " + c.City
	}
	if c.Province != "" {
		address += ", " + c.Province
	}
	if c.PostalCode != "" {
		address += " " + c.PostalCode
	}
	if c.Country != "" && c.Country != "Indonesia" {
		address += ", " + c.Country
	}
	return address
}
