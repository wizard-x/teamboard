package postgres

import (
	"context"
	"fmt"

	"teamboard/internal/domain/entity"
)

type ColumnRepo struct {
	db *DB
}

func NewColumnRepo(db *DB) *ColumnRepo {
	return &ColumnRepo{db: db}
}

func (r *ColumnRepo) Create(ctx context.Context, c *entity.Column) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO board_columns (id, board_id, name, position, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		c.ID, c.BoardID, c.Name, c.Position, c.Status, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting column: %w", err)
	}
	return nil
}

func (r *ColumnRepo) GetByID(ctx context.Context, id string) (*entity.Column, error) {
	var c entity.Column
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, board_id, name, position, status, created_at, updated_at
		 FROM board_columns WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&c.ID, &c.BoardID, &c.Name, &c.Position, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting column: %w", err)
	}
	return &c, nil
}

func (r *ColumnRepo) ListByBoard(ctx context.Context, boardID string) ([]*entity.Column, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, board_id, name, position, status, created_at, updated_at
		 FROM board_columns WHERE board_id = $1 AND deleted_at IS NULL ORDER BY position`, boardID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing columns: %w", err)
	}
	defer rows.Close()

	var columns []*entity.Column
	for rows.Next() {
		var c entity.Column
		if err := rows.Scan(&c.ID, &c.BoardID, &c.Name, &c.Position, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning column: %w", err)
		}
		columns = append(columns, &c)
	}
	return columns, nil
}

func (r *ColumnRepo) Update(ctx context.Context, c *entity.Column) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE board_columns SET name = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`,
		c.Name, c.UpdatedAt, c.ID,
	)
	if err != nil {
		return fmt.Errorf("updating column: %w", err)
	}
	return nil
}

func (r *ColumnRepo) UpdatePositions(ctx context.Context, columns []*entity.Column) error {
	if len(columns) == 0 {
		return nil
	}
	for _, c := range columns {
		_, err := r.db.Pool.Exec(ctx,
			`UPDATE board_columns SET position = $1, updated_at = NOW() WHERE id = $2`,
			c.Position, c.ID,
		)
		if err != nil {
			return fmt.Errorf("updating column position: %w", err)
		}
	}
	return nil
}

func (r *ColumnRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE board_columns SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("deleting column: %w", err)
	}
	return nil
}

func (r *ColumnRepo) CountByBoard(ctx context.Context, boardID string) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM board_columns WHERE board_id = $1 AND deleted_at IS NULL`, boardID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting columns: %w", err)
	}
	return count, nil
}

func (r *ColumnRepo) CountTasks(ctx context.Context, columnID string) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tasks WHERE column_id = $1 AND deleted_at IS NULL`, columnID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting tasks in column: %w", err)
	}
	return count, nil
}

func (r *ColumnRepo) MaxPosition(ctx context.Context, boardID string) (int, error) {
	var maxPos *int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT MAX(position) FROM board_columns WHERE board_id = $1 AND deleted_at IS NULL`, boardID,
	).Scan(&maxPos)
	if err != nil {
		return 0, fmt.Errorf("getting max position: %w", err)
	}
	if maxPos == nil {
		return -1, nil
	}
	return *maxPos, nil
}

func (r *ColumnRepo) ShiftPositions(ctx context.Context, boardID string, fromPosition int) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE board_columns SET position = position + 1, updated_at = NOW()
		 WHERE board_id = $1 AND position >= $2 AND deleted_at IS NULL`,
		boardID, fromPosition,
	)
	if err != nil {
		return fmt.Errorf("shifting column positions: %w", err)
	}
	return nil
}
