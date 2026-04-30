package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	domainErrors "teamboard/internal/domain/errors"
	"teamboard/internal/dto/request"
	"teamboard/internal/dto/response"
	"teamboard/internal/service"
)

// ========== Mock Service Implementations ==========

// MockTeamManager
type mockTeamManager struct {
	registerFn func(ctx context.Context, req *request.RegisterTeamRequest) (*response.RegisterTeamResponse, error)
}

func (m *mockTeamManager) Register(ctx context.Context, req *request.RegisterTeamRequest) (*response.RegisterTeamResponse, error) {
	return m.registerFn(ctx, req)
}

// MockMemberManager
type mockMemberManager struct {
	listByTeamFn       func(ctx context.Context, teamID string) (*response.MemberListResponse, error)
	getCurrentMemberFn func(ctx context.Context, memberID string) (*response.MemberResponse, error)
	createFn           func(ctx context.Context, teamID string, req *request.CreateMemberRequest) (*response.MemberWithKeyResponse, error)
	updateFn           func(ctx context.Context, teamID, memberID string, req *request.UpdateMemberRequest) (*response.MemberResponse, error)
	deleteFn           func(ctx context.Context, teamID, memberID string) error
	regenKeyFn         func(ctx context.Context, teamID, memberID string) (*response.APIKeyResponse, error)
	updateMeFn         func(ctx context.Context, memberID string, req *request.UpdateMeRequest) (*response.MemberResponse, error)
}

func (m *mockMemberManager) ListByTeam(ctx context.Context, teamID string) (*response.MemberListResponse, error) {
	return m.listByTeamFn(ctx, teamID)
}
func (m *mockMemberManager) GetCurrentMember(ctx context.Context, memberID string) (*response.MemberResponse, error) {
	return m.getCurrentMemberFn(ctx, memberID)
}
func (m *mockMemberManager) Create(ctx context.Context, teamID string, req *request.CreateMemberRequest) (*response.MemberWithKeyResponse, error) {
	return m.createFn(ctx, teamID, req)
}
func (m *mockMemberManager) Update(ctx context.Context, teamID, memberID string, req *request.UpdateMemberRequest) (*response.MemberResponse, error) {
	return m.updateFn(ctx, teamID, memberID, req)
}
func (m *mockMemberManager) Delete(ctx context.Context, teamID, memberID string) error {
	return m.deleteFn(ctx, teamID, memberID)
}
func (m *mockMemberManager) RegenerateAPIKey(ctx context.Context, teamID, memberID string) (*response.APIKeyResponse, error) {
	return m.regenKeyFn(ctx, teamID, memberID)
}
func (m *mockMemberManager) UpdateMe(ctx context.Context, memberID string, req *request.UpdateMeRequest) (*response.MemberResponse, error) {
	return m.updateMeFn(ctx, memberID, req)
}

// MockBoardManager
type mockBoardManager struct {
	createFn       func(ctx context.Context, teamID string, req *request.CreateBoardRequest) (*response.BoardDetailResponse, error)
	listByTeamFn   func(ctx context.Context, teamID string) (*response.BoardListResponse, error)
	getByIDFn      func(ctx context.Context, teamID, boardID string) (*response.BoardDetailResponse, error)
	updateFn       func(ctx context.Context, teamID, boardID string, req *request.UpdateBoardRequest) (*response.BoardResponse, error)
	deleteFn       func(ctx context.Context, teamID, boardID string) error
	addColumnFn    func(ctx context.Context, teamID, boardID string, req *request.CreateColumnRequest) (*response.ColumnResponse, error)
	renameColumnFn func(ctx context.Context, teamID, columnID string, req *request.UpdateColumnRequest) (*response.ColumnResponse, error)
	reorderColFn   func(ctx context.Context, teamID, columnID string, req *request.ReorderColumnRequest) (*response.ColumnListResponse, error)
	deleteColumnFn func(ctx context.Context, teamID, columnID string) error
}

func (m *mockBoardManager) Create(ctx context.Context, teamID string, req *request.CreateBoardRequest) (*response.BoardDetailResponse, error) {
	return m.createFn(ctx, teamID, req)
}
func (m *mockBoardManager) ListByTeam(ctx context.Context, teamID string) (*response.BoardListResponse, error) {
	return m.listByTeamFn(ctx, teamID)
}
func (m *mockBoardManager) GetByID(ctx context.Context, teamID, boardID string) (*response.BoardDetailResponse, error) {
	return m.getByIDFn(ctx, teamID, boardID)
}
func (m *mockBoardManager) Update(ctx context.Context, teamID, boardID string, req *request.UpdateBoardRequest) (*response.BoardResponse, error) {
	return m.updateFn(ctx, teamID, boardID, req)
}
func (m *mockBoardManager) Delete(ctx context.Context, teamID, boardID string) error {
	return m.deleteFn(ctx, teamID, boardID)
}
func (m *mockBoardManager) AddColumn(ctx context.Context, teamID, boardID string, req *request.CreateColumnRequest) (*response.ColumnResponse, error) {
	return m.addColumnFn(ctx, teamID, boardID, req)
}
func (m *mockBoardManager) RenameColumn(ctx context.Context, teamID, columnID string, req *request.UpdateColumnRequest) (*response.ColumnResponse, error) {
	return m.renameColumnFn(ctx, teamID, columnID, req)
}
func (m *mockBoardManager) ReorderColumn(ctx context.Context, teamID, columnID string, req *request.ReorderColumnRequest) (*response.ColumnListResponse, error) {
	return m.reorderColFn(ctx, teamID, columnID, req)
}
func (m *mockBoardManager) DeleteColumn(ctx context.Context, teamID, columnID string) error {
	return m.deleteColumnFn(ctx, teamID, columnID)
}

// MockTaskManager
type mockTaskManager struct {
	createFn  func(ctx context.Context, teamID, memberID string, req *request.CreateTaskRequest) (*response.TaskResponse, error)
	getByIDFn func(ctx context.Context, teamID, taskID string) (*response.TaskDetailResponse, error)
	updateFn  func(ctx context.Context, teamID, taskID string, req *request.UpdateTaskRequest) (*response.TaskResponse, error)
	moveFn    func(ctx context.Context, teamID, taskID string, req *request.MoveTaskRequest) (*response.TaskResponse, error)
	deleteFn  func(ctx context.Context, teamID, taskID string) error
}

func (m *mockTaskManager) Create(ctx context.Context, teamID, memberID string, req *request.CreateTaskRequest) (*response.TaskResponse, error) {
	return m.createFn(ctx, teamID, memberID, req)
}
func (m *mockTaskManager) GetByID(ctx context.Context, teamID, taskID string) (*response.TaskDetailResponse, error) {
	return m.getByIDFn(ctx, teamID, taskID)
}
func (m *mockTaskManager) Update(ctx context.Context, teamID, taskID string, req *request.UpdateTaskRequest) (*response.TaskResponse, error) {
	return m.updateFn(ctx, teamID, taskID, req)
}
func (m *mockTaskManager) Move(ctx context.Context, teamID, taskID string, req *request.MoveTaskRequest) (*response.TaskResponse, error) {
	return m.moveFn(ctx, teamID, taskID, req)
}
func (m *mockTaskManager) Delete(ctx context.Context, teamID, taskID string) error {
	return m.deleteFn(ctx, teamID, taskID)
}

// MockCommentManager
type mockCommentManager struct {
	createFn     func(ctx context.Context, teamID, taskID, authorID string, req *request.CreateCommentRequest) (*response.CommentResponse, error)
	listByTaskFn func(ctx context.Context, teamID, taskID string, page, perPage int) (*response.CommentListResponse, error)
	deleteFn     func(ctx context.Context, teamID, commentID, memberID string, isAdmin bool) error
}

func (m *mockCommentManager) Create(ctx context.Context, teamID, taskID, authorID string, req *request.CreateCommentRequest) (*response.CommentResponse, error) {
	return m.createFn(ctx, teamID, taskID, authorID, req)
}
func (m *mockCommentManager) ListByTask(ctx context.Context, teamID, taskID string, page, perPage int) (*response.CommentListResponse, error) {
	return m.listByTaskFn(ctx, teamID, taskID, page, perPage)
}
func (m *mockCommentManager) Delete(ctx context.Context, teamID, commentID, memberID string, isAdmin bool) error {
	return m.deleteFn(ctx, teamID, commentID, memberID, isAdmin)
}

// ========== Helpers ==========

func setAuthContext(c echo.Context, memberID, teamID, role string) {
	c.Set("auth_context", &service.AuthContext{
		MemberID: memberID,
		TeamID:   teamID,
		Role:     role,
	})
}

func parseJSONBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	return result
}

func successReturn() {}
func notFoundFn() (interface{}, error) { return nil, domainErrors.ErrNotFound }
func conflictFn() (interface{}, error)  { return nil, domainErrors.ErrConflict }

// ========== TeamHandler Tests ==========

func TestTeamHandler_Register_Success(t *testing.T) {
	e := echo.New()
	mm := &mockTeamManager{
		registerFn: func(_ context.Context, _ *request.RegisterTeamRequest) (*response.RegisterTeamResponse, error) {
			return &response.RegisterTeamResponse{
				Data: response.RegisterTeamData{
					Team:   response.TeamBrief{ID: "t1", Name: "Team"},
					Member: response.MemberDetail{ID: "m1", Name: "Admin"},
					APIKey: "tb_testkey12345678901234567890123",
				},
			}, nil
		},
	}
	h := NewTeamHandler(mm)

	body := `{"name":"Team","admin_name":"Admin","admin_email":"admin@test.com"}`
	req := httptest.NewRequest(http.MethodPost, "/teams", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Register(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
	result := parseJSONBody(t, rec)
	data := result["data"].(map[string]interface{})
	team := data["team"].(map[string]interface{})
	if team["name"] != "Team" {
		t.Errorf("expected team name 'Team', got %v", team["name"])
	}
}

func TestTeamHandler_Register_ValidationError(t *testing.T) {
	e := echo.New()
	mm := &mockTeamManager{
		registerFn: func(_ context.Context, _ *request.RegisterTeamRequest) (*response.RegisterTeamResponse, error) {
			return nil, domainErrors.ValidationErrors{
				&domainErrors.ValidationError{Field: "name", Message: "is required"},
			}
		},
	}
	h := NewTeamHandler(mm)

	body := `{"name":"","admin_name":"Admin","admin_email":"admin@test.com"}`
	req := httptest.NewRequest(http.MethodPost, "/teams", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Register(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
	result := parseJSONBody(t, rec)
	errBody := result["error"].(map[string]interface{})
	if errBody["code"] != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %v", errBody["code"])
	}
}

// ========== BoardHandler Tests ==========

func TestBoardHandler_Create_Success(t *testing.T) {
	e := echo.New()
	mm := &mockBoardManager{
		createFn: func(_ context.Context, _ string, _ *request.CreateBoardRequest) (*response.BoardDetailResponse, error) {
			return &response.BoardDetailResponse{Data: response.BoardDetail{ID: "b1", Name: "My Board"}}, nil
		},
	}
	h := NewBoardHandler(mm)

	body := `{"name":"My Board"}`
	req := httptest.NewRequest(http.MethodPost, "/boards", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestBoardHandler_List_Success(t *testing.T) {
	e := echo.New()
	mm := &mockBoardManager{
		listByTeamFn: func(_ context.Context, _ string) (*response.BoardListResponse, error) {
			return &response.BoardListResponse{Data: []response.BoardItem{}}, nil
		},
	}
	h := NewBoardHandler(mm)

	req := httptest.NewRequest(http.MethodGet, "/boards", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.List(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestBoardHandler_Get_Success(t *testing.T) {
	e := echo.New()
	mm := &mockBoardManager{
		getByIDFn: func(_ context.Context, _, _ string) (*response.BoardDetailResponse, error) {
			return &response.BoardDetailResponse{Data: response.BoardDetail{ID: "b1", Name: "Board 1"}}, nil
		},
	}
	h := NewBoardHandler(mm)

	req := httptest.NewRequest(http.MethodGet, "/boards/b1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("b1")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.Get(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestBoardHandler_Get_NotFound(t *testing.T) {
	e := echo.New()
	mm := &mockBoardManager{
		getByIDFn: func(_ context.Context, _, _ string) (*response.BoardDetailResponse, error) {
			return nil, domainErrors.ErrNotFound
		},
	}
	h := NewBoardHandler(mm)

	req := httptest.NewRequest(http.MethodGet, "/boards/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.Get(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestBoardHandler_Update_Success(t *testing.T) {
	e := echo.New()
	mm := &mockBoardManager{
		updateFn: func(_ context.Context, _, _ string, _ *request.UpdateBoardRequest) (*response.BoardResponse, error) {
			return &response.BoardResponse{Data: response.BoardItem{ID: "b1", Name: "Updated"}}, nil
		},
	}
	h := NewBoardHandler(mm)

	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/boards/b1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("b1")
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.Update(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestBoardHandler_Delete_AdminAllowed(t *testing.T) {
	e := echo.New()
	mm := &mockBoardManager{
		deleteFn: func(_ context.Context, _, _ string) error { return nil },
	}
	h := NewBoardHandler(mm)

	req := httptest.NewRequest(http.MethodDelete, "/boards/b1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("b1")
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.Delete(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}

func TestBoardHandler_Delete_NonAdminForbidden(t *testing.T) {
	e := echo.New()
	mm := &mockBoardManager{}
	h := NewBoardHandler(mm)

	req := httptest.NewRequest(http.MethodDelete, "/boards/b1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("b1")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.Delete(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
	result := parseJSONBody(t, rec)
	errBody := result["error"].(map[string]interface{})
	if errBody["code"] != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN, got %v", errBody["code"])
	}
}

func TestBoardHandler_AddColumn_Success(t *testing.T) {
	e := echo.New()
	mm := &mockBoardManager{
		addColumnFn: func(_ context.Context, _, _ string, _ *request.CreateColumnRequest) (*response.ColumnResponse, error) {
			return &response.ColumnResponse{Data: response.ColumnItem{ID: "col1", Name: "New Col", BoardID: "b1"}}, nil
		},
	}
	h := NewBoardHandler(mm)

	body := `{"name":"New Col","status":"todo"}`
	req := httptest.NewRequest(http.MethodPost, "/boards/b1/columns", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("boardId")
	c.SetParamValues("b1")
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.AddColumn(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestBoardHandler_RenameColumn_Success(t *testing.T) {
	e := echo.New()
	mm := &mockBoardManager{
		renameColumnFn: func(_ context.Context, _, _ string, _ *request.UpdateColumnRequest) (*response.ColumnResponse, error) {
			return &response.ColumnResponse{Data: response.ColumnItem{ID: "col1", Name: "Renamed"}}, nil
		},
	}
	h := NewBoardHandler(mm)

	body := `{"name":"Renamed"}`
	req := httptest.NewRequest(http.MethodPut, "/columns/col1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("col1")
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.RenameColumn(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestBoardHandler_DeleteColumn_Success(t *testing.T) {
	e := echo.New()
	mm := &mockBoardManager{
		deleteColumnFn: func(_ context.Context, _, _ string) error { return nil },
	}
	h := NewBoardHandler(mm)

	req := httptest.NewRequest(http.MethodDelete, "/columns/col1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("col1")
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.DeleteColumn(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}

// ========== MemberHandler Tests ==========

func TestMemberHandler_Me_Success(t *testing.T) {
	e := echo.New()
	mm := &mockMemberManager{
		getCurrentMemberFn: func(_ context.Context, _ string) (*response.MemberResponse, error) {
			return &response.MemberResponse{Data: response.MemberDetail{ID: "mem1", Name: "Alice", Role: "admin"}}, nil
		},
	}
	h := NewMemberHandler(mm)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.Me(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestMemberHandler_List_Success(t *testing.T) {
	e := echo.New()
	mm := &mockMemberManager{
		listByTeamFn: func(_ context.Context, _ string) (*response.MemberListResponse, error) {
			return &response.MemberListResponse{Data: []response.MemberDetail{}}, nil
		},
	}
	h := NewMemberHandler(mm)

	req := httptest.NewRequest(http.MethodGet, "/members", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.List(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestMemberHandler_Create_AdminAllowed(t *testing.T) {
	e := echo.New()
	mm := &mockMemberManager{
		createFn: func(_ context.Context, _ string, _ *request.CreateMemberRequest) (*response.MemberWithKeyResponse, error) {
			return &response.MemberWithKeyResponse{Data: response.MemberWithKey{ID: "mem2", Name: "Bob", APIKey: "tb_xxx"}}, nil
		},
	}
	h := NewMemberHandler(mm)

	body := `{"name":"Bob","email":"bob@test.com","role":"member"}`
	req := httptest.NewRequest(http.MethodPost, "/members", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestMemberHandler_Create_NonAdminForbidden(t *testing.T) {
	e := echo.New()
	mm := &mockMemberManager{}
	h := NewMemberHandler(mm)

	body := `{"name":"Bob","email":"bob@test.com","role":"member"}`
	req := httptest.NewRequest(http.MethodPost, "/members", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestMemberHandler_Update_AdminAllowed(t *testing.T) {
	e := echo.New()
	mm := &mockMemberManager{
		updateFn: func(_ context.Context, _, _ string, _ *request.UpdateMemberRequest) (*response.MemberResponse, error) {
			return &response.MemberResponse{Data: response.MemberDetail{ID: "mem2", Name: "Updated"}}, nil
		},
	}
	h := NewMemberHandler(mm)

	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/members/mem2", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("mem2")
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.Update(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestMemberHandler_Update_NonAdminForbidden(t *testing.T) {
	e := echo.New()
	mm := &mockMemberManager{}
	h := NewMemberHandler(mm)

	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/members/mem2", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("mem2")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.Update(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestMemberHandler_Delete_AdminAllowed(t *testing.T) {
	e := echo.New()
	mm := &mockMemberManager{
		deleteFn: func(_ context.Context, _, _ string) error { return nil },
	}
	h := NewMemberHandler(mm)

	req := httptest.NewRequest(http.MethodDelete, "/members/mem2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("mem2")
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.Delete(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}

func TestMemberHandler_Delete_NonAdminForbidden(t *testing.T) {
	e := echo.New()
	mm := &mockMemberManager{}
	h := NewMemberHandler(mm)

	req := httptest.NewRequest(http.MethodDelete, "/members/mem2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("mem2")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.Delete(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestMemberHandler_RegenerateKey_AdminAllowed(t *testing.T) {
	e := echo.New()
	mm := &mockMemberManager{
		regenKeyFn: func(_ context.Context, _, _ string) (*response.APIKeyResponse, error) {
			return &response.APIKeyResponse{Data: response.APIKeyData{APIKey: "tb_newkey"}}, nil
		},
	}
	h := NewMemberHandler(mm)

	req := httptest.NewRequest(http.MethodPost, "/members/mem2/regenerate-key", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("mem2")
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.RegenerateKey(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestMemberHandler_RegenerateKey_NonAdminForbidden(t *testing.T) {
	e := echo.New()
	mm := &mockMemberManager{}
	h := NewMemberHandler(mm)

	req := httptest.NewRequest(http.MethodPost, "/members/mem2/regenerate-key", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("mem2")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.RegenerateKey(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

// ========== TaskHandler Tests ==========

func TestTaskHandler_Create_Success(t *testing.T) {
	e := echo.New()
	tm := &mockTaskManager{
		createFn: func(_ context.Context, _, _ string, _ *request.CreateTaskRequest) (*response.TaskResponse, error) {
			return &response.TaskResponse{Data: response.TaskItem{ID: "task1", Title: "My Task"}}, nil
		},
	}
	cm := &mockCommentManager{}
	h := NewTaskHandler(tm, cm)

	body := `{"title":"My Task","priority":"high"}`
	req := httptest.NewRequest(http.MethodPost, "/columns/col1/tasks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("columnId")
	c.SetParamValues("col1")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestTaskHandler_Create_InvalidJSON(t *testing.T) {
	e := echo.New()
	tm := &mockTaskManager{}
	cm := &mockCommentManager{}
	h := NewTaskHandler(tm, cm)

	req := httptest.NewRequest(http.MethodPost, "/columns/col1/tasks", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("columnId")
	c.SetParamValues("col1")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
	result := parseJSONBody(t, rec)
	errBody := result["error"].(map[string]interface{})
	if errBody["code"] != "BAD_REQUEST" {
		t.Errorf("expected BAD_REQUEST, got %v", errBody["code"])
	}
}

func TestTaskHandler_Get_Success(t *testing.T) {
	e := echo.New()
	tm := &mockTaskManager{
		getByIDFn: func(_ context.Context, _, _ string) (*response.TaskDetailResponse, error) {
			return &response.TaskDetailResponse{Data: response.TaskDetailItem{ID: "task1", Title: "Task"}}, nil
		},
	}
	cm := &mockCommentManager{}
	h := NewTaskHandler(tm, cm)

	req := httptest.NewRequest(http.MethodGet, "/tasks/task1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("task1")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.Get(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTaskHandler_Update_Success(t *testing.T) {
	e := echo.New()
	tm := &mockTaskManager{
		updateFn: func(_ context.Context, _, _ string, _ *request.UpdateTaskRequest) (*response.TaskResponse, error) {
			return &response.TaskResponse{Data: response.TaskItem{ID: "task1", Title: "Updated"}}, nil
		},
	}
	cm := &mockCommentManager{}
	h := NewTaskHandler(tm, cm)

	body := `{"title":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/tasks/task1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("task1")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.Update(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTaskHandler_Move_Success(t *testing.T) {
	e := echo.New()
	tm := &mockTaskManager{
		moveFn: func(_ context.Context, _, _ string, _ *request.MoveTaskRequest) (*response.TaskResponse, error) {
			return &response.TaskResponse{Data: response.TaskItem{ID: "task1", ColumnID: "col2"}}, nil
		},
	}
	cm := &mockCommentManager{}
	h := NewTaskHandler(tm, cm)

	body := `{"column_id":"col2"}`
	req := httptest.NewRequest(http.MethodPut, "/tasks/task1/move", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("task1")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.Move(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTaskHandler_Delete_Success(t *testing.T) {
	e := echo.New()
	tm := &mockTaskManager{
		deleteFn: func(_ context.Context, _, _ string) error { return nil },
	}
	cm := &mockCommentManager{}
	h := NewTaskHandler(tm, cm)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/task1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("task1")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.Delete(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}

func TestTaskHandler_CreateComment_Success(t *testing.T) {
	e := echo.New()
	tm := &mockTaskManager{}
	cm := &mockCommentManager{
		createFn: func(_ context.Context, _, _, _ string, _ *request.CreateCommentRequest) (*response.CommentResponse, error) {
			return &response.CommentResponse{Data: response.CommentItem{ID: "c1", Body: "Nice!"}}, nil
		},
	}
	h := NewTaskHandler(tm, cm)

	body := `{"body":"Nice!"}`
	req := httptest.NewRequest(http.MethodPost, "/tasks/task1/comments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("taskId")
	c.SetParamValues("task1")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.CreateComment(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestTaskHandler_ListComments_Success(t *testing.T) {
	e := echo.New()
	tm := &mockTaskManager{}
	cm := &mockCommentManager{
		listByTaskFn: func(_ context.Context, _, _ string, _, _ int) (*response.CommentListResponse, error) {
			return &response.CommentListResponse{
				Data: []response.CommentItem{},
				Meta: response.PaginationMeta{Total: 0, Page: 1, PerPage: 20, TotalPages: 0},
			}, nil
		},
	}
	h := NewTaskHandler(tm, cm)

	req := httptest.NewRequest(http.MethodGet, "/tasks/task1/comments", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("taskId")
	c.SetParamValues("task1")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.ListComments(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTaskHandler_DeleteComment_AsMember(t *testing.T) {
	e := echo.New()
	tm := &mockTaskManager{}
	cm := &mockCommentManager{
		deleteFn: func(_ context.Context, _, _, _ string, isAdmin bool) error {
			if isAdmin {
				t.Error("expected isAdmin=false for member role")
			}
			return nil
		},
	}
	h := NewTaskHandler(tm, cm)

	req := httptest.NewRequest(http.MethodDelete, "/comments/c1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("c1")
	setAuthContext(c, "mem1", "team1", "member")

	if err := h.DeleteComment(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}

func TestTaskHandler_DeleteComment_AsAdmin(t *testing.T) {
	e := echo.New()
	tm := &mockTaskManager{}
	cm := &mockCommentManager{
		deleteFn: func(_ context.Context, _, _, _ string, isAdmin bool) error {
			if !isAdmin {
				t.Error("expected isAdmin=true for admin role")
			}
			return nil
		},
	}
	h := NewTaskHandler(tm, cm)

	req := httptest.NewRequest(http.MethodDelete, "/comments/c1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("c1")
	setAuthContext(c, "admin1", "team1", "admin")

	if err := h.DeleteComment(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}

// ========== Error Mapping Tests ==========

func TestBoardHandler_Create_Conflict(t *testing.T) {
	e := echo.New()
	mm := &mockBoardManager{
		createFn: func(_ context.Context, _ string, _ *request.CreateBoardRequest) (*response.BoardDetailResponse, error) {
			return nil, domainErrors.ErrConflict
		},
	}
	h := NewBoardHandler(mm)

	body := `{"name":"Duplicate"}`
	req := httptest.NewRequest(http.MethodPost, "/boards", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestBoardHandler_Create_InternalError(t *testing.T) {
	e := echo.New()
	mm := &mockBoardManager{
		createFn: func(_ context.Context, _ string, _ *request.CreateBoardRequest) (*response.BoardDetailResponse, error) {
			return nil, errors.New("unexpected db error")
		},
	}
	h := NewBoardHandler(mm)

	body := `{"name":"Board"}`
	req := httptest.NewRequest(http.MethodPost, "/boards", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
	result := parseJSONBody(t, rec)
	errBody := result["error"].(map[string]interface{})
	if errBody["code"] != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR, got %v", errBody["code"])
	}
}

func TestBoardHandler_Create_Unauthorized(t *testing.T) {
	e := echo.New()
	mm := &mockBoardManager{
		createFn: func(_ context.Context, _ string, _ *request.CreateBoardRequest) (*response.BoardDetailResponse, error) {
			return nil, domainErrors.ErrUnauthorized
		},
	}
	h := NewBoardHandler(mm)

	body := `{"name":"Board"}`
	req := httptest.NewRequest(http.MethodPost, "/boards", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestBoardHandler_Create_Forbidden(t *testing.T) {
	e := echo.New()
	mm := &mockBoardManager{
		createFn: func(_ context.Context, _ string, _ *request.CreateBoardRequest) (*response.BoardDetailResponse, error) {
			return nil, domainErrors.ErrForbidden
		},
	}
	h := NewBoardHandler(mm)

	body := `{"name":"Board"}`
	req := httptest.NewRequest(http.MethodPost, "/boards", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setAuthContext(c, "mem1", "team1", "admin")

	if err := h.Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

// Ensure the unused variable doesn't cause issues
var _ = successReturn
var _ = notFoundFn
var _ = conflictFn
