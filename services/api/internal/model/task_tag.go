package model

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// CreateTaskTagRequest is POST /task-tags.
type CreateTaskTagRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// UpdateTaskTagRequest is PATCH /task-tags/:id.
type UpdateTaskTagRequest struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
}

// TaskTagResponse is a colored user tag.
type TaskTagResponse struct {
	ID        string `json:"id"`
	PublicID  string `json:"public_id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func TaskTagFromEntity(t entity.TaskTag) TaskTagResponse {
	resp := TaskTagResponse{
		ID:       t.ID.String(),
		PublicID: t.PublicID,
		Name:     t.Name,
		Color:    t.Color,
	}
	if !t.UpdatedAt.IsZero() {
		resp.UpdatedAt = t.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
	return resp
}

func TaskTagsFromEntities(tags []entity.TaskTag) []TaskTagResponse {
	if len(tags) == 0 {
		return []TaskTagResponse{}
	}
	out := make([]TaskTagResponse, 0, len(tags))
	for _, t := range tags {
		out = append(out, TaskTagFromEntity(t))
	}
	return out
}
