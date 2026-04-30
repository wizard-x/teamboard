package entity

import "time"

type Member struct {
	ID           string     `json:"id"`
	TeamID       string     `json:"team_id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Role         string     `json:"role"`
	APIKeyHash   string     `json:"-"`
	APIKeyPrefix string     `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}
