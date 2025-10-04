package tracking

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewService(db *gorm.DB, redis *redis.Client) *Service {
	return &Service{db: db, redis: redis}
}

func (s *Service) HandleWebSocket(c *gin.Context) {
	c.JSON(200, gin.H{"message": "WebSocket not implemented yet"})
}
