package postgres

import (
	"context"
	"fmt"

	"teamboard/internal/domain/entity"
)

type MemberRepo struct {
	db *DB
}

func NewMemberRepo(db *DB) *MemberRepo {
	return &MemberRepo{db: db}
}

func (r *MemberRepo) Create(ctx context.Context, m *entity.Member) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO members (id, team_id, name, email, role, api_key_hash, api_key_prefix, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		m.ID, m.TeamID, m.Name, m.Email, m.Role, m.APIKeyHash, m.APIKeyPrefix, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting member: %w", err)
	}
	return nil
}

func (r *MemberRepo) GetByID(ctx context.Context, id string) (*entity.Member, error) {
	var m entity.Member
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, team_id, name, email, role, api_key_hash, api_key_prefix, created_at, updated_at
		 FROM members WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&m.ID, &m.TeamID, &m.Name, &m.Email, &m.Role, &m.APIKeyHash, &m.APIKeyPrefix, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting member: %w", err)
	}
	return &m, nil
}

func (r *MemberRepo) GetByAPIKeyHash(ctx context.Context, hash string) (*entity.Member, error) {
	var m entity.Member
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, team_id, name, email, role, api_key_hash, api_key_prefix, created_at, updated_at
		 FROM members WHERE api_key_hash = $1 AND deleted_at IS NULL`, hash,
	).Scan(&m.ID, &m.TeamID, &m.Name, &m.Email, &m.Role, &m.APIKeyHash, &m.APIKeyPrefix, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting member by api key hash: %w", err)
	}
	return &m, nil
}

func (r *MemberRepo) ListByTeam(ctx context.Context, teamID string) ([]*entity.Member, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, team_id, name, email, role, created_at, updated_at
		 FROM members WHERE team_id = $1 AND deleted_at IS NULL ORDER BY created_at`, teamID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}
	defer rows.Close()

	var members []*entity.Member
	for rows.Next() {
		var m entity.Member
		if err := rows.Scan(&m.ID, &m.TeamID, &m.Name, &m.Email, &m.Role, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning member: %w", err)
		}
		members = append(members, &m)
	}
	return members, nil
}

func (r *MemberRepo) Update(ctx context.Context, m *entity.Member) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE members SET name = $1, email = $2, role = $3, api_key_hash = $4, api_key_prefix = $5, updated_at = $6
		 WHERE id = $7 AND deleted_at IS NULL`,
		m.Name, m.Email, m.Role, m.APIKeyHash, m.APIKeyPrefix, m.UpdatedAt, m.ID,
	)
	if err != nil {
		return fmt.Errorf("updating member: %w", err)
	}
	return nil
}

func (r *MemberRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE members SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("deleting member: %w", err)
	}
	return nil
}

func (r *MemberRepo) CountAdmins(ctx context.Context, teamID string) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM members WHERE team_id = $1 AND role = 'admin' AND deleted_at IS NULL`, teamID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting admins: %w", err)
	}
	return count, nil
}
