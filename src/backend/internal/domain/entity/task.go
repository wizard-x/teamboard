package entity

import "time"

type Task struct {
	ID          string       `json:"id"`
	ColumnID    string       `json:"column_id"`
	BoardID     string       `json:"board_id"`
	Title       string       `json:"title"`
	Description *string      `json:"description"`
	Status      string       `json:"status"`
	Priority    string       `json:"priority"`
	Position    int          `json:"position"`
	AssigneeID  *string      `json:"assignee_id"`
	DueDate     *time.Time   `json:"due_date"`
	CreatedBy   string       `json:"created_by"`
	Assignee    *MemberSummary `json:"assignee,omitempty"`
	CommentCount int         `json:"comment_count,omitempty"`
	Comments    []*Comment   `json:"comments,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `json:"-"`
}

type MemberSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
