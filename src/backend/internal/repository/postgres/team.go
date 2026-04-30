package postgres

import (
	"context"
	"fmt"

	"teamboard/internal/domain/entity"
)

type TeamRepo struct {
	db *DB
}

func NewTeamRepo(db *DB) *TeamRepo {
	return &TeamRepo{db: db}
}

func (r *TeamRepo) Create(ctx context.Context, team *entity.Team) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO teams (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		team.ID, team.Name, team.CreatedAt, team.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting team: %w", err)
	}
	return nil
}

func (r *TeamRepo) GetByID(ctx context.Context, id string) (*entity.Team, error) {
	var t entity.Team
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, name, created_at, updated_at FROM teams WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting team by id: %w", err)
	}
	return &t, nil
}

func (r *TeamRepo) GetByName(ctx context.Context, name string) (*entity.Team, error) {
	var t entity.Team
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, name, created_at, updated_at FROM teams WHERE name = $1 AND deleted_at IS NULL`,
		name,
	).Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting team by name: %w", err)
	}
	return &t, nil
}
