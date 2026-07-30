package handler

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/actions"
	"github.com/RandomThacker/donna/services/api/internal/model"
)

func donnaEventFromAction(e actions.EventResult) model.DonnaEventResponse {
	return model.DonnaEventResponse{
		ID:                    e.ID.String(),
		PublicID:              e.PublicID,
		Title:                 e.Title,
		Description:           e.Description,
		StartAt:               e.StartAt.UTC().Format(time.RFC3339Nano),
		EndAt:                 e.EndAt.UTC().Format(time.RFC3339Nano),
		Timezone:              e.Timezone,
		AllDay:                e.AllDay,
		Location:              e.Location,
		ReminderOffsetMinutes: e.ReminderOffsetMinutes,
		RecurrenceRule:        e.RecurrenceRule,
		Status:                e.Status,
		Color:                 e.Color,
		CreatedAt:             e.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:             e.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func donnaReminderFromAction(r actions.ReminderResult) model.DonnaReminderResponse {
	return model.DonnaReminderResponse{
		ID:             r.ID.String(),
		PublicID:       r.PublicID,
		Title:          r.Title,
		Description:    r.Description,
		TriggerAt:      r.TriggerAt.UTC().Format(time.RFC3339Nano),
		Timezone:       r.Timezone,
		RecurrenceRule: r.RecurrenceRule,
		Status:         r.Status,
		Color:          r.Color,
		CreatedAt:      r.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:      r.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func taskOccurrenceFromAction(o actions.TaskOccurrenceResult) model.TaskOccurrenceResponse {
	resp := model.TaskOccurrenceResponse{
		ID:             o.ID.String(),
		PublicID:       o.PublicID,
		TaskID:         o.TaskID.String(),
		Date:           formatCivilDateHTTP(o.Date),
		SortOrder:      o.SortOrder,
		Completed:      o.Completed,
		CarriedForward: o.CarriedForward,
		Source:         o.Source,
		Title:          o.Title,
		Description:    o.Description,
		Priority:       o.Priority,
		Project:        o.Project,
		Labels:         o.Labels,
		RecurrenceRule: o.RecurrenceRule,
		Tags:           taskTagsFromAction(o.Tags),
	}
	if o.CompletedAt != nil {
		v := o.CompletedAt.UTC().Format(time.RFC3339Nano)
		resp.CompletedAt = &v
	}
	return resp
}

func taskFromAction(t actions.TaskResult) model.TaskResponse {
	return model.TaskResponse{
		ID:             t.ID.String(),
		PublicID:       t.PublicID,
		Title:          t.Title,
		Description:    t.Description,
		Priority:       t.Priority,
		Project:        t.Project,
		Labels:         t.Labels,
		Tags:           taskTagsFromAction(t.Tags),
		RecurrenceRule: t.RecurrenceRule,
		UpdatedAt:      t.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func timelineFromAction(r actions.TimelineResult) model.TimelineResponse {
	items := make([]model.TimelineItemResponse, 0, len(r.Items))
	for _, item := range r.Items {
		items = append(items, timelineItemFromAction(item))
	}
	return model.TimelineResponse{
		Items: items,
		From:  r.From.UTC().Format(time.RFC3339Nano),
		To:    r.To.UTC().Format(time.RFC3339Nano),
	}
}

func notificationFromAction(n actions.NotificationResult) model.NotificationResponse {
	resp := model.NotificationResponse{
		ID:                    n.ID.String(),
		PublicID:              n.PublicID,
		TimelineItemParentID:  n.TimelineItemParentID,
		OccurrenceID:          n.OccurrenceID,
		Title:                 n.Title,
		Body:                  n.Body,
		NotificationType:      n.NotificationType,
		Status:                n.Status,
		DeliveryChannels:      n.DeliveryChannels,
		ChannelDeliveryStatus: n.ChannelDeliveryStatus,
		Payload:               n.Payload,
		CreatedAt:             n.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:             n.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if n.DeliveryChannels == nil {
		resp.DeliveryChannels = []string{}
	}
	if n.ScheduledFor != nil {
		v := n.ScheduledFor.UTC().Format(time.RFC3339Nano)
		resp.ScheduledFor = &v
	}
	if n.ReadAt != nil {
		v := n.ReadAt.UTC().Format(time.RFC3339Nano)
		resp.ReadAt = &v
	}
	if n.DismissedAt != nil {
		v := n.DismissedAt.UTC().Format(time.RFC3339Nano)
		resp.DismissedAt = &v
	}
	if n.SentAt != nil {
		v := n.SentAt.UTC().Format(time.RFC3339Nano)
		resp.SentAt = &v
	}
	return resp
}

func notificationsFromAction(items []actions.NotificationResult) []model.NotificationResponse {
	out := make([]model.NotificationResponse, 0, len(items))
	for _, n := range items {
		out = append(out, notificationFromAction(n))
	}
	return out
}

func taskTagsFromAction(tags []actions.TaskTagResult) []model.TaskTagResponse {
	out := make([]model.TaskTagResponse, 0, len(tags))
	for _, t := range tags {
		out = append(out, model.TaskTagResponse{
			ID:       t.ID.String(),
			PublicID: t.PublicID,
			Name:     t.Name,
			Color:    t.Color,
		})
	}
	return out
}

func timelineItemFromAction(item actions.TimelineItemResult) model.TimelineItemResponse {
	resp := model.TimelineItemResponse{
		ID:             item.ID,
		Source:         item.Source,
		Type:           item.Type,
		Status:         item.Status,
		Title:          item.Title,
		Description:    item.Description,
		StartAt:        item.StartAt.UTC().Format(time.RFC3339Nano),
		EndAt:          item.EndAt.UTC().Format(time.RFC3339Nano),
		Timezone:       item.Timezone,
		AllDay:         item.AllDay,
		Color:          item.Color,
		ReadOnly:       item.ReadOnly,
		Metadata:       item.Metadata,
		IsRecurring:    item.IsRecurring,
		RecurrenceRule: item.RecurrenceRule,
		ParentID:       item.ParentID,
		OccurrenceID:   item.OccurrenceID,
	}
	if item.OccurrenceStart != nil {
		v := item.OccurrenceStart.UTC().Format(time.RFC3339Nano)
		resp.OccurrenceStart = &v
	}
	if item.OccurrenceEnd != nil {
		v := item.OccurrenceEnd.UTC().Format(time.RFC3339Nano)
		resp.OccurrenceEnd = &v
	}
	return resp
}

func formatCivilDateHTTP(t time.Time) string {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}
