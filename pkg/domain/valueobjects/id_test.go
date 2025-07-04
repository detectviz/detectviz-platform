package valueobjects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestNewIDVO(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid UUID", validUUID, false},
		{"valid UUID with uppercase", "550E8400-E29B-41D4-A716-446655440000", false},
		{"empty ID", "", true},
		{"invalid format", "not-a-uuid", true},
		{"invalid UUID format", "550e8400-e29b-41d4-a716-44665544000", true},
		{"too short", "550e8400", true},
		{"invalid characters", "550e8400-e29b-41d4-a716-44665544000g", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := NewIDVO(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewIDVO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && id.IsEmpty() {
				t.Error("Valid ID should not be empty")
			}
		})
	}
}

func TestNewIDVOFromUUID(t *testing.T) {
	originalUUID := uuid.New()
	id := NewIDVOFromUUID(originalUUID)

	if id.String() != originalUUID.String() {
		t.Errorf("NewIDVOFromUUID() = %v, want %v", id.String(), originalUUID.String())
	}
}

func TestGenerateNewIDVO(t *testing.T) {
	id1 := GenerateNewIDVO()
	id2 := GenerateNewIDVO()

	if id1.IsEmpty() {
		t.Error("Generated ID should not be empty")
	}

	if id1.Equals(id2) {
		t.Error("Two generated IDs should be different")
	}

	if !id1.IsValidV4() {
		t.Error("Generated ID should be a valid UUID v4")
	}
}

func TestIDVO_String(t *testing.T) {
	uuidStr := "550e8400-e29b-41d4-a716-446655440000"
	id, _ := NewIDVO(uuidStr)

	if id.String() != uuidStr {
		t.Errorf("IDVO.String() = %v, want %v", id.String(), uuidStr)
	}
}

func TestIDVO_Value(t *testing.T) {
	uuidStr := "550e8400-e29b-41d4-a716-446655440000"
	id, _ := NewIDVO(uuidStr)

	if id.Value() != uuidStr {
		t.Errorf("IDVO.Value() = %v, want %v", id.Value(), uuidStr)
	}
}

func TestIDVO_UUID(t *testing.T) {
	originalUUID := uuid.New()
	id := NewIDVOFromUUID(originalUUID)

	parsedUUID, err := id.UUID()
	if err != nil {
		t.Errorf("IDVO.UUID() error = %v", err)
	}

	if parsedUUID != originalUUID {
		t.Errorf("IDVO.UUID() = %v, want %v", parsedUUID, originalUUID)
	}
}

func TestIDVO_Equals(t *testing.T) {
	uuidStr := "550e8400-e29b-41d4-a716-446655440000"
	id1, _ := NewIDVO(uuidStr)
	id2, _ := NewIDVO(uuidStr)
	id3, _ := NewIDVO(uuid.New().String())

	if !id1.Equals(id2) {
		t.Error("Same IDs should be equal")
	}

	if id1.Equals(id3) {
		t.Error("Different IDs should not be equal")
	}
}

func TestIDVO_IsEmpty(t *testing.T) {
	var emptyID IDVO
	id, _ := NewIDVO(uuid.New().String())

	if !emptyID.IsEmpty() {
		t.Error("Empty IDVO should return true for IsEmpty()")
	}

	if id.IsEmpty() {
		t.Error("Non-empty IDVO should return false for IsEmpty()")
	}
}

func TestIDVO_IsNil(t *testing.T) {
	var emptyID IDVO
	nilID := NewIDVOFromUUID(uuid.Nil)
	validID := GenerateNewIDVO()

	if !emptyID.IsNil() {
		t.Error("Empty IDVO should return true for IsNil()")
	}

	if !nilID.IsNil() {
		t.Error("Nil UUID should return true for IsNil()")
	}

	if validID.IsNil() {
		t.Error("Valid IDVO should return false for IsNil()")
	}
}

func TestIDVO_Version(t *testing.T) {
	// UUID v4
	v4ID := GenerateNewIDVO()
	version, err := v4ID.Version()
	if err != nil {
		t.Errorf("IDVO.Version() error = %v", err)
	}

	if version != 4 {
		t.Errorf("Generated UUID should be version 4, got %d", version)
	}
}

func TestIDVO_Variant(t *testing.T) {
	id := GenerateNewIDVO()
	variant, err := id.Variant()
	if err != nil {
		t.Errorf("IDVO.Variant() error = %v", err)
	}

	if variant != uuid.RFC4122 {
		t.Errorf("Generated UUID should be RFC4122 variant, got %v", variant)
	}
}

func TestIDVO_ShortString(t *testing.T) {
	uuidStr := "550e8400-e29b-41d4-a716-446655440000"
	id, _ := NewIDVO(uuidStr)

	expected := "550e8400"
	if id.ShortString() != expected {
		t.Errorf("IDVO.ShortString() = %v, want %v", id.ShortString(), expected)
	}
}

func TestIDVO_IsValidV4(t *testing.T) {
	v4ID := GenerateNewIDVO()

	if !v4ID.IsValidV4() {
		t.Error("Generated UUID should be valid v4")
	}

	// Test with a non-v4 UUID (using nil UUID which is not v4)
	nilID := NewIDVOFromUUID(uuid.Nil)
	if nilID.IsValidV4() {
		t.Error("Nil UUID should not be valid v4")
	}
}

func TestIDVO_ToBytes(t *testing.T) {
	originalUUID := uuid.New()
	id := NewIDVOFromUUID(originalUUID)

	bytes, err := id.ToBytes()
	if err != nil {
		t.Errorf("IDVO.ToBytes() error = %v", err)
	}

	if len(bytes) != 16 {
		t.Errorf("UUID bytes should be 16 bytes long, got %d", len(bytes))
	}

	// Convert back to UUID and compare
	reconstructedUUID, err := uuid.FromBytes(bytes)
	if err != nil {
		t.Errorf("Failed to reconstruct UUID from bytes: %v", err)
	}

	if reconstructedUUID != originalUUID {
		t.Error("Reconstructed UUID should match original")
	}
}

func TestIDVO_MarshalJSON(t *testing.T) {
	uuidStr := "550e8400-e29b-41d4-a716-446655440000"
	id, _ := NewIDVO(uuidStr)

	data, err := json.Marshal(id)
	if err != nil {
		t.Errorf("IDVO.MarshalJSON() error = %v", err)
	}

	expected := `"550e8400-e29b-41d4-a716-446655440000"`
	if string(data) != expected {
		t.Errorf("IDVO.MarshalJSON() = %v, want %v", string(data), expected)
	}
}

func TestIDVO_UnmarshalJSON(t *testing.T) {
	var id IDVO
	data := []byte(`"550e8400-e29b-41d4-a716-446655440000"`)

	err := json.Unmarshal(data, &id)
	if err != nil {
		t.Errorf("IDVO.UnmarshalJSON() error = %v", err)
	}

	if id.Value() != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("IDVO.UnmarshalJSON() = %v, want %v", id.Value(), "550e8400-e29b-41d4-a716-446655440000")
	}
}

func TestIDVO_UnmarshalJSON_Invalid(t *testing.T) {
	var id IDVO
	data := []byte(`"invalid-uuid"`)

	err := json.Unmarshal(data, &id)
	if err == nil {
		t.Error("IDVO.UnmarshalJSON() should return error for invalid UUID")
	}
}

func TestIDVO_MarshalText(t *testing.T) {
	uuidStr := "550e8400-e29b-41d4-a716-446655440000"
	id, _ := NewIDVO(uuidStr)

	data, err := id.MarshalText()
	if err != nil {
		t.Errorf("IDVO.MarshalText() error = %v", err)
	}

	if string(data) != uuidStr {
		t.Errorf("IDVO.MarshalText() = %v, want %v", string(data), uuidStr)
	}
}

func TestIDVO_UnmarshalText(t *testing.T) {
	var id IDVO
	data := []byte("550e8400-e29b-41d4-a716-446655440000")

	err := id.UnmarshalText(data)
	if err != nil {
		t.Errorf("IDVO.UnmarshalText() error = %v", err)
	}

	if id.Value() != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("IDVO.UnmarshalText() = %v, want %v", id.Value(), "550e8400-e29b-41d4-a716-446655440000")
	}
}

func TestIDVO_CaseInsensitive(t *testing.T) {
	id1, _ := NewIDVO("550E8400-E29B-41D4-A716-446655440000")
	id2, _ := NewIDVO("550e8400-e29b-41d4-a716-446655440000")

	if !id1.Equals(id2) {
		t.Error("UUID comparison should be case insensitive")
	}
}

func TestIDVO_WhitespaceHandling(t *testing.T) {
	uuidStr := "550e8400-e29b-41d4-a716-446655440000"
	id, _ := NewIDVO("  " + uuidStr + "  ")

	if id.Value() != uuidStr {
		t.Error("ID should trim whitespace")
	}
}
