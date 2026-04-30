package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"teamboard/internal/service"
)

// ========== Mock Authenticator ==========

type mockAuthenticator struct {
	authCtx *service.AuthContext
	err     error
}

func (m *mockAuthenticator) Authenticate(_ context.Context, _ string) (*service.AuthContext, error) {
	return m.authCtx, m.err
}

type invalidKeyError struct{}

func (e *invalidKeyError) Error() string { return "invalid key" }

// ========== APIKeyAuth Tests ==========

func TestAPIKeyAuth_ValidKey(t *testing.T) {
	e := echo.New()
	auth := &mockAuthenticator{
		authCtx: &service.AuthContext{MemberID: "m1", TeamID: "t1", Role: "admin"},
	}

	called := false
	handler := func(c echo.Context) error {
		called = true
		ac := c.Get("auth_context").(*service.AuthContext)
		if ac.MemberID != "m1" {
			t.Errorf("expected member ID m1, got %s", ac.MemberID)
		}
		if ac.TeamID != "t1" {
			t.Errorf("expected team ID t1, got %s", ac.TeamID)
		}
		if ac.Role != "admin" {
			t.Errorf("expected role admin, got %s", ac.Role)
		}
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "tb_validkey12345678901234567890123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := APIKeyAuth(auth)
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	e := echo.New()
	auth := &mockAuthenticator{}

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := APIKeyAuth(auth)
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected handler NOT to be called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	e := echo.New()
	auth := &mockAuthenticator{
		err: &invalidKeyError{},
	}

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "tb_invalidkey")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := APIKeyAuth(auth)
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected handler NOT to be called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAPIKeyAuth_SetsAuthContext(t *testing.T) {
	e := echo.New()
	expectedCtx := &service.AuthContext{MemberID: "m1", TeamID: "t1", Role: "member"}
	auth := &mockAuthenticator{authCtx: expectedCtx}

	var gotCtx *service.AuthContext
	handler := func(c echo.Context) error {
		gotCtx = c.Get("auth_context").(*service.AuthContext)
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "tb_validkey12345678901234567890123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := APIKeyAuth(auth)
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotCtx != expectedCtx {
		t.Error("auth context not properly set on echo context")
	}
}

// ========== RequireAdmin Tests ==========

func TestRequireAdmin_AdminUser(t *testing.T) {
	e := echo.New()

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("auth_context", &service.AuthContext{Role: "admin"})

	middleware := RequireAdmin()
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestRequireAdmin_MemberUser(t *testing.T) {
	e := echo.New()

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("auth_context", &service.AuthContext{Role: "member"})

	middleware := RequireAdmin()
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected handler NOT to be called")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestRequireAdmin_NoAuthContext(t *testing.T) {
	e := echo.New()

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := RequireAdmin()
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected handler NOT to be called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestRequireAdmin_WrongContextType(t *testing.T) {
	e := echo.New()

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("auth_context", "not the right type")

	middleware := RequireAdmin()
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected handler NOT to be called")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

// ========== CORS Tests ==========

func TestCORS_AllowedOrigin(t *testing.T) {
	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := CORS("http://localhost:3000,http://localhost:4000")
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("expected origin http://localhost:3000, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://evil.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := CORS("http://localhost:3000")
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no Allow-Origin for disallowed origin, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_OptionsRequest(t *testing.T) {
	e := echo.New()

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := CORS("http://localhost:3000")
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("next handler should not be called for OPTIONS")
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}

func TestCORS_AllowMethods(t *testing.T) {
	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := CORS("http://localhost:3000")
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "GET, POST, PUT, DELETE, OPTIONS"
	if rec.Header().Get("Access-Control-Allow-Methods") != expected {
		t.Errorf("expected methods %q, got %q", expected, rec.Header().Get("Access-Control-Allow-Methods"))
	}
}

func TestCORS_AllowHeaders(t *testing.T) {
	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := CORS("http://localhost:3000")
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Content-Type, X-API-Key"
	if rec.Header().Get("Access-Control-Allow-Headers") != expected {
		t.Errorf("expected headers %q, got %q", expected, rec.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestCORS_MaxAge(t *testing.T) {
	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := CORS("http://localhost:3000")
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Header().Get("Access-Control-Max-Age") != "86400" {
		t.Errorf("expected max-age 86400, got %s", rec.Header().Get("Access-Control-Max-Age"))
	}
}

func TestCORS_NoOrigin(t *testing.T) {
	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Origin header
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := CORS("http://localhost:3000")
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should still set other CORS headers
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Allow-Methods header even without Origin")
	}
}

// ========== SecurityHeaders Tests ==========

func TestSecurityHeaders(t *testing.T) {
	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := SecurityHeaders()
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct{ header, expected string }{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Content-Security-Policy", "default-src 'self'"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
	}
	for _, tt := range tests {
		got := rec.Header().Get(tt.header)
		if got != tt.expected {
			t.Errorf("header %s: expected %q, got %q", tt.header, tt.expected, got)
		}
	}
}

func TestSecurityHeaders_CalledBeforeHandler(t *testing.T) {
	e := echo.New()

	// Handler should see the security headers already set
	var handlerSeen string
	handler := func(c echo.Context) error {
		handlerSeen = c.Response().Header().Get("X-Content-Type-Options")
		return c.String(http.StatusOK, "ok")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := SecurityHeaders()
	if err := middleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handlerSeen != "nosniff" {
		t.Errorf("security headers should be set before handler runs, got %q", handlerSeen)
	}
}
