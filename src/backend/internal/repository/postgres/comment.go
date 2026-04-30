package postgres

import (
	"context"
	"fmt"

	"teamboard/internal/domain/entity"
)

type CommentRepo struct {
	db *DB
}

func NewCommentRepo(db *DB) *CommentRepo {
	return &CommentRepo{db: db}
}

func (r *CommentRepo) Create(ctx context.Context, c *entity.Comment) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO comments (id, task_id, author_id, body, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		c.ID, c.TaskID, c.AuthorID, c.Body, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting comment: %w", err)
	}
	return nil
}

func (r *CommentRepo) GetByID(ctx context.Context, id string) (*entity.Comment, error) {
	var c entity.Comment
	var authorName string
	err := r.db.Pool.QueryRow(ctx,
		`SELECT c.id, c.task_id, c.author_id, c.body, c.created_at, c.updated_at, m.name
		 FROM comments c
		 JOIN members m ON c.author_id = m.id AND m.deleted_at IS NULL
		 WHERE c.id = $1 AND c.deleted_at IS NULL`, id,
	).Scan(&c.ID, &c.TaskID, &c.AuthorID, &c.Body, &c.CreatedAt, &c.UpdatedAt, &authorName)
	if err != nil {
		return nil, fmt.Errorf("getting comment: %w", err)
	}
	c.Author = &entity.MemberSummary{ID: c.AuthorID, Name: authorName}
	return &c, nil
}

func (r *CommentRepo) ListByTask(ctx context.Context, taskID string, page, perPage int) ([]*entity.Comment, int, error) {
	// Count total
	var total int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM comments WHERE task_id = $1 AND deleted_at IS NULL`, taskID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting comments: %w", err)
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Pool.Query(ctx,
		`SELECT c.id, c.task_id, c.author_id, c.body, c.created_at, c.updated_at, m.name
		 FROM comments c
		 JOIN members m ON c.author_id = m.id AND m.deleted_at IS NULL
		 WHERE c.task_id = $1 AND c.deleted_at IS NULL
		 ORDER BY c.created_at
		 LIMIT $2 OFFSET $3`, taskID, perPage, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing comments: %w", err)
	}
	defer rows.Close()

	var comments []*entity.Comment
	for rows.Next() {
		var c entity.Comment
		var authorName string
		if err := rows.Scan(&c.ID, &c.TaskID, &c.AuthorID, &c.Body, &c.CreatedAt, &c.UpdatedAt, &authorName); err != nil {
			return nil, 0, fmt.Errorf("scanning comment: %w", err)
		}
		c.Author = &entity.MemberSummary{ID: c.AuthorID, Name: authorName}
		comments = append(comments, &c)
	}
	return comments, total, nil
}

func (r *CommentRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE comments SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("deleting comment: %w", err)
	}
	return nil
}

func (r *CommentRepo) CountByTask(ctx context.Context, taskID string) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM comments WHERE task_id = $1 AND deleted_at IS NULL`, taskID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting comments: %w", err)
	}
	return count, nil
}
