package entity

import "time"

type Column struct {
	ID        string     `json:"id"`
	BoardID   string     `json:"board_id"`
	Name      string     `json:"name"`
	Position  int        `json:"position"`
	Status    string     `json:"status"`
	Tasks     []*Task    `json:"tasks,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}
