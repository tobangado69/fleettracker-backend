package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/database"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/testutil"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.Default()
}

func TestHandler_Register(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	redisClient, _ := database.ConnectRedis("redis://localhost:6379")
	service := NewService(db, redisClient, "test-jwt-secret")
	handler := NewHandler(service)

	router := setupTestRouter()
	router.POST("/auth/register", handler.Register)

	// Create test company
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	tests := []struct {
		name       string
		payload    map[string]interface{}
		wantStatus int
		checkBody  func(*testing.T, map[string]interface{})
	}{
		{
			name: "successful registration",
			payload: map[string]interface{}{
				"company_id": company.ID,
				"email":      "newuser@test.com",
				"username":   "newuser",
				"password":   "SecurePass123!",
				"first_name": "New",
				"last_name":  "User",
				"phone":      "+62 811 1234567",
				"role":       "operator",
			},
			wantStatus: http.StatusCreated,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "newuser@test.com", body["email"])
				assert.Equal(t, "newuser", body["username"])
				assert.NotEmpty(t, body["id"])
			},
		},
		{
			name: "invalid email format",
			payload: map[string]interface{}{
				"company_id": company.ID,
				"email":      "invalid-email",
				"username":   "testuser",
				"password":   "SecurePass123!",
				"first_name": "Test",
				"last_name":  "User",
				"phone":      "+62 811 1234567",
				"role":       "operator",
			},
			wantStatus: http.StatusBadRequest,
			checkBody:  nil,
		},
		{
			name: "missing required fields",
			payload: map[string]interface{}{
				"email": "test@test.com",
			},
			wantStatus: http.StatusBadRequest,
			checkBody:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.checkBody != nil && w.Code == http.StatusCreated {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				tt.checkBody(t, response)
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	redisClient, _ := database.ConnectRedis("redis://localhost:6379")
	service := NewService(db, redisClient, "test-jwt-secret")
	handler := NewHandler(service)

	router := setupTestRouter()
	router.POST("/auth/login", handler.Login)

	// Create test company and user
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	user, err := service.Register(RegisterRequest{
		CompanyID: company.ID,
		Email:     "login@test.com",
		Username:  "loginuser",
		Password:  "SecurePass123!",
		FirstName: "Login",
		LastName:  "User",
		Phone:     "+62 811 1234567",
		Role:      "admin",
	})
	require.NoError(t, err)
	require.NotNil(t, user)

	tests := []struct {
		name       string
		payload    map[string]interface{}
		wantStatus int
		checkBody  func(*testing.T, map[string]interface{})
	}{
		{
			name: "successful login",
			payload: map[string]interface{}{
				"email":    "login@test.com",
				"password": "SecurePass123!",
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.NotEmpty(t, body["access_token"])
				assert.NotEmpty(t, body["refresh_token"])
				assert.Equal(t, "Bearer", body["token_type"])
			},
		},
		{
			name: "invalid password",
			payload: map[string]interface{}{
				"email":    "login@test.com",
				"password": "WrongPassword",
			},
			wantStatus: http.StatusUnauthorized,
			checkBody:  nil,
		},
		{
			name: "non-existent user",
			payload: map[string]interface{}{
				"email":    "nonexistent@test.com",
				"password": "SecurePass123!",
			},
			wantStatus: http.StatusUnauthorized,
			checkBody:  nil,
		},
		{
			name: "missing credentials",
			payload: map[string]interface{}{
				"email": "login@test.com",
			},
			wantStatus: http.StatusBadRequest,
			checkBody:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.checkBody != nil && w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				tt.checkBody(t, response)
			}
		})
	}
}

func TestHandler_GetProfile(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	redisClient, _ := database.ConnectRedis("redis://localhost:6379")
	service := NewService(db, redisClient, "test-jwt-secret")
	handler := NewHandler(service)

	router := setupTestRouter()
	router.GET("/auth/profile", handler.GetProfile)

	// Create test company and user
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	user, err := service.Register(RegisterRequest{
		CompanyID: company.ID,
		Email:     "profile@test.com",
		Username:  "profileuser",
		Password:  "SecurePass123!",
		FirstName: "Profile",
		LastName:  "User",
		Phone:     "+62 811 1234567",
		Role:      "admin",
	})
	require.NoError(t, err)

	// Generate token
	_, tokenResp, err := service.Login(LoginRequest{
		Email:    "profile@test.com",
		Password: "SecurePass123!",
	})
	require.NoError(t, err)

	t.Run("get profile with valid token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/profile", nil)
		req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, user.Email, response["email"])
		assert.Equal(t, user.Username, response["username"])
	})

	t.Run("get profile without token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/profile", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("get profile with invalid token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/profile", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHandler_UpdateProfile(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	redisClient, _ := database.ConnectRedis("redis://localhost:6379")
	service := NewService(db, redisClient, "test-jwt-secret")
	handler := NewHandler(service)

	router := setupTestRouter()
	router.PUT("/auth/profile", handler.UpdateProfile)

	// Create test company and user
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	_, err := service.Register(RegisterRequest{
		CompanyID: company.ID,
		Email:     "update@test.com",
		Username:  "updateuser",
		Password:  "SecurePass123!",
		FirstName: "Update",
		LastName:  "User",
		Phone:     "+62 811 1234567",
		Role:      "admin",
	})
	require.NoError(t, err)

	// Generate token
	_, tokenResp, err := service.Login(LoginRequest{
		Email:    "update@test.com",
		Password: "SecurePass123!",
	})
	require.NoError(t, err)

	t.Run("update profile successfully", func(t *testing.T) {
		payload := map[string]interface{}{
			"first_name": "Updated",
			"last_name":  "Name",
			"phone":      "+62 812 9999999",
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/auth/profile", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Updated", response["first_name"])
		assert.Equal(t, "Name", response["last_name"])
	})
}

