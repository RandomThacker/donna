package handler

import (
	"errors"
	"net/http"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/model"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/RandomThacker/donna/services/api/internal/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler maps Identity HTTP endpoints to the business layer.
type UserHandler struct {
	svc *business.UserService
	log *logger.Logger
}

// NewUserHandler constructs a UserHandler.
func NewUserHandler(svc *business.UserService, log *logger.Logger) *UserHandler {
	return &UserHandler{svc: svc, log: log}
}

// Create handles POST /users.
func (h *UserHandler) Create(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid request body",
			constant.ErrorCodeInvalidRequest,
			err.Error(),
		)
		return
	}

	normalized, err := validation.CreateUser(req)
	if err != nil {
		h.writeError(c, err)
		return
	}

	user, err := h.svc.Create(c.Request.Context(), business.CreateUserInput{
		Email:       normalized.Email,
		DisplayName: normalized.DisplayName,
		AvatarURL:   normalized.AvatarURL,
		Timezone:    *normalized.Timezone,
		Locale:      normalized.Locale,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, constant.MessageUserCreated, model.UserFromEntity(user))
}

// GetByID handles GET /users/:id.
func (h *UserHandler) GetByID(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}

	user, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.OK(c, constant.MessageUserFound, model.UserFromEntity(user))
}

// GetByEmail handles GET /users?email=.
func (h *UserHandler) GetByEmail(c *gin.Context) {
	email, err := validation.EmailQuery(c.Query("email"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	user, err := h.svc.GetByEmail(c.Request.Context(), email)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.OK(c, constant.MessageUserFound, model.UserFromEntity(user))
}

// Update handles PATCH /users/:id.
func (h *UserHandler) Update(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid request body",
			constant.ErrorCodeInvalidRequest,
			err.Error(),
		)
		return
	}

	normalized, err := validation.UpdateUser(req)
	if err != nil {
		h.writeError(c, err)
		return
	}

	user, err := h.svc.Update(c.Request.Context(), id, business.UpdateUserInput{
		DisplayName: normalized.DisplayName,
		AvatarURL:   normalized.AvatarURL,
		Timezone:    normalized.Timezone,
		Locale:      normalized.Locale,
		Status:      normalized.Status,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.OK(c, constant.MessageUserUpdated, model.UserFromEntity(user))
}

// SoftDelete handles DELETE /users/:id.
func (h *UserHandler) SoftDelete(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}

	if err := h.svc.SoftDelete(c.Request.Context(), id); err != nil {
		h.writeError(c, err)
		return
	}

	response.OK(c, constant.MessageUserDeleted, nil)
}

func parseUserID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid user id",
			constant.ErrorCodeInvalidRequest,
			"id must be a UUID",
		)
		return uuid.Nil, false
	}
	return id, true
}

func (h *UserHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrValidation):
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, err.Error())
	case errors.Is(err, apperr.ErrInvalid):
		response.Error(c, http.StatusBadRequest, "invalid request", constant.ErrorCodeInvalidRequest, err.Error())
	case errors.Is(err, apperr.ErrNotFound):
		response.Error(c, http.StatusNotFound, "user not found", constant.ErrorCodeNotFound, err.Error())
	case errors.Is(err, apperr.ErrConflict):
		response.Error(c, http.StatusConflict, "user already exists", constant.ErrorCodeConflict, err.Error())
	default:
		h.log.Error(c.Request.Context(), "identity request failed", constant.LogAttrError, err)
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "unexpected error")
	}
}
