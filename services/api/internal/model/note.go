package model

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// CreateNoteRequest is POST /notes.
type CreateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Color   string `json:"color"`
	Pinned  bool   `json:"pinned"`
}

// UpdateNoteRequest is PATCH /notes/:id.
type UpdateNoteRequest struct {
	Title    *string `json:"title"`
	Content  *string `json:"content"`
	Color    *string `json:"color"`
	Pinned   *bool   `json:"pinned"`
	Archived *bool   `json:"archived"`
}

// NoteResponse is the API shape for a note.
type NoteResponse struct {
	ID        string    `json:"id"`
	PublicID  string    `json:"public_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Color     string    `json:"color"`
	Pinned    bool      `json:"pinned"`
	Archived  bool      `json:"archived"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NoteFromEntity maps an entity note to the API response.
func NoteFromEntity(n entity.Note) NoteResponse {
	return NoteResponse{
		ID:        n.ID.String(),
		PublicID:  n.PublicID,
		Title:     n.Title,
		Content:   n.Content,
		Color:     n.Color,
		Pinned:    n.Pinned,
		Archived:  n.Archived,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

// NotesFromEntities maps a slice of notes.
func NotesFromEntities(notes []entity.Note) []NoteResponse {
	out := make([]NoteResponse, 0, len(notes))
	for _, n := range notes {
		out = append(out, NoteFromEntity(n))
	}
	return out
}
