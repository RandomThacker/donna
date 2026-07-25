package business

import (
	"context"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
)

// UserService orchestrates Identity user operations.
type UserService struct {
	repo repository.UserRepository
	log  *logger.Logger
	now  func() time.Time
}

// NewUserService constructs a UserService.
func NewUserService(repo repository.UserRepository, log *logger.Logger) *UserService {
	return &UserService{
		repo: repo,
		log:  log,
		now:  time.Now,
	}
}

// Create registers a new Donna user.
// email_verified is always false until an IdP verifies it (OAuth not in this module).
func (s *UserService) Create(ctx context.Context, in CreateUserInput) (entity.User, error) {
	if in.Email == "" {
		return entity.User{}, fmt.Errorf("%w: email is required", apperr.ErrValidation)
	}
	if in.Timezone == "" {
		return entity.User{}, fmt.Errorf("%w: timezone is required", apperr.ErrValidation)
	}

	id, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.User{}, fmt.Errorf("generate user id: %w", err)
	}

	now := s.now().UTC()
	user := entity.User{
		ID:            id,
		PublicID:      idgen.PublicID(constant.PublicIDPrefixUser, id),
		Email:         in.Email,
		EmailVerified: false,
		DisplayName:   in.DisplayName,
		AvatarURL:     in.AvatarURL,
		Timezone:      in.Timezone,
		Locale:        in.Locale,
		Status:        constant.UserStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		return entity.User{}, err
	}

	s.log.IdentityEvent(ctx, logger.IdentityEventUserCreated,
		constant.LogAttrUserID, created.ID.String(),
	)
	return created, nil
}

// GetByID returns a live user by internal UUID.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	if id == uuid.Nil {
		return entity.User{}, fmt.Errorf("%w: id is required", apperr.ErrInvalid)
	}
	return s.repo.GetByID(ctx, id)
}

// GetByEmail returns a live user by email.
func (s *UserService) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	if email == "" {
		return entity.User{}, fmt.Errorf("%w: email is required", apperr.ErrValidation)
	}
	return s.repo.GetByEmail(ctx, email)
}

// Update applies a partial update to a live user.
func (s *UserService) Update(ctx context.Context, id uuid.UUID, in UpdateUserInput) (entity.User, error) {
	if id == uuid.Nil {
		return entity.User{}, fmt.Errorf("%w: id is required", apperr.ErrInvalid)
	}
	if in.DisplayName == nil && in.AvatarURL == nil && in.Timezone == nil &&
		in.Locale == nil && in.Status == nil {
		return entity.User{}, fmt.Errorf("%w: at least one field is required", apperr.ErrValidation)
	}

	updated, err := s.repo.Update(ctx, id, repository.UserUpdateFields{
		DisplayName: in.DisplayName,
		AvatarURL:   in.AvatarURL,
		Timezone:    in.Timezone,
		Locale:      in.Locale,
		Status:      in.Status,
	}, s.now().UTC())
	if err != nil {
		return entity.User{}, err
	}

	s.log.IdentityEvent(ctx, logger.IdentityEventUserUpdated,
		constant.LogAttrUserID, updated.ID.String(),
	)
	return updated, nil
}

// SoftDelete marks a user as pending deletion and sets deleted_at.
func (s *UserService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: id is required", apperr.ErrInvalid)
	}

	if err := s.repo.SoftDelete(ctx, id, constant.UserStatusPendingDeletion, s.now().UTC()); err != nil {
		return err
	}

	s.log.IdentityEvent(ctx, logger.IdentityEventUserDeleted,
		constant.LogAttrUserID, id.String(),
	)
	return nil
}
