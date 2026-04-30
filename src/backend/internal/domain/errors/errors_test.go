package errors

import (
	"testing"
)

func TestValidationErrors_Error(t *testing.T) {
	errs := ValidationErrors{
		&ValidationError{Field: "name", Message: "is required"},
	}
	if errs.Error() != "name: is required" {
		t.Errorf("unexpected error message: %s", errs.Error())
	}
}

func TestValidationErrors_Empty(t *testing.T) {
	errs := ValidationErrors{}
	if errs.Error() != "validation error" {
		t.Errorf("unexpected error message for empty: %s", errs.Error())
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{Field: "email", Message: "is invalid"}
	if err.Error() != "email: is invalid" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		msg   string
	}{
		{"ErrNotFound", ErrNotFound, "resource not found"},
		{"ErrUnauthorized", ErrUnauthorized, "unauthorized"},
		{"ErrForbidden", ErrForbidden, "forbidden"},
		{"ErrConflict", ErrConflict, "resource already exists"},
		{"ErrBadRequest", ErrBadRequest, "bad request"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("expected %q, got %q", tt.msg, tt.err.Error())
			}
		})
	}
}
