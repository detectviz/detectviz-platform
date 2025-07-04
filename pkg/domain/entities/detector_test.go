package entities

import (
	"testing"
	"time"
)

func TestDetector_FieldValidation(t *testing.T) {
	// Test direct field access and validation
	now := time.Now()
	detector := &Detector{
		ID:          "detector123",
		Name:        "Test Detector",
		Description: "A test detector for validation",
		OwnerID:     "user123",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Test field access
	if detector.ID != "detector123" {
		t.Errorf("Expected ID detector123, got %s", detector.ID)
	}
	if detector.Name != "Test Detector" {
		t.Errorf("Expected Name 'Test Detector', got %s", detector.Name)
	}
	if detector.Description != "A test detector for validation" {
		t.Errorf("Expected Description 'A test detector for validation', got %s", detector.Description)
	}
	if detector.OwnerID != "user123" {
		t.Errorf("Expected OwnerID 'user123', got %s", detector.OwnerID)
	}
	if !detector.CreatedAt.Equal(now) {
		t.Errorf("Expected CreatedAt %v, got %v", now, detector.CreatedAt)
	}
	if !detector.UpdatedAt.Equal(now) {
		t.Errorf("Expected UpdatedAt %v, got %v", now, detector.UpdatedAt)
	}
}

func TestDetector_EmptyValues(t *testing.T) {
	// Test behavior with empty values
	detector := &Detector{}

	if detector.ID != "" {
		t.Errorf("Expected empty ID, got %s", detector.ID)
	}
	if detector.Name != "" {
		t.Errorf("Expected empty Name, got %s", detector.Name)
	}
	if detector.Description != "" {
		t.Errorf("Expected empty Description, got %s", detector.Description)
	}
	if detector.OwnerID != "" {
		t.Errorf("Expected empty OwnerID, got %s", detector.OwnerID)
	}
	if !detector.CreatedAt.IsZero() {
		t.Errorf("Expected zero CreatedAt, got %v", detector.CreatedAt)
	}
	if !detector.UpdatedAt.IsZero() {
		t.Errorf("Expected zero UpdatedAt, got %v", detector.UpdatedAt)
	}
}

func TestDetector_TimestampBehavior(t *testing.T) {
	now1 := time.Now()
	detector1 := &Detector{
		ID:        "detector1",
		Name:      "Detector One",
		CreatedAt: now1,
		UpdatedAt: now1,
	}

	// Wait a bit to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	now2 := time.Now()
	detector2 := &Detector{
		ID:        "detector2",
		Name:      "Detector Two",
		CreatedAt: now2,
		UpdatedAt: now2,
	}

	// Verify that timestamps are different
	if !detector2.CreatedAt.After(detector1.CreatedAt) {
		t.Errorf("Expected detector2 CreatedAt to be after detector1 CreatedAt")
	}
	if !detector2.UpdatedAt.After(detector1.UpdatedAt) {
		t.Errorf("Expected detector2 UpdatedAt to be after detector1 UpdatedAt")
	}
}

func TestDetector_StructureIntegrity(t *testing.T) {
	// Test that the detector maintains its structure integrity
	detector := &Detector{
		ID:          "test-detector-123",
		Name:        "Integration Test Detector",
		Description: "This detector is used for integration testing purposes",
		OwnerID:     "owner-456",
		CreatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}

	// Verify all fields are preserved
	expectedID := "test-detector-123"
	expectedName := "Integration Test Detector"
	expectedDescription := "This detector is used for integration testing purposes"
	expectedOwnerID := "owner-456"
	expectedCreatedAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	expectedUpdatedAt := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)

	if detector.ID != expectedID {
		t.Errorf("ID mismatch: expected %s, got %s", expectedID, detector.ID)
	}
	if detector.Name != expectedName {
		t.Errorf("Name mismatch: expected %s, got %s", expectedName, detector.Name)
	}
	if detector.Description != expectedDescription {
		t.Errorf("Description mismatch: expected %s, got %s", expectedDescription, detector.Description)
	}
	if detector.OwnerID != expectedOwnerID {
		t.Errorf("OwnerID mismatch: expected %s, got %s", expectedOwnerID, detector.OwnerID)
	}
	if !detector.CreatedAt.Equal(expectedCreatedAt) {
		t.Errorf("CreatedAt mismatch: expected %v, got %v", expectedCreatedAt, detector.CreatedAt)
	}
	if !detector.UpdatedAt.Equal(expectedUpdatedAt) {
		t.Errorf("UpdatedAt mismatch: expected %v, got %v", expectedUpdatedAt, detector.UpdatedAt)
	}
}

func TestDetector_OwnershipValidation(t *testing.T) {
	// Test different ownership scenarios
	tests := []struct {
		name    string
		ownerID string
		valid   bool
	}{
		{
			name:    "Valid owner ID",
			ownerID: "user123",
			valid:   true,
		},
		{
			name:    "Empty owner ID",
			ownerID: "",
			valid:   true, // Current implementation allows empty owner ID
		},
		{
			name:    "UUID-style owner ID",
			ownerID: "550e8400-e29b-41d4-a716-446655440000",
			valid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &Detector{
				ID:        "detector123",
				Name:      "Test Detector",
				OwnerID:   tt.ownerID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// For now, all owner IDs are considered valid
			// In the future, we might add validation logic
			if detector.OwnerID != tt.ownerID {
				t.Errorf("Expected OwnerID %s, got %s", tt.ownerID, detector.OwnerID)
			}
		})
	}
}

// Note: This test suite documents the current Detector entity structure.
// As the business logic evolves, additional validation methods and business rules
// should be added to the Detector entity, and corresponding tests should be created.
