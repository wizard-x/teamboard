package repository

import (
	"testing"
)

// Test that repository interfaces exist and have the right method signatures
// by attempting to assign nil implementations to them at compile time.
func TestTeamRepositoryInterface(t *testing.T) {
	var _ TeamRepository = (TeamRepository)(nil)
}

func TestMemberRepositoryInterface(t *testing.T) {
	var _ MemberRepository = (MemberRepository)(nil)
}

func TestBoardRepositoryInterface(t *testing.T) {
	var _ BoardRepository = (BoardRepository)(nil)
}

func TestColumnRepositoryInterface(t *testing.T) {
	var _ ColumnRepository = (ColumnRepository)(nil)
}

func TestTaskRepositoryInterface(t *testing.T) {
	var _ TaskRepository = (TaskRepository)(nil)
}

func TestCommentRepositoryInterface(t *testing.T) {
	var _ CommentRepository = (CommentRepository)(nil)
}
