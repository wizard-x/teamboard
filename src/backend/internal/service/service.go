package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"

	domainErrors "teamboard/internal/domain/errors"
	"teamboard/internal/domain/entity"
	"teamboard/internal/dto/request"
	"teamboard/internal/dto/response"
	"teamboard/internal/repository"
	"teamboard/internal/repository/cache"
)

var (
	defaultColumns = []struct {
		Name     string
		Status   string
	}{
		{"Todo", "todo"},
		{"In Progress", "in_progress"},
		{"Review", "review"},
		{"Done", "done"},
	}
)

func newULID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
}

// GenerateAPIKey creates a new API key with format tb_<random base62> (total 40 chars).
func GenerateAPIKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const prefix = "tb_"
	keyLen := 40 - len(prefix)
	b := make([]byte, keyLen)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return prefix + string(b)
}

// HashAPIKey returns SHA-256 hex of the API key.
func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// --- Auth Service ---

type AuthService struct {
	memberRepo repository.MemberRepository
	cache      *cache.Cache
}

func NewAuthService(memberRepo repository.MemberRepository, cache *cache.Cache) *AuthService {
	return &AuthService{memberRepo: memberRepo, cache: cache}
}

type AuthContext struct {
	MemberID string
	TeamID   string
	Role     string
}

func (s *AuthService) Authenticate(ctx context.Context, apiKey string) (*AuthContext, error) {
	if apiKey == "" {
		return nil, domainErrors.ErrUnauthorized
	}

	hash := HashAPIKey(apiKey)
	prefix := apiKey[:10]

	// Check cache first
	cached, err := s.cache.GetAuthContext(ctx, prefix, hash)
	if err != nil {
		// Log but continue to DB
		_ = err
	}
	if cached != nil {
		return &AuthContext{
			MemberID: cached.MemberID,
			TeamID:   cached.TeamID,
			Role:     cached.Role,
		}, nil
	}

	// DB lookup
	member, err := s.memberRepo.GetByAPIKeyHash(ctx, hash)
	if err != nil {
		return nil, domainErrors.ErrUnauthorized
	}

	authCtx := &cache.AuthContext{
		MemberID: member.ID,
		TeamID:   member.TeamID,
		Role:     member.Role,
	}

	// Cache it
	_ = s.cache.SetAuthContext(ctx, prefix, hash, authCtx)

	return &AuthContext{
		MemberID: member.ID,
		TeamID:   member.TeamID,
		Role:     member.Role,
	}, nil
}

// --- Team Service ---

type TeamService struct {
	teamRepo   repository.TeamRepository
	memberRepo repository.MemberRepository
}

func NewTeamService(teamRepo repository.TeamRepository, memberRepo repository.MemberRepository) *TeamService {
	return &TeamService{teamRepo: teamRepo, memberRepo: memberRepo}
}

func (s *TeamService) Register(ctx context.Context, req *request.RegisterTeamRequest) (*response.RegisterTeamResponse, error) {
	// Validate
	var errs domainErrors.ValidationErrors
	if len(strings.TrimSpace(req.Name)) < 1 || len(req.Name) > 100 {
		errs = append(errs, &domainErrors.ValidationError{Field: "name", Message: "must be 1-100 characters"})
	}
	if len(strings.TrimSpace(req.AdminName)) < 1 || len(req.AdminName) > 100 {
		errs = append(errs, &domainErrors.ValidationError{Field: "admin_name", Message: "must be 1-100 characters"})
	}
	if !isValidEmail(req.AdminEmail) {
		errs = append(errs, &domainErrors.ValidationError{Field: "admin_email", Message: "must be a valid email"})
	}
	if len(errs) > 0 {
		return nil, errs
	}

	// Check unique team name
	_, err := s.teamRepo.GetByName(ctx, req.Name)
	if err == nil {
		return nil, domainErrors.ErrConflict
	}

	now := time.Now()

	// Create team
	team := &entity.Team{
		ID:        newULID(),
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, fmt.Errorf("creating team: %w", err)
	}

	// Create admin member
	apiKey := GenerateAPIKey()
	apiKeyHash := HashAPIKey(apiKey)

	member := &entity.Member{
		ID:           newULID(),
		TeamID:       team.ID,
		Name:         req.AdminName,
		Email:        req.AdminEmail,
		Role:         "admin",
		APIKeyHash:   apiKeyHash,
		APIKeyPrefix: apiKey[:10],
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("creating admin member: %w", err)
	}

	return &response.RegisterTeamResponse{
		Data: response.RegisterTeamData{
			Team: response.TeamBrief{
				ID:        team.ID,
				Name:      team.Name,
				CreatedAt: team.CreatedAt,
			},
			Member: response.MemberDetail{
				ID:        member.ID,
				Name:      member.Name,
				Email:     member.Email,
				Role:      member.Role,
				CreatedAt: member.CreatedAt,
			},
			APIKey: apiKey,
		},
	}, nil
}

// --- Member Service ---

type MemberService struct {
	memberRepo repository.MemberRepository
	cache      *cache.Cache
}

func NewMemberService(memberRepo repository.MemberRepository, cache *cache.Cache) *MemberService {
	return &MemberService{memberRepo: memberRepo, cache: cache}
}

func (s *MemberService) GetCurrentMember(ctx context.Context, memberID string) (*response.MemberResponse, error) {
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}
	return &response.MemberResponse{
		Data: response.MemberDetail{
			ID:        member.ID,
			Name:      member.Name,
			Email:     member.Email,
			Role:      member.Role,
			CreatedAt: member.CreatedAt,
		},
	}, nil
}

func (s *MemberService) ListByTeam(ctx context.Context, teamID string) (*response.MemberListResponse, error) {
	members, err := s.memberRepo.ListByTeam(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}

	items := make([]response.MemberDetail, 0, len(members))
	for _, m := range members {
		items = append(items, response.MemberDetail{
			ID:        m.ID,
			Name:      m.Name,
			Email:     m.Email,
			Role:      m.Role,
			CreatedAt: m.CreatedAt,
		})
	}
	return &response.MemberListResponse{Data: items}, nil
}

func (s *MemberService) Create(ctx context.Context, teamID string, req *request.CreateMemberRequest) (*response.MemberWithKeyResponse, error) {
	var errs domainErrors.ValidationErrors
	if len(strings.TrimSpace(req.Name)) < 1 || len(req.Name) > 100 {
		errs = append(errs, &domainErrors.ValidationError{Field: "name", Message: "must be 1-100 characters"})
	}
	if !isValidEmail(req.Email) {
		errs = append(errs, &domainErrors.ValidationError{Field: "email", Message: "must be a valid email"})
	}
	if req.Role != "admin" && req.Role != "member" {
		errs = append(errs, &domainErrors.ValidationError{Field: "role", Message: "must be admin or member"})
	}
	if len(errs) > 0 {
		return nil, errs
	}

	now := time.Now()
	apiKey := GenerateAPIKey()
	apiKeyHash := HashAPIKey(apiKey)

	member := &entity.Member{
		ID:           newULID(),
		TeamID:       teamID,
		Name:         req.Name,
		Email:        req.Email,
		Role:         req.Role,
		APIKeyHash:   apiKeyHash,
		APIKeyPrefix: apiKey[:10],
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, domainErrors.ErrConflict
	}

	return &response.MemberWithKeyResponse{
		Data: response.MemberWithKey{
			ID:        member.ID,
			Name:      member.Name,
			Email:     member.Email,
			Role:      member.Role,
			APIKey:    apiKey,
			CreatedAt: member.CreatedAt,
		},
	}, nil
}

func (s *MemberService) Update(ctx context.Context, teamID, memberID string, req *request.UpdateMemberRequest) (*response.MemberResponse, error) {
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}
	if member.TeamID != teamID {
		return nil, domainErrors.ErrNotFound
	}

	if req.Name != "" {
		member.Name = req.Name
	}
	if req.Role != "" {
		if req.Role != "admin" && req.Role != "member" {
			return nil, domainErrors.ErrBadRequest
		}
		// Check if demoting last admin
		if member.Role == "admin" && req.Role == "member" {
			admins, _ := s.memberRepo.CountAdmins(ctx, teamID)
			if admins <= 1 {
				return nil, domainErrors.ErrBadRequest
			}
		}
		member.Role = req.Role
	}

	member.UpdatedAt = time.Now()
	if err := s.memberRepo.Update(ctx, member); err != nil {
		return nil, fmt.Errorf("updating member: %w", err)
	}

	// Invalidate cache
	_ = s.cache.InvalidateAuthContext(ctx, member.APIKeyPrefix, member.APIKeyHash)

	return &response.MemberResponse{
		Data: response.MemberDetail{
			ID:        member.ID,
			Name:      member.Name,
			Email:     member.Email,
			Role:      member.Role,
			CreatedAt: member.CreatedAt,
		},
	}, nil
}

func (s *MemberService) Delete(ctx context.Context, teamID, memberID string) error {
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return domainErrors.ErrNotFound
	}
	if member.TeamID != teamID {
		return domainErrors.ErrNotFound
	}

	// Cannot delete last admin
	if member.Role == "admin" {
		admins, _ := s.memberRepo.CountAdmins(ctx, teamID)
		if admins <= 1 {
			return domainErrors.ErrBadRequest
		}
	}

	if err := s.memberRepo.Delete(ctx, memberID); err != nil {
		return fmt.Errorf("deleting member: %w", err)
	}

	// Invalidate auth cache
	_ = s.cache.InvalidateAuthContext(ctx, member.APIKeyPrefix, member.APIKeyHash)

	return nil
}

func (s *MemberService) UpdateMe(ctx context.Context, memberID string, req *request.UpdateMeRequest) (*response.MemberResponse, error) {
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}

	if len(strings.TrimSpace(req.Name)) < 1 || len(req.Name) > 100 {
		return nil, domainErrors.ValidationErrors{
			&domainErrors.ValidationError{Field: "name", Message: "must be 1-100 characters"},
		}
	}

	member.Name = req.Name
	member.UpdatedAt = time.Now()
	if err := s.memberRepo.Update(ctx, member); err != nil {
		return nil, fmt.Errorf("updating member profile: %w", err)
	}

	// Invalidate cache
	_ = s.cache.InvalidateAuthContext(ctx, member.APIKeyPrefix, member.APIKeyHash)

	return &response.MemberResponse{
		Data: response.MemberDetail{
			ID:        member.ID,
			Name:      member.Name,
			Email:     member.Email,
			Role:      member.Role,
			CreatedAt: member.CreatedAt,
		},
	}, nil
}

func (s *MemberService) RegenerateAPIKey(ctx context.Context, teamID, memberID string) (*response.APIKeyResponse, error) {
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}
	if member.TeamID != teamID {
		return nil, domainErrors.ErrNotFound
	}

	// Invalidate old cache
	_ = s.cache.InvalidateAuthContext(ctx, member.APIKeyPrefix, member.APIKeyHash)

	newKey := GenerateAPIKey()
	member.APIKeyHash = HashAPIKey(newKey)
	member.APIKeyPrefix = newKey[:10]
	member.UpdatedAt = time.Now()

	if err := s.memberRepo.Update(ctx, member); err != nil {
		return nil, fmt.Errorf("updating member key: %w", err)
	}

	return &response.APIKeyResponse{
		Data: response.APIKeyData{APIKey: newKey},
	}, nil
}

// --- Board Service ---

type BoardService struct {
	boardRepo  repository.BoardRepository
	columnRepo repository.ColumnRepository
	taskRepo   repository.TaskRepository
	memberRepo repository.MemberRepository
	commentRepo repository.CommentRepository
	cache      *cache.Cache
}

func NewBoardService(
	boardRepo repository.BoardRepository,
	columnRepo repository.ColumnRepository,
	taskRepo repository.TaskRepository,
	memberRepo repository.MemberRepository,
	commentRepo repository.CommentRepository,
	cache *cache.Cache,
) *BoardService {
	return &BoardService{
		boardRepo:   boardRepo,
		columnRepo:  columnRepo,
		taskRepo:    taskRepo,
		memberRepo:  memberRepo,
		commentRepo: commentRepo,
		cache:       cache,
	}
}

func (s *BoardService) Create(ctx context.Context, teamID string, req *request.CreateBoardRequest) (*response.BoardDetailResponse, error) {
	if len(strings.TrimSpace(req.Name)) < 1 || len(req.Name) > 100 {
		return nil, domainErrors.ValidationErrors{
			&domainErrors.ValidationError{Field: "name", Message: "must be 1-100 characters"},
		}
	}
	if req.Description != nil && len(*req.Description) > 500 {
		return nil, domainErrors.ValidationErrors{
			&domainErrors.ValidationError{Field: "description", Message: "must be at most 500 characters"},
		}
	}

	now := time.Now()
	board := &entity.Board{
		ID:          newULID(),
		TeamID:      teamID,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.boardRepo.Create(ctx, board); err != nil {
		return nil, fmt.Errorf("creating board: %w", err)
	}

	// Create default columns
	var columns []response.ColumnDetail
	for i, dc := range defaultColumns {
		col := &entity.Column{
			ID:        newULID(),
			BoardID:   board.ID,
			Name:      dc.Name,
			Position:  i,
			Status:    dc.Status,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.columnRepo.Create(ctx, col); err != nil {
			return nil, fmt.Errorf("creating default column: %w", err)
		}
		columns = append(columns, response.ColumnDetail{
			ID:       col.ID,
			Name:     col.Name,
			Position: col.Position,
			Status:   col.Status,
			Tasks:    []response.TaskItem{},
		})
	}

	_ = s.cache.InvalidateBoardList(ctx, teamID)

	return &response.BoardDetailResponse{
		Data: response.BoardDetail{
			ID:          board.ID,
			Name:        board.Name,
			Description: board.Description,
			Columns:     columns,
			CreatedAt:   board.CreatedAt,
			UpdatedAt:   board.UpdatedAt,
		},
	}, nil
}

func (s *BoardService) ListByTeam(ctx context.Context, teamID string) (*response.BoardListResponse, error) {
	boards, err := s.boardRepo.ListByTeam(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("listing boards: %w", err)
	}

	items := make([]response.BoardItem, 0, len(boards))
	for _, b := range boards {
		items = append(items, response.BoardItem{
			ID:          b.ID,
			Name:        b.Name,
			Description: b.Description,
			CreatedAt:   b.CreatedAt,
			UpdatedAt:   b.UpdatedAt,
		})
	}
	return &response.BoardListResponse{Data: items}, nil
}

func (s *BoardService) GetByID(ctx context.Context, teamID, boardID string) (*response.BoardDetailResponse, error) {
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}
	if board.TeamID != teamID {
		return nil, domainErrors.ErrNotFound
	}

	columns, err := s.columnRepo.ListByBoard(ctx, boardID)
	if err != nil {
		return nil, fmt.Errorf("listing columns: %w", err)
	}

	tasks, err := s.taskRepo.ListByBoard(ctx, boardID)
	if err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}

	// Group tasks by column
	tasksByColumn := make(map[string][]response.TaskItem)
	for _, t := range tasks {
		item := taskToResponse(t)
		tasksByColumn[t.ColumnID] = append(tasksByColumn[t.ColumnID], item)
	}

	colDetails := make([]response.ColumnDetail, 0, len(columns))
	for _, c := range columns {
		colTasks := tasksByColumn[c.ID]
		if colTasks == nil {
			colTasks = []response.TaskItem{}
		}
		colDetails = append(colDetails, response.ColumnDetail{
			ID:       c.ID,
			Name:     c.Name,
			Position: c.Position,
			Status:   c.Status,
			Tasks:    colTasks,
		})
	}

	return &response.BoardDetailResponse{
		Data: response.BoardDetail{
			ID:          board.ID,
			Name:        board.Name,
			Description: board.Description,
			Columns:     colDetails,
			CreatedAt:   board.CreatedAt,
			UpdatedAt:   board.UpdatedAt,
		},
	}, nil
}

func (s *BoardService) Update(ctx context.Context, teamID, boardID string, req *request.UpdateBoardRequest) (*response.BoardResponse, error) {
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}
	if board.TeamID != teamID {
		return nil, domainErrors.ErrNotFound
	}

	if req.Name != nil {
		if len(strings.TrimSpace(*req.Name)) < 1 || len(*req.Name) > 100 {
			return nil, domainErrors.ValidationErrors{
				&domainErrors.ValidationError{Field: "name", Message: "must be 1-100 characters"},
			}
		}
		board.Name = *req.Name
	}
	if req.Description != nil {
		board.Description = req.Description
	}
	board.UpdatedAt = time.Now()

	if err := s.boardRepo.Update(ctx, board); err != nil {
		return nil, fmt.Errorf("updating board: %w", err)
	}

	_ = s.cache.InvalidateBoard(ctx, boardID)
	_ = s.cache.InvalidateBoardList(ctx, teamID)

	return &response.BoardResponse{
		Data: response.BoardItem{
			ID:          board.ID,
			Name:        board.Name,
			Description: board.Description,
			CreatedAt:   board.CreatedAt,
			UpdatedAt:   board.UpdatedAt,
		},
	}, nil
}

func (s *BoardService) Delete(ctx context.Context, teamID, boardID string) error {
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return domainErrors.ErrNotFound
	}
	if board.TeamID != teamID {
		return domainErrors.ErrNotFound
	}

	if err := s.boardRepo.Delete(ctx, boardID); err != nil {
		return fmt.Errorf("deleting board: %w", err)
	}

	_ = s.cache.InvalidateBoard(ctx, boardID)
	_ = s.cache.InvalidateBoardList(ctx, teamID)

	return nil
}

// --- Column operations (part of board context) ---

func (s *BoardService) AddColumn(ctx context.Context, teamID, boardID string, req *request.CreateColumnRequest) (*response.ColumnResponse, error) {
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}
	if board.TeamID != teamID {
		return nil, domainErrors.ErrNotFound
	}

	if len(strings.TrimSpace(req.Name)) < 1 || len(req.Name) > 50 {
		return nil, domainErrors.ValidationErrors{
			&domainErrors.ValidationError{Field: "name", Message: "must be 1-50 characters"},
		}
	}

	status := req.Status
	if status == "" {
		status = "todo"
	}
	if !isValidStatus(status) {
		return nil, domainErrors.ValidationErrors{
			&domainErrors.ValidationError{Field: "status", Message: "must be one of: todo, in_progress, review, done"},
		}
	}

	position := 0
	if req.Position != nil && *req.Position >= 0 {
		position = *req.Position
		// Shift existing columns at this position or later by +1
		if err := s.columnRepo.ShiftPositions(ctx, boardID, position); err != nil {
			return nil, fmt.Errorf("shifting column positions: %w", err)
		}
	} else {
		maxPos, _ := s.columnRepo.MaxPosition(ctx, boardID)
		position = maxPos + 1
	}

	now := time.Now()
	column := &entity.Column{
		ID:        newULID(),
		BoardID:   boardID,
		Name:      req.Name,
		Position:  position,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.columnRepo.Create(ctx, column); err != nil {
		return nil, fmt.Errorf("creating column: %w", err)
	}

	_ = s.cache.InvalidateBoard(ctx, boardID)

	return &response.ColumnResponse{
		Data: response.ColumnItem{
			ID:       column.ID,
			Name:     column.Name,
			Position: column.Position,
			Status:   column.Status,
			BoardID:  column.BoardID,
		},
	}, nil
}

func (s *BoardService) RenameColumn(ctx context.Context, teamID, columnID string, req *request.UpdateColumnRequest) (*response.ColumnResponse, error) {
	column, err := s.columnRepo.GetByID(ctx, columnID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}

	// Verify team ownership via board
	board, err := s.boardRepo.GetByID(ctx, column.BoardID)
	if err != nil || board.TeamID != teamID {
		return nil, domainErrors.ErrNotFound
	}

	column.Name = req.Name
	column.UpdatedAt = time.Now()
	if err := s.columnRepo.Update(ctx, column); err != nil {
		return nil, fmt.Errorf("renaming column: %w", err)
	}

	_ = s.cache.InvalidateBoard(ctx, column.BoardID)

	return &response.ColumnResponse{
		Data: response.ColumnItem{
			ID:       column.ID,
			Name:     column.Name,
			Position: column.Position,
			Status:   column.Status,
			BoardID:  column.BoardID,
		},
	}, nil
}

func (s *BoardService) ReorderColumn(ctx context.Context, teamID, columnID string, req *request.ReorderColumnRequest) (*response.ColumnListResponse, error) {
	column, err := s.columnRepo.GetByID(ctx, columnID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}

	board, err := s.boardRepo.GetByID(ctx, column.BoardID)
	if err != nil || board.TeamID != teamID {
		return nil, domainErrors.ErrNotFound
	}

	if req.Position < 0 {
		return nil, domainErrors.ErrBadRequest
	}

	// Get all columns, reorder
	columns, err := s.columnRepo.ListByBoard(ctx, column.BoardID)
	if err != nil {
		return nil, fmt.Errorf("listing columns: %w", err)
	}

	// Remove the column from its current position, insert at new position
	newPos := req.Position

	var reordered []*entity.Column
	now := time.Now()
	for _, c := range columns {
		if c.ID == columnID {
			continue
		}
		reordered = append(reordered, c)
	}

	// Insert at new position
	if newPos >= len(reordered) {
		reordered = append(reordered, column)
	} else {
		reordered = append(reordered[:newPos], append([]*entity.Column{column}, reordered[newPos:]...)...)
	}

	// Update positions
	for i, c := range reordered {
		c.Position = i
		c.UpdatedAt = now
	}
	if err := s.columnRepo.UpdatePositions(ctx, reordered); err != nil {
		return nil, fmt.Errorf("updating column positions: %w", err)
	}

	_ = s.cache.InvalidateBoard(ctx, column.BoardID)

	items := make([]response.ColumnItem, 0, len(reordered))
	for _, c := range reordered {
		items = append(items, response.ColumnItem{
			ID:       c.ID,
			Name:     c.Name,
			Position: c.Position,
			Status:   c.Status,
			BoardID:  c.BoardID,
		})
	}
	return &response.ColumnListResponse{Data: items}, nil
}

func (s *BoardService) DeleteColumn(ctx context.Context, teamID, columnID string) error {
	column, err := s.columnRepo.GetByID(ctx, columnID)
	if err != nil {
		return domainErrors.ErrNotFound
	}

	board, err := s.boardRepo.GetByID(ctx, column.BoardID)
	if err != nil || board.TeamID != teamID {
		return domainErrors.ErrNotFound
	}

	// Check if last column
	colCount, _ := s.columnRepo.CountByBoard(ctx, column.BoardID)
	if colCount <= 1 {
		return domainErrors.ErrBadRequest
	}

	// Check if has tasks
	taskCount, _ := s.columnRepo.CountTasks(ctx, columnID)
	if taskCount > 0 {
		return domainErrors.ErrBadRequest
	}

	if err := s.columnRepo.Delete(ctx, columnID); err != nil {
		return fmt.Errorf("deleting column: %w", err)
	}

	// Re-sequence remaining columns
	remainingCols, err := s.columnRepo.ListByBoard(ctx, column.BoardID)
	if err == nil {
		now := time.Now()
		for i, c := range remainingCols {
			c.Position = i
			c.UpdatedAt = now
		}
		_ = s.columnRepo.UpdatePositions(ctx, remainingCols)
	}

	_ = s.cache.InvalidateBoard(ctx, column.BoardID)
	return nil
}

// --- Task Service ---

type TaskService struct {
	taskRepo    repository.TaskRepository
	columnRepo  repository.ColumnRepository
	boardRepo   repository.BoardRepository
	memberRepo  repository.MemberRepository
	commentRepo repository.CommentRepository
	cache       *cache.Cache
}

func NewTaskService(
	taskRepo repository.TaskRepository,
	columnRepo repository.ColumnRepository,
	boardRepo repository.BoardRepository,
	memberRepo repository.MemberRepository,
	commentRepo repository.CommentRepository,
	cache *cache.Cache,
) *TaskService {
	return &TaskService{
		taskRepo:    taskRepo,
		columnRepo:  columnRepo,
		boardRepo:   boardRepo,
		memberRepo:  memberRepo,
		commentRepo: commentRepo,
		cache:       cache,
	}
}

func (s *TaskService) Create(ctx context.Context, teamID, memberID string, req *request.CreateTaskRequest) (*response.TaskResponse, error) {
	if len(strings.TrimSpace(req.Title)) < 1 || len(req.Title) > 200 {
		return nil, domainErrors.ValidationErrors{
			&domainErrors.ValidationError{Field: "title", Message: "must be 1-200 characters"},
		}
	}
	if req.Description != nil && len(*req.Description) > 2000 {
		return nil, domainErrors.ValidationErrors{
			&domainErrors.ValidationError{Field: "description", Message: "must be at most 2000 characters"},
		}
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	if !isValidPriority(priority) {
		return nil, domainErrors.ValidationErrors{
			&domainErrors.ValidationError{Field: "priority", Message: "must be one of: low, medium, high, critical"},
		}
	}

	// Validate column belongs to team
	column, err := s.columnRepo.GetByID(ctx, req.ColumnID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}
	board, err := s.boardRepo.GetByID(ctx, column.BoardID)
	if err != nil || board.TeamID != teamID {
		return nil, domainErrors.ErrNotFound
	}

	// Validate assignee belongs to team
	if req.AssigneeID != nil && *req.AssigneeID != "" {
		assignee, err := s.memberRepo.GetByID(ctx, *req.AssigneeID)
		if err != nil || assignee.TeamID != teamID {
			return nil, domainErrors.ValidationErrors{
				&domainErrors.ValidationError{Field: "assignee_id", Message: "invalid assignee"},
			}
		}
	}

	// Parse due date
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		parsed, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			return nil, domainErrors.ValidationErrors{
				&domainErrors.ValidationError{Field: "due_date", Message: "must be ISO 8601 format"},
			}
		}
		dueDate = &parsed
	}

	// Determine position
	position := 0
	maxPos, _ := s.taskRepo.MaxPosition(ctx, req.ColumnID)
	position = maxPos + 1

	now := time.Now()
	task := &entity.Task{
		ID:          newULID(),
		ColumnID:    req.ColumnID,
		BoardID:     column.BoardID,
		Title:       req.Title,
		Description: req.Description,
		Status:      column.Status,
		Priority:    priority,
		Position:    position,
		AssigneeID:  nilIfEmpty(req.AssigneeID),
		DueDate:     dueDate,
		CreatedBy:   memberID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("creating task: %w", err)
	}

	_ = s.cache.InvalidateBoard(ctx, task.BoardID)

	// Fetch with assignee info
	created, _ := s.taskRepo.GetByID(ctx, task.ID)

	return &response.TaskResponse{Data: taskToResponse(created)}, nil
}

func (s *TaskService) GetByID(ctx context.Context, teamID, taskID string) (*response.TaskDetailResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}

	board, err := s.boardRepo.GetByID(ctx, task.BoardID)
	if err != nil || board.TeamID != teamID {
		return nil, domainErrors.ErrNotFound
	}

	// Get comments
	comments, _, err := s.commentRepo.ListByTask(ctx, taskID, 1, 100)
	if err != nil {
		return nil, fmt.Errorf("listing comments: %w", err)
	}

	commentItems := make([]response.CommentItem, 0, len(comments))
	for _, c := range comments {
		commentItems = append(commentItems, commentToResponse(c))
	}

	return &response.TaskDetailResponse{
		Data: response.TaskDetailItem{
			ID:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Status:      task.Status,
			Priority:    task.Priority,
			Position:    task.Position,
			ColumnID:    task.ColumnID,
			BoardID:     task.BoardID,
			Assignee:    taskToResponse(task).Assignee,
			DueDate:     task.DueDate,
			Comments:    commentItems,
			CreatedAt:   task.CreatedAt,
			UpdatedAt:   task.UpdatedAt,
		},
	}, nil
}

func (s *TaskService) Update(ctx context.Context, teamID, taskID string, req *request.UpdateTaskRequest) (*response.TaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}

	board, err := s.boardRepo.GetByID(ctx, task.BoardID)
	if err != nil || board.TeamID != teamID {
		return nil, domainErrors.ErrNotFound
	}

	if req.Title != nil {
		if len(strings.TrimSpace(*req.Title)) < 1 || len(*req.Title) > 200 {
			return nil, domainErrors.ValidationErrors{
				&domainErrors.ValidationError{Field: "title", Message: "must be 1-200 characters"},
			}
		}
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = req.Description
	}
	if req.Priority != nil {
		if !isValidPriority(*req.Priority) {
			return nil, domainErrors.ValidationErrors{
				&domainErrors.ValidationError{Field: "priority", Message: "invalid priority"},
			}
		}
		task.Priority = *req.Priority
	}
	if req.AssigneeID != nil {
		if *req.AssigneeID != "" {
			assignee, err := s.memberRepo.GetByID(ctx, *req.AssigneeID)
			if err != nil || assignee.TeamID != teamID {
				return nil, domainErrors.ValidationErrors{
					&domainErrors.ValidationError{Field: "assignee_id", Message: "invalid assignee"},
				}
			}
		}
		task.AssigneeID = nilIfEmpty(req.AssigneeID)
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			task.DueDate = nil
		} else {
			parsed, err := time.Parse(time.RFC3339, *req.DueDate)
			if err != nil {
				return nil, domainErrors.ValidationErrors{
					&domainErrors.ValidationError{Field: "due_date", Message: "must be ISO 8601 format"},
				}
			}
			task.DueDate = &parsed
		}
	}

	task.UpdatedAt = time.Now()
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("updating task: %w", err)
	}

	_ = s.cache.InvalidateBoard(ctx, task.BoardID)

	updated, _ := s.taskRepo.GetByID(ctx, task.ID)
	return &response.TaskResponse{Data: taskToResponse(updated)}, nil
}

func (s *TaskService) Move(ctx context.Context, teamID, taskID string, req *request.MoveTaskRequest) (*response.TaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}

	board, err := s.boardRepo.GetByID(ctx, task.BoardID)
	if err != nil || board.TeamID != teamID {
		return nil, domainErrors.ErrNotFound
	}

	// Validate target column
	targetCol, err := s.columnRepo.GetByID(ctx, req.ColumnID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}

	// Target column must be on the same board
	targetBoard, err := s.boardRepo.GetByID(ctx, targetCol.BoardID)
	if err != nil || targetBoard.TeamID != teamID || targetCol.BoardID != task.BoardID {
		return nil, domainErrors.ErrBadRequest
	}

	position := 0
	if req.Position != nil && *req.Position >= 0 {
		position = *req.Position
	} else {
		maxPos, _ := s.taskRepo.MaxPosition(ctx, req.ColumnID)
		position = maxPos + 1
	}

	if err := s.taskRepo.Move(ctx, taskID, req.ColumnID, position); err != nil {
		return nil, fmt.Errorf("moving task: %w", err)
	}

	// Update task status to match column status
	_ = s.taskRepo.UpdateStatus(ctx, taskID, targetCol.Status)

	// Re-sequence old column if different
	if task.ColumnID != req.ColumnID {
		_ = s.taskRepo.ReSequencePositions(ctx, task.ColumnID)
	}
	_ = s.taskRepo.ReSequencePositions(ctx, req.ColumnID)

	_ = s.cache.InvalidateBoard(ctx, task.BoardID)

	updated, _ := s.taskRepo.GetByID(ctx, taskID)
	return &response.TaskResponse{Data: taskToResponse(updated)}, nil
}

func (s *TaskService) Delete(ctx context.Context, teamID, taskID string) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return domainErrors.ErrNotFound
	}

	board, err := s.boardRepo.GetByID(ctx, task.BoardID)
	if err != nil || board.TeamID != teamID {
		return domainErrors.ErrNotFound
	}

	columnID := task.ColumnID
	boardID := task.BoardID

	if err := s.taskRepo.Delete(ctx, taskID); err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}

	_ = s.taskRepo.ReSequencePositions(ctx, columnID)
	_ = s.cache.InvalidateBoard(ctx, boardID)

	return nil
}

// --- Comment Service ---

type CommentService struct {
	commentRepo repository.CommentRepository
	taskRepo    repository.TaskRepository
	boardRepo   repository.BoardRepository
	memberRepo  repository.MemberRepository
	cache       *cache.Cache
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	taskRepo repository.TaskRepository,
	boardRepo repository.BoardRepository,
	memberRepo repository.MemberRepository,
	cache *cache.Cache,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		taskRepo:    taskRepo,
		boardRepo:   boardRepo,
		memberRepo:  memberRepo,
		cache:       cache,
	}
}

func (s *CommentService) Create(ctx context.Context, teamID, taskID, authorID string, req *request.CreateCommentRequest) (*response.CommentResponse, error) {
	if len(strings.TrimSpace(req.Body)) < 1 || len(req.Body) > 2000 {
		return nil, domainErrors.ValidationErrors{
			&domainErrors.ValidationError{Field: "body", Message: "must be 1-2000 characters"},
		}
	}

	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}

	board, err := s.boardRepo.GetByID(ctx, task.BoardID)
	if err != nil || board.TeamID != teamID {
		return nil, domainErrors.ErrNotFound
	}

	author, err := s.memberRepo.GetByID(ctx, authorID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}

	now := time.Now()
	comment := &entity.Comment{
		ID:        newULID(),
		TaskID:    taskID,
		AuthorID:  authorID,
		Body:      req.Body,
		Author:    &entity.MemberSummary{ID: author.ID, Name: author.Name},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("creating comment: %w", err)
	}

	_ = s.cache.InvalidateBoard(ctx, task.BoardID)

	return &response.CommentResponse{
		Data: commentToResponse(comment),
	}, nil
}

func (s *CommentService) ListByTask(ctx context.Context, teamID, taskID string, page, perPage int) (*response.CommentListResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, domainErrors.ErrNotFound
	}

	board, err := s.boardRepo.GetByID(ctx, task.BoardID)
	if err != nil || board.TeamID != teamID {
		return nil, domainErrors.ErrNotFound
	}

	if perPage > 100 {
		perPage = 100
	}
	if perPage < 1 {
		perPage = 20
	}
	if page < 1 {
		page = 1
	}

	comments, total, err := s.commentRepo.ListByTask(ctx, taskID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("listing comments: %w", err)
	}

	totalPages := total / perPage
	if total%perPage > 0 {
		totalPages++
	}

	items := make([]response.CommentItem, 0, len(comments))
	for _, c := range comments {
		items = append(items, commentToResponse(c))
	}

	return &response.CommentListResponse{
		Data: items,
		Meta: response.PaginationMeta{
			Total:      total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *CommentService) Delete(ctx context.Context, teamID, commentID, memberID string, isAdmin bool) error {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return domainErrors.ErrNotFound
	}

	task, err := s.taskRepo.GetByID(ctx, comment.TaskID)
	if err != nil {
		return domainErrors.ErrNotFound
	}

	board, err := s.boardRepo.GetByID(ctx, task.BoardID)
	if err != nil || board.TeamID != teamID {
		return domainErrors.ErrNotFound
	}

	if !isAdmin && comment.AuthorID != memberID {
		return domainErrors.ErrForbidden
	}

	if err := s.commentRepo.Delete(ctx, commentID); err != nil {
		return fmt.Errorf("deleting comment: %w", err)
	}

	_ = s.cache.InvalidateBoard(ctx, task.BoardID)
	return nil
}

// --- Helper functions ---

func taskToResponse(t *entity.Task) response.TaskItem {
	item := response.TaskItem{
		ID:           t.ID,
		Title:        t.Title,
		Description:  t.Description,
		Status:       t.Status,
		Priority:     t.Priority,
		Position:     t.Position,
		ColumnID:     t.ColumnID,
		BoardID:      t.BoardID,
		DueDate:      t.DueDate,
		CommentCount: t.CommentCount,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
	if t.Assignee != nil {
		item.Assignee = &response.MemberSummary{
			ID:   t.Assignee.ID,
			Name: t.Assignee.Name,
		}
	}
	return item
}

func commentToResponse(c *entity.Comment) response.CommentItem {
	item := response.CommentItem{
		ID:        c.ID,
		Body:      c.Body,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	if c.Author != nil {
		item.Author = &response.MemberSummary{
			ID:   c.Author.ID,
			Name: c.Author.Name,
		}
	}
	return item
}

func isValidEmail(email string) bool {
	if len(email) < 3 || len(email) > 255 {
		return false
	}
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return false
	}
	local := parts[0]
	domain := parts[1]
	if len(local) == 0 || len(domain) < 3 {
		return false
	}
	// Domain must contain at least one dot with chars on both sides
	dotIdx := strings.LastIndex(domain, ".")
	if dotIdx <= 0 || dotIdx == len(domain)-1 {
		return false
	}
	return true
}

func isValidStatus(status string) bool {
	return status == "todo" || status == "in_progress" || status == "review" || status == "done"
}

func isValidPriority(priority string) bool {
	return priority == "low" || priority == "medium" || priority == "high" || priority == "critical"
}

func nilIfEmpty(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

// Ensure services implement their interfaces at compile time.
// (We reference redis/v9 in imports so the module resolution works.)
var _ = redis.Client{}
