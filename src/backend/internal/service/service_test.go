package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestGenerateAPIKey(t *testing.T) {
	key := GenerateAPIKey()
	if len(key) != 40 {
		t.Errorf("expected API key length 40, got %d", len(key))
	}
	if key[:3] != "tb_" {
		t.Errorf("expected API key prefix 'tb_', got '%s'", key[:3])
	}

	// Ensure uniqueness
	key2 := GenerateAPIKey()
	if key == key2 {
		t.Error("two generated API keys should not be equal")
	}
}

func TestHashAPIKey(t *testing.T) {
	key := "tb_testkey12345678901234567890123"
	hash := HashAPIKey(key)

	// Verify it's SHA-256
	expected := sha256.Sum256([]byte(key))
	expectedHex := hex.EncodeToString(expected[:])
	if hash != expectedHex {
		t.Errorf("hash mismatch: expected %s, got %s", expectedHex, hash)
	}

	// Length should be 64 (SHA-256 hex)
	if len(hash) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash))
	}
}

func TestHashAPIKey_Deterministic(t *testing.T) {
	key := "tb_testkey12345678901234567890123"
	hash1 := HashAPIKey(key)
	hash2 := HashAPIKey(key)
	if hash1 != hash2 {
		t.Error("hashing the same key should produce the same hash")
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		want  bool
	}{
		{"test@example.com", true},
		{"user@domain.org", true},
		{"a@b.co", true},
		{"", false},
		{"@", false},
		{"user@", false},
		{"@domain.com", false},
		{"user@domain", false},
		{"user@.com", false},
		{"plaintext", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			if got := isValidEmail(tt.email); got != tt.want {
				t.Errorf("isValidEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"todo", true},
		{"in_progress", true},
		{"review", true},
		{"done", true},
		{"", false},
		{"unknown", false},
		{"TODO", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := isValidStatus(tt.status); got != tt.want {
				t.Errorf("isValidStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestIsValidPriority(t *testing.T) {
	tests := []struct {
		priority string
		want     bool
	}{
		{"low", true},
		{"medium", true},
		{"high", true},
		{"critical", true},
		{"", false},
		{"urgent", false},
		{"HIGH", false},
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			if got := isValidPriority(tt.priority); got != tt.want {
				t.Errorf("isValidPriority(%q) = %v, want %v", tt.priority, got, tt.want)
			}
		})
	}
}

func TestNilIfEmpty(t *testing.T) {
	some := "value"
	empty := ""

	tests := []struct {
		name  string
		input *string
		want  *string
	}{
		{"nil input", nil, nil},
		{"empty string", &empty, nil},
		{"non-empty string", &some, &some},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nilIfEmpty(tt.input)
			if tt.want == nil && got != nil {
				t.Errorf("expected nil, got %v", *got)
			}
			if tt.want != nil && got == nil {
				t.Error("expected non-nil, got nil")
			}
			if tt.want != nil && got != nil && *got != *tt.want {
				t.Errorf("expected %v, got %v", *tt.want, *got)
			}
		})
	}
}

func TestNewULID(t *testing.T) {
	id := newULID()
	if len(id) != 26 {
		t.Errorf("expected ULID length 26, got %d", len(id))
	}

	id2 := newULID()
	if id == id2 {
		t.Error("two ULIDs should not be equal")
	}
}

func TestAuthService_Authenticate_EmptyKey(t *testing.T) {
	svc := &AuthService{}
	_, err := svc.Authenticate(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty API key")
	}
}
