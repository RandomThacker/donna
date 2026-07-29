package business

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
)

var taskTagColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// CreateTaskTagInput creates a user tag.
type CreateTaskTagInput struct {
	Name  string
	Color string
}

// UpdateTaskTagInput patches a user tag.
type UpdateTaskTagInput struct {
	Name  *string
	Color *string
}

// ListTaskTags returns all tags for a user.
func (s *TaskJournalService) ListTaskTags(ctx context.Context, userID uuid.UUID) ([]entity.TaskTag, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	return s.tags.ListByUser(ctx, userID)
}

// CreateTaskTag adds a colored tag.
func (s *TaskJournalService) CreateTaskTag(ctx context.Context, userID uuid.UUID, in CreateTaskTagInput) (entity.TaskTag, error) {
	name, color, err := normalizeTaskTagInput(in.Name, in.Color)
	if err != nil {
		return entity.TaskTag{}, err
	}
	if userID == uuid.Nil {
		return entity.TaskTag{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	now := s.now().UTC()
	id, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.TaskTag{}, err
	}
	return s.tags.Create(ctx, entity.TaskTag{
		ID:        id,
		PublicID:  idgen.PublicID(constant.PublicIDPrefixTaskTag, id),
		UserID:    userID,
		Name:      name,
		Color:     color,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

// UpdateTaskTag patches a tag.
func (s *TaskJournalService) UpdateTaskTag(ctx context.Context, userID, tagID uuid.UUID, in UpdateTaskTagInput) (entity.TaskTag, error) {
	if userID == uuid.Nil || tagID == uuid.Nil {
		return entity.TaskTag{}, fmt.Errorf("%w: user and tag id are required", apperr.ErrValidation)
	}
	existing, err := s.tags.GetByID(ctx, tagID)
	if err != nil {
		return entity.TaskTag{}, err
	}
	if existing.UserID != userID {
		return entity.TaskTag{}, apperr.ErrForbidden
	}
	fields := repository.TaskTagUpdateFields{}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return entity.TaskTag{}, fmt.Errorf("%w: tag name is required", apperr.ErrValidation)
		}
		fields.Name = &name
	}
	if in.Color != nil {
		color := strings.TrimSpace(*in.Color)
		if !taskTagColorPattern.MatchString(color) {
			return entity.TaskTag{}, fmt.Errorf("%w: tag color must be #RRGGBB", apperr.ErrValidation)
		}
		fields.Color = &color
	}
	return s.tags.Update(ctx, tagID, userID, fields, s.now().UTC())
}

// DeleteTaskTag removes a tag and its assignments.
func (s *TaskJournalService) DeleteTaskTag(ctx context.Context, userID, tagID uuid.UUID) error {
	if userID == uuid.Nil || tagID == uuid.Nil {
		return fmt.Errorf("%w: user and tag id are required", apperr.ErrValidation)
	}
	existing, err := s.tags.GetByID(ctx, tagID)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return apperr.ErrForbidden
	}
	return s.tags.Delete(ctx, tagID, userID)
}

func normalizeTaskTagInput(name, color string) (string, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", "", fmt.Errorf("%w: tag name is required", apperr.ErrValidation)
	}
	color = strings.TrimSpace(color)
	if !taskTagColorPattern.MatchString(color) {
		return "", "", fmt.Errorf("%w: tag color must be #RRGGBB", apperr.ErrValidation)
	}
	return name, color, nil
}

func (s *TaskJournalService) attachTagsToOccurrences(
	ctx context.Context,
	userID uuid.UUID,
	occurrences []entity.TaskOccurrenceWithTask,
) ([]entity.TaskOccurrenceWithTask, error) {
	if len(occurrences) == 0 {
		return occurrences, nil
	}
	taskIDs := make([]uuid.UUID, len(occurrences))
	for i, o := range occurrences {
		taskIDs[i] = o.TaskID
	}
	byTask, err := s.tags.ListByTaskIDs(ctx, userID, taskIDs)
	if err != nil {
		return nil, err
	}
	for i := range occurrences {
		occurrences[i].Tags = byTask[occurrences[i].TaskID]
		if occurrences[i].Tags == nil {
			occurrences[i].Tags = []entity.TaskTag{}
		}
	}
	return occurrences, nil
}

func (s *TaskJournalService) occurrenceWithTags(
	ctx context.Context,
	userID uuid.UUID,
	occ entity.TaskOccurrenceWithTask,
) (entity.TaskOccurrenceWithTask, error) {
	list, err := s.attachTagsToOccurrences(ctx, userID, []entity.TaskOccurrenceWithTask{occ})
	if err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}
	return list[0], nil
}

func (s *TaskJournalService) validateAndAssignTags(
	ctx context.Context,
	userID uuid.UUID,
	taskID uuid.UUID,
	tagIDs []uuid.UUID,
) error {
	if len(tagIDs) == 0 {
		return s.tags.ReplaceTaskTags(ctx, taskID, nil, s.now().UTC())
	}
	seen := make(map[uuid.UUID]struct{}, len(tagIDs))
	unique := make([]uuid.UUID, 0, len(tagIDs))
	for _, id := range tagIDs {
		if id == uuid.Nil {
			return fmt.Errorf("%w: invalid tag id", apperr.ErrValidation)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		tag, err := s.tags.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if tag.UserID != userID {
			return apperr.ErrForbidden
		}
		unique = append(unique, id)
	}
	return s.tags.ReplaceTaskTags(ctx, taskID, unique, s.now().UTC())
}

func (s *TaskJournalService) tagsForTask(ctx context.Context, userID, taskID uuid.UUID) ([]entity.TaskTag, error) {
	byTask, err := s.tags.ListByTaskIDs(ctx, userID, []uuid.UUID{taskID})
	if err != nil {
		return nil, err
	}
	tags := byTask[taskID]
	if tags == nil {
		return []entity.TaskTag{}, nil
	}
	return tags, nil
}

// ListTaskTagsForTask returns tags assigned to a task.
func (s *TaskJournalService) ListTaskTagsForTask(ctx context.Context, userID, taskID uuid.UUID) ([]entity.TaskTag, error) {
	if userID == uuid.Nil || taskID == uuid.Nil {
		return nil, fmt.Errorf("%w: user and task id are required", apperr.ErrValidation)
	}
	return s.tagsForTask(ctx, userID, taskID)
}
