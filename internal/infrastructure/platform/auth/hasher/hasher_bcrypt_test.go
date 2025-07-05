package hasher

import (
	"context"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestNewBcryptPasswordHasher(t *testing.T) {
	tests := []struct {
		name        string
		cost        int
		expectError bool
	}{
		{
			name:        "Valid cost - minimum",
			cost:        bcrypt.MinCost,
			expectError: false,
		},
		{
			name:        "Valid cost - default",
			cost:        bcrypt.DefaultCost,
			expectError: false,
		},
		{
			name:        "Valid cost - maximum",
			cost:        bcrypt.MaxCost,
			expectError: false,
		},
		{
			name:        "Invalid cost - too low",
			cost:        bcrypt.MinCost - 1,
			expectError: true,
		},
		{
			name:        "Invalid cost - too high",
			cost:        bcrypt.MaxCost + 1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher, err := NewBcryptPasswordHasher(tt.cost)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if hasher != nil {
					t.Errorf("Expected nil hasher when error occurs, got %v", hasher)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if hasher == nil {
					t.Errorf("Expected hasher but got nil")
				}
				if hasher.cost != tt.cost {
					t.Errorf("Expected cost %d, got %d", tt.cost, hasher.cost)
				}
			}
		})
	}
}

func TestNewDefaultBcryptPasswordHasher(t *testing.T) {
	hasher := NewDefaultBcryptPasswordHasher()

	if hasher == nil {
		t.Errorf("Expected hasher but got nil")
	}
	if hasher.cost != bcrypt.DefaultCost {
		t.Errorf("Expected cost %d, got %d", bcrypt.DefaultCost, hasher.cost)
	}
}

func TestBcryptPasswordHasher_HashPassword(t *testing.T) {
	hasher := NewDefaultBcryptPasswordHasher()
	ctx := context.Background()

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "Valid password",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "Empty password",
			password:    "",
			expectError: true,
		},
		{
			name:        "Long password",
			password:    "this_is_a_very_long_password_that_should_still_work_fine_with_bcrypt",
			expectError: false,
		},
		{
			name:        "Password with special characters",
			password:    "p@ssw0rd!@#$%^&*()",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := hasher.HashPassword(ctx, tt.password)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if hashedPassword != "" {
					t.Errorf("Expected empty hash when error occurs, got %s", hashedPassword)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if hashedPassword == "" {
					t.Errorf("Expected non-empty hash")
				}
				if hashedPassword == tt.password {
					t.Errorf("Hash should not equal plain password")
				}
			}
		})
	}
}

func TestBcryptPasswordHasher_VerifyPassword(t *testing.T) {
	hasher := NewDefaultBcryptPasswordHasher()
	ctx := context.Background()

	// Generate a test hash
	plainPassword := "test_password_123"
	hashedPassword, err := hasher.HashPassword(ctx, plainPassword)
	if err != nil {
		t.Fatalf("Failed to generate test hash: %v", err)
	}

	tests := []struct {
		name           string
		plainPassword  string
		hashedPassword string
		expectMatch    bool
		expectError    bool
	}{
		{
			name:           "Correct password",
			plainPassword:  plainPassword,
			hashedPassword: hashedPassword,
			expectMatch:    true,
			expectError:    false,
		},
		{
			name:           "Incorrect password",
			plainPassword:  "wrong_password",
			hashedPassword: hashedPassword,
			expectMatch:    false,
			expectError:    false,
		},
		{
			name:           "Empty plain password",
			plainPassword:  "",
			hashedPassword: hashedPassword,
			expectMatch:    false,
			expectError:    true,
		},
		{
			name:           "Empty hashed password",
			plainPassword:  plainPassword,
			hashedPassword: "",
			expectMatch:    false,
			expectError:    true,
		},
		{
			name:           "Invalid hash format",
			plainPassword:  plainPassword,
			hashedPassword: "invalid_hash",
			expectMatch:    false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := hasher.VerifyPassword(ctx, tt.plainPassword, tt.hashedPassword)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if match != tt.expectMatch {
					t.Errorf("Expected match %v, got %v", tt.expectMatch, match)
				}
			}
		})
	}
}

func TestBcryptPasswordHasher_GetName(t *testing.T) {
	tests := []struct {
		name         string
		cost         int
		expectedName string
	}{
		{
			name:         "Default cost",
			cost:         bcrypt.DefaultCost,
			expectedName: "bcrypt_hasher_cost_10",
		},
		{
			name:         "Minimum cost",
			cost:         bcrypt.MinCost,
			expectedName: "bcrypt_hasher_cost_4",
		},
		{
			name:         "Custom cost",
			cost:         12,
			expectedName: "bcrypt_hasher_cost_12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher, err := NewBcryptPasswordHasher(tt.cost)
			if err != nil {
				t.Fatalf("Failed to create hasher: %v", err)
			}

			name := hasher.GetName()
			if name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, name)
			}
		})
	}
}

func TestBcryptPasswordHasher_ContextCancellation(t *testing.T) {
	hasher := NewDefaultBcryptPasswordHasher()

	// Test context cancellation during hashing
	t.Run("Hash with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := hasher.HashPassword(ctx, "password")
		if err == nil {
			t.Errorf("Expected error due to cancelled context")
		}
	})

	// Test context cancellation during verification
	t.Run("Verify with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := hasher.VerifyPassword(ctx, "password", "hash")
		if err == nil {
			t.Errorf("Expected error due to cancelled context")
		}
	})
}

func TestBcryptPasswordHasher_ContextTimeout(t *testing.T) {
	hasher := NewDefaultBcryptPasswordHasher()

	// Test context timeout during hashing
	t.Run("Hash with timeout context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(1 * time.Millisecond) // Ensure timeout

		_, err := hasher.HashPassword(ctx, "password")
		if err == nil {
			t.Errorf("Expected error due to context timeout")
		}
	})
}

func TestBcryptPasswordHasher_Integration(t *testing.T) {
	hasher := NewDefaultBcryptPasswordHasher()
	ctx := context.Background()

	// Test full cycle: hash and verify
	passwords := []string{
		"simple",
		"complex_password_123!@#",
		"中文密碼",
		"emoji🔒password",
	}

	for _, password := range passwords {
		t.Run("Integration_"+password, func(t *testing.T) {
			// Hash the password
			hashedPassword, err := hasher.HashPassword(ctx, password)
			if err != nil {
				t.Fatalf("Failed to hash password: %v", err)
			}

			// Verify the password
			match, err := hasher.VerifyPassword(ctx, password, hashedPassword)
			if err != nil {
				t.Fatalf("Failed to verify password: %v", err)
			}

			if !match {
				t.Errorf("Password verification failed for: %s", password)
			}

			// Verify wrong password doesn't match
			wrongMatch, err := hasher.VerifyPassword(ctx, password+"wrong", hashedPassword)
			if err != nil {
				t.Fatalf("Failed to verify wrong password: %v", err)
			}

			if wrongMatch {
				t.Errorf("Wrong password should not match for: %s", password)
			}
		})
	}
}
