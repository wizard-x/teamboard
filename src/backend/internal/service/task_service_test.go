package service

import (
	"context"
	"errors"
	"testing"

	domainErrors "teamboard/internal/domain/errors"
	"teamboard/internal/dto/request"
)

// ========== TaskService Tests ==========

func newTaskService() (*TaskService, *mockTaskRepo, *mockColumnRepo, *mockBoardRepo, *mockMemberRepo, *mockCommentRepo) {
	tr := newMockTaskRepo()
	cr := newMockColumnRepo()
	br := newMockBoardRepo()
	mr := newMockMemberRepo()
	cmr := newMockCommentRepo()
	c, _ := newTestCache()
	svc := NewTaskService(tr, cr, br, mr, cmr, c)
	return svc, tr, cr, br, mr, cmr
}

// --- Create ---

func TestTaskService_Create_ValidRequest(t *testing.T) {
	svc, _, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	// Need column
	colRepo := svc.columnRepo.(*mockColumnRepo)
	seedColumn(colRepo, "col1", "b1", "Todo", 0, "todo")

	req := &request.CreateTaskRequest{
		ColumnID: "col1",
		Title:    "My Task",
		Priority: "high",
	}

	resp, err := svc.Create(context.Background(), "team1", "mem1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Title != "My Task" {
		t.Errorf("expected title 'My Task', got %s", resp.Data.Title)
	}
	if resp.Data.Priority != "high" {
		t.Errorf("expected priority 'high', got %s", resp.Data.Priority)
	}
	if resp.Data.Status != "todo" {
		t.Errorf("expected status from column 'todo', got %s", resp.Data.Status)
	}
}

func TestTaskService_Create_EmptyTitle(t *testing.T) {
	svc, _, _, _, _, _ := newTaskService()

	req := &request.CreateTaskRequest{
		ColumnID: "col1",
		Title:    "",
	}

	_, err := svc.Create(context.Background(), "team1", "mem1", req)
	if err == nil {
		t.Fatal("expected error for empty title")
	}
	var valErrs domainErrors.ValidationErrors
	if !errors.As(err, &valErrs) {
		t.Errorf("expected ValidationErrors, got %T", err)
	}
}

func TestTaskService_Create_TitleTooLong(t *testing.T) {
	svc, _, _, _, _, _ := newTaskService()

	longTitle := make([]byte, 201)
	for i := range longTitle {
		longTitle[i] = 'a'
	}

	req := &request.CreateTaskRequest{
		ColumnID: "col1",
		Title:    string(longTitle),
	}

	_, err := svc.Create(context.Background(), "team1", "mem1", req)
	if err == nil {
		t.Fatal("expected error for title too long")
	}
}

func TestTaskService_Create_InvalidPriority(t *testing.T) {
	svc, _, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")

	req := &request.CreateTaskRequest{
		ColumnID: "col1",
		Title:    "Task",
		Priority: "urgent",
	}

	_, err := svc.Create(context.Background(), "team1", "mem1", req)
	if err == nil {
		t.Fatal("expected error for invalid priority")
	}
}

func TestTaskService_Create_DefaultPriority(t *testing.T) {
	svc, _, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	colRepo := svc.columnRepo.(*mockColumnRepo)
	seedColumn(colRepo, "col1", "b1", "Todo", 0, "todo")

	req := &request.CreateTaskRequest{
		ColumnID: "col1",
		Title:    "Task",
	}

	resp, err := svc.Create(context.Background(), "team1", "mem1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Priority != "medium" {
		t.Errorf("expected default priority 'medium', got %s", resp.Data.Priority)
	}
}

func TestTaskService_Create_ColumnNotFound(t *testing.T) {
	svc, _, _, _, _, _ := newTaskService()

	req := &request.CreateTaskRequest{
		ColumnID: "nonexistent",
		Title:    "Task",
	}

	_, err := svc.Create(context.Background(), "team1", "mem1", req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskService_Create_InvalidAssignee(t *testing.T) {
	svc, _, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	colRepo := svc.columnRepo.(*mockColumnRepo)
	seedColumn(colRepo, "col1", "b1", "Todo", 0, "todo")

	assigneeID := "nonexistent"
	req := &request.CreateTaskRequest{
		ColumnID:   "col1",
		Title:      "Task",
		AssigneeID: &assigneeID,
	}

	_, err := svc.Create(context.Background(), "team1", "mem1", req)
	if err == nil {
		t.Fatal("expected error for invalid assignee")
	}
}

func TestTaskService_Create_ValidAssignee(t *testing.T) {
	svc, _, _, br, mr, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	colRepo := svc.columnRepo.(*mockColumnRepo)
	seedColumn(colRepo, "col1", "b1", "Todo", 0, "todo")
	seedMember(mr, "mem2", "team1", "Bob", "bob@test.com", "member")

	assigneeID := "mem2"
	req := &request.CreateTaskRequest{
		ColumnID:   "col1",
		Title:      "Task",
		AssigneeID: &assigneeID,
	}

	resp, err := svc.Create(context.Background(), "team1", "mem1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Title != "Task" {
		t.Errorf("expected title 'Task', got %s", resp.Data.Title)
	}
}

func TestTaskService_Create_DescriptionTooLong(t *testing.T) {
	svc, _, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	colRepo := svc.columnRepo.(*mockColumnRepo)
	seedColumn(colRepo, "col1", "b1", "Todo", 0, "todo")

	longDesc := make([]byte, 2001)
	for i := range longDesc {
		longDesc[i] = 'a'
	}
	desc := string(longDesc)

	req := &request.CreateTaskRequest{
		ColumnID:    "col1",
		Title:       "Task",
		Description: &desc,
	}

	_, err := svc.Create(context.Background(), "team1", "mem1", req)
	if err == nil {
		t.Fatal("expected error for description too long")
	}
}

func TestTaskService_Create_InvalidDueDate(t *testing.T) {
	svc, _, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	colRepo := svc.columnRepo.(*mockColumnRepo)
	seedColumn(colRepo, "col1", "b1", "Todo", 0, "todo")

	badDate := "not-a-date"
	req := &request.CreateTaskRequest{
		ColumnID: "col1",
		Title:    "Task",
		DueDate:  &badDate,
	}

	_, err := svc.Create(context.Background(), "team1", "mem1", req)
	if err == nil {
		t.Fatal("expected error for invalid due date")
	}
}

func TestTaskService_Create_ValidDueDate(t *testing.T) {
	svc, _, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	colRepo := svc.columnRepo.(*mockColumnRepo)
	seedColumn(colRepo, "col1", "b1", "Todo", 0, "todo")

	dateStr := "2025-12-31T23:59:59Z"
	req := &request.CreateTaskRequest{
		ColumnID: "col1",
		Title:    "Task",
		DueDate:  &dateStr,
	}

	resp, err := svc.Create(context.Background(), "team1", "mem1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.DueDate == nil {
		t.Error("expected non-nil due date")
	}
}

// --- GetByID ---

func TestTaskService_GetByID_Success(t *testing.T) {
	svc, tr, _, br, _, cmr := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)
	seedComment(cmr, "c1", "task1", "mem1", "A comment")

	resp, err := svc.GetByID(context.Background(), "team1", "task1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Title != "Task" {
		t.Errorf("expected title 'Task', got %s", resp.Data.Title)
	}
	if len(resp.Data.Comments) != 1 {
		t.Errorf("expected 1 comment, got %d", len(resp.Data.Comments))
	}
}

func TestTaskService_GetByID_NotFound(t *testing.T) {
	svc, _, _, _, _, _ := newTaskService()

	_, err := svc.GetByID(context.Background(), "team1", "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskService_GetByID_WrongTeam(t *testing.T) {
	svc, tr, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team2", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)

	_, err := svc.GetByID(context.Background(), "team1", "task1")
	if err == nil {
		t.Fatal("expected error for wrong team")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- Update ---

func TestTaskService_Update_Title(t *testing.T) {
	svc, tr, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Old Title", "todo", "medium", 0)

	newTitle := "New Title"
	req := &request.UpdateTaskRequest{
		Title: &newTitle,
	}

	resp, err := svc.Update(context.Background(), "team1", "task1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Title != "New Title" {
		t.Errorf("expected title 'New Title', got %s", resp.Data.Title)
	}
}

func TestTaskService_Update_Priority(t *testing.T) {
	svc, tr, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)

	newPriority := "critical"
	req := &request.UpdateTaskRequest{
		Priority: &newPriority,
	}

	resp, err := svc.Update(context.Background(), "team1", "task1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Priority != "critical" {
		t.Errorf("expected priority 'critical', got %s", resp.Data.Priority)
	}
}

func TestTaskService_Update_InvalidPriority(t *testing.T) {
	svc, tr, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)

	badPriority := "urgent"
	req := &request.UpdateTaskRequest{
		Priority: &badPriority,
	}

	_, err := svc.Update(context.Background(), "team1", "task1", req)
	if err == nil {
		t.Fatal("expected error for invalid priority")
	}
}

func TestTaskService_Update_NotFound(t *testing.T) {
	svc, _, _, _, _, _ := newTaskService()

	newTitle := "Updated"
	req := &request.UpdateTaskRequest{
		Title: &newTitle,
	}

	_, err := svc.Update(context.Background(), "team1", "nonexistent", req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskService_Update_ClearDueDate(t *testing.T) {
	svc, tr, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)

	emptyDate := ""
	req := &request.UpdateTaskRequest{
		DueDate: &emptyDate,
	}

	resp, err := svc.Update(context.Background(), "team1", "task1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.DueDate != nil {
		t.Error("expected nil due date after clearing")
	}
}

func TestTaskService_Update_InvalidDueDate(t *testing.T) {
	svc, tr, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)

	badDate := "not-a-date"
	req := &request.UpdateTaskRequest{
		DueDate: &badDate,
	}

	_, err := svc.Update(context.Background(), "team1", "task1", req)
	if err == nil {
		t.Fatal("expected error for invalid due date")
	}
}

func TestTaskService_Update_EmptyTitle(t *testing.T) {
	svc, tr, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)

	emptyTitle := ""
	req := &request.UpdateTaskRequest{
		Title: &emptyTitle,
	}

	_, err := svc.Update(context.Background(), "team1", "task1", req)
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

// --- Move ---

func TestTaskService_Move_Success(t *testing.T) {
	svc, tr, cr, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	seedColumn(cr, "col1", "b1", "Todo", 0, "todo")
	seedColumn(cr, "col2", "b1", "Done", 1, "done")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)

	req := &request.MoveTaskRequest{
		ColumnID: "col2",
	}

	resp, err := svc.Move(context.Background(), "team1", "task1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.ColumnID != "col2" {
		t.Errorf("expected column_id col2, got %s", resp.Data.ColumnID)
	}
}

func TestTaskService_Move_TaskNotFound(t *testing.T) {
	svc, _, _, _, _, _ := newTaskService()

	req := &request.MoveTaskRequest{
		ColumnID: "col2",
	}

	_, err := svc.Move(context.Background(), "team1", "nonexistent", req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskService_Move_ColumnNotFound(t *testing.T) {
	svc, tr, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)

	req := &request.MoveTaskRequest{
		ColumnID: "nonexistent",
	}

	_, err := svc.Move(context.Background(), "team1", "task1", req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- Delete ---

func TestTaskService_Delete_Success(t *testing.T) {
	svc, tr, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team1", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)

	err := svc.Delete(context.Background(), "team1", "task1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskService_Delete_NotFound(t *testing.T) {
	svc, _, _, _, _, _ := newTaskService()

	err := svc.Delete(context.Background(), "team1", "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskService_Delete_WrongTeam(t *testing.T) {
	svc, tr, _, br, _, _ := newTaskService()
	seedBoard(br, "b1", "team2", "Board")
	seedTask(tr, "task1", "col1", "b1", "Task", "todo", "medium", 0)

	err := svc.Delete(context.Background(), "team1", "task1")
	if err == nil {
		t.Fatal("expected error for wrong team")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
