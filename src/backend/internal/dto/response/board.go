package response

import "time"

// --- Board ---

type BoardResponse struct {
	Data BoardItem `json:"data"`
}

type BoardItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BoardListResponse struct {
	Data []BoardItem `json:"data"`
}

type BoardDetailResponse struct {
	Data BoardDetail `json:"data"`
}

type BoardDetail struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description *string        `json:"description"`
	Columns     []ColumnDetail `json:"columns"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type ColumnDetail struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Position  int           `json:"position"`
	Status    string        `json:"status"`
	Tasks     []TaskItem    `json:"tasks"`
}

// --- Column standalone ---

type ColumnResponse struct {
	Data ColumnItem `json:"data"`
}

type ColumnItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
	Status   string `json:"status"`
	BoardID  string `json:"board_id"`
}

type ColumnListResponse struct {
	Data []ColumnItem `json:"data"`
}

// --- Task ---

type TaskResponse struct {
	Data TaskItem `json:"data"`
}

type TaskItem struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	Description  *string         `json:"description"`
	Status       string          `json:"status"`
	Priority     string          `json:"priority"`
	Position     int             `json:"position"`
	ColumnID     string          `json:"column_id"`
	BoardID      string          `json:"board_id"`
	Assignee     *MemberSummary  `json:"assignee"`
	DueDate      *time.Time      `json:"due_date"`
	CommentCount int             `json:"comment_count"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type TaskDetailResponse struct {
	Data TaskDetailItem `json:"data"`
}

type TaskDetailItem struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Description *string         `json:"description"`
	Status      string          `json:"status"`
	Priority    string          `json:"priority"`
	Position    int             `json:"position"`
	ColumnID    string          `json:"column_id"`
	BoardID     string          `json:"board_id"`
	Assignee    *MemberSummary  `json:"assignee"`
	DueDate     *time.Time      `json:"due_date"`
	Comments    []CommentItem   `json:"comments"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// --- Comment ---

type CommentResponse struct {
	Data CommentItem `json:"data"`
}

type CommentItem struct {
	ID        string         `json:"id"`
	Body      string         `json:"body"`
	Author    *MemberSummary `json:"author"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type CommentListResponse struct {
	Data []CommentItem  `json:"data"`
	Meta PaginationMeta `json:"meta"`
}
