package postgres

import (
	"context"
	"fmt"

	"teamboard/internal/domain/entity"
)

type BoardRepo struct {
	db *DB
}

func NewBoardRepo(db *DB) *BoardRepo {
	return &BoardRepo{db: db}
}

func (r *BoardRepo) Create(ctx context.Context, b *entity.Board) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO boards (id, team_id, name, description, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		b.ID, b.TeamID, b.Name, b.Description, b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting board: %w", err)
	}
	return nil
}

func (r *BoardRepo) GetByID(ctx context.Context, id string) (*entity.Board, error) {
	var b entity.Board
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, team_id, name, description, created_at, updated_at
		 FROM boards WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&b.ID, &b.TeamID, &b.Name, &b.Description, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting board: %w", err)
	}
	return &b, nil
}

func (r *BoardRepo) ListByTeam(ctx context.Context, teamID string) ([]*entity.Board, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, team_id, name, description, created_at, updated_at
		 FROM boards WHERE team_id = $1 AND deleted_at IS NULL ORDER BY created_at`, teamID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing boards: %w", err)
	}
	defer rows.Close()

	var boards []*entity.Board
	for rows.Next() {
		var b entity.Board
		if err := rows.Scan(&b.ID, &b.TeamID, &b.Name, &b.Description, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning board: %w", err)
		}
		boards = append(boards, &b)
	}
	return boards, nil
}

func (r *BoardRepo) Update(ctx context.Context, b *entity.Board) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE boards SET name = $1, description = $2, updated_at = $3 WHERE id = $4 AND deleted_at IS NULL`,
		b.Name, b.Description, b.UpdatedAt, b.ID,
	)
	if err != nil {
		return fmt.Errorf("updating board: %w", err)
	}
	return nil
}

func (r *BoardRepo) Delete(ctx context.Context, id string) error {
	// Cascade soft-delete: comments → tasks → columns → board
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning tx: %w", err)
	}
	defer tx.Rollback(ctx)

	now := "NOW()"

	// Delete comments for tasks in this board's columns
	_, err = tx.Exec(ctx,
		`UPDATE comments SET deleted_at = `+now+`, updated_at = `+now+`
		 WHERE task_id IN (SELECT t.id FROM tasks t JOIN board_columns c ON t.column_id = c.id WHERE c.board_id = $1 AND t.deleted_at IS NULL)`,
		id,
	)
	if err != nil {
		return fmt.Errorf("deleting board comments: %w", err)
	}

	// Delete tasks in this board's columns
	_, err = tx.Exec(ctx,
		`UPDATE tasks SET deleted_at = `+now+`, updated_at = `+now+`
		 WHERE board_id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("deleting board tasks: %w", err)
	}

	// Delete columns
	_, err = tx.Exec(ctx,
		`UPDATE board_columns SET deleted_at = `+now+`, updated_at = `+now+`
		 WHERE board_id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("deleting board columns: %w", err)
	}

	// Delete board
	tag, err := tx.Exec(ctx,
		`UPDATE boards SET deleted_at = `+now+`, updated_at = `+now+`
		 WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("deleting board: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("board not found")
	}

	return tx.Commit(ctx)
}
