package business_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type mockUserRepo struct {
	createFn     func(ctx context.Context, user entity.User) (entity.User, error)
	getByIDFn    func(ctx context.Context, id uuid.UUID) (entity.User, error)
	getByEmailFn func(ctx context.Context, email string) (entity.User, error)
	updateFn     func(ctx context.Context, id uuid.UUID, fields repository.UserUpdateFields, updatedAt time.Time) (entity.User, error)
	softDeleteFn func(ctx context.Context, id uuid.UUID, status string, deletedAt time.Time) error
}

func (m mockUserRepo) Create(ctx context.Context, user entity.User) (entity.User, error) {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return user, nil
}

func (m mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return entity.User{}, apperr.ErrNotFound
}

func (m mockUserRepo) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return entity.User{}, apperr.ErrNotFound
}

func (m mockUserRepo) Update(ctx context.Context, id uuid.UUID, fields repository.UserUpdateFields, updatedAt time.Time) (entity.User, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, fields, updatedAt)
	}
	return entity.User{}, apperr.ErrNotFound
}

func (m mockUserRepo) SoftDelete(ctx context.Context, id uuid.UUID, status string, deletedAt time.Time) error {
	if m.softDeleteFn != nil {
		return m.softDeleteFn(ctx, id, status, deletedAt)
	}
	return apperr.ErrNotFound
}

func (m mockUserRepo) TouchLastLogin(ctx context.Context, id uuid.UUID, at time.Time) (entity.User, error) {
	if m.getByIDFn != nil {
		u, err := m.getByIDFn(ctx, id)
		if err != nil {
			return entity.User{}, err
		}
		u.LastLoginAt = &at
		return u, nil
	}
	return entity.User{ID: id, LastLoginAt: &at}, nil
}

func (m mockUserRepo) WithTx(pgx.Tx) repository.UserRepository {
	return m
}

func testIdentityLog(t *testing.T) *logger.Logger {
	t.Helper()
	return logger.NewFactory(logger.Options{
		Environment: constant.EnvDevelopment,
		Level:       "error",
		Output:      io.Discard,
	}).Module(constant.ModuleIdentity)
}

func TestCreateUserSuccess(t *testing.T) {
	var stored entity.User
	repo := mockUserRepo{
		createFn: func(_ context.Context, user entity.User) (entity.User, error) {
			stored = user
			return user, nil
		},
	}
	svc := business.NewUserService(repo, testIdentityLog(t))

	name := "Ada"
	got, err := svc.Create(context.Background(), business.CreateUserInput{
		Email:       "ada@example.com",
		DisplayName: &name,
		Timezone:    "America/New_York",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.Email != "ada@example.com" {
		t.Fatalf("email = %q", got.Email)
	}
	if got.EmailVerified {
		t.Fatal("email_verified must be false without IdP")
	}
	if got.Timezone != "America/New_York" {
		t.Fatalf("timezone = %q", got.Timezone)
	}
	if got.Status != constant.UserStatusActive {
		t.Fatalf("status = %q", got.Status)
	}
	if !strings.HasPrefix(got.PublicID, constant.PublicIDPrefixUser) {
		t.Fatalf("public_id = %q", got.PublicID)
	}
	if stored.ID == uuid.Nil || stored.PublicID == "" {
		t.Fatal("expected generated ids on create")
	}
}

func TestCreateUserRequiresEmail(t *testing.T) {
	svc := business.NewUserService(mockUserRepo{}, testIdentityLog(t))
	_, err := svc.Create(context.Background(), business.CreateUserInput{Timezone: "UTC"})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateUserConflict(t *testing.T) {
	repo := mockUserRepo{
		createFn: func(context.Context, entity.User) (entity.User, error) {
			return entity.User{}, apperr.ErrConflict
		},
	}
	svc := business.NewUserService(repo, testIdentityLog(t))
	_, err := svc.Create(context.Background(), business.CreateUserInput{
		Email:    "a@b.com",
		Timezone: "UTC",
	})
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetByID(t *testing.T) {
	id := uuid.MustParse("01900000-0000-7000-8000-000000000001")
	repo := mockUserRepo{
		getByIDFn: func(_ context.Context, got uuid.UUID) (entity.User, error) {
			if got != id {
				t.Fatalf("id = %s", got)
			}
			return entity.User{ID: id, Email: "a@b.com"}, nil
		},
	}
	svc := business.NewUserService(repo, testIdentityLog(t))
	user, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if user.Email != "a@b.com" {
		t.Fatalf("email = %q", user.Email)
	}
}

func TestGetByIDNil(t *testing.T) {
	svc := business.NewUserService(mockUserRepo{}, testIdentityLog(t))
	_, err := svc.GetByID(context.Background(), uuid.Nil)
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetByEmail(t *testing.T) {
	repo := mockUserRepo{
		getByEmailFn: func(_ context.Context, email string) (entity.User, error) {
			if email != "a@b.com" {
				t.Fatalf("email = %q", email)
			}
			return entity.User{Email: email}, nil
		},
	}
	svc := business.NewUserService(repo, testIdentityLog(t))
	_, err := svc.GetByEmail(context.Background(), "a@b.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	id := uuid.MustParse("01900000-0000-7000-8000-000000000002")
	tz := "UTC"
	repo := mockUserRepo{
		updateFn: func(_ context.Context, got uuid.UUID, fields repository.UserUpdateFields, _ time.Time) (entity.User, error) {
			if got != id {
				t.Fatalf("id = %s", got)
			}
			if fields.Timezone == nil || *fields.Timezone != "UTC" {
				t.Fatalf("timezone = %#v", fields.Timezone)
			}
			return entity.User{ID: id, Timezone: "UTC"}, nil
		},
	}
	svc := business.NewUserService(repo, testIdentityLog(t))
	user, err := svc.Update(context.Background(), id, business.UpdateUserInput{Timezone: &tz})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if user.Timezone != "UTC" {
		t.Fatalf("timezone = %q", user.Timezone)
	}
}

func TestUpdateUserEmptyBody(t *testing.T) {
	id := uuid.MustParse("01900000-0000-7000-8000-000000000003")
	svc := business.NewUserService(mockUserRepo{}, testIdentityLog(t))
	_, err := svc.Update(context.Background(), id, business.UpdateUserInput{})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}

func TestSoftDelete(t *testing.T) {
	id := uuid.MustParse("01900000-0000-7000-8000-000000000004")
	repo := mockUserRepo{
		softDeleteFn: func(_ context.Context, got uuid.UUID, status string, _ time.Time) error {
			if got != id {
				t.Fatalf("id = %s", got)
			}
			if status != constant.UserStatusPendingDeletion {
				t.Fatalf("status = %q", status)
			}
			return nil
		},
	}
	svc := business.NewUserService(repo, testIdentityLog(t))
	if err := svc.SoftDelete(context.Background(), id); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}
}

func TestSoftDeleteNotFound(t *testing.T) {
	id := uuid.MustParse("01900000-0000-7000-8000-000000000005")
	svc := business.NewUserService(mockUserRepo{}, testIdentityLog(t))
	err := svc.SoftDelete(context.Background(), id)
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}
