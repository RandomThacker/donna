package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/chat"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/gin-gonic/gin"
)

// ChatHandler maps chat command HTTP endpoints to the chat executor.
type ChatHandler struct {
	executor *chat.Executor
	users    *business.UserService
	log      *logger.Logger
}

// NewChatHandler constructs a ChatHandler.
func NewChatHandler(executor *chat.Executor, users *business.UserService, log *logger.Logger) *ChatHandler {
	return &ChatHandler{executor: executor, users: users, log: log}
}

// Command handles POST /chat/command.
func (h *ChatHandler) Command(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var req chat.CommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, "message is required")
		return
	}

	tz := constant.DefaultUserTimezone
	if h.users != nil {
		if user, err := h.users.GetByID(c.Request.Context(), userID); err == nil && strings.TrimSpace(user.Timezone) != "" {
			tz = user.Timezone
		}
	}

	result := h.executor.Execute(c.Request.Context(), chat.ExecuteInput{
		UserID:   userID,
		Timezone: tz,
		Now:      time.Now(),
		Message:  msg,
	})
	response.OK(c, constant.MessageOK, result)
}
