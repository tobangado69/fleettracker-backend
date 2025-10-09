package validators

import (
	"testing"
)

func TestValidateNIK(t *testing.T) {
	tests := []struct {
		name    string
		nik     string
		wantErr bool
	}{
		{
			name:    "valid NIK",
			nik:     "3201012345678901",
			wantErr: false,
		},
		{
			name:    "valid NIK with female code",
			nik:     "3241012345678901",
			wantErr: false,
		},
		{
			name:    "too short",
			nik:     "320101234567890",
			wantErr: true,
		},
		{
			name:    "too long",
			nik:     "32010123456789012",
			wantErr: true,
		},
		{
			name:    "contains non-digits",
			nik:     "3201ABC234567890",
			wantErr: true,
		},
		{
			name:    "invalid month code",
			nik:     "3213012345678901", // Month 13 invalid
			wantErr: true,
		},
		{
			name:    "invalid district code",
			nik:     "0001012345678901", // District 00 invalid
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNIK(tt.nik)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNIK() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSIM(t *testing.T) {
	tests := []struct {
		name    string
		sim     string
		wantErr bool
	}{
		{
			name:    "valid SIM",
			sim:     "123456789012",
			wantErr: false,
		},
		{
			name:    "too short",
			sim:     "12345678901",
			wantErr: true,
		},
		{
			name:    "too long",
			sim:     "1234567890123",
			wantErr: true,
		},
		{
			name:    "contains letters",
			sim:     "123456ABC012",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSIM(tt.sim)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSIM() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePlateNumber(t *testing.T) {
	tests := []struct {
		name    string
		plate   string
		wantErr bool
	}{
		{
			name:    "valid Jakarta plate",
			plate:   "B 1234 ABC",
			wantErr: false,
		},
		{
			name:    "valid without spaces",
			plate:   "B1234ABC",
			wantErr: false,
		},
		{
			name:    "valid 2-letter prefix",
			plate:   "DK 1234 AB",
			wantErr: false,
		},
		{
			name:    "valid lowercase",
			plate:   "b 1234 abc",
			wantErr: false,
		},
		{
			name:    "invalid format - no letters",
			plate:   "1234",
			wantErr: true,
		},
		{
			name:    "invalid prefix",
			plate:   "X 1234 ABC", // X not a valid province code
			wantErr: true,
		},
		{
			name:    "invalid - only 3 chars",
			plate:   "B12",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlateNumber(tt.plate)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlateNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNPWP(t *testing.T) {
	tests := []struct {
		name    string
		npwp    string
		wantErr bool
	}{
		{
			name:    "valid NPWP without formatting",
			npwp:    "123456789012345",
			wantErr: false,
		},
		{
			name:    "valid NPWP with formatting",
			npwp:    "12.345.678.9-012.345",
			wantErr: false,
		},
		{
			name:    "too short",
			npwp:    "12345678901234",
			wantErr: true,
		},
		{
			name:    "too long",
			npwp:    "1234567890123456",
			wantErr: true,
		},
		{
			name:    "contains letters",
			npwp:    "123456ABC012345",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNPWP(tt.npwp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNPWP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePhoneNumber(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{
			name:    "valid international format",
			phone:   "+628123456789",
			wantErr: false,
		},
		{
			name:    "valid local format",
			phone:   "08123456789",
			wantErr: false,
		},
		{
			name:    "valid without prefix",
			phone:   "8123456789",
			wantErr: false,
		},
		{
			name:    "valid with spaces",
			phone:   "+62 812 3456 789",
			wantErr: false,
		},
		{
			name:    "too short",
			phone:   "+628123456",
			wantErr: true,
		},
		{
			name:    "invalid prefix",
			phone:   "+6212345678", // Should start with 8
			wantErr: true,
		},
		{
			name:    "contains letters",
			phone:   "+62812ABC6789",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePhoneNumber(tt.phone)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePhoneNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatPlateNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "add spaces",
			input:    "B1234ABC",
			expected: "B 1234 ABC",
		},
		{
			name:     "already formatted",
			input:    "B 1234 ABC",
			expected: "B 1234 ABC",
		},
		{
			name:     "lowercase to uppercase",
			input:    "b 1234 abc",
			expected: "B 1234 ABC",
		},
		{
			name:     "two-letter prefix",
			input:    "DK1234AB",
			expected: "DK 1234 AB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPlateNumber(tt.input)
			if result != tt.expected {
				t.Errorf("FormatPlateNumber() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatPhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "local to international",
			input:    "08123456789",
			expected: "+628123456789",
		},
		{
			name:     "without prefix",
			input:    "8123456789",
			expected: "+628123456789",
		},
		{
			name:     "already international",
			input:    "+628123456789",
			expected: "+628123456789",
		},
		{
			name:     "with spaces",
			input:    "0812 3456 789",
			expected: "+628123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPhoneNumber(tt.input)
			if result != tt.expected {
				t.Errorf("FormatPhoneNumber() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "valid with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "valid with plus",
			email:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "missing @",
			email:   "userexample.com",
			wantErr: true,
		},
		{
			name:    "missing domain",
			email:   "user@",
			wantErr: true,
		},
		{
			name:    "missing user",
			email:   "@example.com",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			email:   "user name@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "Password123",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Pass1",
			wantErr:  true,
		},
		{
			name:     "no uppercase",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "no lowercase",
			password: "PASSWORD123",
			wantErr:  true,
		},
		{
			name:     "no digit",
			password: "Password",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

