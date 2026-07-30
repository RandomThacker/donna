package business

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
)

// TaskJournalService orchestrates the daily task notebook.
type TaskJournalService struct {
	tasks       repository.TaskRepository
	occurrences repository.TaskOccurrenceRepository
	notes       repository.DailyNoteRepository
	tags        repository.TaskTagRepository
	now         func() time.Time
}

// NewTaskJournalService constructs a TaskJournalService.
func NewTaskJournalService(
	tasks repository.TaskRepository,
	occurrences repository.TaskOccurrenceRepository,
	notes repository.DailyNoteRepository,
	tags repository.TaskTagRepository,
) *TaskJournalService {
	return &TaskJournalService{
		tasks:       tasks,
		occurrences: occurrences,
		notes:       notes,
		tags:        tags,
		now:         time.Now,
	}
}

// DayView is the full journal page for one civil day.
type DayView = entity.TaskJournalDay

// DayStatistics holds analytics for one day.
type DayStatistics = entity.TaskDayStatistics

// CreateTaskInput creates a permanent task and today's occurrence.
type CreateTaskInput struct {
	Title          string
	Description    *string
	Priority       *string
	Project        *string
	Labels         []string
	TagIDs         []uuid.UUID
	RecurrenceRule *string
	Date           time.Time
	Source         string
}

// UpdateTaskInput patches permanent task fields.
type UpdateTaskInput struct {
	Title          *string
	Description    *string
	Priority       *string
	Project        *string
	Labels         []string
	TagIDs         *[]uuid.UUID
	RecurrenceRule *string
}

// ReorderOccurrencesInput sets sort_order for one day.
type ReorderOccurrencesInput struct {
	Date           time.Time
	OccurrenceIDs  []uuid.UUID
}

// GetDay returns (and materializes) a journal day.
func (s *TaskJournalService) GetDay(ctx context.Context, userID uuid.UUID, date time.Time) (DayView, error) {
	if userID == uuid.Nil {
		return DayView{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	date = civilDate(date)
	if err := s.purgeFutureCarryForwards(ctx, userID); err != nil {
		return DayView{}, err
	}
	if err := s.EnsureDay(ctx, userID, date); err != nil {
		return DayView{}, err
	}
	occurrences, err := s.occurrences.ListByUserDate(ctx, userID, date)
	if err != nil {
		return DayView{}, err
	}
	occurrences, err = s.attachTagsToOccurrences(ctx, userID, occurrences)
	if err != nil {
		return DayView{}, err
	}
	note, err := s.noteForDay(ctx, userID, date)
	if err != nil {
		return DayView{}, err
	}
	stats := computeDayStatistics(occurrences)
	return DayView{
		Date:        date,
		Note:        note,
		Statistics:  stats,
		Occurrences: occurrences,
	}, nil
}

// EnsureDay materializes incomplete tasks from the previous day onto today/past.
// Future dates stay empty until the user adds a task there.
// Carry-forward is idempotent: tasks already present on the day are skipped.
func (s *TaskJournalService) EnsureDay(ctx context.Context, userID uuid.UUID, date time.Time) error {
	date = civilDate(date)
	today := civilDate(s.now())
	if date.After(today) {
		return nil
	}
	return s.carryForwardTo(ctx, userID, date)
}

// CarryForward explicitly runs carry-forward for a target day (today or past only).
func (s *TaskJournalService) CarryForward(ctx context.Context, userID uuid.UUID, date time.Time) error {
	if userID == uuid.Nil {
		return fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	date = civilDate(date)
	today := civilDate(s.now())
	if date.After(today) {
		return fmt.Errorf("%w: cannot carry forward into a future day", apperr.ErrValidation)
	}
	return s.carryForwardTo(ctx, userID, date)
}

// purgeFutureCarryForwards removes auto-cloned backlog rows that should never
// have been materialized on days after today.
func (s *TaskJournalService) purgeFutureCarryForwards(ctx context.Context, userID uuid.UUID) error {
	_, err := s.occurrences.DeleteCarryForwardAfter(ctx, userID, civilDate(s.now()))
	return err
}

func (s *TaskJournalService) carryForwardTo(ctx context.Context, userID uuid.UUID, date time.Time) error {
	// Walk backwards up to 30 days to find the most recent day with incomplete
	// tasks. This handles gaps where the user didn't open the app for one or
	// more days.
	const maxLookback = 30
	var incomplete []entity.TaskOccurrence
	for offset := 1; offset <= maxLookback; offset++ {
		lookback := date.AddDate(0, 0, -offset)
		tasks, err := s.occurrences.ListIncompleteByUserDate(ctx, userID, lookback)
		if err != nil {
			return err
		}
		if len(tasks) > 0 {
			incomplete = tasks
			break
		}
	}
	if len(incomplete) == 0 {
		return nil
	}
	now := s.now().UTC()
	sortOrder, err := s.occurrences.MaxSortOrder(ctx, userID, date)
	if err != nil {
		return err
	}
	for _, prev := range incomplete {
		exists, err := s.occurrences.ExistsForTaskDate(ctx, prev.TaskID, date)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		sortOrder++
		id, err := idgen.NewUUIDv7()
		if err != nil {
			return err
		}
		_, err = s.occurrences.Create(ctx, entity.TaskOccurrence{
			ID:             id,
			PublicID:       idgen.PublicID(constant.PublicIDPrefixTaskOccurrence, id),
			TaskID:         prev.TaskID,
			UserID:         userID,
			Date:           date,
			SortOrder:      sortOrder,
			Completed:      false,
			CarriedForward: true,
			Source:         constant.TaskOccurrenceSourceCarryForward,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateTask adds a permanent task and an occurrence on the given day.
func (s *TaskJournalService) CreateTask(ctx context.Context, userID uuid.UUID, in CreateTaskInput) (entity.TaskOccurrenceWithTask, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return entity.TaskOccurrenceWithTask{}, fmt.Errorf("%w: title is required", apperr.ErrValidation)
	}
	if userID == uuid.Nil {
		return entity.TaskOccurrenceWithTask{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	date := civilDate(in.Date)
	source := strings.TrimSpace(in.Source)
	if source == "" {
		source = constant.TaskOccurrenceSourceManual
	}
	now := s.now().UTC()

	taskID, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}
	task, err := s.tasks.Create(ctx, entity.Task{
		ID:          taskID,
		PublicID:    idgen.PublicID(constant.PublicIDPrefixTask, taskID),
		UserID:      userID,
		Title:       title,
		Description: in.Description,
		Status:      constant.TaskStatusOpen,
		Priority:    in.Priority,
		Project:     in.Project,
		Labels:      in.Labels,
		IsBacklog:   false,
		RecurrenceRule: in.RecurrenceRule,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}

	sortOrder := 0
	if err := s.occurrences.BumpSortOrders(ctx, userID, date, 1, now); err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}
	occID, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}
	_, err = s.occurrences.Create(ctx, entity.TaskOccurrence{
		ID:        occID,
		PublicID:  idgen.PublicID(constant.PublicIDPrefixTaskOccurrence, occID),
		TaskID:    task.ID,
		UserID:    userID,
		Date:      date,
		SortOrder: sortOrder,
		Completed: false,
		Source:    source,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}
	if len(in.TagIDs) > 0 {
		if err := s.validateAndAssignTags(ctx, userID, task.ID, in.TagIDs); err != nil {
			return entity.TaskOccurrenceWithTask{}, err
		}
	}
	occ, err := s.occurrences.GetByID(ctx, occID)
	if err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}
	return s.occurrenceWithTags(ctx, userID, occ)
}

// UpdateTask patches the permanent task (does not rewrite history).
func (s *TaskJournalService) UpdateTask(ctx context.Context, userID, taskID uuid.UUID, in UpdateTaskInput) (entity.Task, error) {
	if userID == uuid.Nil || taskID == uuid.Nil {
		return entity.Task{}, fmt.Errorf("%w: user and task id are required", apperr.ErrValidation)
	}
	task, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return entity.Task{}, err
	}
	if task.UserID != userID {
		return entity.Task{}, apperr.ErrForbidden
	}
	task, err = s.tasks.Update(ctx, taskID, userID, repository.TaskUpdateFields{
		Title:          in.Title,
		Description:    in.Description,
		Priority:       in.Priority,
		Project:        in.Project,
		Labels:         in.Labels,
		RecurrenceRule: in.RecurrenceRule,
	}, s.now().UTC())
	if err != nil {
		return entity.Task{}, err
	}
	if in.TagIDs != nil {
		if err := s.validateAndAssignTags(ctx, userID, taskID, *in.TagIDs); err != nil {
			return entity.Task{}, err
		}
	}
	return task, nil
}

// UpdateOccurrence toggles completion for a journal row.
func (s *TaskJournalService) UpdateOccurrence(ctx context.Context, userID, occurrenceID uuid.UUID, completed bool) (entity.TaskOccurrenceWithTask, error) {
	if userID == uuid.Nil || occurrenceID == uuid.Nil {
		return entity.TaskOccurrenceWithTask{}, fmt.Errorf("%w: user and occurrence id are required", apperr.ErrValidation)
	}
	existing, err := s.occurrences.GetByID(ctx, occurrenceID)
	if err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}
	if existing.UserID != userID {
		return entity.TaskOccurrenceWithTask{}, apperr.ErrForbidden
	}
	now := s.now().UTC()
	var completedAt *time.Time
	if completed {
		t := now
		completedAt = &t
	}
	_, err = s.occurrences.UpdateCompletion(ctx, occurrenceID, userID, completed, completedAt, now)
	if err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}
	// Keep completed rows at the bottom of the day list.
	if completed {
		maxSort, maxErr := s.occurrences.MaxSortOrder(ctx, userID, existing.Date)
		if maxErr != nil {
			return entity.TaskOccurrenceWithTask{}, maxErr
		}
		if err := s.occurrences.UpdateSortOrder(ctx, occurrenceID, userID, maxSort+1, existing.Date, now); err != nil {
			return entity.TaskOccurrenceWithTask{}, err
		}
	}
	occ, err := s.occurrences.GetByID(ctx, occurrenceID)
	if err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}
	return s.occurrenceWithTags(ctx, userID, occ)
}

// RescheduleOccurrence moves a journal row to another civil date (prepends there).
func (s *TaskJournalService) RescheduleOccurrence(ctx context.Context, userID, occurrenceID uuid.UUID, newDate time.Time) (entity.TaskOccurrenceWithTask, error) {
	if userID == uuid.Nil || occurrenceID == uuid.Nil {
		return entity.TaskOccurrenceWithTask{}, fmt.Errorf("%w: user and occurrence id are required", apperr.ErrValidation)
	}
	existing, err := s.occurrences.GetByID(ctx, occurrenceID)
	if err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}
	if existing.UserID != userID {
		return entity.TaskOccurrenceWithTask{}, apperr.ErrForbidden
	}
	target := civilDate(newDate)
	if civilDate(existing.Date).Equal(target) {
		return s.occurrenceWithTags(ctx, userID, existing)
	}
	exists, err := s.occurrences.ExistsForTaskDate(ctx, existing.TaskID, target)
	if err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}
	if exists {
		return entity.TaskOccurrenceWithTask{}, fmt.Errorf("%w: task already exists on that day", apperr.ErrValidation)
	}
	now := s.now().UTC()
	if err := s.occurrences.BumpSortOrders(ctx, userID, target, 1, now); err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}
	if _, err := s.occurrences.UpdateDateAndSort(ctx, occurrenceID, userID, target, 0, now); err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}
	occ, err := s.occurrences.GetByID(ctx, occurrenceID)
	if err != nil {
		return entity.TaskOccurrenceWithTask{}, err
	}
	return s.occurrenceWithTags(ctx, userID, occ)
}

// DeleteTask soft-deletes the permanent task and removes all journal rows.
func (s *TaskJournalService) DeleteTask(ctx context.Context, userID, taskID uuid.UUID) error {
	if userID == uuid.Nil || taskID == uuid.Nil {
		return fmt.Errorf("%w: user and task id are required", apperr.ErrValidation)
	}
	task, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task.UserID != userID {
		return apperr.ErrForbidden
	}
	if _, err := s.occurrences.DeleteByTaskID(ctx, taskID, userID); err != nil {
		return err
	}
	return s.tasks.SoftDelete(ctx, taskID, userID, s.now().UTC())
}

// ReorderOccurrences updates sort_order for a day without touching other days.
func (s *TaskJournalService) ReorderOccurrences(ctx context.Context, userID uuid.UUID, in ReorderOccurrencesInput) error {
	if userID == uuid.Nil {
		return fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	date := civilDate(in.Date)
	now := s.now().UTC()
	for i, id := range in.OccurrenceIDs {
		occ, err := s.occurrences.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if occ.UserID != userID {
			return apperr.ErrForbidden
		}
		if !sameCivilDay(occ.Date, date) {
			return fmt.Errorf("%w: occurrence does not belong to this day", apperr.ErrValidation)
		}
		if err := s.occurrences.UpdateSortOrder(ctx, id, userID, i, date, now); err != nil {
			return err
		}
	}
	return nil
}

// GetHistory returns per-day summaries for the mini calendar.
func (s *TaskJournalService) GetHistory(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.TaskDaySummary, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if err := s.purgeFutureCarryForwards(ctx, userID); err != nil {
		return nil, err
	}
	return s.occurrences.SummariesByDateRange(ctx, userID, civilDate(from), civilDate(to))
}

// UpsertDailyNote saves markdown for a day.
func (s *TaskJournalService) UpsertDailyNote(ctx context.Context, userID uuid.UUID, date time.Time, content string) (entity.DailyNote, error) {
	if userID == uuid.Nil {
		return entity.DailyNote{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	date = civilDate(date)
	now := s.now().UTC()
	id, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.DailyNote{}, err
	}
	return s.notes.Upsert(ctx, entity.DailyNote{
		ID:        id,
		PublicID:  idgen.PublicID(constant.PublicIDPrefixDailyNote, id),
		UserID:    userID,
		Date:      date,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *TaskJournalService) noteForDay(ctx context.Context, userID uuid.UUID, date time.Time) (entity.DailyNote, error) {
	note, err := s.notes.GetByUserDate(ctx, userID, date)
	if errors.Is(err, apperr.ErrNotFound) {
		return entity.DailyNote{UserID: userID, Date: date, Content: ""}, nil
	}
	return note, err
}

func computeDayStatistics(occurrences []entity.TaskOccurrenceWithTask) entity.TaskDayStatistics {
	total := len(occurrences)
	completed := 0
	carried := 0
	var completionMinutes []float64
	for _, o := range occurrences {
		if o.Completed {
			completed++
			if o.CompletedAt != nil {
				minutes := o.CompletedAt.Sub(o.CreatedAt).Minutes()
				if minutes >= 0 {
					completionMinutes = append(completionMinutes, minutes)
				}
			}
		}
		if o.CarriedForward {
			carried++
		}
	}
	pending := total - completed
	var pct float64
	if total > 0 {
		pct = float64(completed) / float64(total) * 100
	}
	var avg *float64
	if len(completionMinutes) > 0 {
		sum := 0.0
		for _, m := range completionMinutes {
			sum += m
		}
		v := sum / float64(len(completionMinutes))
		avg = &v
	}
	streak := 0
	if total > 0 && pending == 0 {
		streak = 1
	}
	return entity.TaskDayStatistics{
		Total:                total,
		Completed:            completed,
		Pending:              pending,
		Carried:              carried,
		CompletionPct:        pct,
		CompletedToday:       completed,
		CarriedForward:       carried,
		LongestCarriedStreak: longestCarried(occurrences),
		AverageCompletionMin: avg,
		Streak:               streak,
	}
}

func longestCarried(occurrences []entity.TaskOccurrenceWithTask) int {
	max := 0
	for _, o := range occurrences {
		if o.CarriedForward {
			max++
		}
	}
	return max
}

func civilDate(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func sameCivilDay(a, b time.Time) bool {
	return civilDate(a).Equal(civilDate(b))
}

// ParseCivilDate parses YYYY-MM-DD.
func ParseCivilDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fmt.Errorf("%w: date is required", apperr.ErrValidation)
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: invalid date", apperr.ErrValidation)
	}
	return civilDate(t), nil
}
