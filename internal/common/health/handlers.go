package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP handlers for health checks
type Handler struct {
	checker *HealthChecker
}

// NewHandler creates a new health check handler
func NewHandler(checker *HealthChecker) *Handler {
	return &Handler{
		checker: checker,
	}
}

// HandleHealth handles basic health check (liveness probe)
// @Summary Health check
// @Description Basic health check endpoint (liveness probe)
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *Handler) HandleHealth(c *gin.Context) {
	response := h.checker.Check()
	c.JSON(http.StatusOK, response)
}

// HandleLiveness handles Kubernetes liveness probe
// @Summary Liveness probe
// @Description Kubernetes liveness probe endpoint
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health/live [get]
func (h *Handler) HandleLiveness(c *gin.Context) {
	response := h.checker.CheckLiveness()
	c.JSON(http.StatusOK, response)
}

// HandleReadiness handles Kubernetes readiness probe
// @Summary Readiness probe
// @Description Kubernetes readiness probe endpoint with dependency checks
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse "Service is ready"
// @Success 503 {object} HealthResponse "Service is not ready"
// @Router /health/ready [get]
func (h *Handler) HandleReadiness(c *gin.Context) {
	response := h.checker.CheckReadiness(c.Request.Context())
	
	// Return appropriate HTTP status based on health
	statusCode := http.StatusOK
	switch response.Status {
case StatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	case StatusDegraded:
		statusCode = http.StatusOK // Still return 200 for degraded (service works but slower)
	}
	
	c.JSON(statusCode, response)
}

// HandleDetailed handles detailed health check
// @Summary Detailed health check
// @Description Comprehensive health check with all system details
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health/detailed [get]
func (h *Handler) HandleDetailed(c *gin.Context) {
	response := h.checker.CheckReadiness(c.Request.Context())
	c.JSON(http.StatusOK, response)
}

// SetupHealthRoutes sets up health check routes
func SetupHealthRoutes(r *gin.Engine, handler *Handler) {
	// Basic health check (for load balancers, simple monitoring)
	r.GET("/health", handler.HandleHealth)
	
	// Kubernetes probes
	r.GET("/health/live", handler.HandleLiveness)
	r.GET("/health/ready", handler.HandleReadiness)
	
	// Detailed health check (for ops/debugging)
	r.GET("/health/detailed", handler.HandleDetailed)
}

