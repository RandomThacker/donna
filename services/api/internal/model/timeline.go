package model

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// TimelineItemResponse is the unified timeline DTO for API clients.
type TimelineItemResponse struct {
	ID              string         `json:"id"`
	Source          string         `json:"source"`
	Type            string         `json:"type"`
	Status          string         `json:"status"`
	Title           string         `json:"title"`
	Description     *string        `json:"description,omitempty"`
	StartAt         string         `json:"start_at"`
	EndAt           string         `json:"end_at"`
	Timezone        string         `json:"timezone"`
	AllDay          bool           `json:"all_day"`
	Color           *string        `json:"color,omitempty"`
	ReadOnly        bool           `json:"read_only"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	IsRecurring     bool           `json:"is_recurring"`
	RecurrenceRule  *string        `json:"recurrence_rule,omitempty"`
	ParentID        *string        `json:"parent_id,omitempty"`
	OccurrenceID    string         `json:"occurrence_id"`
	OccurrenceStart *string        `json:"occurrence_start,omitempty"`
	OccurrenceEnd   *string        `json:"occurrence_end,omitempty"`
}

// TimelineResponse wraps a sorted list of timeline items.
type TimelineResponse struct {
	Items []TimelineItemResponse `json:"items"`
	From  string                 `json:"from"`
	To    string                 `json:"to"`
}

// TimelineItemFromEntity maps a domain timeline item to the API response.
func TimelineItemFromEntity(item entity.TimelineItem) TimelineItemResponse {
	resp := TimelineItemResponse{
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

// TimelineItemsFromEntities maps a slice of timeline items.
func TimelineItemsFromEntities(items []entity.TimelineItem) []TimelineItemResponse {
	out := make([]TimelineItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, TimelineItemFromEntity(item))
	}
	return out
}
