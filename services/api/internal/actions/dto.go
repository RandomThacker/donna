package actions

import (
	"encoding/json"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

// EventResult is the domain DTO returned by event actions.
type EventResult struct {
	ID                    uuid.UUID
	PublicID              string
	UserID                uuid.UUID
	Title                 string
	Description           *string
	StartAt               time.Time
	EndAt                 time.Time
	Timezone              string
	AllDay                bool
	Location              *string
	ReminderOffsetMinutes *int
	RecurrenceRule        *string
	Status                string
	Color                 *string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// ReminderResult is the domain DTO returned by reminder actions.
type ReminderResult struct {
	ID             uuid.UUID
	PublicID       string
	UserID         uuid.UUID
	Title          string
	Description    *string
	TriggerAt      time.Time
	Timezone       string
	RecurrenceRule *string
	Status         string
	Color          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TaskTagResult is a tag attached to a task.
type TaskTagResult struct {
	ID       uuid.UUID
	PublicID string
	Name     string
	Color    string
}

// TaskResult is the domain DTO for a permanent task (with optional tags).
type TaskResult struct {
	ID             uuid.UUID
	PublicID       string
	UserID         uuid.UUID
	Title          string
	Description    *string
	Priority       *string
	Project        *string
	Labels         []string
	RecurrenceRule *string
	Tags           []TaskTagResult
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TaskOccurrenceResult is a day occurrence with denormalized task fields.
type TaskOccurrenceResult struct {
	ID             uuid.UUID
	PublicID       string
	UserID         uuid.UUID
	TaskID         uuid.UUID
	Date           time.Time
	Completed      bool
	CompletedAt    *time.Time
	SortOrder      int
	CarriedForward bool
	Source         string
	Title          string
	Description    *string
	Priority       *string
	Project        *string
	Labels         []string
	RecurrenceRule *string
	Tags           []TaskTagResult
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TimelineItemResult is one unified timeline entry.
type TimelineItemResult struct {
	ID              string
	Source          string
	Type            string
	Status          string
	Title           string
	Description     *string
	StartAt         time.Time
	EndAt           time.Time
	Timezone        string
	AllDay          bool
	Color           *string
	ReadOnly        bool
	Metadata        map[string]any
	IsRecurring     bool
	RecurrenceRule  *string
	ParentID        *string
	OccurrenceID    string
	OccurrenceStart *time.Time
	OccurrenceEnd   *time.Time
}

// TimelineResult is a queried window of timeline items.
type TimelineResult struct {
	Items []TimelineItemResult
	From  time.Time
	To    time.Time
}

// NotificationResult is a queued notification DTO.
type NotificationResult struct {
	ID                   uuid.UUID
	PublicID             string
	UserID               uuid.UUID
	TimelineItemParentID *string
	OccurrenceID         *string
	Title                string
	Body                 string
	NotificationType     *string
	ScheduledFor         *time.Time
	Status               string
	DeliveryChannels     []string
	ChannelDeliveryStatus json.RawMessage
	Payload              json.RawMessage
	ReadAt               *time.Time
	DismissedAt          *time.Time
	SentAt               *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func eventFromEntity(e entity.DonnaEvent) EventResult {
	return EventResult{
		ID: e.ID, PublicID: e.PublicID, UserID: e.UserID,
		Title: e.Title, Description: e.Description,
		StartAt: e.StartAt, EndAt: e.EndAt, Timezone: e.Timezone, AllDay: e.AllDay,
		Location: e.Location, ReminderOffsetMinutes: e.ReminderOffsetMinutes,
		RecurrenceRule: e.RecurrenceRule, Status: e.Status, Color: e.Color,
		CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}

func reminderFromEntity(e entity.DonnaReminder) ReminderResult {
	return ReminderResult{
		ID: e.ID, PublicID: e.PublicID, UserID: e.UserID,
		Title: e.Title, Description: e.Description, TriggerAt: e.TriggerAt,
		Timezone: e.Timezone, RecurrenceRule: e.RecurrenceRule, Status: e.Status,
		Color: e.Color, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}

func taskTagFromEntity(t entity.TaskTag) TaskTagResult {
	return TaskTagResult{ID: t.ID, PublicID: t.PublicID, Name: t.Name, Color: t.Color}
}

func taskFromEntity(t entity.Task, tags []entity.TaskTag) TaskResult {
	out := TaskResult{
		ID: t.ID, PublicID: t.PublicID, UserID: t.UserID,
		Title: t.Title, Description: t.Description, Priority: t.Priority,
		Project: t.Project, Labels: t.Labels, RecurrenceRule: t.RecurrenceRule,
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
	if out.Labels == nil {
		out.Labels = []string{}
	}
	out.Tags = make([]TaskTagResult, 0, len(tags))
	for _, tag := range tags {
		out.Tags = append(out.Tags, taskTagFromEntity(tag))
	}
	return out
}

func taskOccurrenceFromEntity(o entity.TaskOccurrenceWithTask) TaskOccurrenceResult {
	labels := o.Labels
	if labels == nil {
		labels = []string{}
	}
	tags := make([]TaskTagResult, 0, len(o.Tags))
	for _, tag := range o.Tags {
		tags = append(tags, taskTagFromEntity(tag))
	}
	return TaskOccurrenceResult{
		ID: o.ID, PublicID: o.PublicID, UserID: o.UserID, TaskID: o.TaskID,
		Date: o.Date, Completed: o.Completed, CompletedAt: o.CompletedAt,
		SortOrder: o.SortOrder, CarriedForward: o.CarriedForward, Source: o.Source,
		Title: o.Title, Description: o.Description, Priority: o.Priority,
		Project: o.Project, Labels: labels, RecurrenceRule: o.RecurrenceRule,
		Tags: tags, CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt,
	}
}

func timelineItemFromEntity(item entity.TimelineItem) TimelineItemResult {
	return TimelineItemResult{
		ID: item.ID, Source: item.Source, Type: item.Type, Status: item.Status,
		Title: item.Title, Description: item.Description,
		StartAt: item.StartAt, EndAt: item.EndAt, Timezone: item.Timezone,
		AllDay: item.AllDay, Color: item.Color, ReadOnly: item.ReadOnly,
		Metadata: item.Metadata, IsRecurring: item.IsRecurring,
		RecurrenceRule: item.RecurrenceRule, ParentID: item.ParentID,
		OccurrenceID: item.OccurrenceID, OccurrenceStart: item.OccurrenceStart,
		OccurrenceEnd: item.OccurrenceEnd,
	}
}

func notificationFromEntity(n entity.Notification) NotificationResult {
	channels := n.DeliveryChannels
	if channels == nil {
		channels = []string{}
	}
	return NotificationResult{
		ID: n.ID, PublicID: n.PublicID, UserID: n.UserID,
		TimelineItemParentID: n.TimelineItemParentID, OccurrenceID: n.OccurrenceID,
		Title: n.Title, Body: n.Body, NotificationType: n.NotificationType,
		ScheduledFor: n.ScheduledFor, Status: n.Status, DeliveryChannels: channels,
		ChannelDeliveryStatus: n.ChannelDeliveryStatus, Payload: n.Payload,
		ReadAt: n.ReadAt, DismissedAt: n.DismissedAt, SentAt: n.SentAt,
		CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt,
	}
}
