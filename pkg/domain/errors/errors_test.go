package errors

import (
	"testing"
)

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		message  string
		details  map[string]string
		expected string
	}{
		{
			name:     "simple validation error",
			field:    "email",
			message:  "invalid format",
			expected: "validation failed for field 'email': invalid format",
		},
		{
			name:    "validation error with details",
			field:   "password",
			message: "too weak",
			details: map[string]string{
				"min_length": "8",
				"required":   "uppercase,lowercase,number",
			},
			expected: "validation failed for field 'password': too weak (details: map[min_length:8 required:uppercase,lowercase,number])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err ValidationError
			if tt.details != nil {
				err = NewValidationError(tt.field, tt.message, tt.details)
			} else {
				err = NewValidationError(tt.field, tt.message)
			}

			if err.Error() != tt.expected {
				t.Errorf("ValidationError.Error() = %v, want %v", err.Error(), tt.expected)
			}

			if !IsValidationError(err) {
				t.Error("IsValidationError() should return true for ValidationError")
			}
		})
	}
}

func TestBusinessError(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		message  string
		details  map[string]string
		expected string
	}{
		{
			name:     "simple business error",
			code:     "USER_NOT_FOUND",
			message:  "user with ID 123 not found",
			expected: "business error [USER_NOT_FOUND]: user with ID 123 not found",
		},
		{
			name:    "business error with details",
			code:    "QUOTA_EXCEEDED",
			message: "user has exceeded quota",
			details: map[string]string{
				"current": "105",
				"limit":   "100",
			},
			expected: "business error [QUOTA_EXCEEDED]: user has exceeded quota (details: map[current:105 limit:100])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err BusinessError
			if tt.details != nil {
				err = NewBusinessError(tt.code, tt.message, tt.details)
			} else {
				err = NewBusinessError(tt.code, tt.message)
			}

			if err.Error() != tt.expected {
				t.Errorf("BusinessError.Error() = %v, want %v", err.Error(), tt.expected)
			}

			if !IsBusinessError(err) {
				t.Error("IsBusinessError() should return true for BusinessError")
			}
		})
	}
}

func TestInfrastructureError(t *testing.T) {
	tests := []struct {
		name      string
		component string
		operation string
		message   string
		details   map[string]string
		expected  string
	}{
		{
			name:      "simple infrastructure error",
			component: "database",
			operation: "connect",
			message:   "connection timeout",
			expected:  "infrastructure error in database.connect: connection timeout",
		},
		{
			name:      "infrastructure error with details",
			component: "redis",
			operation: "set",
			message:   "write failed",
			details: map[string]string{
				"key":     "user:123",
				"timeout": "5s",
			},
			expected: "infrastructure error in redis.set: write failed (details: map[key:user:123 timeout:5s])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err InfrastructureError
			if tt.details != nil {
				err = NewInfrastructureError(tt.component, tt.operation, tt.message, tt.details)
			} else {
				err = NewInfrastructureError(tt.component, tt.operation, tt.message)
			}

			if err.Error() != tt.expected {
				t.Errorf("InfrastructureError.Error() = %v, want %v", err.Error(), tt.expected)
			}

			if !IsInfrastructureError(err) {
				t.Error("IsInfrastructureError() should return true for InfrastructureError")
			}
		})
	}
}

func TestPluginError(t *testing.T) {
	tests := []struct {
		name       string
		pluginName string
		phase      string
		message    string
		details    map[string]string
		expected   string
	}{
		{
			name:       "simple plugin error",
			pluginName: "csv_importer",
			phase:      "Init",
			message:    "invalid configuration",
			expected:   "plugin error in csv_importer during Init: invalid configuration",
		},
		{
			name:       "plugin error with details",
			pluginName: "threshold_detector",
			phase:      "Execute",
			message:    "threshold validation failed",
			details: map[string]string{
				"field":     "temperature",
				"value":     "150",
				"threshold": "100",
			},
			expected: "plugin error in threshold_detector during Execute: threshold validation failed (details: map[field:temperature threshold:100 value:150])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err PluginError
			if tt.details != nil {
				err = NewPluginError(tt.pluginName, tt.phase, tt.message, tt.details)
			} else {
				err = NewPluginError(tt.pluginName, tt.phase, tt.message)
			}

			if err.Error() != tt.expected {
				t.Errorf("PluginError.Error() = %v, want %v", err.Error(), tt.expected)
			}

			if !IsPluginError(err) {
				t.Error("IsPluginError() should return true for PluginError")
			}
		})
	}
}

func TestErrorTypeChecking(t *testing.T) {
	validationErr := NewValidationError("email", "invalid format")
	businessErr := NewBusinessError("USER_NOT_FOUND", "user not found")
	infraErr := NewInfrastructureError("database", "connect", "timeout")
	pluginErr := NewPluginError("csv_importer", "Init", "config error")

	// Test cross-type checking
	if IsBusinessError(validationErr) {
		t.Error("ValidationError should not be identified as BusinessError")
	}
	if IsInfrastructureError(validationErr) {
		t.Error("ValidationError should not be identified as InfrastructureError")
	}
	if IsPluginError(validationErr) {
		t.Error("ValidationError should not be identified as PluginError")
	}

	if IsValidationError(businessErr) {
		t.Error("BusinessError should not be identified as ValidationError")
	}
	if IsInfrastructureError(businessErr) {
		t.Error("BusinessError should not be identified as InfrastructureError")
	}
	if IsPluginError(businessErr) {
		t.Error("BusinessError should not be identified as PluginError")
	}

	if IsValidationError(infraErr) {
		t.Error("InfrastructureError should not be identified as ValidationError")
	}
	if IsBusinessError(infraErr) {
		t.Error("InfrastructureError should not be identified as BusinessError")
	}
	if IsPluginError(infraErr) {
		t.Error("InfrastructureError should not be identified as PluginError")
	}

	if IsValidationError(pluginErr) {
		t.Error("PluginError should not be identified as ValidationError")
	}
	if IsBusinessError(pluginErr) {
		t.Error("PluginError should not be identified as BusinessError")
	}
	if IsInfrastructureError(pluginErr) {
		t.Error("PluginError should not be identified as InfrastructureError")
	}
}
