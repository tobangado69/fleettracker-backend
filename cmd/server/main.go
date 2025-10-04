package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/internal/auth"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/config"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/database"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/repository"
	"github.com/tobangado69/fleettracker-pro/backend/internal/driver"
	"github.com/tobangado69/fleettracker-pro/backend/internal/payment"
	"github.com/tobangado69/fleettracker-pro/backend/internal/tracking"
	"github.com/tobangado69/fleettracker-pro/backend/internal/vehicle"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"

	_ "github.com/tobangado69/fleettracker-pro/backend/docs"
)

// @title FleetTracker Pro API
// @version 1.0
// @description Comprehensive Driver Tracking SaaS Application for Indonesian Fleet Management
// @termsOfService https://fleettracker.id/terms

// @contact.name FleetTracker Pro Support
// @contact.url https://fleettracker.id/support
// @contact.email support@fleettracker.id

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name auth
// @tag.description Authentication endpoints
// @tag.name vehicles
// @tag.description Vehicle management endpoints
// @tag.name drivers
// @tag.description Driver management endpoints
// @tag.name tracking
// @tag.description GPS tracking endpoints
// @tag.name payments
// @tag.description Payment integration endpoints
// @tag.name analytics
// @tag.description Analytics and reporting endpoints
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close(db)

	// Initialize Redis for caching
	redisClient, err := database.ConnectRedis(cfg.RedisURL)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// Auto-migrate database models
	log.Println("Running database migrations...")
	err = db.AutoMigrate(
		&models.Company{},
		&models.User{},
		&models.Session{},
		&models.AuditLog{},
		&models.PasswordResetToken{},
		&models.Vehicle{},
		&models.MaintenanceLog{},
		&models.FuelLog{},
		&models.Driver{},
		&models.DriverEvent{},
		&models.PerformanceLog{},
		&models.GPSTrack{},
		&models.Trip{},
		&models.Geofence{},
		&models.Subscription{},
		&models.Payment{},
		&models.Invoice{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migrations completed successfully")

	// Initialize repository manager
	repoManager := repository.NewRepositoryManager(db)
	log.Println("✅ Repository manager initialized successfully")

	// Initialize Gin router
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS configuration for Indonesian domains
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Security middleware
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RateLimit(cfg.RateLimitRequestsPerMinute))

	// Initialize services
	authService := auth.NewService(db, cfg.JWTSecret)
	trackingService := tracking.NewService(db, redisClient)
	vehicleService := vehicle.NewService(db)
	driverService := driver.NewService(db)
	paymentService := payment.NewService(db, cfg)

	// Initialize handlers
	authHandler := auth.NewHandler(authService)
	trackingHandler := tracking.NewHandler(trackingService)
	vehicleHandler := vehicle.NewHandler(vehicleService)
	driverHandler := driver.NewHandler(driverService)
	paymentHandler := payment.NewHandler(paymentService)

	// Setup routes
	setupRoutes(r, authHandler, trackingHandler, vehicleHandler, driverHandler, paymentHandler, cfg, db, repoManager)

	// Setup WebSocket for real-time tracking
	setupWebSocket(r, trackingService)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   "FleetTracker Pro API",
			"version":   "1.0.0",
		})
	})

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		log.Printf("🚛 FleetTracker Pro API starting on port %s", cfg.Port)
		log.Printf("📊 Health check: http://localhost:%s/health", cfg.Port)
		log.Printf("📚 API docs: http://localhost:%s/swagger/index.html", cfg.Port)
		log.Printf("🇮🇩 Indonesian Fleet Management SaaS Ready!")
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited gracefully")
}

func setupRoutes(
	r *gin.Engine,
	authHandler *auth.Handler,
	trackingHandler *tracking.Handler,
	vehicleHandler *vehicle.Handler,
	driverHandler *driver.Handler,
	paymentHandler *payment.Handler,
	cfg *config.Config,
	db *gorm.DB,
	repoManager *repository.RepositoryManager,
) {
	// API documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.AuthRequired(cfg.JWTSecret, db), authHandler.Logout)
			auth.GET("/profile", middleware.AuthRequired(cfg.JWTSecret, db), authHandler.GetProfile)
			auth.PUT("/profile", middleware.AuthRequired(cfg.JWTSecret, db), authHandler.UpdateProfile)
			auth.PUT("/change-password", middleware.AuthRequired(cfg.JWTSecret, db), authHandler.ChangePassword)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired(cfg.JWTSecret, db))
		{
			// Vehicle management
			vehicles := protected.Group("/vehicles")
			{
				vehicles.GET("", vehicleHandler.ListVehicles)              // List vehicles with filters
				vehicles.POST("", vehicleHandler.CreateVehicle)            // Create vehicle
				vehicles.GET("/:id", vehicleHandler.GetVehicle)            // Get vehicle details
				vehicles.PUT("/:id", vehicleHandler.UpdateVehicle)         // Update vehicle
				vehicles.DELETE("/:id", vehicleHandler.DeleteVehicle)      // Delete vehicle
				vehicles.PUT("/:id/status", vehicleHandler.UpdateVehicleStatus)     // Update vehicle status
				vehicles.POST("/:id/assign-driver", vehicleHandler.AssignDriver)    // Assign driver
				vehicles.DELETE("/:id/driver", vehicleHandler.UnassignDriver)       // Unassign driver
				vehicles.GET("/:id/driver", vehicleHandler.GetVehicleDriver)        // Get vehicle driver
				vehicles.PUT("/:id/inspection", vehicleHandler.UpdateInspectionDate) // Update inspection date
				
				// Legacy endpoints for backward compatibility
				vehicles.GET("/:id/status", vehicleHandler.GetVehicleStatus)
			}

			// Driver management
			drivers := protected.Group("/drivers")
			{
				drivers.GET("", driverHandler.ListDrivers)              // List drivers with filters
				drivers.POST("", driverHandler.CreateDriver)            // Create driver
				drivers.GET("/:id", driverHandler.GetDriver)            // Get driver details
				drivers.PUT("/:id", driverHandler.UpdateDriver)         // Update driver
				drivers.DELETE("/:id", driverHandler.DeleteDriver)      // Delete driver
				drivers.PUT("/:id/status", driverHandler.UpdateDriverStatus)     // Update driver status
				drivers.GET("/:id/performance", driverHandler.GetDriverPerformance) // Get performance data
				drivers.PUT("/:id/performance", driverHandler.UpdateDriverPerformance) // Update performance scores
				drivers.POST("/:id/assign-vehicle", driverHandler.AssignVehicle)    // Assign vehicle
				drivers.DELETE("/:id/vehicle", driverHandler.UnassignVehicle)       // Unassign vehicle
				drivers.GET("/:id/vehicle", driverHandler.GetDriverVehicle)        // Get assigned vehicle
				drivers.PUT("/:id/medical", driverHandler.UpdateMedicalCheckup)     // Update medical checkup
				drivers.PUT("/:id/training", driverHandler.UpdateTrainingStatus)    // Update training status
				
				// Legacy endpoints for backward compatibility
				drivers.GET("/:id/trips", driverHandler.GetDriverTrips)
			}

			// GPS tracking
			tracking := protected.Group("/tracking")
			{
				// GPS Data Management
				tracking.POST("/gps", trackingHandler.ProcessGPSData)                    // Submit GPS data
				tracking.GET("/vehicles/:id/current", trackingHandler.GetCurrentLocation) // Get current location
				tracking.GET("/vehicles/:id/history", trackingHandler.GetLocationHistory) // Get location history
				tracking.GET("/vehicles/:id/route", trackingHandler.GetRoute)            // Get route data
				
				// Driver Event Management
				tracking.POST("/events", trackingHandler.ProcessDriverEvent)             // Submit driver event
				tracking.GET("/events", trackingHandler.GetDriverEvents)                 // Get driver events
				
				// Trip Management
				tracking.POST("/trips", trackingHandler.StartTrip)                       // Start/end trip
				tracking.GET("/trips", trackingHandler.GetTrips)                         // Get trip history
				
				// Geofence Management
				tracking.POST("/geofences", trackingHandler.CreateGeofence)              // Create geofence
				tracking.GET("/geofences", trackingHandler.GetGeofences)                 // List geofences
				tracking.PUT("/geofences/:id", trackingHandler.UpdateGeofence)           // Update geofence
				tracking.DELETE("/geofences/:id", trackingHandler.DeleteGeofence)        // Delete geofence
				
				// WebSocket for real-time tracking
				tracking.GET("/ws/:vehicle_id", trackingHandler.HandleWebSocket)         // WebSocket connection
				
				// Analytics and Reporting
				tracking.GET("/dashboard/stats", trackingHandler.GetDashboardStats)     // Dashboard statistics
				tracking.GET("/analytics/fuel", trackingHandler.GetFuelConsumption)      // Fuel analytics
				tracking.GET("/analytics/drivers", trackingHandler.GetDriverPerformance) // Driver performance
				tracking.POST("/reports/generate", trackingHandler.GenerateReport)       // Generate reports
				tracking.GET("/reports/compliance", trackingHandler.GetComplianceReport) // Compliance report
			}

			// Payment integration
			payments := protected.Group("/payments")
			{
				payments.POST("/qris", paymentHandler.CreateQRISPayment)
				payments.POST("/bank-transfer", paymentHandler.CreateBankTransfer)
				payments.POST("/e-wallet", paymentHandler.CreateEWalletPayment)
				payments.GET("/subscriptions", paymentHandler.GetSubscriptions)
				payments.POST("/subscriptions", paymentHandler.CreateSubscription)
				payments.GET("/invoices", paymentHandler.GetInvoices)
			}

			// Analytics and reporting
			analytics := protected.Group("/analytics")
			{
				analytics.GET("/dashboard", trackingHandler.GetDashboardStats)
				analytics.GET("/fuel-consumption", trackingHandler.GetFuelConsumption)
				analytics.GET("/driver-performance", trackingHandler.GetDriverPerformance)
				analytics.GET("/reports", trackingHandler.GenerateReport)
				analytics.GET("/compliance", trackingHandler.GetComplianceReport)
			}

			// Repository health check (admin only)
			repo := protected.Group("/repository")
			repo.Use(middleware.RoleRequired("admin"))
			{
				repo.GET("/health", func(c *gin.Context) {
					if err := repoManager.HealthCheck(c.Request.Context()); err != nil {
						c.JSON(http.StatusServiceUnavailable, gin.H{
							"status": "unhealthy",
							"error":  err.Error(),
						})
						return
					}
					
					stats, err := repoManager.GetStats()
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"status": "error",
							"error":  err.Error(),
						})
						return
					}
					
					c.JSON(http.StatusOK, gin.H{
						"status": "healthy",
						"stats":  stats,
					})
				})
			}
		}
	}
}

func setupWebSocket(r *gin.Engine, trackingService *tracking.Service) {
	// WebSocket endpoint for real-time GPS tracking
	r.GET("/ws/tracking", trackingService.HandleWebSocket)
}
