package actions

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/automationcatalog"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

// EventService is the service port used by event actions.
type EventService interface {
	Create(ctx context.Context, userID uuid.UUID, in business.CreateDonnaEventInput) (entity.DonnaEvent, error)
	Update(ctx context.Context, userID, eventID uuid.UUID, in business.UpdateDonnaEventInput) (entity.DonnaEvent, error)
	Delete(ctx context.Context, userID, eventID uuid.UUID) error
}

// ReminderService is the service port used by reminder actions.
type ReminderService interface {
	Create(ctx context.Context, userID uuid.UUID, in business.CreateDonnaReminderInput) (entity.DonnaReminder, error)
	Update(ctx context.Context, userID, reminderID uuid.UUID, in business.UpdateDonnaReminderInput) (entity.DonnaReminder, error)
	Delete(ctx context.Context, userID, reminderID uuid.UUID) error
}

// AutomationService is the service port used by automation actions.
type AutomationService interface {
	Create(ctx context.Context, userID uuid.UUID, in business.CreateAutomationInput) (entity.Automation, error)
	Update(ctx context.Context, userID, autoID uuid.UUID, in business.UpdateAutomationInput) (entity.Automation, error)
	Delete(ctx context.Context, userID, autoID uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID) ([]entity.Automation, error)
	ListTemplates() ([]automationcatalog.Template, error)
	GetOwned(ctx context.Context, userID, autoID uuid.UUID) (entity.Automation, error)
}

// AutomationExecutionQueryService is the port for history / metrics / analytics.
type AutomationExecutionQueryService interface {
	GetExecution(ctx context.Context, userID, executionID uuid.UUID) (entity.AutomationExecution, error)
	ListHistoryForAutomation(ctx context.Context, userID, automationID uuid.UUID, limit int) ([]entity.AutomationExecution, error)
	ListHistoryForUser(ctx context.Context, userID uuid.UUID, limit int) ([]entity.AutomationExecution, error)
	MetricsForAutomations(ctx context.Context, userID uuid.UUID, automationIDs []uuid.UUID) (map[uuid.UUID]entity.AutomationRunMetrics, error)
	Analytics(ctx context.Context, userID uuid.UUID) (business.AutomationAnalytics, error)
}

// TaskService is the service port used by task actions.
type TaskService interface {
	CreateTask(ctx context.Context, userID uuid.UUID, in business.CreateTaskInput) (entity.TaskOccurrenceWithTask, error)
	UpdateTask(ctx context.Context, userID, taskID uuid.UUID, in business.UpdateTaskInput) (entity.Task, error)
	UpdateOccurrence(ctx context.Context, userID, occurrenceID uuid.UUID, completed bool) (entity.TaskOccurrenceWithTask, error)
	DeleteTask(ctx context.Context, userID, taskID uuid.UUID) error
	ListTaskTagsForTask(ctx context.Context, userID, taskID uuid.UUID) ([]entity.TaskTag, error)
}

// TimelineQueryService is the service port used by QueryTimelineAction.
type TimelineQueryService interface {
	List(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.TimelineItem, error)
}

// NotificationQueryService is the service port used by notification actions.
type NotificationQueryService interface {
	List(ctx context.Context, userID uuid.UUID, statuses []string) ([]entity.Notification, error)
	MarkRead(ctx context.Context, userID, id uuid.UUID) (entity.Notification, error)
	MarkDismissed(ctx context.Context, userID, id uuid.UUID) (entity.Notification, error)
}
