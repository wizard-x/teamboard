package postgres

import (
	"context"
	"fmt"

	"teamboard/internal/domain/entity"
)

type TaskRepo struct {
	db *DB
}

func NewTaskRepo(db *DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, t *entity.Task) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO tasks (id, column_id, board_id, title, description, status, priority, position, assignee_id, due_date, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		t.ID, t.ColumnID, t.BoardID, t.Title, t.Description, t.Status, t.Priority, t.Position, t.AssigneeID, t.DueDate, t.CreatedBy, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting task: %w", err)
	}
	return nil
}

func (r *TaskRepo) GetByID(ctx context.Context, id string) (*entity.Task, error) {
	var t entity.Task
	var assigneeName *string
	err := r.db.Pool.QueryRow(ctx,
		`SELECT t.id, t.column_id, t.board_id, t.title, t.description, t.status, t.priority,
		        t.position, t.assignee_id, t.due_date, t.created_by, t.created_at, t.updated_at,
		        m.name as assignee_name,
		        (SELECT COUNT(*) FROM comments c WHERE c.task_id = t.id AND c.deleted_at IS NULL) as comment_count
		 FROM tasks t
		 LEFT JOIN members m ON t.assignee_id = m.id AND m.deleted_at IS NULL
		 WHERE t.id = $1 AND t.deleted_at IS NULL`, id,
	).Scan(&t.ID, &t.ColumnID, &t.BoardID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.Position, &t.AssigneeID, &t.DueDate, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
		&assigneeName, &t.CommentCount,
	)
	if err != nil {
		return nil, fmt.Errorf("getting task: %w", err)
	}
	if t.AssigneeID != nil && assigneeName != nil {
		t.Assignee = &entity.MemberSummary{ID: *t.AssigneeID, Name: *assigneeName}
	}
	return &t, nil
}

func (r *TaskRepo) ListByBoard(ctx context.Context, boardID string) ([]*entity.Task, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT t.id, t.column_id, t.board_id, t.title, t.description, t.status, t.priority,
		        t.position, t.assignee_id, t.due_date, t.created_by, t.created_at, t.updated_at,
		        m.name as assignee_name,
		        (SELECT COUNT(*) FROM comments c WHERE c.task_id = t.id AND c.deleted_at IS NULL) as comment_count
		 FROM tasks t
		 LEFT JOIN members m ON t.assignee_id = m.id AND m.deleted_at IS NULL
		 WHERE t.board_id = $1 AND t.deleted_at IS NULL
		 ORDER BY t.column_id, t.position, t.created_at`, boardID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing tasks by board: %w", err)
	}
	defer rows.Close()

	var tasks []*entity.Task
	for rows.Next() {
		var t entity.Task
		var assigneeName *string
		if err := rows.Scan(&t.ID, &t.ColumnID, &t.BoardID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.Position, &t.AssigneeID, &t.DueDate, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
			&assigneeName, &t.CommentCount,
		); err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}
		if t.AssigneeID != nil && assigneeName != nil {
			t.Assignee = &entity.MemberSummary{ID: *t.AssigneeID, Name: *assigneeName}
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

func (r *TaskRepo) ListByColumn(ctx context.Context, columnID string) ([]*entity.Task, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, column_id, board_id, title, description, status, priority,
		        position, assignee_id, due_date, created_by, created_at, updated_at
		 FROM tasks WHERE column_id = $1 AND deleted_at IS NULL ORDER BY position, created_at`, columnID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing tasks by column: %w", err)
	}
	defer rows.Close()

	var tasks []*entity.Task
	for rows.Next() {
		var t entity.Task
		if err := rows.Scan(&t.ID, &t.ColumnID, &t.BoardID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.Position, &t.AssigneeID, &t.DueDate, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

func (r *TaskRepo) Update(ctx context.Context, t *entity.Task) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE tasks SET title = $1, description = $2, priority = $3, assignee_id = $4, due_date = $5, updated_at = $6
		 WHERE id = $7 AND deleted_at IS NULL`,
		t.Title, t.Description, t.Priority, t.AssigneeID, t.DueDate, t.UpdatedAt, t.ID,
	)
	if err != nil {
		return fmt.Errorf("updating task: %w", err)
	}
	return nil
}

func (r *TaskRepo) UpdateStatus(ctx context.Context, taskID, status string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE tasks SET status = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`,
		status, taskID,
	)
	if err != nil {
		return fmt.Errorf("updating task status: %w", err)
	}
	return nil
}

func (r *TaskRepo) Move(ctx context.Context, taskID, columnID string, position int) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE tasks SET column_id = $1, position = $2, updated_at = NOW() WHERE id = $3 AND deleted_at IS NULL`,
		columnID, position, taskID,
	)
	if err != nil {
		return fmt.Errorf("moving task: %w", err)
	}
	return nil
}

func (r *TaskRepo) Delete(ctx context.Context, id string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Soft-delete comments first
	_, err = tx.Exec(ctx,
		`UPDATE comments SET deleted_at = NOW(), updated_at = NOW() WHERE task_id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("deleting task comments: %w", err)
	}

	// Soft-delete task
	_, err = tx.Exec(ctx,
		`UPDATE tasks SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *TaskRepo) ReSequencePositions(ctx context.Context, columnID string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE tasks SET position = sub.new_pos
		 FROM (
		   SELECT id, ROW_NUMBER() OVER (ORDER BY position, created_at) - 1 AS new_pos
		   FROM tasks WHERE column_id = $1 AND deleted_at IS NULL
		 ) sub
		 WHERE tasks.id = sub.id`, columnID,
	)
	if err != nil {
		return fmt.Errorf("re-sequencing positions: %w", err)
	}
	return nil
}

func (r *TaskRepo) UnassignByMember(ctx context.Context, memberID string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE tasks SET assignee_id = NULL, updated_at = NOW() WHERE assignee_id = $1 AND deleted_at IS NULL`, memberID,
	)
	if err != nil {
		return fmt.Errorf("unassigning member tasks: %w", err)
	}
	return nil
}

func (r *TaskRepo) CountByColumn(ctx context.Context, columnID string) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tasks WHERE column_id = $1 AND deleted_at IS NULL`, columnID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting tasks: %w", err)
	}
	return count, nil
}

func (r *TaskRepo) MaxPosition(ctx context.Context, columnID string) (int, error) {
	var maxPos *int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT MAX(position) FROM tasks WHERE column_id = $1 AND deleted_at IS NULL`, columnID,
	).Scan(&maxPos)
	if err != nil {
		return 0, fmt.Errorf("getting max position: %w", err)
	}
	if maxPos == nil {
		return -1, nil
	}
	return *maxPos, nil
}
