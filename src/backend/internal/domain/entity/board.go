package entity

import "time"

type Board struct {
	ID          string     `json:"id"`
	TeamID      string     `json:"team_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Columns     []*Column  `json:"columns,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"-"`
}
