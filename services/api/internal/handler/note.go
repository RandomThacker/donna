package handler

import (
	"errors"
	"net/http"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/RandomThacker/donna/services/api/internal/model"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NoteHandler maps notes HTTP endpoints to the business layer.
type NoteHandler struct {
	svc *business.NoteService
	log *logger.Logger
}

// NewNoteHandler constructs a NoteHandler.
func NewNoteHandler(svc *business.NoteService, log *logger.Logger) *NoteHandler {
	return &NoteHandler{svc: svc, log: log}
}

// List handles GET /notes.
func (h *NoteHandler) List(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	notes, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		h.writeNoteError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"notes": model.NotesFromEntities(notes)})
}

// Create handles POST /notes.
func (h *NoteHandler) Create(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var req model.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	note, err := h.svc.Create(c.Request.Context(), userID, business.CreateNoteInput{
		Title:   req.Title,
		Content: req.Content,
		Color:   req.Color,
		Pinned:  req.Pinned,
	})
	if err != nil {
		h.writeNoteError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, constant.MessageOK, model.NoteFromEntity(note))
}

// Update handles PATCH /notes/:id.
func (h *NoteHandler) Update(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid note id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	var req model.UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	note, err := h.svc.Update(c.Request.Context(), userID, noteID, business.UpdateNoteInput{
		Title:    req.Title,
		Content:  req.Content,
		Color:    req.Color,
		Pinned:   req.Pinned,
		Archived: req.Archived,
	})
	if err != nil {
		h.writeNoteError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, model.NoteFromEntity(note))
}

// Delete handles DELETE /notes/:id.
func (h *NoteHandler) Delete(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid note id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	if err := h.svc.Delete(c.Request.Context(), userID, noteID); err != nil {
		h.writeNoteError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"deleted": true})
}

func (h *NoteHandler) writeNoteError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not found", constant.ErrorCodeNotFound, err.Error())
	case errors.Is(err, apperr.ErrForbidden):
		response.Error(c, http.StatusForbidden, "forbidden", constant.ErrorCodeForbidden, err.Error())
	case errors.Is(err, apperr.ErrValidation):
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeInvalidRequest, err.Error())
	default:
		h.log.Error(c.Request.Context(), "notes error", constant.LogAttrError, err)
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "notes failed")
	}
}
