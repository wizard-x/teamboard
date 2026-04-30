package service

import (
	"context"
	"errors"
	"testing"

	domainErrors "teamboard/internal/domain/errors"
	"teamboard/internal/dto/request"
)

// ========== BoardService Tests ==========

func newBoardService() (*BoardService, *mockBoardRepo, *mockColumnRepo, *mockTaskRepo, *mockMemberRepo, *mockCommentRepo) {
	br := newMockBoardRepo()
	cr := newMockColumnRepo()
	tr := newMockTaskRepo()
	mr := newMockMemberRepo()
	cmtr := newMockCommentRepo()
	c, _ := newTestCache()
	svc := NewBoardService(br, cr, tr, mr, cmtr, c)
	return svc, br, cr, tr, mr, cmtr
}

// --- Create ---

func TestBoardService_Create_ValidRequest(t *testing.T) {
	svc, _, _, _, _, _ := newBoardService()

	desc := "A test board"
	req := &request.CreateBoardRequest{
		Name:        "My Board",
		Description: &desc,
	}

	resp, err := svc.Create(context.Background(), "team1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Name != "My Board" {
		t.Errorf("expected name 'My Board', got %s", resp.Data.Name)
	}
	if resp.Data.Description == nil || *resp.Data.Description != "A test board" {
		t.Error("expected description 'A test board'")
	}
	// Should have default columns
	if len(resp.Data.Columns) != 4 {
		t.Errorf("expected 4 default columns, got %d", len(resp.Data.Columns))
	}
	expectedNames := []string{"Todo", "In Progress", "Review", "Done"}
	for i, col := range resp.Data.Columns {
		if col.Name != expectedNames[i] {
			t.Errorf("column %d: expected name %s, got %s", i, expectedNames[i], col.Name)
		}
		if len(col.Tasks) != 0 {
			t.Errorf("column %d: expected empty tasks, got %d", i, len(col.Tasks))
		}
	}
}

func TestBoardService_Create_EmptyName(t *testing.T) {
	svc, _, _, _, _, _ := newBoardService()

	req := &request.CreateBoardRequest{
		Name: "",
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

func TestBoardService_Create_NameTooLong(t *testing.T) {
	svc, _, _, _, _, _ := newBoardService()

	longName := make([]byte, 101)
	for i := range longName {
		longName[i] = 'a'
	}

	req := &request.CreateBoardRequest{
		Name: string(longName),
	}

	_, err := svc.Create(context.Background(), "team1", req)
	if err == nil {
		t.Fatal("expected error for name too long")
	}
}

func TestBoardService_Create_DescriptionTooLong(t *testing.T) {
	svc, _, _, _, _, _ := newBoardService()

	longDesc := make([]byte, 501)
	for i := range longDesc {
		longDesc[i] = 'a'
	}
	desc := string(longDesc)

	req := &request.CreateBoardRequest{
		Name:        "Board",
		Description: &desc,
	}

	_, err := svc.Create(context.Background(), "team1", req)
	if err == nil {
		t.Fatal("expected error for description too long")
	}
}

func TestBoardService_Create_NoDescription(t *testing.T) {
	svc, _, _, _, _, _ := newBoardService()

	req := &request.CreateBoardRequest{
		Name: "Board",
	}

	resp, err := svc.Create(context.Background(), "team1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Description != nil {
		t.Errorf("expected nil description, got %v", resp.Data.Description)
	}
}

// --- ListByTeam ---

func TestBoardService_ListByTeam(t *testing.T) {
	svc, br, _, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")
	seedBoard(br, "b2", "team1", "Board 2")
	seedBoard(br, "b3", "team2", "Other Board")

	resp, err := svc.ListByTeam(context.Background(), "team1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 boards, got %d", len(resp.Data))
	}
}

func TestBoardService_ListByTeam_Empty(t *testing.T) {
	svc, _, _, _, _, _ := newBoardService()

	resp, err := svc.ListByTeam(context.Background(), "team1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0 boards, got %d", len(resp.Data))
	}
}

// --- GetByID ---

func TestBoardService_GetByID_Success(t *testing.T) {
	svc, br, cr, tr, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")
	seedColumn(cr, "col1", "b1", "Todo", 0, "todo")
	seedColumn(cr, "col2", "b1", "Done", 1, "done")
	seedTask(tr, "task1", "col1", "b1", "My Task", "todo", "medium", 0)

	resp, err := svc.GetByID(context.Background(), "team1", "b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Name != "Board 1" {
		t.Errorf("expected name 'Board 1', got %s", resp.Data.Name)
	}
	if len(resp.Data.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(resp.Data.Columns))
	}
	// First column should have 1 task
	if len(resp.Data.Columns[0].Tasks) != 1 {
		t.Errorf("expected 1 task in first column, got %d", len(resp.Data.Columns[0].Tasks))
	}
	// Second column should have empty tasks slice
	if resp.Data.Columns[1].Tasks == nil {
		t.Error("expected non-nil empty tasks slice")
	}
}

func TestBoardService_GetByID_NotFound(t *testing.T) {
	svc, _, _, _, _, _ := newBoardService()

	_, err := svc.GetByID(context.Background(), "team1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent board")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestBoardService_GetByID_WrongTeam(t *testing.T) {
	svc, br, _, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")

	_, err := svc.GetByID(context.Background(), "team2", "b1")
	if err == nil {
		t.Fatal("expected error for wrong team")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- Update ---

func TestBoardService_Update_Name(t *testing.T) {
	svc, br, _, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")

	newName := "Updated Board"
	req := &request.UpdateBoardRequest{
		Name: &newName,
	}

	resp, err := svc.Update(context.Background(), "team1", "b1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Name != "Updated Board" {
		t.Errorf("expected name 'Updated Board', got %s", resp.Data.Name)
	}
}

func TestBoardService_Update_Description(t *testing.T) {
	svc, br, _, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")

	newDesc := "New description"
	req := &request.UpdateBoardRequest{
		Description: &newDesc,
	}

	resp, err := svc.Update(context.Background(), "team1", "b1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Description == nil || *resp.Data.Description != "New description" {
		t.Error("expected description 'New description'")
	}
}

func TestBoardService_Update_NotFound(t *testing.T) {
	svc, _, _, _, _, _ := newBoardService()

	newName := "Updated"
	req := &request.UpdateBoardRequest{
		Name: &newName,
	}

	_, err := svc.Update(context.Background(), "team1", "nonexistent", req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestBoardService_Update_EmptyName(t *testing.T) {
	svc, br, _, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")

	emptyName := ""
	req := &request.UpdateBoardRequest{
		Name: &emptyName,
	}

	_, err := svc.Update(context.Background(), "team1", "b1", req)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

// --- Delete ---

func TestBoardService_Delete_Success(t *testing.T) {
	svc, br, _, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")

	err := svc.Delete(context.Background(), "team1", "b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBoardService_Delete_NotFound(t *testing.T) {
	svc, _, _, _, _, _ := newBoardService()

	err := svc.Delete(context.Background(), "team1", "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestBoardService_Delete_WrongTeam(t *testing.T) {
	svc, br, _, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")

	err := svc.Delete(context.Background(), "team2", "b1")
	if err == nil {
		t.Fatal("expected error for wrong team")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- AddColumn ---

func TestBoardService_AddColumn_ValidRequest(t *testing.T) {
	svc, br, _, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")

	req := &request.CreateColumnRequest{
		Name:   "New Column",
		Status: "todo",
	}

	resp, err := svc.AddColumn(context.Background(), "team1", "b1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Name != "New Column" {
		t.Errorf("expected name 'New Column', got %s", resp.Data.Name)
	}
	if resp.Data.BoardID != "b1" {
		t.Errorf("expected board_id b1, got %s", resp.Data.BoardID)
	}
}

func TestBoardService_AddColumn_EmptyName(t *testing.T) {
	svc, br, _, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")

	req := &request.CreateColumnRequest{
		Name: "",
	}

	_, err := svc.AddColumn(context.Background(), "team1", "b1", req)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestBoardService_AddColumn_NameTooLong(t *testing.T) {
	svc, br, _, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")

	longName := make([]byte, 51)
	for i := range longName {
		longName[i] = 'a'
	}

	req := &request.CreateColumnRequest{
		Name: string(longName),
	}

	_, err := svc.AddColumn(context.Background(), "team1", "b1", req)
	if err == nil {
		t.Fatal("expected error for name too long")
	}
}

func TestBoardService_AddColumn_InvalidStatus(t *testing.T) {
	svc, br, _, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")

	req := &request.CreateColumnRequest{
		Name:   "Column",
		Status: "invalid",
	}

	_, err := svc.AddColumn(context.Background(), "team1", "b1", req)
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestBoardService_AddColumn_DefaultStatus(t *testing.T) {
	svc, br, _, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")

	req := &request.CreateColumnRequest{
		Name: "Column",
	}

	resp, err := svc.AddColumn(context.Background(), "team1", "b1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Status != "todo" {
		t.Errorf("expected default status 'todo', got %s", resp.Data.Status)
	}
}

func TestBoardService_AddColumn_BoardNotFound(t *testing.T) {
	svc, _, _, _, _, _ := newBoardService()

	req := &request.CreateColumnRequest{
		Name: "Column",
	}

	_, err := svc.AddColumn(context.Background(), "team1", "nonexistent", req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestBoardService_AddColumn_WithPosition(t *testing.T) {
	svc, br, cr, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")
	seedColumn(cr, "col1", "b1", "Existing", 0, "todo")

	pos := 0
	req := &request.CreateColumnRequest{
		Name:     "Inserted",
		Position: &pos,
		Status:   "todo",
	}

	resp, err := svc.AddColumn(context.Background(), "team1", "b1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Position != 0 {
		t.Errorf("expected position 0, got %d", resp.Data.Position)
	}
}

// --- RenameColumn ---

func TestBoardService_RenameColumn_Success(t *testing.T) {
	svc, br, cr, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board 1")
	seedColumn(cr, "col1", "b1", "Old Name", 0, "todo")

	req := &request.UpdateColumnRequest{
		Name: "New Name",
	}

	resp, err := svc.RenameColumn(context.Background(), "team1", "col1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Name != "New Name" {
		t.Errorf("expected name 'New Name', got %s", resp.Data.Name)
	}
}

func TestBoardService_RenameColumn_NotFound(t *testing.T) {
	svc, _, _, _, _, _ := newBoardService()

	req := &request.UpdateColumnRequest{
		Name: "New Name",
	}

	_, err := svc.RenameColumn(context.Background(), "team1", "nonexistent", req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestBoardService_RenameColumn_WrongTeam(t *testing.T) {
	svc, br, cr, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team2", "Board")
	seedColumn(cr, "col1", "b1", "Column", 0, "todo")

	req := &request.UpdateColumnRequest{
		Name: "New Name",
	}

	_, err := svc.RenameColumn(context.Background(), "team1", "col1", req)
	if err == nil {
		t.Fatal("expected error for wrong team")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- ReorderColumn ---

func TestBoardService_ReorderColumn_Success(t *testing.T) {
	svc, br, cr, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board")
	seedColumn(cr, "col1", "b1", "Col 1", 0, "todo")
	seedColumn(cr, "col2", "b1", "Col 2", 1, "in_progress")
	seedColumn(cr, "col3", "b1", "Col 3", 2, "done")

	req := &request.ReorderColumnRequest{
		Position: 0,
	}

	resp, err := svc.ReorderColumn(context.Background(), "team1", "col3", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(resp.Data))
	}
	// col3 should be at position 0 now
	if resp.Data[0].ID != "col3" {
		t.Errorf("expected col3 at position 0, got %s", resp.Data[0].ID)
	}
}

func TestBoardService_ReorderColumn_NegativePosition(t *testing.T) {
	svc, br, cr, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board")
	seedColumn(cr, "col1", "b1", "Col 1", 0, "todo")

	req := &request.ReorderColumnRequest{
		Position: -1,
	}

	_, err := svc.ReorderColumn(context.Background(), "team1", "col1", req)
	if err == nil {
		t.Fatal("expected error for negative position")
	}
	if !errors.Is(err, domainErrors.ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

// --- DeleteColumn ---

func TestBoardService_DeleteColumn_Success(t *testing.T) {
	svc, br, cr, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board")
	seedColumn(cr, "col1", "b1", "Col 1", 0, "todo")
	seedColumn(cr, "col2", "b1", "Col 2", 1, "done")

	err := svc.DeleteColumn(context.Background(), "team1", "col2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBoardService_DeleteColumn_LastColumn(t *testing.T) {
	svc, br, cr, _, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board")
	seedColumn(cr, "col1", "b1", "Only Col", 0, "todo")

	err := svc.DeleteColumn(context.Background(), "team1", "col1")
	if err == nil {
		t.Fatal("expected error when deleting last column")
	}
	if !errors.Is(err, domainErrors.ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

func TestBoardService_DeleteColumn_HasTasks(t *testing.T) {
	svc, br, cr, tr, _, _ := newBoardService()
	seedBoard(br, "b1", "team1", "Board")
	seedColumn(cr, "col1", "b1", "Col 1", 0, "todo")
	seedColumn(cr, "col2", "b1", "Col 2", 1, "done")
	seedTask(tr, "task1", "col2", "b1", "Task", "done", "medium", 0)

	// The mock CountTasks returns 0, so we need to override the column repo
	// to simulate tasks existing. Override CountTasks:
	cr2 := &columnRepoWithTasks{mockColumnRepo: cr, taskCounts: map[string]int{"col2": 1}}
	svc2 := &BoardService{
		boardRepo: br, columnRepo: cr2, taskRepo: tr,
		memberRepo: newMockMemberRepo(), commentRepo: newMockCommentRepo(),
		cache: svc.cache,
	}

	err := svc2.DeleteColumn(context.Background(), "team1", "col2")
	if err == nil {
		t.Fatal("expected error when column has tasks")
	}
	if !errors.Is(err, domainErrors.ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

// columnRepoWithTasks wraps mockColumnRepo to override CountTasks
type columnRepoWithTasks struct {
	*mockColumnRepo
	taskCounts map[string]int
}

func (c *columnRepoWithTasks) CountTasks(_ context.Context, columnID string) (int, error) {
	if count, ok := c.taskCounts[columnID]; ok {
		return count, nil
	}
	return 0, nil
}

func TestBoardService_DeleteColumn_NotFound(t *testing.T) {
	svc, _, _, _, _, _ := newBoardService()

	err := svc.DeleteColumn(context.Background(), "team1", "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domainErrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
