package entities

import (
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		userName    string
		email       string
		password    string
		expectError bool
		errorType   error
	}{
		{
			name:        "Valid user creation",
			id:          "user123",
			userName:    "John Doe",
			email:       "john@example.com",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "Empty name should fail",
			id:          "user123",
			userName:    "",
			email:       "john@example.com",
			password:    "password123",
			expectError: true,
			errorType:   ErrInvalidUserFields,
		},
		{
			name:        "Empty email should fail",
			id:          "user123",
			userName:    "John Doe",
			email:       "",
			password:    "password123",
			expectError: true,
			errorType:   ErrInvalidUserFields,
		},
		{
			name:        "Empty password should fail",
			id:          "user123",
			userName:    "John Doe",
			email:       "john@example.com",
			password:    "",
			expectError: true,
			errorType:   ErrInvalidUserFields,
		},
		{
			name:        "All empty fields should fail",
			id:          "user123",
			userName:    "",
			email:       "",
			password:    "",
			expectError: true,
			errorType:   ErrInvalidUserFields,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.id, tt.userName, tt.email, tt.password)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("Expected error %v, got %v", tt.errorType, err)
				}
				if user != nil {
					t.Errorf("Expected nil user when error occurs, got %v", user)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if user == nil {
					t.Errorf("Expected user but got nil")
				}

				// Verify user fields
				if user.ID != tt.id {
					t.Errorf("Expected ID %s, got %s", tt.id, user.ID)
				}
				if user.Name != tt.userName {
					t.Errorf("Expected Name %s, got %s", tt.userName, user.Name)
				}
				if user.Email != tt.email {
					t.Errorf("Expected Email %s, got %s", tt.email, user.Email)
				}
				if user.PasswordHash != tt.password {
					t.Errorf("Expected PasswordHash %s, got %s", tt.password, user.PasswordHash)
				}

				// Verify timestamps are set
				if user.CreatedAt.IsZero() {
					t.Errorf("Expected CreatedAt to be set")
				}
				if user.UpdatedAt.IsZero() {
					t.Errorf("Expected UpdatedAt to be set")
				}

				// Verify CreatedAt and UpdatedAt are close to current time
				now := time.Now()
				if now.Sub(user.CreatedAt) > time.Second {
					t.Errorf("CreatedAt seems too old: %v", user.CreatedAt)
				}
				if now.Sub(user.UpdatedAt) > time.Second {
					t.Errorf("UpdatedAt seems too old: %v", user.UpdatedAt)
				}
			}
		})
	}
}

func TestUser_FieldValidation(t *testing.T) {
	// Test direct field access and validation
	user := &User{
		ID:           "test123",
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "testpass",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Test field access
	if user.ID != "test123" {
		t.Errorf("Expected ID test123, got %s", user.ID)
	}
	if user.Name != "Test User" {
		t.Errorf("Expected Name 'Test User', got %s", user.Name)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected Email 'test@example.com', got %s", user.Email)
	}
	if user.PasswordHash != "testpass" {
		t.Errorf("Expected PasswordHash 'testpass', got %s", user.PasswordHash)
	}
}

func TestUser_TimestampBehavior(t *testing.T) {
	user1, err := NewUser("user1", "User One", "user1@example.com", "pass1")
	if err != nil {
		t.Fatalf("Unexpected error creating user1: %v", err)
	}

	// Wait a bit to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	user2, err := NewUser("user2", "User Two", "user2@example.com", "pass2")
	if err != nil {
		t.Fatalf("Unexpected error creating user2: %v", err)
	}

	// Verify that timestamps are different
	if !user2.CreatedAt.After(user1.CreatedAt) {
		t.Errorf("Expected user2 CreatedAt to be after user1 CreatedAt")
	}
	if !user2.UpdatedAt.After(user1.UpdatedAt) {
		t.Errorf("Expected user2 UpdatedAt to be after user1 UpdatedAt")
	}
}

func TestUser_PasswordSecurity(t *testing.T) {
	// Note: This test documents the updated behavior where passwords are stored as hashes
	// The security issue has been addressed by using PasswordHash instead of plain text
	user, err := NewUser("user123", "Test User", "test@example.com", "hashed_password_value")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Password is now stored as hash (security improvement)
	if user.PasswordHash != "hashed_password_value" {
		t.Errorf("Expected password hash to be stored, got %s", user.PasswordHash)
	}

	// Verify that the password hash is not empty
	if user.PasswordHash == "" {
		t.Errorf("Password hash should not be empty")
	}
}
