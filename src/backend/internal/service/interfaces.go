package service

import (
	"context"

	"teamboard/internal/dto/request"
	"teamboard/internal/dto/response"
)

// Authenticator defines the interface for authentication operations.
type Authenticator interface {
	Authenticate(ctx context.Context, apiKey string) (*AuthContext, error)
}

// TeamManager defines the interface for team operations.
type TeamManager interface {
	Register(ctx context.Context, req *request.RegisterTeamRequest) (*response.RegisterTeamResponse, error)
}

// MemberManager defines the interface for member operations.
type MemberManager interface {
	ListByTeam(ctx context.Context, teamID string) (*response.MemberListResponse, error)
	GetCurrentMember(ctx context.Context, memberID string) (*response.MemberResponse, error)
	Create(ctx context.Context, teamID string, req *request.CreateMemberRequest) (*response.MemberWithKeyResponse, error)
	Update(ctx context.Context, teamID, memberID string, req *request.UpdateMemberRequest) (*response.MemberResponse, error)
	Delete(ctx context.Context, teamID, memberID string) error
	RegenerateAPIKey(ctx context.Context, teamID, memberID string) (*response.APIKeyResponse, error)
	UpdateMe(ctx context.Context, memberID string, req *request.UpdateMeRequest) (*response.MemberResponse, error)
}

// BoardManager defines the interface for board and column operations.
type BoardManager interface {
	Create(ctx context.Context, teamID string, req *request.CreateBoardRequest) (*response.BoardDetailResponse, error)
	ListByTeam(ctx context.Context, teamID string) (*response.BoardListResponse, error)
	GetByID(ctx context.Context, teamID, boardID string) (*response.BoardDetailResponse, error)
	Update(ctx context.Context, teamID, boardID string, req *request.UpdateBoardRequest) (*response.BoardResponse, error)
	Delete(ctx context.Context, teamID, boardID string) error
	AddColumn(ctx context.Context, teamID, boardID string, req *request.CreateColumnRequest) (*response.ColumnResponse, error)
	RenameColumn(ctx context.Context, teamID, columnID string, req *request.UpdateColumnRequest) (*response.ColumnResponse, error)
	ReorderColumn(ctx context.Context, teamID, columnID string, req *request.ReorderColumnRequest) (*response.ColumnListResponse, error)
	DeleteColumn(ctx context.Context, teamID, columnID string) error
}

// TaskManager defines the interface for task operations.
type TaskManager interface {
	Create(ctx context.Context, teamID, memberID string, req *request.CreateTaskRequest) (*response.TaskResponse, error)
	GetByID(ctx context.Context, teamID, taskID string) (*response.TaskDetailResponse, error)
	Update(ctx context.Context, teamID, taskID string, req *request.UpdateTaskRequest) (*response.TaskResponse, error)
	Move(ctx context.Context, teamID, taskID string, req *request.MoveTaskRequest) (*response.TaskResponse, error)
	Delete(ctx context.Context, teamID, taskID string) error
}

// CommentManager defines the interface for comment operations.
type CommentManager interface {
	Create(ctx context.Context, teamID, taskID, authorID string, req *request.CreateCommentRequest) (*response.CommentResponse, error)
	ListByTask(ctx context.Context, teamID, taskID string, page, perPage int) (*response.CommentListResponse, error)
	Delete(ctx context.Context, teamID, commentID, memberID string, isAdmin bool) error
}
