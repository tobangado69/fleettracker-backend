package driver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetDrivers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"drivers": []map[string]interface{}{}, "message": "Driver list not implemented yet"})
}
func (h *Handler) CreateDriver(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "Driver creation not implemented yet"})
}
func (h *Handler) GetDriver(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get driver not implemented yet"})
}
func (h *Handler) UpdateDriver(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Driver update not implemented yet"})
}
func (h *Handler) DeleteDriver(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Driver deletion not implemented yet"})
}
func (h *Handler) GetDriverPerformance(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Driver performance not implemented yet"})
}
func (h *Handler) GetDriverTrips(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Driver trips not implemented yet"})
}
