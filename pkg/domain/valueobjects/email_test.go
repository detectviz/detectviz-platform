package valueobjects

import (
	"encoding/json"
	"testing"
)

func TestNewEmailVO(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"valid email with numbers", "user123@example123.com", false},
		{"valid email with special chars", "user.name+tag@example.com", false},
		{"empty email", "", true},
		{"invalid format - no @", "testexample.com", true},
		{"invalid format - no domain", "test@", true},
		{"invalid format - no local part", "@example.com", true},
		{"invalid format - multiple @", "test@@example.com", true},
		{"too long email", "a" + string(make([]byte, 250)) + "@example.com", true},
		{"local part too long", string(make([]byte, 65)) + "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmailVO(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEmailVO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && email.IsEmpty() {
				t.Error("Valid email should not be empty")
			}
		})
	}
}

func TestEmailVO_String(t *testing.T) {
	email, _ := NewEmailVO("Test@Example.COM")
	expected := "test@example.com"

	if email.String() != expected {
		t.Errorf("EmailVO.String() = %v, want %v", email.String(), expected)
	}
}

func TestEmailVO_Value(t *testing.T) {
	email, _ := NewEmailVO("test@example.com")

	if email.Value() != "test@example.com" {
		t.Errorf("EmailVO.Value() = %v, want %v", email.Value(), "test@example.com")
	}
}

func TestEmailVO_Equals(t *testing.T) {
	email1, _ := NewEmailVO("test@example.com")
	email2, _ := NewEmailVO("test@example.com")
	email3, _ := NewEmailVO("other@example.com")

	if !email1.Equals(email2) {
		t.Error("Same emails should be equal")
	}

	if email1.Equals(email3) {
		t.Error("Different emails should not be equal")
	}
}

func TestEmailVO_IsEmpty(t *testing.T) {
	var emptyEmail EmailVO
	email, _ := NewEmailVO("test@example.com")

	if !emptyEmail.IsEmpty() {
		t.Error("Empty EmailVO should return true for IsEmpty()")
	}

	if email.IsEmpty() {
		t.Error("Non-empty EmailVO should return false for IsEmpty()")
	}
}

func TestEmailVO_Domain(t *testing.T) {
	email, _ := NewEmailVO("test@example.com")

	if email.Domain() != "example.com" {
		t.Errorf("EmailVO.Domain() = %v, want %v", email.Domain(), "example.com")
	}
}

func TestEmailVO_LocalPart(t *testing.T) {
	email, _ := NewEmailVO("test.user@example.com")

	if email.LocalPart() != "test.user" {
		t.Errorf("EmailVO.LocalPart() = %v, want %v", email.LocalPart(), "test.user")
	}
}

func TestEmailVO_IsGmailAddress(t *testing.T) {
	gmail, _ := NewEmailVO("test@gmail.com")
	notGmail, _ := NewEmailVO("test@example.com")

	if !gmail.IsGmailAddress() {
		t.Error("Gmail address should return true for IsGmailAddress()")
	}

	if notGmail.IsGmailAddress() {
		t.Error("Non-Gmail address should return false for IsGmailAddress()")
	}
}

func TestEmailVO_IsCompanyEmail(t *testing.T) {
	companyEmail, _ := NewEmailVO("employee@company.com")
	gmailEmail, _ := NewEmailVO("user@gmail.com")
	yahooEmail, _ := NewEmailVO("user@yahoo.com")

	if !companyEmail.IsCompanyEmail() {
		t.Error("Company email should return true for IsCompanyEmail()")
	}

	if gmailEmail.IsCompanyEmail() {
		t.Error("Gmail address should return false for IsCompanyEmail()")
	}

	if yahooEmail.IsCompanyEmail() {
		t.Error("Yahoo address should return false for IsCompanyEmail()")
	}
}

func TestEmailVO_MarshalJSON(t *testing.T) {
	email, _ := NewEmailVO("test@example.com")

	data, err := json.Marshal(email)
	if err != nil {
		t.Errorf("EmailVO.MarshalJSON() error = %v", err)
	}

	expected := `"test@example.com"`
	if string(data) != expected {
		t.Errorf("EmailVO.MarshalJSON() = %v, want %v", string(data), expected)
	}
}

func TestEmailVO_UnmarshalJSON(t *testing.T) {
	var email EmailVO
	data := []byte(`"test@example.com"`)

	err := json.Unmarshal(data, &email)
	if err != nil {
		t.Errorf("EmailVO.UnmarshalJSON() error = %v", err)
	}

	if email.Value() != "test@example.com" {
		t.Errorf("EmailVO.UnmarshalJSON() = %v, want %v", email.Value(), "test@example.com")
	}
}

func TestEmailVO_UnmarshalJSON_Invalid(t *testing.T) {
	var email EmailVO
	data := []byte(`"invalid-email"`)

	err := json.Unmarshal(data, &email)
	if err == nil {
		t.Error("EmailVO.UnmarshalJSON() should return error for invalid email")
	}
}

func TestEmailVO_CaseInsensitive(t *testing.T) {
	email1, _ := NewEmailVO("Test@Example.COM")
	email2, _ := NewEmailVO("test@example.com")

	if !email1.Equals(email2) {
		t.Error("Email comparison should be case insensitive")
	}

	if email1.String() != "test@example.com" {
		t.Error("Email should be normalized to lowercase")
	}
}

func TestEmailVO_WhitespaceHandling(t *testing.T) {
	email, _ := NewEmailVO("  test@example.com  ")

	if email.Value() != "test@example.com" {
		t.Error("Email should trim whitespace")
	}
}
