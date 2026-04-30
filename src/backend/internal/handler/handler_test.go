package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	domainErrors "teamboard/internal/domain/errors"
	"teamboard/internal/dto/response"
)

func TestHandleError_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handleError(c, domainErrors.ErrNotFound)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHandleError_Unauthorized(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handleError(c, domainErrors.ErrUnauthorized)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestHandleError_Forbidden(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handleError(c, domainErrors.ErrForbidden)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestHandleError_Conflict(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handleError(c, domainErrors.ErrConflict)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestHandleError_BadRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handleError(c, domainErrors.ErrBadRequest)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestHandleError_ValidationErrors(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	validationErrs := domainErrors.ValidationErrors{
		&domainErrors.ValidationError{Field: "name", Message: "is required"},
		&domainErrors.ValidationError{Field: "email", Message: "is invalid"},
	}

	err := handleError(c, validationErrs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	errorBody := body["error"].(map[string]interface{})
	if errorBody["code"] != "VALIDATION_ERROR" {
		t.Errorf("expected error code VALIDATION_ERROR, got %v", errorBody["code"])
	}
	details := errorBody["details"].([]interface{})
	if len(details) != 2 {
		t.Errorf("expected 2 validation errors, got %d", len(details))
	}
}

func TestHandleError_InternalError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handleError(c, fmt.Errorf("some internal error"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestNewErrorResponse(t *testing.T) {
	errResp := newErrorResponse("TEST_CODE", "Test message")
	if errResp.Error.Code != "TEST_CODE" {
		t.Errorf("expected code TEST_CODE, got %s", errResp.Error.Code)
	}
	if errResp.Error.Message != "Test message" {
		t.Errorf("expected message 'Test message', got %s", errResp.Error.Message)
	}
}

func TestGetPage(t *testing.T) {
	e := echo.New()

	tests := []struct {
		param    string
		expected int
	}{
		{"", 1},
		{"0", 1},
		{"-1", 1},
		{"1", 1},
		{"5", 5},
		{"abc", 1},
	}

	for _, tt := range tests {
		t.Run(tt.param, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?page="+tt.param, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if got := getPage(c); got != tt.expected {
				t.Errorf("getPage(%q) = %d, want %d", tt.param, got, tt.expected)
			}
		})
	}
}

func TestGetPerPage(t *testing.T) {
	e := echo.New()

	tests := []struct {
		param    string
		expected int
	}{
		{"", 20},
		{"0", 20},
		{"-1", 20},
		{"50", 50},
		{"200", 100},
		{"abc", 20},
	}

	for _, tt := range tests {
		t.Run(tt.param, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?per_page="+tt.param, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if got := getPerPage(c); got != tt.expected {
				t.Errorf("getPerPage(%q) = %d, want %d", tt.param, got, tt.expected)
			}
		})
	}
}

func TestGetAuthContext_Nil(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	result := getAuthContext(c)
	if result != nil {
		t.Error("expected nil auth context")
	}
}

func TestGetAuthContext_WithServiceContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set a wrong type - should return nil
	c.Set("auth_context", "wrong type")
	result := getAuthContext(c)
	if result != nil {
		t.Error("expected nil for wrong type")
	}
}

// Ensure ErrorResponse and FieldError types work correctly
func TestResponseTypes(t *testing.T) {
	errResp := response.ErrorResponse{
		Error: response.ErrorBody{
			Code:    "TEST",
			Message: "test message",
			Details: []response.FieldError{
				{Field: "f1", Message: "m1"},
			},
		},
	}
	if errResp.Error.Code != "TEST" {
		t.Error("code mismatch")
	}
	if len(errResp.Error.Details) != 1 {
		t.Error("details length mismatch")
	}
	if errResp.Error.Details[0].Field != "f1" {
		t.Error("field mismatch")
	}
}
