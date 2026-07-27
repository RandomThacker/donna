package business

import (
	"context"
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

// NoteService orchestrates Keep-style notes.
type NoteService struct {
	notes repository.NoteRepository
	now   func() time.Time
}

// NewNoteService constructs a NoteService.
func NewNoteService(notes repository.NoteRepository) *NoteService {
	return &NoteService{notes: notes, now: time.Now}
}

// CreateNoteInput creates a new note.
type CreateNoteInput struct {
	Title   string
	Content string
	Color   string
	Pinned  bool
}

// UpdateNoteInput patches an existing note.
type UpdateNoteInput struct {
	Title    *string
	Content  *string
	Color    *string
	Pinned   *bool
	Archived *bool
}

// List returns live, non-archived notes for the user.
func (s *NoteService) List(ctx context.Context, userID uuid.UUID) ([]entity.Note, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	return s.notes.ListByUser(ctx, userID)
}

// Create adds a note. Title or content must be non-empty after trim.
func (s *NoteService) Create(ctx context.Context, userID uuid.UUID, in CreateNoteInput) (entity.Note, error) {
	if userID == uuid.Nil {
		return entity.Note{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	title := strings.TrimSpace(in.Title)
	content := strings.TrimSpace(in.Content)
	if title == "" && content == "" {
		return entity.Note{}, fmt.Errorf("%w: title or content is required", apperr.ErrValidation)
	}
	color := strings.TrimSpace(in.Color)
	if color == "" {
		color = constant.NoteColorDefault
	}
	if _, ok := constant.ValidNoteColors[color]; !ok {
		return entity.Note{}, fmt.Errorf("%w: invalid color", apperr.ErrValidation)
	}
	id, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.Note{}, err
	}
	now := s.now().UTC()
	return s.notes.Create(ctx, entity.Note{
		ID:        id,
		PublicID:  idgen.PublicID(constant.PublicIDPrefixNote, id),
		UserID:    userID,
		Title:     title,
		Content:   content,
		Color:     color,
		Pinned:    in.Pinned,
		Archived:  false,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

// Update patches a note owned by the user.
func (s *NoteService) Update(ctx context.Context, userID, noteID uuid.UUID, in UpdateNoteInput) (entity.Note, error) {
	if userID == uuid.Nil || noteID == uuid.Nil {
		return entity.Note{}, fmt.Errorf("%w: user and note id are required", apperr.ErrValidation)
	}
	existing, err := s.notes.GetByID(ctx, noteID)
	if err != nil {
		return entity.Note{}, err
	}
	if existing.UserID != userID {
		return entity.Note{}, apperr.ErrForbidden
	}
	fields := repository.NoteUpdateFields{
		Title:    in.Title,
		Content:  in.Content,
		Color:    in.Color,
		Pinned:   in.Pinned,
		Archived: in.Archived,
	}
	if in.Title != nil {
		trimmed := strings.TrimSpace(*in.Title)
		fields.Title = &trimmed
	}
	if in.Content != nil {
		trimmed := strings.TrimSpace(*in.Content)
		fields.Content = &trimmed
	}
	if in.Color != nil {
		color := strings.TrimSpace(*in.Color)
		if _, ok := constant.ValidNoteColors[color]; !ok {
			return entity.Note{}, fmt.Errorf("%w: invalid color", apperr.ErrValidation)
		}
		fields.Color = &color
	}
	nextTitle := existing.Title
	nextContent := existing.Content
	if fields.Title != nil {
		nextTitle = *fields.Title
	}
	if fields.Content != nil {
		nextContent = *fields.Content
	}
	if nextTitle == "" && nextContent == "" {
		return entity.Note{}, fmt.Errorf("%w: title or content is required", apperr.ErrValidation)
	}
	return s.notes.Update(ctx, noteID, userID, fields, s.now().UTC())
}

// Delete soft-deletes a note owned by the user.
func (s *NoteService) Delete(ctx context.Context, userID, noteID uuid.UUID) error {
	if userID == uuid.Nil || noteID == uuid.Nil {
		return fmt.Errorf("%w: user and note id are required", apperr.ErrValidation)
	}
	existing, err := s.notes.GetByID(ctx, noteID)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return apperr.ErrForbidden
	}
	return s.notes.SoftDelete(ctx, noteID, userID, s.now().UTC())
}
