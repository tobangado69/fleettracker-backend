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

	"github.com/tobangado69/fleettracker-pro/backend/internal/auth"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/config"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/database"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
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
	setupRoutes(r, authHandler, trackingHandler, vehicleHandler, driverHandler, paymentHandler)

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
			auth.POST("/logout", middleware.AuthRequired(), authHandler.Logout)
			auth.GET("/profile", middleware.AuthRequired(), authHandler.GetProfile)
			auth.PUT("/profile", middleware.AuthRequired(), authHandler.UpdateProfile)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired())
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
				drivers.GET("", driverHandler.GetDrivers)
				drivers.POST("", driverHandler.CreateDriver)
				drivers.GET("/:id", driverHandler.GetDriver)
				drivers.PUT("/:id", driverHandler.UpdateDriver)
				drivers.DELETE("/:id", driverHandler.DeleteDriver)
				drivers.GET("/:id/performance", driverHandler.GetDriverPerformance)
				drivers.GET("/:id/trips", driverHandler.GetDriverTrips)
			}

			// GPS tracking
			tracking := protected.Group("/tracking")
			{
				tracking.POST("/gps", trackingHandler.ProcessGPSData)
				tracking.GET("/vehicles/:id/current", trackingHandler.GetCurrentLocation)
				tracking.GET("/vehicles/:id/history", trackingHandler.GetLocationHistory)
				tracking.GET("/vehicles/:id/route", trackingHandler.GetRoute)
				tracking.POST("/events", trackingHandler.ProcessDriverEvent)
				tracking.GET("/events", trackingHandler.GetDriverEvents)
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
		}
	}
}

func setupWebSocket(r *gin.Engine, trackingService *tracking.Service) {
	// WebSocket endpoint for real-time GPS tracking
	r.GET("/ws/tracking", trackingService.HandleWebSocket)
}
