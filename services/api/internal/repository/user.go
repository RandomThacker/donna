package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	pgUniqueViolation = "23505"

	sqlInsertUser = `
INSERT INTO users (
	id, public_id, email, email_verified, display_name, avatar_url,
	timezone, locale, status, created_at, updated_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING
	id, public_id, email, email_verified, display_name, avatar_url,
	timezone, locale, status, last_login_at, created_at, updated_at, deleted_at`

	sqlSelectUserByID = `
SELECT
	id, public_id, email, email_verified, display_name, avatar_url,
	timezone, locale, status, last_login_at, created_at, updated_at, deleted_at
FROM users
WHERE id = $1 AND deleted_at IS NULL`

	sqlSelectUserByEmail = `
SELECT
	id, public_id, email, email_verified, display_name, avatar_url,
	timezone, locale, status, last_login_at, created_at, updated_at, deleted_at
FROM users
WHERE email = $1 AND deleted_at IS NULL`

	sqlUpdateUser = `
UPDATE users SET
	display_name = COALESCE($2, display_name),
	avatar_url   = COALESCE($3, avatar_url),
	timezone     = COALESCE($4, timezone),
	locale       = COALESCE($5, locale),
	status       = COALESCE($6, status),
	updated_at   = $7
WHERE id = $1 AND deleted_at IS NULL
RETURNING
	id, public_id, email, email_verified, display_name, avatar_url,
	timezone, locale, status, last_login_at, created_at, updated_at, deleted_at`

	sqlSoftDeleteUser = `
UPDATE users SET
	status     = $2,
	deleted_at = $3,
	updated_at = $3
WHERE id = $1 AND deleted_at IS NULL`

	sqlTouchLastLogin = `
UPDATE users SET last_login_at = $2, updated_at = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING
	id, public_id, email, email_verified, display_name, avatar_url,
	timezone, locale, status, last_login_at, created_at, updated_at, deleted_at`

	sqlListActiveUserIDs = `
SELECT id
FROM users
WHERE deleted_at IS NULL
  AND status = 'active'
ORDER BY id`
)

// UserUpdateFields holds nullable columns for a partial update.
// A nil pointer means "leave unchanged". Empty string pointers clear nullable text columns.
type UserUpdateFields struct {
	DisplayName *string
	AvatarURL   *string
	Timezone    *string
	Locale      *string
	Status      *string
}

// UserRepository persists users.
type UserRepository interface {
	Create(ctx context.Context, user entity.User) (entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.User, error)
	GetByEmail(ctx context.Context, email string) (entity.User, error)
	Update(ctx context.Context, id uuid.UUID, fields UserUpdateFields, updatedAt time.Time) (entity.User, error)
	SoftDelete(ctx context.Context, id uuid.UUID, status string, deletedAt time.Time) error
	TouchLastLogin(ctx context.Context, id uuid.UUID, at time.Time) (entity.User, error)
	ListActiveIDs(ctx context.Context) ([]uuid.UUID, error)
	WithTx(tx pgx.Tx) UserRepository
}

type userRepository struct {
	q Querier
}

// NewUserRepository constructs a UserRepository backed by pgxpool.
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{q: pool}
}

func (r *userRepository) WithTx(tx pgx.Tx) UserRepository {
	return &userRepository{q: tx}
}

func (r *userRepository) Create(ctx context.Context, user entity.User) (entity.User, error) {
	row := r.q.QueryRow(ctx, sqlInsertUser,
		user.ID,
		user.PublicID,
		user.Email,
		user.EmailVerified,
		user.DisplayName,
		user.AvatarURL,
		user.Timezone,
		user.Locale,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	created, err := scanUser(row)
	if err != nil {
		if isUniqueViolation(err) {
			return entity.User{}, apperr.ErrConflict
		}
		return entity.User{}, fmt.Errorf("insert user: %w", err)
	}
	return created, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	user, err := scanUser(r.q.QueryRow(ctx, sqlSelectUserByID, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, apperr.ErrNotFound
		}
		return entity.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	user, err := scanUser(r.q.QueryRow(ctx, sqlSelectUserByEmail, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, apperr.ErrNotFound
		}
		return entity.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *userRepository) Update(ctx context.Context, id uuid.UUID, fields UserUpdateFields, updatedAt time.Time) (entity.User, error) {
	user, err := scanUser(r.q.QueryRow(ctx, sqlUpdateUser,
		id,
		fields.DisplayName,
		fields.AvatarURL,
		fields.Timezone,
		fields.Locale,
		fields.Status,
		updatedAt,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, apperr.ErrNotFound
		}
		return entity.User{}, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

func (r *userRepository) SoftDelete(ctx context.Context, id uuid.UUID, status string, deletedAt time.Time) error {
	tag, err := r.q.Exec(ctx, sqlSoftDeleteUser, id, status, deletedAt)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func (r *userRepository) TouchLastLogin(ctx context.Context, id uuid.UUID, at time.Time) (entity.User, error) {
	user, err := scanUser(r.q.QueryRow(ctx, sqlTouchLastLogin, id, at))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, apperr.ErrNotFound
		}
		return entity.User{}, fmt.Errorf("touch last login: %w", err)
	}
	return user, nil
}

func (r *userRepository) ListActiveIDs(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := r.q.Query(ctx, sqlListActiveUserIDs)
	if err != nil {
		return nil, fmt.Errorf("list active user ids: %w", err)
	}
	defer rows.Close()
	out := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanUser(row scannable) (entity.User, error) {
	var u entity.User
	err := row.Scan(
		&u.ID,
		&u.PublicID,
		&u.Email,
		&u.EmailVerified,
		&u.DisplayName,
		&u.AvatarURL,
		&u.Timezone,
		&u.Locale,
		&u.Status,
		&u.LastLoginAt,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)
	return u, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}
