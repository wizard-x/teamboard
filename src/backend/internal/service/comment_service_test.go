package service

import (
	"context"
	"errors"
	"testing"

	domainErrors "teamboard/internal/domain/errors"
	"teamboard/internal/dto/request"
)

// ========== CommentService Tests ==========

func newCommentService() (*CommentService, *mockCommentRepo, *mockTaskRepo, *mockBoardRepo, *mockMemberRepo) {
	cmr := newMockCommentRepo()
	tr := newMockTaskRepo()
	br := newMockBoardRepo()
	mr := newMockMemberRepo()
	c, _ := newTestCache()
	svc := NewCommentService(cmr, tr, br, mr, c)
	return svc, cmr, tr, br, mr
}

func TestCommentService_Create_ValidRequest(t *testing.T) {
	svc, _, tr, br, mr := newCommentService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)
	seedMember(mr, "mem1", "team1", "Author", "auth@test.com", "member")

	req := &request.CreateCommentRequest{
		Body: "Great work!",
	}

	resp, err := svc.Create(context.Background(), "team1", "task1", "mem1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Body != "Great work!" {
		t.Errorf("expected body 'Great work!', got %s", resp.Data.Body)
	}
	if resp.Data.Author == nil || resp.Data.Author.Name != "Author" {
		t.Error("expected author name 'Author'")
	}
}

func TestCommentService_Create_EmptyBody(t *testing.T) {
	svc, _, _, _, _ := newCommentService()

	req := &request.CreateCommentRequest{
		Body: "",
	}

	_, err := svc.Create(context.Background(), "team1", "task1", "mem1", req)
	if err == nil {
		t.Fatal("expected error for empty body")
	}
}

func TestCommentService_Create_BodyTooLong(t *testing.T) {
	svc, _, _, _, _ := newCommentService()

	longBody := make([]byte, 2001)
	for i := range longBody {
		longBody[i] = 'a'
	}

	req := &request.CreateCommentRequest{
		Body: string(longBody),
	}

	_, err := svc.Create(context.Background(), "team1", "task1", "mem1", req)
	if err == nil {
		t.Fatal("expected error for body too long")
	}
}

func TestCommentService_Create_TaskNotFound(t *testing.T) {
	svc, _, _, _, _ := newCommentService()

	req := &request.CreateCommentRequest{
		Body: "Comment",
	}

	_, err := svc.Create(context.Background(), "team1", "nonexistent", "mem1", req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCommentService_Create_WrongTeam(t *testing.T) {
	svc, _, tr, br, _ := newCommentService()
	seedBoard(br, "b1", "team2", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)

	req := &request.CreateCommentRequest{
		Body: "Comment",
	}

	_, err := svc.Create(context.Background(), "team1", "task1", "mem1", req)
	if err == nil {
		t.Fatal("expected error for wrong team")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCommentService_Create_AuthorNotFound(t *testing.T) {
	svc, _, tr, br, _ := newCommentService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)

	req := &request.CreateCommentRequest{
		Body: "Comment",
	}

	_, err := svc.Create(context.Background(), "team1", "task1", "nonexistent", req)
	if err == nil {
		t.Fatal("expected error for nonexistent author")
	}
}

// --- ListByTask ---

func TestCommentService_ListByTask_Success(t *testing.T) {
	svc, cmr, tr, br, _ := newCommentService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)
	seedComment(cmr, "c1", "task1", "mem1", "First")
	seedComment(cmr, "c2", "task1", "mem1", "Second")

	resp, err := svc.ListByTask(context.Background(), "team1", "task1", 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(resp.Data))
	}
	if resp.Meta.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Meta.Total)
	}
}

func TestCommentService_ListByTask_Pagination(t *testing.T) {
	svc, cmr, tr, br, _ := newCommentService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)
	for i := 0; i < 5; i++ {
		seedComment(cmr, "c"+string(rune('0'+i)), "task1", "mem1", "Comment")
	}

	resp, err := svc.ListByTask(context.Background(), "team1", "task1", 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Errorf("expected 2 comments on page 1, got %d", len(resp.Data))
	}
	if resp.Meta.Total != 5 {
		t.Errorf("expected total 5, got %d", resp.Meta.Total)
	}
	if resp.Meta.TotalPages != 3 {
		t.Errorf("expected 3 total pages, got %d", resp.Meta.TotalPages)
	}
}

func TestCommentService_ListByTask_TaskNotFound(t *testing.T) {
	svc, _, _, _, _ := newCommentService()

	_, err := svc.ListByTask(context.Background(), "team1", "nonexistent", 1, 20)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCommentService_ListByTask_PerPageClamped(t *testing.T) {
	svc, cmr, tr, br, _ := newCommentService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)
	seedComment(cmr, "c1", "task1", "mem1", "Comment")

	resp, err := svc.ListByTask(context.Background(), "team1", "task1", 1, 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Meta.PerPage != 100 {
		t.Errorf("expected perPage clamped to 100, got %d", resp.Meta.PerPage)
	}
}

func TestCommentService_ListByTask_DefaultPagination(t *testing.T) {
	svc, cmr, tr, br, _ := newCommentService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)
	seedComment(cmr, "c1", "task1", "mem1", "Comment")

	resp, err := svc.ListByTask(context.Background(), "team1", "task1", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Meta.Page != 1 {
		t.Errorf("expected page 1, got %d", resp.Meta.Page)
	}
	if resp.Meta.PerPage != 20 {
		t.Errorf("expected perPage 20, got %d", resp.Meta.PerPage)
	}
}

// --- Delete ---

func TestCommentService_Delete_AsAuthor(t *testing.T) {
	svc, cmr, tr, br, _ := newCommentService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)
	seedComment(cmr, "c1", "task1", "mem1", "Comment")

	err := svc.Delete(context.Background(), "team1", "c1", "mem1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCommentService_Delete_AsAdmin(t *testing.T) {
	svc, cmr, tr, br, _ := newCommentService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)
	seedComment(cmr, "c1", "task1", "mem1", "Comment")

	err := svc.Delete(context.Background(), "team1", "c1", "admin1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCommentService_Delete_NotAuthor_NotAdmin(t *testing.T) {
	svc, cmr, tr, br, _ := newCommentService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)
	seedComment(cmr, "c1", "task1", "mem1", "Comment")

	err := svc.Delete(context.Background(), "team1", "c1", "mem2", false)
	if err == nil {
		t.Fatal("expected error when non-author non-admin deletes comment")
	}
	if !errors.Is(err, domainErrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestCommentService_Delete_CommentNotFound(t *testing.T) {
	svc, _, _, _, _ := newCommentService()

	err := svc.Delete(context.Background(), "team1", "nonexistent", "mem1", false)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
