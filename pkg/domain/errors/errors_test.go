package errors

import (
	"testing"
)

func TestDomainError(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		message  string
		field    string
		details  map[string]string
		expected string
	}{
		{
			name:     "validation error",
			code:     ErrorCodeValidation,
			field:    "email",
			message:  "invalid format",
			expected: "validation failed for field 'email': invalid format",
		},
		{
			name:     "business error",
			code:     ErrorCodeBusiness,
			message:  "user not found",
			expected: "domain error [BUSINESS_ERROR]: user not found",
		},
		{
			name:    "validation error with details",
			code:    ErrorCodeValidation,
			field:   "password",
			message: "password too weak",
			details: map[string]string{
				"min_length": "8",
				"required":   "uppercase,lowercase,number",
			},
			expected: "validation failed for field 'password': password too weak",
		},
		{
			name:     "plugin error",
			code:     ErrorCodePlugin,
			message:  "plugin initialization failed",
			expected: "plugin error: plugin initialization failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DomainError{
				Code:    tt.code,
				Message: tt.message,
				Field:   tt.field,
				Details: tt.details,
			}

			if err.Error() != tt.expected {
				t.Errorf("DomainError.Error() = %v, want %v", err.Error(), tt.expected)
			}

			if !IsDomainError(err) {
				t.Error("IsDomainError() should return true for DomainError")
			}

			if err.Code != tt.code {
				t.Errorf("DomainError.Code = %v, want %v", err.Code, tt.code)
			}
		})
	}
}

func TestInfrastructureError(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		component string
		operation string
		message   string
		details   map[string]string
		expected  string
	}{
		{
			name:      "database error",
			code:      ErrorCodeDatabase,
			component: "mysql",
			operation: "connect",
			message:   "connection timeout",
			expected:  "infrastructure error [DATABASE_ERROR] in mysql.connect: connection timeout",
		},
		{
			name:      "network error with details",
			code:      ErrorCodeNetwork,
			component: "redis",
			operation: "set",
			message:   "write failed",
			details: map[string]string{
				"key":     "user:123",
				"timeout": "5s",
			},
			expected: "infrastructure error [NETWORK_ERROR] in redis.set: write failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InfrastructureError{
				Code:      tt.code,
				Component: tt.component,
				Operation: tt.operation,
				Message:   tt.message,
				Details:   tt.details,
			}

			if err.Error() != tt.expected {
				t.Errorf("InfrastructureError.Error() = %v, want %v", err.Error(), tt.expected)
			}

			if !IsInfrastructureError(err) {
				t.Error("IsInfrastructureError() should return true for InfrastructureError")
			}

			if err.Component != tt.component {
				t.Errorf("InfrastructureError.Component = %v, want %v", err.Component, tt.component)
			}

			if err.Operation != tt.operation {
				t.Errorf("InfrastructureError.Operation = %v, want %v", err.Operation, tt.operation)
			}
		})
	}
}

func TestErrorCreationFunctions(t *testing.T) {
	// Test specific error creation functions
	validationErr := NewValidationError("email", "invalid format")
	if validationErr.Code != ErrorCodeValidation {
		t.Errorf("NewValidationError should create error with VALIDATION code")
	}

	businessErr := NewBusinessError("user not found")
	if businessErr.Code != ErrorCodeBusiness {
		t.Errorf("NewBusinessError should create error with BUSINESS code")
	}

	pluginErr := NewPluginError("csv_importer", "Init", "config error")
	if pluginErr.Code != ErrorCodePlugin {
		t.Errorf("NewPluginError should create error with PLUGIN code")
	}

	authErr := NewAuthError("invalid token")
	if authErr.Code != ErrorCodeAuth {
		t.Errorf("NewAuthError should create error with AUTH code")
	}

	dbErr := NewDatabaseError("mysql", "connect", "timeout")
	if dbErr.Code != ErrorCodeDatabase {
		t.Errorf("NewDatabaseError should create error with DATABASE code")
	}

	networkErr := NewNetworkError("redis", "get", "connection failed")
	if networkErr.Code != ErrorCodeNetwork {
		t.Errorf("NewNetworkError should create error with NETWORK code")
	}
}

func TestErrorTypeChecking(t *testing.T) {
	domainErr := NewValidationError("field", "validation failed")
	infraErr := NewDatabaseError("mysql", "connect", "timeout")

	// Test type checking
	if !IsDomainError(domainErr) {
		t.Error("IsDomainError() should return true for DomainError")
	}
	if IsDomainError(infraErr) {
		t.Error("IsDomainError() should return false for InfrastructureError")
	}

	if !IsInfrastructureError(infraErr) {
		t.Error("IsInfrastructureError() should return true for InfrastructureError")
	}
	if IsInfrastructureError(domainErr) {
		t.Error("IsInfrastructureError() should return false for DomainError")
	}

	// Test specific error type checking
	validationErr := NewValidationError("field", "message")
	if !IsValidationError(validationErr) {
		t.Error("IsValidationError() should return true for validation error")
	}
	if IsValidationError(infraErr) {
		t.Error("IsValidationError() should return false for infrastructure error")
	}

	businessErr := NewBusinessError("message")
	if !IsBusinessError(businessErr) {
		t.Error("IsBusinessError() should return true for business error")
	}
	if IsBusinessError(validationErr) {
		t.Error("IsBusinessError() should return false for validation error")
	}

	pluginErr := NewPluginError("plugin", "phase", "message")
	if !IsPluginError(pluginErr) {
		t.Error("IsPluginError() should return true for plugin error")
	}
	if IsPluginError(businessErr) {
		t.Error("IsPluginError() should return false for business error")
	}

	authErr := NewAuthError("message")
	if !IsAuthError(authErr) {
		t.Error("IsAuthError() should return true for auth error")
	}
	if IsAuthError(pluginErr) {
		t.Error("IsAuthError() should return false for plugin error")
	}

	dbErr := NewDatabaseError("db", "op", "message")
	if !IsDatabaseError(dbErr) {
		t.Error("IsDatabaseError() should return true for database error")
	}
	if IsDatabaseError(domainErr) {
		t.Error("IsDatabaseError() should return false for domain error")
	}

	networkErr := NewNetworkError("net", "op", "message")
	if !IsNetworkError(networkErr) {
		t.Error("IsNetworkError() should return true for network error")
	}
	if IsNetworkError(dbErr) {
		t.Error("IsNetworkError() should return false for database error")
	}
}

func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"validation", ErrorCodeValidation},
		{"business", ErrorCodeBusiness},
		{"plugin", ErrorCodePlugin},
		{"auth", ErrorCodeAuth},
		{"not_found", ErrorCodeNotFound},
		{"database", ErrorCodeDatabase},
		{"network", ErrorCodeNetwork},
		{"filesystem", ErrorCodeFileSystem},
		{"external", ErrorCodeExternal},
		{"config", ErrorCodeConfig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code == "" {
				t.Errorf("Error code %s should not be empty", tt.name)
			}
		})
	}
}
