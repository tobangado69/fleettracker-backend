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
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/internal/analytics"
	"github.com/tobangado69/fleettracker-pro/backend/internal/auth"
	advancedanalytics "github.com/tobangado69/fleettracker-pro/backend/internal/common/analytics"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/config"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/database"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/export"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/fleet"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/geofencing"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/health"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/jobs"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/logging"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/ratelimit"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/repository"
	"github.com/tobangado69/fleettracker-pro/backend/internal/driver"
	"github.com/tobangado69/fleettracker-pro/backend/internal/payment"
	"github.com/tobangado69/fleettracker-pro/backend/internal/tracking"
	"github.com/tobangado69/fleettracker-pro/backend/internal/vehicle"

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

	// Initialize structured logging
	loggerConfig := &logging.LoggerConfig{
		Level:      logging.LogLevel(getEnv("LOG_LEVEL", "info")),
		Format:     "json", // JSON format for production
		Output:     os.Stdout,
		AddSource:  true,
		TimeFormat: "2006-01-02T15:04:05.000Z07:00",
	}
	logger := logging.NewLogger(loggerConfig)
	logging.InitDefaultLogger(loggerConfig)
	
	logger.Info("Starting FleetTracker Pro API",
		"version", "1.0.0",
		"environment", getEnv("ENVIRONMENT", "development"),
	)

	// Initialize database
	logger.Info("Connecting to database...")
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close(db)
	logger.Info("✅ Database connected successfully")

	// Configure GORM with slow query logging
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	
	// Set up slow query logger (queries > 100ms)
	slowQueryLogger := logging.NewSlowQueryLogger(logger, 100*time.Millisecond)
	db.Logger = slowQueryLogger

	// Initialize Redis for caching
	logger.Info("Connecting to Redis...")
	redisClient, err := database.ConnectRedis(cfg.RedisURL)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()
	logger.Info("✅ Redis connected successfully")

	// Database migrations are handled via SQL migration files
	// See migrations/ directory and use: make migrate-up
	// AutoMigrate disabled to avoid UUID default syntax issues
	logger.Info("⏭️  Skipping AutoMigrate - use 'make migrate-up' to run migrations")
	logger.Info("💡 Note: Database schema is managed via SQL migrations in migrations/ directory")

	// Initialize repository manager
	repoManager := repository.NewRepositoryManager(db)
	log.Println("✅ Repository manager initialized successfully")

	// Initialize audit logger
	auditLogger := logging.NewAuditLogger(logger, db)
	logger.Info("✅ Audit logging initialized")

	// Initialize health checker
	healthChecker := health.NewHealthChecker(db, redisClient, "FleetTracker Pro API", "1.0.0")
	healthHandler := health.NewHandler(healthChecker)
	metricsHandler := health.NewMetricsHandler(healthChecker)
	logger.Info("✅ Health check system initialized")

	// Initialize Gin router
	r := gin.New()
	
	// Compression middleware (60-80% bandwidth reduction)
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	logger.Info("✅ Response compression enabled (gzip)")
	
	// Logging middleware (replaces gin.Logger)
	r.Use(logging.RequestLoggingMiddleware(logger))
	r.Use(logging.PerformanceLoggingMiddleware(logger, 1*time.Second))
	r.Use(logging.ErrorLoggingMiddleware(logger))
	r.Use(logging.RecoveryLoggingMiddleware(logger))
	
	logger.Info("✅ Logging middleware initialized")

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
	
	// API versioning middleware
	apiVersionConfig := middleware.DefaultAPIVersionConfig()
	r.Use(middleware.APIVersionMiddleware(apiVersionConfig))
	logger.Info("✅ API versioning headers enabled", "version", apiVersionConfig.Version)
	
	// Audit middleware for state-changing operations
	r.Use(logging.AuditMiddleware(auditLogger))
	
	// Initialize rate limiting system
	rateLimitManager := ratelimit.NewRateLimitManager(redisClient, nil)
	rateLimitMonitor := ratelimit.NewRateLimitMonitor(redisClient)
	
	// Apply comprehensive rate limiting middleware
	r.Use(ratelimit.MonitoredRateLimitMiddleware(rateLimitManager, rateLimitMonitor))

	// Initialize export service with caching
	exportCacheService := export.NewExportCacheService(redisClient)
	exportService := export.NewExportService(db, exportCacheService)
	
	// Initialize job processing system
	log.Println("Initializing job processing system...")
	jobManager := jobs.NewManager(db, redisClient, jobs.DefaultManagerConfig())
	
	// Start job manager (workers and scheduler)
	if err := jobManager.Start(); err != nil {
		log.Fatal("Failed to start job manager:", err)
	}
	log.Println("✅ Export service with caching initialized successfully")

	// Initialize services
	authService := auth.NewService(db, redisClient, cfg.JWTSecret)
	trackingService := tracking.NewService(db, redisClient)
	vehicleService := vehicle.NewService(db, redisClient)
	vehicleHistoryService := vehicle.NewVehicleHistoryService(db, repoManager)
	driverService := driver.NewService(db, redisClient)
	paymentService := payment.NewService(db, redisClient, cfg, repoManager)
	analyticsService := analytics.NewService(db, redisClient, repoManager)
	
	// Initialize fleet management system
	fleetManager := fleet.NewFleetManager(db, redisClient)
	fleetAPI := fleet.NewFleetAPI(fleetManager)
	log.Println("✅ Advanced Fleet Management system initialized successfully")
	
	// Initialize geofencing system
	geofenceManager := geofencing.NewGeofenceManager(db, redisClient)
	geofenceAPI := geofencing.NewGeofenceAPI(geofenceManager)
	geofenceMonitor := geofencing.NewGeofenceMonitor(db, redisClient, geofenceManager)
	
	// Start geofence monitoring
	geofenceMonitor.StartMonitoring(context.Background(), nil)
	log.Println("✅ Advanced Geofencing Management system initialized successfully")
	
	// Initialize advanced analytics system
	analyticsEngine := advancedanalytics.NewAnalyticsEngine(db, redisClient)
	analyticsAPI := advancedanalytics.NewAnalyticsAPI(analyticsEngine)
	log.Println("✅ Advanced Analytics system initialized successfully")

	// Initialize handlers
	authHandler := auth.NewHandler(authService)
	trackingHandler := tracking.NewHandler(trackingService)
	vehicleHandler := vehicle.NewHandler(vehicleService)
	vehicleHistoryHandler := vehicle.NewVehicleHistoryHandler(vehicleHistoryService)
	driverHandler := driver.NewHandler(driverService)
	paymentHandler := payment.NewHandler(paymentService)
	analyticsHandler := analytics.NewHandler(analyticsService)

	// Setup routes
	setupRoutes(r, authHandler, trackingHandler, vehicleHandler, vehicleHistoryHandler, driverHandler, paymentHandler, analyticsHandler, fleetAPI, geofenceAPI, analyticsAPI, cfg, db, repoManager, rateLimitManager, rateLimitMonitor, jobManager, exportService)

	// Setup WebSocket for real-time tracking
	setupWebSocket(r, trackingService)

	// Setup health check endpoints
	health.SetupHealthRoutes(r, healthHandler)
	health.SetupMetricsRoutes(r, metricsHandler)
	logger.Info("✅ Health check endpoints configured")

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		logger.Info("🚛 FleetTracker Pro API starting",
			"port", cfg.Port,
			"health_check", "http://localhost:"+cfg.Port+"/health",
			"api_docs", "http://localhost:"+cfg.Port+"/swagger/index.html",
		)
		logger.Info("🇮🇩 Indonesian Fleet Management SaaS Ready!")
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Warn("🛑 Shutting down server...")
	
	// Stop geofence monitoring
	logger.Info("Stopping geofence monitoring...")
	geofenceMonitor.StopMonitoring()
	logger.Info("✅ Geofence monitoring stopped")
	
	// Stop job processing system
	logger.Info("Stopping job processing system...")
	jobManager.Stop()
	logger.Info("✅ Job processing system stopped")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		log.Fatal("Server forced to shutdown:", err)
	}

	logger.Info("✅ Server exited gracefully")
}

func setupRoutes(
	r *gin.Engine,
	authHandler *auth.Handler,
	trackingHandler *tracking.Handler,
	vehicleHandler *vehicle.Handler,
	vehicleHistoryHandler *vehicle.VehicleHistoryHandler,
	driverHandler *driver.Handler,
	paymentHandler *payment.Handler,
	analyticsHandler *analytics.Handler,
	fleetAPI *fleet.FleetAPI,
	geofenceAPI *geofencing.GeofenceAPI,
	analyticsAPI *advancedanalytics.AnalyticsAPI,
	cfg *config.Config,
	db *gorm.DB,
	repoManager *repository.RepositoryManager,
	rateLimitManager *ratelimit.RateLimitManager,
	rateLimitMonitor *ratelimit.RateLimitMonitor,
	jobManager *jobs.Manager,
	exportService *export.ExportService,
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
				
				// Vehicle History Management 🚧 **NEW**
				vehicles.GET("/:id/history", vehicleHistoryHandler.GetVehicleHistory)                    // Get vehicle history
				vehicles.POST("/:id/history", vehicleHistoryHandler.AddVehicleHistory)                   // Add history entry
				vehicles.GET("/:id/history/:historyId", vehicleHistoryHandler.GetVehicleHistoryByID)     // Get specific history entry
				vehicles.PUT("/:id/history/:historyId", vehicleHistoryHandler.UpdateVehicleHistory)      // Update history entry
				vehicles.DELETE("/:id/history/:historyId", vehicleHistoryHandler.DeleteVehicleHistory)   // Delete history entry
				vehicles.GET("/:id/maintenance", vehicleHistoryHandler.GetMaintenanceHistory)            // Get maintenance history
				vehicles.GET("/:id/costs", vehicleHistoryHandler.GetCostSummary)                         // Get cost summary
				vehicles.GET("/:id/trends", vehicleHistoryHandler.GetMaintenanceTrends)                  // Get maintenance trends
				
				// Legacy endpoints for backward compatibility
				vehicles.GET("/:id/status", vehicleHandler.GetVehicleStatus)
			}
			
			// Vehicle Maintenance Management 🚧 **NEW**
			maintenance := protected.Group("/vehicles/maintenance")
			{
				maintenance.GET("/upcoming", vehicleHistoryHandler.GetUpcomingMaintenance)               // Get upcoming maintenance
				maintenance.GET("/overdue", vehicleHistoryHandler.GetOverdueMaintenance)                 // Get overdue maintenance
				maintenance.PUT("/:historyId/schedule", vehicleHistoryHandler.UpdateMaintenanceSchedule) // Update maintenance schedule
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
				// Manual bank transfer with invoice generation
				payments.POST("/invoices", paymentHandler.GenerateInvoice)                    // Generate invoice
				payments.POST("/invoices/:id/confirm", paymentHandler.ConfirmPayment)        // Confirm payment
				payments.GET("/invoices", paymentHandler.GetInvoices)                        // List invoices
				payments.GET("/invoices/:id/instructions", paymentHandler.GetPaymentInstructions) // Get payment instructions
				payments.POST("/subscriptions/billing", paymentHandler.GenerateSubscriptionBilling) // Generate subscription billing
				
				// Legacy endpoints (not implemented for manual bank transfer)
				payments.POST("/qris", paymentHandler.CreateQRISPayment)
				payments.POST("/bank-transfer", paymentHandler.CreateBankTransfer)
				payments.POST("/e-wallet", paymentHandler.CreateEWalletPayment)
				payments.GET("/subscriptions", paymentHandler.GetSubscriptions)
				payments.POST("/subscriptions", paymentHandler.CreateSubscription)
			}

		// User Management (admin-only endpoints)
		users := protected.Group("/users")
		{
			// Get allowed roles for current user
			users.GET("/allowed-roles", authHandler.GetAllowedRoles)
			
			// User management (super-admin/owner/admin only)
			users.POST("", authHandler.CreateUser)             // Create new user
			users.GET("", authHandler.ListUsers)               // List company users
			users.GET("/:id", authHandler.GetUserByID)         // Get user details
			users.PUT("/:id", authHandler.UpdateUser)          // Update user
			users.DELETE("/:id", authHandler.DeactivateUser)   // Deactivate user (owner/super-admin)
			users.PUT("/:id/role", authHandler.ChangeUserRole) // Change user role
		}

		// Analytics and reporting
		analytics := protected.Group("/analytics")
		{
			// Dashboard
			analytics.GET("/dashboard", analyticsHandler.GetDashboard)
			analytics.GET("/dashboard/realtime", analyticsHandler.GetRealTimeDashboard)
			
			// Fuel Analytics
			analytics.GET("/fuel/consumption", analyticsHandler.GetFuelConsumption)
			analytics.GET("/fuel/efficiency", analyticsHandler.GetFuelEfficiency)
			analytics.GET("/fuel/theft", analyticsHandler.GetFuelTheftAlerts)
			analytics.GET("/fuel/optimization", analyticsHandler.GetFuelOptimization)
			
			// Driver Performance
			analytics.GET("/drivers/performance", analyticsHandler.GetDriverPerformance)
			analytics.GET("/drivers/ranking", analyticsHandler.GetDriverRanking)
			analytics.GET("/drivers/behavior", analyticsHandler.GetDriverBehavior)
			analytics.GET("/drivers/recommendations", analyticsHandler.GetDriverRecommendations)
			
			// Fleet Operations
			analytics.GET("/fleet/utilization", analyticsHandler.GetFleetUtilization)
			analytics.GET("/fleet/costs", analyticsHandler.GetFleetCosts)
			analytics.GET("/fleet/maintenance", analyticsHandler.GetMaintenanceInsights)
			
			// Reports
			analytics.POST("/reports/generate", analyticsHandler.GenerateReport)
			analytics.GET("/reports/compliance", analyticsHandler.GetComplianceReport)
			analytics.GET("/reports/export/:id", analyticsHandler.ExportReport)
		}

		// Fleet Management System
		fleet.SetupFleetRoutes(protected, fleetAPI)

		// Geofencing Management System
		geofencing.SetupGeofenceRoutes(protected, geofenceAPI)
		
		// Advanced Analytics System
		advancedanalytics.SetupAnalyticsRoutes(protected, analyticsAPI)

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
		
		// Rate limiting management endpoints (admin only)
		admin := v1.Group("/admin")
		admin.Use(middleware.RoleRequired("admin"))
		{
			rateLimit := admin.Group("/rate-limit")
			{
				rateLimit.GET("/metrics", ratelimit.RateLimitMetricsHandler(rateLimitMonitor))
				rateLimit.GET("/health", ratelimit.RateLimitHealthHandler(rateLimitMonitor))
				rateLimit.GET("/stats", ratelimit.RateLimitStatsHandler(rateLimitMonitor))
				rateLimit.GET("/config", ratelimit.RateLimitConfigHandler(rateLimitManager))
				rateLimit.POST("/config", ratelimit.RateLimitConfigHandler(rateLimitManager))
				rateLimit.PUT("/config/:path/:method", ratelimit.RateLimitConfigHandler(rateLimitManager))
				rateLimit.DELETE("/config/:path/:method", ratelimit.RateLimitConfigHandler(rateLimitManager))
				rateLimit.POST("/reset", ratelimit.RateLimitResetHandler(rateLimitManager))
			}
			
		// Job management endpoints (admin only)
		jobAPI := jobs.NewJobAPI(jobManager)
		jobs.SetupJobRoutes(admin, jobAPI)
		}
		
		// Export endpoints (authenticated users)
		exportAPI := export.NewExportAPI(exportService)
		export.SetupExportRoutes(v1, exportAPI)
	}
}

func setupWebSocket(r *gin.Engine, trackingService *tracking.Service) {
	// WebSocket endpoint for real-time GPS tracking
	r.GET("/ws/tracking", trackingService.HandleWebSocket)
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
