package analytics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
)

// GenerateReport godoc
// @Summary Generate analytics report
// @Description Generate various analytics reports
// @Tags analytics
// @Produce json
// @Param type query string false "Report type (fuel, driver, fleet, compliance)"
// @Param format query string false "Report format (json, csv, pdf)"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/reports/generate [post]
// @Security BearerAuth
func (h *Handler) GenerateReport(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	reportType := c.DefaultQuery("type", "fleet")
	format := c.DefaultQuery("format", "json")

	// Generate report based on type
	var reportData interface{}

	switch reportType {
	case "fuel":
		analytics, err := h.service.GetFuelConsumption(c.Request.Context(), companyID.(string), 
			time.Now().AddDate(0, 0, -30), time.Now())
		if err != nil {
			middleware.AbortWithInternal(c, "Failed to generate fuel report", err)
			return
		}
		reportData = analytics
	case "compliance":
		report, err := h.service.GetComplianceReport(c.Request.Context(), companyID.(string), "monthly")
		if err != nil {
			middleware.AbortWithInternal(c, "Failed to generate compliance report", err)
			return
		}
		reportData = report
	default: // fleet
		dashboard, err := h.service.GetFleetDashboard(c.Request.Context(), companyID.(string))
		if err != nil {
			middleware.AbortWithInternal(c, "Failed to generate fleet report", err)
			return
		}
		reportData = dashboard
	}

	// For now, return JSON format
	// In real implementation, this would generate PDF/CSV based on format parameter
	responseData := map[string]interface{}{
		"report_type": reportType,
		"format":     format,
		"data":       reportData,
		"generated_at": time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    responseData,
		Message: "Report generated successfully",
	})
}

// GetComplianceReport godoc
// @Summary Get compliance report
// @Description Get Indonesian regulatory compliance report
// @Tags analytics
// @Produce json
// @Param period query string false "Report period (monthly, quarterly, yearly)"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/reports/compliance [get]
// @Security BearerAuth
func (h *Handler) GetComplianceReport(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	period := c.DefaultQuery("period", "monthly")

	report, err := h.service.GetComplianceReport(c.Request.Context(), companyID.(string), period)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get compliance report", err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    report,
	})
}

// ExportReport godoc
// @Summary Export analytics report
// @Description Export analytics report in various formats
// @Tags analytics
// @Produce json
// @Param id path string true "Report ID"
// @Param format query string false "Export format (csv, pdf, excel)"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/reports/export/{id} [get]
// @Security BearerAuth
func (h *Handler) ExportReport(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	reportID := c.Param("id")
	_ = companyID // Use companyID for future implementation
	format := c.DefaultQuery("format", "csv")

	// For now, return a placeholder response
	// In real implementation, this would generate and return the actual file
	exportData := map[string]interface{}{
		"report_id":    reportID,
		"format":      format,
		"download_url": "/api/v1/analytics/reports/download/" + reportID + "." + format,
		"expires_at":  time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    exportData,
		Message: "Export link generated successfully",
	})
}

