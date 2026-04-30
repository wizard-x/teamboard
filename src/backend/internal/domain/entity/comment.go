package entity

import "time"

type Comment struct {
	ID        string         `json:"id"`
	TaskID    string         `json:"task_id"`
	AuthorID  string         `json:"author_id"`
	Body      string         `json:"body"`
	Author    *MemberSummary `json:"author,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt *time.Time     `json:"-"`
}
