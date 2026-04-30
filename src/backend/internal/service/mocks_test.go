package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"teamboard/internal/domain/entity"
	"teamboard/internal/repository/cache"
)

// newTestCache creates a cache backed by an in-memory miniredis instance.
func newTestCache() (*cache.Cache, *miniredis.Miniredis) {
	mr := miniredis.NewMiniRedis()
	_ = mr.Start()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return cache.New(rdb, 60), mr
}

// --- Mock Team Repository ---

type mockTeamRepo struct {
	mu     sync.RWMutex
	teams  map[string]*entity.Team
	byName map[string]*entity.Team
	err    error
}

func newMockTeamRepo() *mockTeamRepo {
	return &mockTeamRepo{
		teams:  make(map[string]*entity.Team),
		byName: make(map[string]*entity.Team),
	}
}

func (m *mockTeamRepo) Create(_ context.Context, team *entity.Team) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.teams[team.ID] = team
	m.byName[team.Name] = team
	return nil
}

func (m *mockTeamRepo) GetByID(_ context.Context, id string) (*entity.Team, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.teams[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return t, nil
}

func (m *mockTeamRepo) GetByName(_ context.Context, name string) (*entity.Team, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.byName[name]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return t, nil
}

// --- Mock Member Repository ---

type mockMemberRepo struct {
	mu      sync.RWMutex
	members map[string]*entity.Member
	byHash  map[string]*entity.Member
	err     error
}

func newMockMemberRepo() *mockMemberRepo {
	return &mockMemberRepo{
		members: make(map[string]*entity.Member),
		byHash:  make(map[string]*entity.Member),
	}
}

func (m *mockMemberRepo) Create(_ context.Context, member *entity.Member) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.members[member.ID] = member
	m.byHash[member.APIKeyHash] = member
	return nil
}

func (m *mockMemberRepo) GetByID(_ context.Context, id string) (*entity.Member, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	mb, ok := m.members[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return mb, nil
}

func (m *mockMemberRepo) GetByAPIKeyHash(_ context.Context, hash string) (*entity.Member, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	mb, ok := m.byHash[hash]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return mb, nil
}

func (m *mockMemberRepo) ListByTeam(_ context.Context, teamID string) ([]*entity.Member, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*entity.Member
	for _, mb := range m.members {
		if mb.TeamID == teamID {
			result = append(result, mb)
		}
	}
	return result, nil
}

func (m *mockMemberRepo) Update(_ context.Context, member *entity.Member) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	// Remove old hash mapping if APIKeyHash changed
	if old, ok := m.members[member.ID]; ok {
		delete(m.byHash, old.APIKeyHash)
	}
	m.members[member.ID] = member
	m.byHash[member.APIKeyHash] = member
	return nil
}

func (m *mockMemberRepo) Delete(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	mb, ok := m.members[id]
	if !ok {
		return fmt.Errorf("not found")
	}
	delete(m.byHash, mb.APIKeyHash)
	delete(m.members, id)
	return nil
}

func (m *mockMemberRepo) CountAdmins(_ context.Context, teamID string) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, mb := range m.members {
		if mb.TeamID == teamID && mb.Role == "admin" {
			count++
		}
	}
	return count, nil
}

// --- Mock Board Repository ---

type mockBoardRepo struct {
	mu     sync.RWMutex
	boards map[string]*entity.Board
	err    error
}

func newMockBoardRepo() *mockBoardRepo {
	return &mockBoardRepo{boards: make(map[string]*entity.Board)}
}

func (m *mockBoardRepo) Create(_ context.Context, board *entity.Board) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.boards[board.ID] = board
	return nil
}

func (m *mockBoardRepo) GetByID(_ context.Context, id string) (*entity.Board, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.boards[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return b, nil
}

func (m *mockBoardRepo) ListByTeam(_ context.Context, teamID string) ([]*entity.Board, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*entity.Board
	for _, b := range m.boards {
		if b.TeamID == teamID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (m *mockBoardRepo) Update(_ context.Context, board *entity.Board) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.boards[board.ID] = board
	return nil
}

func (m *mockBoardRepo) Delete(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.boards, id)
	return nil
}

// --- Mock Column Repository ---

type mockColumnRepo struct {
	mu      sync.RWMutex
	columns map[string]*entity.Column
	err     error
}

func newMockColumnRepo() *mockColumnRepo {
	return &mockColumnRepo{columns: make(map[string]*entity.Column)}
}

func (m *mockColumnRepo) Create(_ context.Context, col *entity.Column) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.columns[col.ID] = col
	return nil
}

func (m *mockColumnRepo) GetByID(_ context.Context, id string) (*entity.Column, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.columns[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

func (m *mockColumnRepo) ListByBoard(_ context.Context, boardID string) ([]*entity.Column, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*entity.Column
	for _, c := range m.columns {
		if c.BoardID == boardID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockColumnRepo) Update(_ context.Context, col *entity.Column) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.columns[col.ID] = col
	return nil
}

func (m *mockColumnRepo) UpdatePositions(_ context.Context, cols []*entity.Column) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range cols {
		m.columns[c.ID] = c
	}
	return nil
}

func (m *mockColumnRepo) ShiftPositions(_ context.Context, boardID string, fromPosition int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.columns {
		if c.BoardID == boardID && c.Position >= fromPosition {
			c.Position++
		}
	}
	return nil
}

func (m *mockColumnRepo) Delete(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.columns, id)
	return nil
}

func (m *mockColumnRepo) CountByBoard(_ context.Context, boardID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, c := range m.columns {
		if c.BoardID == boardID {
			count++
		}
	}
	return count, nil
}

func (m *mockColumnRepo) CountTasks(_ context.Context, columnID string) (int, error) {
	return 0, nil
}

func (m *mockColumnRepo) MaxPosition(_ context.Context, boardID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	max := -1
	for _, c := range m.columns {
		if c.BoardID == boardID && c.Position > max {
			max = c.Position
		}
	}
	return max, nil
}

// --- Mock Task Repository ---

type mockTaskRepo struct {
	mu    sync.RWMutex
	tasks map[string]*entity.Task
	err   error
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{tasks: make(map[string]*entity.Task)}
}

func (m *mockTaskRepo) Create(_ context.Context, task *entity.Task) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepo) GetByID(_ context.Context, id string) (*entity.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tasks[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return t, nil
}

func (m *mockTaskRepo) ListByBoard(_ context.Context, boardID string) ([]*entity.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*entity.Task
	for _, t := range m.tasks {
		if t.BoardID == boardID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTaskRepo) ListByColumn(_ context.Context, columnID string) ([]*entity.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*entity.Task
	for _, t := range m.tasks {
		if t.ColumnID == columnID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTaskRepo) Update(_ context.Context, task *entity.Task) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepo) UpdateStatus(_ context.Context, taskID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("not found")
	}
	t.Status = status
	return nil
}

func (m *mockTaskRepo) Move(_ context.Context, taskID, columnID string, position int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("not found")
	}
	t.ColumnID = columnID
	t.Position = position
	return nil
}

func (m *mockTaskRepo) Delete(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tasks, id)
	return nil
}

func (m *mockTaskRepo) ReSequencePositions(_ context.Context, columnID string) error {
	return nil
}

func (m *mockTaskRepo) UnassignByMember(_ context.Context, memberID string) error {
	return nil
}

func (m *mockTaskRepo) CountByColumn(_ context.Context, columnID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, t := range m.tasks {
		if t.ColumnID == columnID {
			count++
		}
	}
	return count, nil
}

func (m *mockTaskRepo) MaxPosition(_ context.Context, columnID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	max := -1
	for _, t := range m.tasks {
		if t.ColumnID == columnID && t.Position > max {
			max = t.Position
		}
	}
	return max, nil
}

// --- Mock Comment Repository ---

type mockCommentRepo struct {
	mu       sync.RWMutex
	comments map[string]*entity.Comment
	err      error
}

func newMockCommentRepo() *mockCommentRepo {
	return &mockCommentRepo{comments: make(map[string]*entity.Comment)}
}

func (m *mockCommentRepo) Create(_ context.Context, comment *entity.Comment) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.comments[comment.ID] = comment
	return nil
}

func (m *mockCommentRepo) GetByID(_ context.Context, id string) (*entity.Comment, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.comments[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

func (m *mockCommentRepo) ListByTask(_ context.Context, taskID string, page, perPage int) ([]*entity.Comment, int, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var all []*entity.Comment
	for _, c := range m.comments {
		if c.TaskID == taskID {
			all = append(all, c)
		}
	}
	total := len(all)
	start := (page - 1) * perPage
	if start >= total {
		return []*entity.Comment{}, total, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (m *mockCommentRepo) Delete(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.comments, id)
	return nil
}

func (m *mockCommentRepo) CountByTask(_ context.Context, taskID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, c := range m.comments {
		if c.TaskID == taskID {
			count++
		}
	}
	return count, nil
}

// --- Helpers ---

func seedMember(repo *mockMemberRepo, id, teamID, name, email, role string) *entity.Member {
	key := GenerateAPIKey()
	m := &entity.Member{
		ID:           id,
		TeamID:       teamID,
		Name:         name,
		Email:        email,
		Role:         role,
		APIKeyHash:   HashAPIKey(key),
		APIKeyPrefix: key[:10],
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.members[id] = m
	repo.byHash[m.APIKeyHash] = m
	return m
}

func seedTeam(repo *mockTeamRepo, id, name string) *entity.Team {
	t := &entity.Team{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.teams[id] = t
	repo.byName[name] = t
	return t
}

func seedBoard(repo *mockBoardRepo, id, teamID, name string) *entity.Board {
	b := &entity.Board{
		ID:        id,
		TeamID:    teamID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.boards[id] = b
	return b
}

func seedColumn(repo *mockColumnRepo, id, boardID, name string, position int, status string) *entity.Column {
	c := &entity.Column{
		ID:        id,
		BoardID:   boardID,
		Name:      name,
		Position:  position,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.columns[id] = c
	return c
}

func seedTask(repo *mockTaskRepo, id, columnID, boardID, title, status, priority string, position int) *entity.Task {
	t := &entity.Task{
		ID:        id,
		ColumnID:  columnID,
		BoardID:   boardID,
		Title:     title,
		Status:    status,
		Priority:  priority,
		Position:  position,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.tasks[id] = t
	return t
}

func seedComment(repo *mockCommentRepo, id, taskID, authorID, body string) *entity.Comment {
	c := &entity.Comment{
		ID:        id,
		TaskID:    taskID,
		AuthorID:  authorID,
		Body:      body,
		Author:    &entity.MemberSummary{ID: authorID, Name: "Author"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.comments[id] = c
	return c
}
