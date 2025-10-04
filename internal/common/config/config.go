package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for our application
type Config struct {
	// Server Configuration
	Port         string
	Host         string
	Environment  string

	// Database Configuration
	DatabaseURL  string
	TimeseriesURL string
	RedisURL     string

	// Authentication & Security
	JWTSecret               string
	JWTAccessExpiry         time.Duration
	JWTRefreshExpiry        time.Duration
	BcryptCost              int

	// Indonesian Payment Integration
	QRISAPIURL              string
	QRISAPIKey              string
	QRISMerchantID          string

	// Bank Transfer APIs
	BCAAPIURL               string
	BCAAPIKey               string
	MandiriAPIURL           string
	MandiriAPIKey           string
	BNIAPIURL               string
	BNIAPIKey               string
	BRIAPIURL               string
	BRIAPIKey               string

	// E-Wallet APIs
	GoPayAPIURL             string
	GoPayAPIKey             string
	OVOAPIURL               string
	OVOAPIKey               string
	DANAAPIURL              string
	DANAAPIKey              string
	ShopeePayAPIURL         string
	ShopeePayAPIKey         string

	// External APIs
	GoogleMapsAPIKey        string
	GoogleMapsAPIURL        string
	HEREAPIKey              string
	HEREAPIURL              string
	WhatsAppAPIURL          string
	WhatsAppAccessToken     string
	WhatsAppPhoneNumberID   string
	SMSAPIURL               string
	SMSAPIKey               string

	// GPS Tracking Configuration
	GPSUpdateInterval       int
	GPSAccuracyThreshold    float64
	DefaultSpeedLimit       int
	HighwaySpeedLimit       int
	CitySpeedLimit          int
	HarshBrakingThreshold   float64
	RapidAccelerationThreshold float64

	// Indonesian Market Configuration
	DefaultCurrency         string
	DefaultLocale           string
	DefaultTimezone         string
	PhoneNumberRegex        string
	LicensePlateRegex       string

	// Monitoring & Logging
	LogLevel                string
	LogFormat               string
	MetricsEnabled          bool
	MetricsPort             string
	HealthCheckInterval     time.Duration
	HealthCheckTimeout      time.Duration

	// Rate Limiting
	RateLimitEnabled        bool
	RateLimitRequestsPerMinute int
	RateLimitBurst          int
	GPSRateLimitRequestsPerMinute int
	GPSRateLimitBurst       int

	// File Upload
	MaxFileSize             int64
	AllowedFileTypes        []string
	UploadDir               string

	// Email Configuration
	SMTPHost                string
	SMTPPort                int
	SMTPUsername            string
	SMTPPassword            string
	SMTPFromAddress         string

	// Development Configuration
	Debug                   bool
	VerboseLogging          bool
	CORSAllowedOrigins      []string
	CORSAllowedMethods      []string
	CORSAllowedHeaders      []string

	// Database Connection Pooling
	DBMaxOpenConns          int
	DBMaxIdleConns          int
	DBConnMaxLifetime       time.Duration

	// Cache Configuration
	CacheTTL                time.Duration
	CacheMaxSize            int

	// Indonesian Compliance
	DataResidencyEnforced   bool
	AuditLoggingEnabled     bool
	AuditLogRetentionDays   int
	PrivacyModeEnabled      bool
	PIIEncryptionEnabled    bool
}

// Load loads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		// Server Configuration
		Port:        getEnv("PORT", "8080"),
		Host:        getEnv("HOST", "localhost"),
		Environment: getEnv("ENVIRONMENT", "development"),

		// Database Configuration
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://fleettracker:password123@localhost:5432/fleettracker?sslmode=disable"),
		TimeseriesURL: getEnv("TIMESERIES_URL", "postgres://fleettracker:password123@localhost:5433/fleettracker_timeseries?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),

		// Authentication & Security
		JWTSecret:        getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		JWTAccessExpiry:  getDurationEnv("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshExpiry: getDurationEnv("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
		BcryptCost:       getIntEnv("BCRYPT_COST", 12),

		// Indonesian Payment Integration
		QRISAPIURL:     getEnv("QRIS_API_URL", "https://api.qris.id"),
		QRISAPIKey:     getEnv("QRIS_API_KEY", ""),
		QRISMerchantID: getEnv("QRIS_MERCHANT_ID", ""),

		// Bank Transfer APIs
		BCAAPIURL:     getEnv("BCA_API_URL", "https://api.bca.co.id"),
		BCAAPIKey:     getEnv("BCA_API_KEY", ""),
		MandiriAPIURL: getEnv("MANDIRI_API_URL", "https://api.mandiri.co.id"),
		MandiriAPIKey: getEnv("MANDIRI_API_KEY", ""),
		BNIAPIURL:     getEnv("BNI_API_URL", "https://api.bni.co.id"),
		BNIAPIKey:     getEnv("BNI_API_KEY", ""),
		BRIAPIURL:     getEnv("BRI_API_URL", "https://api.bri.co.id"),
		BRIAPIKey:     getEnv("BRI_API_KEY", ""),

		// E-Wallet APIs
		GoPayAPIURL:     getEnv("GOPAY_API_URL", "https://api.gopay.co.id"),
		GoPayAPIKey:     getEnv("GOPAY_API_KEY", ""),
		OVOAPIURL:       getEnv("OVO_API_URL", "https://api.ovo.id"),
		OVOAPIKey:       getEnv("OVO_API_KEY", ""),
		DANAAPIURL:      getEnv("DANA_API_URL", "https://api.dana.id"),
		DANAAPIKey:      getEnv("DANA_API_KEY", ""),
		ShopeePayAPIURL: getEnv("SHOPEEPAY_API_URL", "https://api.shopee.co.id"),
		ShopeePayAPIKey: getEnv("SHOPEEPAY_API_KEY", ""),

		// External APIs
		GoogleMapsAPIKey:      getEnv("GOOGLE_MAPS_API_KEY", ""),
		GoogleMapsAPIURL:      getEnv("GOOGLE_MAPS_API_URL", "https://maps.googleapis.com"),
		HEREAPIKey:            getEnv("HERE_API_KEY", ""),
		HEREAPIURL:            getEnv("HERE_API_URL", "https://api.here.com"),
		WhatsAppAPIURL:        getEnv("WHATSAPP_API_URL", "https://graph.facebook.com/v18.0"),
		WhatsAppAccessToken:   getEnv("WHATSAPP_ACCESS_TOKEN", ""),
		WhatsAppPhoneNumberID: getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
		SMSAPIURL:             getEnv("SMS_API_URL", "https://api.sms-gateway.id"),
		SMSAPIKey:             getEnv("SMS_API_KEY", ""),

		// GPS Tracking Configuration
		GPSUpdateInterval:         getIntEnv("GPS_UPDATE_INTERVAL", 30),
		GPSAccuracyThreshold:      getFloatEnv("GPS_ACCURACY_THRESHOLD", 5.0),
		DefaultSpeedLimit:         getIntEnv("DEFAULT_SPEED_LIMIT", 100),
		HighwaySpeedLimit:         getIntEnv("HIGHWAY_SPEED_LIMIT", 120),
		CitySpeedLimit:            getIntEnv("CITY_SPEED_LIMIT", 60),
		HarshBrakingThreshold:     getFloatEnv("HARSH_BRAKING_THRESHOLD", 0.4),
		RapidAccelerationThreshold: getFloatEnv("RAPID_ACCELERATION_THRESHOLD", 0.3),

		// Indonesian Market Configuration
		DefaultCurrency:   getEnv("DEFAULT_CURRENCY", "IDR"),
		DefaultLocale:     getEnv("DEFAULT_LOCALE", "id_ID"),
		DefaultTimezone:   getEnv("DEFAULT_TIMEZONE", "Asia/Jakarta"),
		PhoneNumberRegex:  getEnv("PHONE_NUMBER_REGEX", `^(\+62|62|0)[0-9]{9,12}$`),
		LicensePlateRegex: getEnv("LICENSE_PLATE_REGEX", `^[A-Z]{1,2}\s*[0-9]{1,4}\s*[A-Z]{1,3}$`),

		// Monitoring & Logging
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		LogFormat:           getEnv("LOG_FORMAT", "json"),
		MetricsEnabled:      getBoolEnv("METRICS_ENABLED", true),
		MetricsPort:         getEnv("METRICS_PORT", "9090"),
		HealthCheckInterval: getDurationEnv("HEALTH_CHECK_INTERVAL", 30*time.Second),
		HealthCheckTimeout:  getDurationEnv("HEALTH_CHECK_TIMEOUT", 5*time.Second),

		// Rate Limiting
		RateLimitEnabled:              getBoolEnv("RATE_LIMIT_ENABLED", true),
		RateLimitRequestsPerMinute:    getIntEnv("RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
		RateLimitBurst:                getIntEnv("RATE_LIMIT_BURST", 20),
		GPSRateLimitRequestsPerMinute: getIntEnv("GPS_RATE_LIMIT_REQUESTS_PER_MINUTE", 1000),
		GPSRateLimitBurst:             getIntEnv("GPS_RATE_LIMIT_BURST", 100),

		// File Upload
		MaxFileSize:      getInt64Env("MAX_FILE_SIZE", 10*1024*1024), // 10MB
		AllowedFileTypes: strings.Split(getEnv("ALLOWED_FILE_TYPES", "jpg,jpeg,png,pdf,doc,docx"), ","),
		UploadDir:        getEnv("UPLOAD_DIR", "./uploads"),

		// Email Configuration
		SMTPHost:        getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:        getIntEnv("SMTP_PORT", 587),
		SMTPUsername:    getEnv("SMTP_USERNAME", ""),
		SMTPPassword:    getEnv("SMTP_PASSWORD", ""),
		SMTPFromAddress: getEnv("SMTP_FROM_ADDRESS", "noreply@fleettracker.id"),

		// Development Configuration
		Debug:               getBoolEnv("DEBUG", true),
		VerboseLogging:      getBoolEnv("VERBOSE_LOGGING", false),
		CORSAllowedOrigins:  strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000"), ","),
		CORSAllowedMethods:  strings.Split(getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"), ","),
		CORSAllowedHeaders:  strings.Split(getEnv("CORS_ALLOWED_HEADERS", "Origin,Content-Type,Authorization"), ","),

		// Database Connection Pooling
		DBMaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 100),
		DBMaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", time.Hour),

		// Cache Configuration
		CacheTTL:     getDurationEnv("CACHE_TTL", 5*time.Minute),
		CacheMaxSize: getIntEnv("CACHE_MAX_SIZE", 1000),

		// Indonesian Compliance
		DataResidencyEnforced: getBoolEnv("DATA_RESIDENCY_ENFORCED", true),
		AuditLoggingEnabled:   getBoolEnv("AUDIT_LOGGING_ENABLED", true),
		AuditLogRetentionDays: getIntEnv("AUDIT_LOG_RETENTION_DAYS", 2555), // 7 years
		PrivacyModeEnabled:    getBoolEnv("PRIVACY_MODE_ENABLED", true),
		PIIEncryptionEnabled:  getBoolEnv("PII_ENCRYPTION_ENABLED", true),
	}

	return cfg
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
