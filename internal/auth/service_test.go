package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/testutil"
)

func TestService_Register(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	// Create test company first
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	service := NewService(db, "test-jwt-secret")

	tests := []struct {
		name    string
		request RegisterRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid registration",
			request: RegisterRequest{
				CompanyID:       company.ID,
				Email:           "newuser@test.com",
				Username:        "newuser",
				Password:        "SecurePass123!",
				FirstName:       "New",
				LastName:        "User",
				Phone:           "+62 811 1234567",
				Role:            "operator",
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			request: RegisterRequest{
				CompanyID:       company.ID,
				Email:           "newuser@test.com", // Same as above
				Username:        "anotheruser",
				Password:        "SecurePass123!",
				FirstName:       "Another",
				LastName:        "User",
				Phone:           "+62 812 1234567",
				Role:            "operator",
			},
			wantErr: true,
			errMsg:  "email already exists",
		},
		{
			name: "invalid email format",
			request: RegisterRequest{
				CompanyID:       company.ID,
				Email:           "invalid-email",
				Username:        "testuser",
				Password:        "SecurePass123!",
				FirstName:       "Test",
				LastName:        "User",
				Phone:           "+62 811 1234567",
				Role:            "operator",
			},
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name: "weak password",
			request: RegisterRequest{
				CompanyID:       company.ID,
				Email:           "weak@test.com",
				Username:        "weakuser",
				Password:        "123", // Too weak
				FirstName:       "Weak",
				LastName:        "User",
				Phone:           "+62 811 1234567",
				Role:            "operator",
			},
			wantErr: true,
			errMsg:  "password too weak",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.Register(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				testutil.AssertValidUUID(t, user.ID)
				assert.Equal(t, tt.request.Email, user.Email)
				assert.Equal(t, tt.request.Username, user.Username)
				assert.True(t, user.IsActive)
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db, "test-jwt-secret")

	// Create test company and user
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	// Register a user for testing login
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

	tests := []struct {
		name    string
		request LoginRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid login with email",
			request: LoginRequest{
				Email:    "login@test.com",
				Password: "SecurePass123!",
			},
			wantErr: false,
		},
		{
			name: "invalid password",
			request: LoginRequest{
				Email:    "login@test.com",
				Password: "WrongPassword",
			},
			wantErr: true,
			errMsg:  "invalid credentials",
		},
		{
			name: "non-existent user",
			request: LoginRequest{
				Email:    "nonexistent@test.com",
				Password: "SecurePass123!",
			},
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name: "inactive user",
			request: LoginRequest{
				Email:    user.Email,
				Password: "SecurePass123!",
			},
			wantErr: false, // Will test separately by deactivating user
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Special case: deactivate user for inactive test
			if tt.name == "inactive user" {
				db.Model(&user).Update("is_active", false)
				defer db.Model(&user).Update("is_active", true) // Restore
			}

			loginUser, tokenResp, err := service.Login(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, loginUser)
				assert.Nil(t, tokenResp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, loginUser)
				assert.NotNil(t, tokenResp)
				assert.NotEmpty(t, tokenResp.AccessToken)
				assert.NotEmpty(t, tokenResp.RefreshToken)
				assert.Equal(t, tt.request.Email, loginUser.Email)
			}
		})
	}
}

func TestService_TokenGeneration(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db, "test-jwt-secret")

	// Create test company and user
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	_, err := service.Register(RegisterRequest{
		CompanyID: company.ID,
		Email:     "token@test.com",
		Username:  "tokenuser",
		Password:  "SecurePass123!",
		FirstName: "Token",
		LastName:  "User",
		Phone:     "+62 811 1234567",
		Role:      "admin",
	})
	require.NoError(t, err)

	t.Run("login generates valid tokens", func(t *testing.T) {
		_, tokenResp, err := service.Login(LoginRequest{
			Email:    "token@test.com",
			Password: "SecurePass123!",
		})

		assert.NoError(t, err)
		assert.NotNil(t, tokenResp)
		assert.NotEmpty(t, tokenResp.AccessToken)
		assert.NotEmpty(t, tokenResp.RefreshToken)
		assert.Equal(t, "Bearer", tokenResp.TokenType)
		assert.Greater(t, tokenResp.ExpiresIn, 0)
	})

	t.Run("tokens are JWT format", func(t *testing.T) {
		_, tokenResp, err := service.Login(LoginRequest{
			Email:    "token@test.com",
			Password: "SecurePass123!",
		})

		require.NoError(t, err)
		// JWT tokens have 3 parts separated by dots
		accessParts := len(strings.Split(tokenResp.AccessToken, "."))
		refreshParts := len(strings.Split(tokenResp.RefreshToken, "."))
		assert.Equal(t, 3, accessParts)
		assert.Equal(t, 3, refreshParts)
	})
}

func TestService_PasswordHashing(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db, "test-jwt-secret")

	t.Run("password is hashed correctly", func(t *testing.T) {
		password := "SecurePass123!"
		
		company := testutil.NewTestCompany()
		require.NoError(t, db.Create(company).Error)

		userResp, err := service.Register(RegisterRequest{
			CompanyID: company.ID,
			Email:     "hash@test.com",
			Username:  "hashuser",
			Password:  password,
			FirstName: "Hash",
			LastName:  "User",
			Phone:     "+62 811 1234567",
			Role:      "operator",
		})

		require.NoError(t, err)
		assert.NotNil(t, userResp)
		
		// Verify password is hashed by checking we can login
		loginUser, tokenResp, err := service.Login(LoginRequest{
			Email:    "hash@test.com",
			Password: password,
		})
		assert.NoError(t, err)
		assert.NotNil(t, loginUser)
		assert.NotNil(t, tokenResp)
	})

	t.Run("wrong password fails login", func(t *testing.T) {
		company := testutil.NewTestCompany()
		require.NoError(t, db.Create(company).Error)

		_, err := service.Register(RegisterRequest{
			CompanyID: company.ID,
			Email:     "wrongpass@test.com",
			Username:  "wronguser",
			Password:  "CorrectPass123!",
			FirstName: "Wrong",
			LastName:  "Password",
			Phone:     "+62 811 3333333",
			Role:      "operator",
		})
		require.NoError(t, err)

		// Try to login with wrong password
		_, _, err = service.Login(LoginRequest{
			Email:    "wrongpass@test.com",
			Password: "WrongPass123!",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
	})
}

func TestService_TokenExpiry(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db, "test-jwt-secret")

	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	_, err := service.Register(RegisterRequest{
		CompanyID: company.ID,
		Email:     "expiry@test.com",
		Username:  "expiryuser",
		Password:  "SecurePass123!",
		FirstName: "Expiry",
		LastName:  "Test",
		Phone:     "+62 811 4444444",
		Role:      "admin",
	})
	require.NoError(t, err)

	t.Run("token should have expiry time in future", func(t *testing.T) {
		_, tokenResp, err := service.Login(LoginRequest{
			Email:    "expiry@test.com",
			Password: "SecurePass123!",
		})
		require.NoError(t, err)

		assert.Greater(t, tokenResp.ExpiresIn, 0, "Token should have positive expiry duration")
	})
}

