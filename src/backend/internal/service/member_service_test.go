package service

import (
	"context"
	"errors"
	"testing"

	domainErrors "teamboard/internal/domain/errors"
	"teamboard/internal/dto/request"
)

// ========== MemberService Tests ==========

func newMemberService() (*MemberService, *mockMemberRepo) {
	mr := newMockMemberRepo()
	c, _ := newTestCache()
	svc := NewMemberService(mr, c)
	return svc, mr
}

func TestMemberService_GetCurrentMember_Found(t *testing.T) {
	svc, mr := newMemberService()
	seedMember(mr, "mem1", "team1", "Alice", "alice@test.com", "admin")

	resp, err := svc.GetCurrentMember(context.Background(), "mem1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.ID != "mem1" {
		t.Errorf("expected ID mem1, got %s", resp.Data.ID)
	}
	if resp.Data.Name != "Alice" {
		t.Errorf("expected name Alice, got %s", resp.Data.Name)
	}
	if resp.Data.Role != "admin" {
		t.Errorf("expected role admin, got %s", resp.Data.Role)
	}
}

func TestMemberService_GetCurrentMember_NotFound(t *testing.T) {
	svc, _ := newMemberService()

	_, err := svc.GetCurrentMember(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent member")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemberService_ListByTeam(t *testing.T) {
	svc, mr := newMemberService()
	seedMember(mr, "mem1", "team1", "Alice", "alice@test.com", "admin")
	seedMember(mr, "mem2", "team1", "Bob", "bob@test.com", "member")
	seedMember(mr, "mem3", "team2", "Charlie", "charlie@test.com", "admin")

	resp, err := svc.ListByTeam(context.Background(), "team1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 members, got %d", len(resp.Data))
	}
}

func TestMemberService_ListByTeam_Empty(t *testing.T) {
	svc, _ := newMemberService()

	resp, err := svc.ListByTeam(context.Background(), "team1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0 members, got %d", len(resp.Data))
	}
}

func TestMemberService_Create_ValidRequest(t *testing.T) {
	svc, _ := newMemberService()

	req := &request.CreateMemberRequest{
		Name:  "Bob",
		Email: "bob@test.com",
		Role:  "member",
	}

	resp, err := svc.Create(context.Background(), "team1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Name != "Bob" {
		t.Errorf("expected name Bob, got %s", resp.Data.Name)
	}
	if resp.Data.Role != "member" {
		t.Errorf("expected role member, got %s", resp.Data.Role)
	}
	if resp.Data.APIKey == "" {
		t.Error("expected non-empty API key")
	}
}

func TestMemberService_Create_EmptyName(t *testing.T) {
	svc, _ := newMemberService()

	req := &request.CreateMemberRequest{
		Name:  "",
		Email: "bob@test.com",
		Role:  "member",
	}

	_, err := svc.Create(context.Background(), "team1", req)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	var valErrs domainErrors.ValidationErrors
	if !errors.As(err, &valErrs) {
		t.Errorf("expected ValidationErrors, got %T", err)
	}
}

func TestMemberService_Create_InvalidEmail(t *testing.T) {
	svc, _ := newMemberService()

	req := &request.CreateMemberRequest{
		Name:  "Bob",
		Email: "bad",
		Role:  "member",
	}

	_, err := svc.Create(context.Background(), "team1", req)
	if err == nil {
		t.Fatal("expected error for invalid email")
	}
}

func TestMemberService_Create_InvalidRole(t *testing.T) {
	svc, _ := newMemberService()

	req := &request.CreateMemberRequest{
		Name:  "Bob",
		Email: "bob@test.com",
		Role:  "superadmin",
	}

	_, err := svc.Create(context.Background(), "team1", req)
	if err == nil {
		t.Fatal("expected error for invalid role")
	}
}

func TestMemberService_Create_AdminRole(t *testing.T) {
	svc, _ := newMemberService()

	req := &request.CreateMemberRequest{
		Name:  "Carol",
		Email: "carol@test.com",
		Role:  "admin",
	}

	resp, err := svc.Create(context.Background(), "team1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Role != "admin" {
		t.Errorf("expected role admin, got %s", resp.Data.Role)
	}
}

func TestMemberService_Update_Name(t *testing.T) {
	svc, mr := newMemberService()
	seedMember(mr, "mem1", "team1", "Alice", "alice@test.com", "admin")

	req := &request.UpdateMemberRequest{
		Name: "Alice Updated",
	}

	resp, err := svc.Update(context.Background(), "team1", "mem1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Name != "Alice Updated" {
		t.Errorf("expected name 'Alice Updated', got %s", resp.Data.Name)
	}
}

func TestMemberService_Update_RoleToMember(t *testing.T) {
	svc, mr := newMemberService()
	seedMember(mr, "mem1", "team1", "Alice", "alice@test.com", "admin")
	seedMember(mr, "mem2", "team1", "Bob", "bob@test.com", "admin")

	req := &request.UpdateMemberRequest{
		Role: "member",
	}

	resp, err := svc.Update(context.Background(), "team1", "mem1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Role != "member" {
		t.Errorf("expected role member, got %s", resp.Data.Role)
	}
}

func TestMemberService_Update_DemoteLastAdmin(t *testing.T) {
	svc, mr := newMemberService()
	seedMember(mr, "mem1", "team1", "Alice", "alice@test.com", "admin")

	req := &request.UpdateMemberRequest{
		Role: "member",
	}

	_, err := svc.Update(context.Background(), "team1", "mem1", req)
	if err == nil {
		t.Fatal("expected error when demoting last admin")
	}
	if !errors.Is(err, domainErrors.ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

func TestMemberService_Update_InvalidRole(t *testing.T) {
	svc, mr := newMemberService()
	seedMember(mr, "mem1", "team1", "Alice", "alice@test.com", "admin")
	seedMember(mr, "mem2", "team1", "Bob", "bob@test.com", "admin")

	req := &request.UpdateMemberRequest{
		Role: "superadmin",
	}

	_, err := svc.Update(context.Background(), "team1", "mem1", req)
	if err == nil {
		t.Fatal("expected error for invalid role")
	}
	if !errors.Is(err, domainErrors.ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

func TestMemberService_Update_NotFound(t *testing.T) {
	svc, _ := newMemberService()

	req := &request.UpdateMemberRequest{
		Name: "Nobody",
	}

	_, err := svc.Update(context.Background(), "team1", "nonexistent", req)
	if err == nil {
		t.Fatal("expected error for nonexistent member")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemberService_Update_WrongTeam(t *testing.T) {
	svc, mr := newMemberService()
	seedMember(mr, "mem1", "team1", "Alice", "alice@test.com", "admin")

	req := &request.UpdateMemberRequest{
		Name: "Updated",
	}

	_, err := svc.Update(context.Background(), "team2", "mem1", req)
	if err == nil {
		t.Fatal("expected error for wrong team")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemberService_Delete_Success(t *testing.T) {
	svc, mr := newMemberService()
	seedMember(mr, "mem1", "team1", "Alice", "alice@test.com", "admin")
	seedMember(mr, "mem2", "team1", "Bob", "bob@test.com", "member")

	err := svc.Delete(context.Background(), "team1", "mem2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemberService_Delete_NotFound(t *testing.T) {
	svc, _ := newMemberService()

	err := svc.Delete(context.Background(), "team1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent member")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemberService_Delete_LastAdmin(t *testing.T) {
	svc, mr := newMemberService()
	seedMember(mr, "mem1", "team1", "Alice", "alice@test.com", "admin")

	err := svc.Delete(context.Background(), "team1", "mem1")
	if err == nil {
		t.Fatal("expected error when deleting last admin")
	}
	if !errors.Is(err, domainErrors.ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

func TestMemberService_Delete_WrongTeam(t *testing.T) {
	svc, mr := newMemberService()
	seedMember(mr, "mem1", "team1", "Alice", "alice@test.com", "admin")

	err := svc.Delete(context.Background(), "team2", "mem1")
	if err == nil {
		t.Fatal("expected error for wrong team")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemberService_RegenerateAPIKey_Success(t *testing.T) {
	svc, mr := newMemberService()
	seedMember(mr, "mem1", "team1", "Alice", "alice@test.com", "admin")

	resp, err := svc.RegenerateAPIKey(context.Background(), "team1", "mem1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.APIKey == "" {
		t.Error("expected non-empty API key")
	}
}

func TestMemberService_RegenerateAPIKey_NotFound(t *testing.T) {
	svc, _ := newMemberService()

	_, err := svc.RegenerateAPIKey(context.Background(), "team1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent member")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemberService_RegenerateAPIKey_WrongTeam(t *testing.T) {
	svc, mr := newMemberService()
	seedMember(mr, "mem1", "team1", "Alice", "alice@test.com", "admin")

	_, err := svc.RegenerateAPIKey(context.Background(), "team2", "mem1")
	if err == nil {
		t.Fatal("expected error for wrong team")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
