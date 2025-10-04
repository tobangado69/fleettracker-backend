package payment

import (
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/config"
	"gorm.io/gorm"
)

type Service struct {
	db   *gorm.DB
	cfg  *config.Config
}

func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{db: db, cfg: cfg}
}
