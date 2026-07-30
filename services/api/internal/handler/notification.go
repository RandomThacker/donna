package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/RandomThacker/donna/services/api/internal/actions"
	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NotificationHandler maps notification HTTP endpoints to Actions.
type NotificationHandler struct {
	list     *actions.GetNotificationsAction
	markRead *actions.MarkNotificationReadAction
	dismiss  *actions.DismissNotificationAction
	log      *logger.Logger
}

// NewNotificationHandler constructs a NotificationHandler.
func NewNotificationHandler(
	list *actions.GetNotificationsAction,
	markRead *actions.MarkNotificationReadAction,
	dismiss *actions.DismissNotificationAction,
	log *logger.Logger,
) *NotificationHandler {
	return &NotificationHandler{list: list, markRead: markRead, dismiss: dismiss, log: log}
}

// List handles GET /notifications?status=PENDING,SENT
func (h *NotificationHandler) List(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var statuses []string
	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			s := strings.TrimSpace(strings.ToUpper(part))
			if s != "" {
				statuses = append(statuses, s)
			}
		}
	}
	items, err := h.list.Execute(c.Request.Context(), actions.GetNotificationsRequest{
		UserID:   userID,
		Statuses: statuses,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{
		"notifications": notificationsFromAction(items),
	})
}

// MarkRead handles PATCH /notifications/:id/read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid notification id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	n, err := h.markRead.Execute(c.Request.Context(), actions.MarkNotificationReadRequest{
		UserID:         userID,
		NotificationID: id,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, notificationFromAction(n))
}

// MarkDismissed handles PATCH /notifications/:id/dismiss
func (h *NotificationHandler) MarkDismissed(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid notification id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	n, err := h.dismiss.Execute(c.Request.Context(), actions.DismissNotificationRequest{
		UserID:         userID,
		NotificationID: id,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, notificationFromAction(n))
}

func (h *NotificationHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not found", constant.ErrorCodeNotFound, err.Error())
	case errors.Is(err, apperr.ErrForbidden):
		response.Error(c, http.StatusForbidden, "forbidden", constant.ErrorCodeForbidden, err.Error())
	case errors.Is(err, apperr.ErrValidation):
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, err.Error())
	default:
		if h.log != nil {
			h.log.Error(c.Request.Context(), "notification request failed", constant.LogAttrError, err)
		}
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "notification failed")
	}
}
