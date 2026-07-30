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
	"github.com/RandomThacker/donna/services/api/internal/model"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/gin-gonic/gin"
)

// ChatHandler maps chat command HTTP endpoints to the chat executor.
type ChatHandler struct {
	executor      *chat.Executor
	conversations *business.ConversationService
	users         *business.UserService
	log           *logger.Logger
}

// NewChatHandler constructs a ChatHandler.
func NewChatHandler(
	executor *chat.Executor,
	conversations *business.ConversationService,
	users *business.UserService,
	log *logger.Logger,
) *ChatHandler {
	return &ChatHandler{executor: executor, conversations: conversations, users: users, log: log}
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

	out := model.ChatCommandResponse{
		Reply:  result.Reply,
		Intent: string(result.Intent),
	}

	if h.conversations != nil {
		var clientID *string
		if trimmed := strings.TrimSpace(req.ClientMessageID); trimmed != "" {
			clientID = &trimmed
		}
		persisted, err := h.conversations.AppendTurn(c.Request.Context(), userID, msg, result.Reply, clientID)
		if err != nil {
			if h.log != nil {
				h.log.Warn(c.Request.Context(), "chat history persist failed",
					constant.LogAttrError, err,
				)
			}
		} else {
			out.ConversationPublicID = persisted.Conversation.PublicID
			out.UserMessagePublicID = persisted.UserMessage.PublicID
			out.ReplyMessagePublicID = persisted.AssistantMessage.PublicID
		}
	}

	response.OK(c, constant.MessageOK, out)
}

// Messages handles GET /chat/messages — primary web thread history.
func (h *ChatHandler) Messages(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	if h.conversations == nil {
		response.OK(c, constant.MessageOK, model.ChatHistoryResponse{Messages: []model.ChatMessageResponse{}})
		return
	}
	history, err := h.conversations.GetPrimaryHistory(c.Request.Context(), userID)
	if err != nil {
		if h.log != nil {
			h.log.Error(c.Request.Context(), "chat history load failed", constant.LogAttrError, err)
		}
		response.Error(c, http.StatusInternalServerError, "failed to load chat history", constant.ErrorCodeInternal, err.Error())
		return
	}
	response.OK(c, constant.MessageOK, model.ChatHistoryFromEntities(history.Conversation, history.Messages))
}
