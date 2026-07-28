package model

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// CreateTaskRequest is POST /tasks.
type CreateTaskRequest struct {
	Title          string   `json:"title"`
	Description    *string  `json:"description,omitempty"`
	Priority       *string  `json:"priority,omitempty"`
	Project        *string  `json:"project,omitempty"`
	Labels         []string `json:"labels,omitempty"`
	RecurrenceRule *string  `json:"recurrence_rule,omitempty"`
	Date           string   `json:"date"`
	Source         string   `json:"source,omitempty"`
}

// UpdateTaskRequest is PATCH /tasks/:id.
type UpdateTaskRequest struct {
	Title          *string  `json:"title,omitempty"`
	Description    *string  `json:"description,omitempty"`
	Priority       *string  `json:"priority,omitempty"`
	Project        *string  `json:"project,omitempty"`
	Labels         []string `json:"labels,omitempty"`
	RecurrenceRule *string  `json:"recurrence_rule,omitempty"`
}

// UpdateTaskOccurrenceRequest is PATCH /task-occurrences/:id.
type UpdateTaskOccurrenceRequest struct {
	Completed *bool `json:"completed"`
}

// ReorderTaskOccurrencesRequest is PATCH /task-occurrences/reorder.
type ReorderTaskOccurrencesRequest struct {
	Date          string   `json:"date"`
	OccurrenceIDs []string `json:"occurrence_ids"`
}

// UpsertDailyNoteRequest is PUT/PATCH daily note.
type UpsertDailyNoteRequest struct {
	Content string `json:"content"`
}

// CarryForwardRequest is POST /tasks/carry-forward.
type CarryForwardRequest struct {
	Date string `json:"date"`
}

// TaskDayResponse is GET /tasks/day/:date.
type TaskDayResponse struct {
	Date        string                    `json:"date"`
	Note        DailyNoteResponse         `json:"note"`
	Statistics  TaskDayStatisticsResponse `json:"statistics"`
	Occurrences []TaskOccurrenceResponse  `json:"occurrences"`
}

type DailyNoteResponse struct {
	Content   string `json:"content"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type TaskDayStatisticsResponse struct {
	Total                int      `json:"total"`
	Completed            int      `json:"completed"`
	Pending              int      `json:"pending"`
	Carried              int      `json:"carried"`
	CompletionPct        float64  `json:"completion_pct"`
	CompletedToday       int      `json:"completed_today"`
	CarriedForward       int      `json:"carried_forward"`
	LongestCarriedStreak int      `json:"longest_carried_streak"`
	AverageCompletionMin *float64 `json:"average_completion_min,omitempty"`
	Streak               int      `json:"streak"`
}

type TaskOccurrenceResponse struct {
	ID             string   `json:"id"`
	PublicID       string   `json:"public_id"`
	TaskID         string   `json:"task_id"`
	Date           string   `json:"date"`
	SortOrder      int      `json:"sort_order"`
	Completed      bool     `json:"completed"`
	CompletedAt    *string  `json:"completed_at,omitempty"`
	CarriedForward bool     `json:"carried_forward"`
	Source         string   `json:"source"`
	Title          string   `json:"title"`
	Description    *string  `json:"description,omitempty"`
	Priority       *string  `json:"priority,omitempty"`
	Project        *string  `json:"project,omitempty"`
	Labels         []string `json:"labels,omitempty"`
	RecurrenceRule *string  `json:"recurrence_rule,omitempty"`
}

type TaskHistoryDayResponse struct {
	Date      string `json:"date"`
	Total     int    `json:"total"`
	Completed int    `json:"completed"`
	Pending   int    `json:"pending"`
	Carried   int    `json:"carried"`
}

type TaskResponse struct {
	ID             string   `json:"id"`
	PublicID       string   `json:"public_id"`
	Title          string   `json:"title"`
	Description    *string  `json:"description,omitempty"`
	Priority       *string  `json:"priority,omitempty"`
	Project        *string  `json:"project,omitempty"`
	Labels         []string `json:"labels,omitempty"`
	RecurrenceRule *string  `json:"recurrence_rule,omitempty"`
	UpdatedAt      string   `json:"updated_at"`
}

func TaskDayFromEntity(view entity.TaskJournalDay) TaskDayResponse {
	occurrences := make([]TaskOccurrenceResponse, 0, len(view.Occurrences))
	for _, o := range view.Occurrences {
		occurrences = append(occurrences, TaskOccurrenceFromEntity(o))
	}
	return TaskDayResponse{
		Date:        formatCivilDate(view.Date),
		Note:        DailyNoteFromEntity(view.Note),
		Statistics:  TaskStatisticsFromEntity(view.Statistics),
		Occurrences: occurrences,
	}
}

func TaskOccurrenceFromEntity(o entity.TaskOccurrenceWithTask) TaskOccurrenceResponse {
	resp := TaskOccurrenceResponse{
		ID:             o.ID.String(),
		PublicID:       o.PublicID,
		TaskID:         o.TaskID.String(),
		Date:           formatCivilDate(o.Date),
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
	}
	if o.CompletedAt != nil {
		v := o.CompletedAt.UTC().Format(time.RFC3339Nano)
		resp.CompletedAt = &v
	}
	return resp
}

func TaskFromEntity(t entity.Task) TaskResponse {
	return TaskResponse{
		ID:             t.ID.String(),
		PublicID:       t.PublicID,
		Title:          t.Title,
		Description:    t.Description,
		Priority:       t.Priority,
		Project:        t.Project,
		Labels:         t.Labels,
		RecurrenceRule: t.RecurrenceRule,
		UpdatedAt:      t.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func DailyNoteFromEntity(n entity.DailyNote) DailyNoteResponse {
	resp := DailyNoteResponse{Content: n.Content}
	if !n.UpdatedAt.IsZero() {
		resp.UpdatedAt = n.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
	return resp
}

func TaskStatisticsFromEntity(s entity.TaskDayStatistics) TaskDayStatisticsResponse {
	return TaskDayStatisticsResponse{
		Total:                s.Total,
		Completed:            s.Completed,
		Pending:              s.Pending,
		Carried:              s.Carried,
		CompletionPct:        s.CompletionPct,
		CompletedToday:       s.CompletedToday,
		CarriedForward:       s.CarriedForward,
		LongestCarriedStreak: s.LongestCarriedStreak,
		AverageCompletionMin: s.AverageCompletionMin,
		Streak:               s.Streak,
	}
}

func TaskHistoryFromSummaries(summaries []entity.TaskDaySummary) []TaskHistoryDayResponse {
	out := make([]TaskHistoryDayResponse, 0, len(summaries))
	for _, s := range summaries {
		out = append(out, TaskHistoryDayResponse{
			Date:      formatCivilDate(s.Date),
			Total:     s.Total,
			Completed: s.Completed,
			Pending:   s.Pending,
			Carried:   s.Carried,
		})
	}
	return out
}

func formatCivilDate(t time.Time) string {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}
