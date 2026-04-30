package repository

import (
	"context"

	"teamboard/internal/domain/entity"
)

type TeamRepository interface {
	Create(ctx context.Context, team *entity.Team) error
	GetByID(ctx context.Context, id string) (*entity.Team, error)
	GetByName(ctx context.Context, name string) (*entity.Team, error)
}

type MemberRepository interface {
	Create(ctx context.Context, member *entity.Member) error
	GetByID(ctx context.Context, id string) (*entity.Member, error)
	GetByAPIKeyHash(ctx context.Context, hash string) (*entity.Member, error)
	ListByTeam(ctx context.Context, teamID string) ([]*entity.Member, error)
	Update(ctx context.Context, member *entity.Member) error
	Delete(ctx context.Context, id string) error
	CountAdmins(ctx context.Context, teamID string) (int, error)
}

type BoardRepository interface {
	Create(ctx context.Context, board *entity.Board) error
	GetByID(ctx context.Context, id string) (*entity.Board, error)
	ListByTeam(ctx context.Context, teamID string) ([]*entity.Board, error)
	Update(ctx context.Context, board *entity.Board) error
	Delete(ctx context.Context, id string) error
}

type ColumnRepository interface {
	Create(ctx context.Context, column *entity.Column) error
	GetByID(ctx context.Context, id string) (*entity.Column, error)
	ListByBoard(ctx context.Context, boardID string) ([]*entity.Column, error)
	Update(ctx context.Context, column *entity.Column) error
	UpdatePositions(ctx context.Context, columns []*entity.Column) error
	ShiftPositions(ctx context.Context, boardID string, fromPosition int) error
	Delete(ctx context.Context, id string) error
	CountByBoard(ctx context.Context, boardID string) (int, error)
	CountTasks(ctx context.Context, columnID string) (int, error)
	MaxPosition(ctx context.Context, boardID string) (int, error)
}

type TaskRepository interface {
	Create(ctx context.Context, task *entity.Task) error
	GetByID(ctx context.Context, id string) (*entity.Task, error)
	ListByBoard(ctx context.Context, boardID string) ([]*entity.Task, error)
	ListByColumn(ctx context.Context, columnID string) ([]*entity.Task, error)
	Update(ctx context.Context, task *entity.Task) error
	UpdateStatus(ctx context.Context, taskID, status string) error
	Move(ctx context.Context, taskID, columnID string, position int) error
	Delete(ctx context.Context, id string) error
	ReSequencePositions(ctx context.Context, columnID string) error
	UnassignByMember(ctx context.Context, memberID string) error
	CountByColumn(ctx context.Context, columnID string) (int, error)
	MaxPosition(ctx context.Context, columnID string) (int, error)
}

type CommentRepository interface {
	Create(ctx context.Context, comment *entity.Comment) error
	GetByID(ctx context.Context, id string) (*entity.Comment, error)
	ListByTask(ctx context.Context, taskID string, page, perPage int) ([]*entity.Comment, int, error)
	Delete(ctx context.Context, id string) error
	CountByTask(ctx context.Context, taskID string) (int, error)
}
