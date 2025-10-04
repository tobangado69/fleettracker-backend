package vehicle

import (
	"gorm.io/gorm"
)

// Service handles vehicle operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new vehicle service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}
