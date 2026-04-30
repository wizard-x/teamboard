package service

import (
	"context"
	"errors"
	"testing"

	domainErrors "teamboard/internal/domain/errors"
	"teamboard/internal/dto/request"
)

// ========== AuthService Tests ==========

func TestAuthService_Authenticate_ValidKey(t *testing.T) {
	mr := newMockMemberRepo()
	c, mrClose := newTestCache()
	defer mrClose.Close()

	apiKey := GenerateAPIKey()
	m := seedMember(mr, "mem1", "team1", "Admin", "a@b.com", "admin")
	m.APIKeyHash = HashAPIKey(apiKey)
	m.APIKeyPrefix = apiKey[:10]
	mr.byHash[HashAPIKey(apiKey)] = m

	svc := NewAuthService(mr, c)
	authCtx, err := svc.Authenticate(context.Background(), apiKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if authCtx.MemberID != "mem1" {
		t.Errorf("expected member ID mem1, got %s", authCtx.MemberID)
	}
	if authCtx.TeamID != "team1" {
		t.Errorf("expected team ID team1, got %s", authCtx.TeamID)
	}
	if authCtx.Role != "admin" {
		t.Errorf("expected role admin, got %s", authCtx.Role)
	}
}

func TestAuthService_Authenticate_InvalidKey(t *testing.T) {
	mr := newMockMemberRepo()
	c, mrClose := newTestCache()
	defer mrClose.Close()
	svc := NewAuthService(mr, c)

	_, err := svc.Authenticate(context.Background(), "tb_nonexistentkey12345678901234567")
	if err == nil {
		t.Fatal("expected error for invalid API key")
	}
	if !errors.Is(err, domainErrors.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAuthService_Authenticate_CacheHit(t *testing.T) {
	mr := newMockMemberRepo()
	c, mrClose := newTestCache()
	defer mrClose.Close()
	svc := NewAuthService(mr, c)

	apiKey := GenerateAPIKey()
	seedMember(mr, "mem1", "team1", "Admin", "a@b.com", "admin")
	// Set the hash to match
	mr.members["mem1"].APIKeyHash = HashAPIKey(apiKey)
	mr.members["mem1"].APIKeyPrefix = apiKey[:10]
	mr.byHash[HashAPIKey(apiKey)] = mr.members["mem1"]

	// First call - DB hit + cache set
	authCtx1, err := svc.Authenticate(context.Background(), apiKey)
	if err != nil {
		t.Fatalf("first call: unexpected error: %v", err)
	}

	// Delete from mock repo to simulate DB miss
	delete(mr.members, "mem1")
	delete(mr.byHash, HashAPIKey(apiKey))

	// Second call - should hit cache
	authCtx2, err := svc.Authenticate(context.Background(), apiKey)
	if err != nil {
		t.Fatalf("second call (cache): unexpected error: %v", err)
	}
	if authCtx2.MemberID != authCtx1.MemberID {
		t.Errorf("cache hit returned different member: got %s, want %s", authCtx2.MemberID, authCtx1.MemberID)
	}
}

// ========== TeamService Tests ==========

func TestTeamService_Register_ValidRequest(t *testing.T) {
	tr := newMockTeamRepo()
	mr := newMockMemberRepo()
	svc := NewTeamService(tr, mr)

	req := &request.RegisterTeamRequest{
		Name:       "My Team",
		AdminName:  "Alice",
		AdminEmail: "alice@example.com",
	}

	resp, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Team.Name != "My Team" {
		t.Errorf("expected team name 'My Team', got %s", resp.Data.Team.Name)
	}
	if resp.Data.Member.Name != "Alice" {
		t.Errorf("expected member name 'Alice', got %s", resp.Data.Member.Name)
	}
	if resp.Data.APIKey == "" {
		t.Error("expected non-empty API key")
	}
	if resp.Data.Team.ID == "" {
		t.Error("expected non-empty team ID")
	}
	if resp.Data.Member.ID == "" {
		t.Error("expected non-empty member ID")
	}
}

func TestTeamService_Register_EmptyName(t *testing.T) {
	tr := newMockTeamRepo()
	mr := newMockMemberRepo()
	svc := NewTeamService(tr, mr)

	req := &request.RegisterTeamRequest{
		Name:       "",
		AdminName:  "Alice",
		AdminEmail: "alice@example.com",
	}

	_, err := svc.Register(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty team name")
	}
	var valErrs domainErrors.ValidationErrors
	if !errors.As(err, &valErrs) {
		t.Errorf("expected ValidationErrors, got %T: %v", err, err)
	}
}

func TestTeamService_Register_InvalidEmail(t *testing.T) {
	tr := newMockTeamRepo()
	mr := newMockMemberRepo()
	svc := NewTeamService(tr, mr)

	req := &request.RegisterTeamRequest{
		Name:       "Team",
		AdminName:  "Alice",
		AdminEmail: "notanemail",
	}

	_, err := svc.Register(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid email")
	}
	var valErrs domainErrors.ValidationErrors
	if !errors.As(err, &valErrs) {
		t.Errorf("expected ValidationErrors, got %T: %v", err, err)
	}
}

func TestTeamService_Register_EmptyAdminName(t *testing.T) {
	tr := newMockTeamRepo()
	mr := newMockMemberRepo()
	svc := NewTeamService(tr, mr)

	req := &request.RegisterTeamRequest{
		Name:       "Team",
		AdminName:  "",
		AdminEmail: "alice@example.com",
	}

	_, err := svc.Register(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty admin name")
	}
}

func TestTeamService_Register_DuplicateTeamName(t *testing.T) {
	tr := newMockTeamRepo()
	mr := newMockMemberRepo()
	svc := NewTeamService(tr, mr)

	seedTeam(tr, "team1", "Existing Team")

	req := &request.RegisterTeamRequest{
		Name:       "Existing Team",
		AdminName:  "Alice",
		AdminEmail: "alice@example.com",
	}

	_, err := svc.Register(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for duplicate team name")
	}
	if !errors.Is(err, domainErrors.ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestTeamService_Register_MultipleValidationErrors(t *testing.T) {
	tr := newMockTeamRepo()
	mr := newMockMemberRepo()
	svc := NewTeamService(tr, mr)

	req := &request.RegisterTeamRequest{
		Name:       "",
		AdminName:  "",
		AdminEmail: "bad",
	}

	_, err := svc.Register(context.Background(), req)
	if err == nil {
		t.Fatal("expected error")
	}
	var valErrs domainErrors.ValidationErrors
	if !errors.As(err, &valErrs) {
		t.Errorf("expected ValidationErrors, got %T", err)
	}
	if len(valErrs) < 3 {
		t.Errorf("expected at least 3 validation errors, got %d", len(valErrs))
	}
}

func TestTeamService_Register_TeamCreationFails(t *testing.T) {
	tr := newMockTeamRepo()
	tr.err = errors.New("db error")
	mr := newMockMemberRepo()
	svc := NewTeamService(tr, mr)

	req := &request.RegisterTeamRequest{
		Name:       "Team",
		AdminName:  "Alice",
		AdminEmail: "alice@example.com",
	}

	_, err := svc.Register(context.Background(), req)
	if err == nil {
		t.Fatal("expected error when team creation fails")
	}
}
