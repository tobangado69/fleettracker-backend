package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/errors"
)

// ErrorResponse represents a standardized error response.
type ErrorResponse struct {
	Success bool                   `json:"success"`
	Error   *ErrorDetail           `json:"error"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// ErrorDetail contains error information.
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorHandler middleware handles errors and returns standardized error responses.
// It should be one of the last middleware in the chain.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Execute request handlers
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error (most recent)
			err := c.Errors.Last().Err

			// Convert to AppError
			appErr := errors.GetAppError(err)

			// Log error with context
			logError(c, appErr)

			// Don't write response if headers already sent
			if c.Writer.Written() {
				return
			}

			// Build error response
			response := ErrorResponse{
				Success: false,
				Error: &ErrorDetail{
					Code:    appErr.Code,
					Message: appErr.Message,
					Details: appErr.Details,
				},
				Meta: buildErrorMeta(c),
			}

			// Send error response
			c.JSON(appErr.Status, response)
		}
	}
}

// RecoveryHandler recovers from panics and returns a 500 error.
func RecoveryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log panic with stack trace
				log.Printf("[PANIC] %v\n%s", err, debug.Stack())

				// Don't write response if headers already sent
				if c.Writer.Written() {
					return
				}

				// Return 500 error
				response := ErrorResponse{
					Success: false,
					Error: &ErrorDetail{
						Code:    "INTERNAL_ERROR",
						Message: "Internal server error",
					},
					Meta: buildErrorMeta(c),
				}

				c.JSON(http.StatusInternalServerError, response)
			}
		}()

		c.Next()
	}
}

// AbortWithError is a helper to abort request with AppError.
func AbortWithError(c *gin.Context, err *errors.AppError) {
	c.Error(err)
	c.Abort()
}

// AbortWithNotFound aborts with 404 error.
func AbortWithNotFound(c *gin.Context, resource string) {
	AbortWithError(c, errors.NewNotFoundError(resource))
}

// AbortWithUnauthorized aborts with 401 error.
func AbortWithUnauthorized(c *gin.Context, message string) {
	AbortWithError(c, errors.NewUnauthorizedError(message))
}

// AbortWithForbidden aborts with 403 error.
func AbortWithForbidden(c *gin.Context, message string) {
	AbortWithError(c, errors.NewForbiddenError(message))
}

// AbortWithValidation aborts with 400 validation error.
func AbortWithValidation(c *gin.Context, message string) {
	AbortWithError(c, errors.NewValidationError(message))
}

// AbortWithBadRequest aborts with 400 bad request error.
func AbortWithBadRequest(c *gin.Context, message string) {
	AbortWithError(c, errors.NewBadRequestError(message))
}

// AbortWithConflict aborts with 409 conflict error.
func AbortWithConflict(c *gin.Context, message string) {
	AbortWithError(c, errors.NewConflictError(message))
}

// AbortWithInternal aborts with 500 internal error.
func AbortWithInternal(c *gin.Context, message string, err error) {
	appErr := errors.NewInternalError(message)
	if err != nil {
		appErr = appErr.WithInternal(err)
	}
	AbortWithError(c, appErr)
}

// logError logs the error with request context.
func logError(c *gin.Context, err *errors.AppError) {
	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = "unknown"
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	companyID := c.GetString("company_id")
	if companyID == "" {
		companyID = "none"
	}

	log.Printf(
		"[ERROR] [%s] %s %s | User: %s | Company: %s | Code: %s | Message: %s | Internal: %v",
		requestID,
		c.Request.Method,
		c.Request.URL.Path,
		userID,
		companyID,
		err.Code,
		err.Message,
		err.InternalErr,
	)

	// Log stack trace for internal errors
	if err.Status >= 500 && err.InternalErr != nil {
		log.Printf("[ERROR] [%s] Stack trace: %s", requestID, debug.Stack())
	}
}

// buildErrorMeta builds metadata for error response.
func buildErrorMeta(c *gin.Context) map[string]interface{} {
	meta := make(map[string]interface{})

	// Add request ID if available
	if requestID := c.GetString("request_id"); requestID != "" {
		meta["request_id"] = requestID
	}

	// Add timestamp
	meta["timestamp"] = c.GetTime("request_time").Format("2006-01-02T15:04:05Z07:00")

	return meta
}

