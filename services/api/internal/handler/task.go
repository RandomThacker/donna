package handler

import (
	"errors"
	"net/http"

	"github.com/RandomThacker/donna/services/api/internal/actions"
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

// TaskHandler maps task journal HTTP endpoints to Actions / services.
type TaskHandler struct {
	svc          *business.TaskJournalService
	createTask   *actions.CreateTaskAction
	updateTask   *actions.UpdateTaskAction
	completeTask *actions.CompleteTaskAction
	deleteTask   *actions.DeleteTaskAction
	log          *logger.Logger
}

// NewTaskHandler constructs a TaskHandler.
func NewTaskHandler(
	svc *business.TaskJournalService,
	createTask *actions.CreateTaskAction,
	updateTask *actions.UpdateTaskAction,
	completeTask *actions.CompleteTaskAction,
	deleteTask *actions.DeleteTaskAction,
	log *logger.Logger,
) *TaskHandler {
	return &TaskHandler{
		svc:          svc,
		createTask:   createTask,
		updateTask:   updateTask,
		completeTask: completeTask,
		deleteTask:   deleteTask,
		log:          log,
	}
}

// GetDay handles GET /tasks/day/:date.
func (h *TaskHandler) GetDay(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	date, err := business.ParseCivilDate(c.Param("date"))
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	view, err := h.svc.GetDay(c.Request.Context(), userID, date)
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, model.TaskDayFromEntity(view))
}

// CreateTask handles POST /tasks.
func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var req model.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	date, err := business.ParseCivilDate(req.Date)
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	occ, err := h.createTask.Execute(c.Request.Context(), actions.CreateTaskRequest{
		UserID:         userID,
		Title:          req.Title,
		Description:    req.Description,
		Priority:       req.Priority,
		Project:        req.Project,
		Labels:         req.Labels,
		TagIDs:         parseUUIDList(req.TagIDs),
		RecurrenceRule: req.RecurrenceRule,
		Date:           date,
		Source:         req.Source,
	})
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, constant.MessageOK, taskOccurrenceFromAction(occ))
}

// UpdateTask handles PATCH /tasks/:id.
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid task id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	var req model.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	update := actions.UpdateTaskRequest{
		UserID:         userID,
		TaskID:         taskID,
		Title:          req.Title,
		Description:    req.Description,
		Priority:       req.Priority,
		Project:        req.Project,
		Labels:         req.Labels,
		RecurrenceRule: req.RecurrenceRule,
	}
	if req.TagIDs != nil {
		ids := parseUUIDList(*req.TagIDs)
		update.TagIDs = &ids
	}
	task, err := h.updateTask.Execute(c.Request.Context(), update)
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, taskFromAction(task))
}

// DeleteTask handles DELETE /tasks/:id.
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid task id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	if err := h.deleteTask.Execute(c.Request.Context(), actions.DeleteTaskRequest{
		UserID: userID,
		TaskID: taskID,
	}); err != nil {
		h.writeTaskError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"deleted": true})
}

// UpdateOccurrence handles PATCH /task-occurrences/:id.
func (h *TaskHandler) UpdateOccurrence(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	occID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid occurrence id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	var req model.UpdateTaskOccurrenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	if req.Completed == nil && req.Date == nil {
		response.Error(c, http.StatusBadRequest, "completed or date is required", constant.ErrorCodeInvalidRequest, "missing fields")
		return
	}

	var body model.TaskOccurrenceResponse
	if req.Date != nil {
		date, parseErr := business.ParseCivilDate(*req.Date)
		if parseErr != nil {
			h.writeTaskError(c, parseErr)
			return
		}
		moved, moveErr := h.svc.RescheduleOccurrence(c.Request.Context(), userID, occID, date)
		if moveErr != nil {
			h.writeTaskError(c, moveErr)
			return
		}
		body = model.TaskOccurrenceFromEntity(moved)
	}

	if req.Completed != nil {
		updated, completeErr := h.completeTask.Execute(c.Request.Context(), actions.CompleteTaskRequest{
			UserID:       userID,
			OccurrenceID: occID,
			Completed:    *req.Completed,
		})
		if completeErr != nil {
			h.writeTaskError(c, completeErr)
			return
		}
		body = taskOccurrenceFromAction(updated)
	}

	response.OK(c, constant.MessageOK, body)
}

// ReorderOccurrences handles PATCH /task-occurrences/reorder.
func (h *TaskHandler) ReorderOccurrences(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var req model.ReorderTaskOccurrencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	date, err := business.ParseCivilDate(req.Date)
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	ids := make([]uuid.UUID, 0, len(req.OccurrenceIDs))
	for _, raw := range req.OccurrenceIDs {
		id, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			response.Error(c, http.StatusBadRequest, "invalid occurrence id", constant.ErrorCodeInvalidRequest, parseErr.Error())
			return
		}
		ids = append(ids, id)
	}
	if err := h.svc.ReorderOccurrences(c.Request.Context(), userID, business.ReorderOccurrencesInput{
		Date:          date,
		OccurrenceIDs: ids,
	}); err != nil {
		h.writeTaskError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"reordered": len(ids)})
}

// GetHistory handles GET /tasks/history.
func (h *TaskHandler) GetHistory(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	from, err := business.ParseCivilDate(c.Query("from"))
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	to, err := business.ParseCivilDate(c.Query("to"))
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	summaries, err := h.svc.GetHistory(c.Request.Context(), userID, from, to)
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"days": model.TaskHistoryFromSummaries(summaries)})
}

// CarryForward handles POST /tasks/carry-forward.
func (h *TaskHandler) CarryForward(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var req model.CarryForwardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	date, err := business.ParseCivilDate(req.Date)
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	if err := h.svc.CarryForward(c.Request.Context(), userID, date); err != nil {
		h.writeTaskError(c, err)
		return
	}
	view, err := h.svc.GetDay(c.Request.Context(), userID, date)
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, model.TaskDayFromEntity(view))
}

// UpsertDailyNote handles PUT /daily-notes/:date.
func (h *TaskHandler) UpsertDailyNote(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	date, err := business.ParseCivilDate(c.Param("date"))
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	var req model.UpsertDailyNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	note, err := h.svc.UpsertDailyNote(c.Request.Context(), userID, date, req.Content)
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, model.DailyNoteFromEntity(note))
}

// ListTaskTags handles GET /task-tags.
func (h *TaskHandler) ListTaskTags(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	tags, err := h.svc.ListTaskTags(c.Request.Context(), userID)
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"tags": model.TaskTagsFromEntities(tags)})
}

// CreateTaskTag handles POST /task-tags.
func (h *TaskHandler) CreateTaskTag(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var req model.CreateTaskTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	tag, err := h.svc.CreateTaskTag(c.Request.Context(), userID, business.CreateTaskTagInput{
		Name:  req.Name,
		Color: req.Color,
	})
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, constant.MessageOK, model.TaskTagFromEntity(tag))
}

// UpdateTaskTag handles PATCH /task-tags/:id.
func (h *TaskHandler) UpdateTaskTag(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	tagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid tag id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	var req model.UpdateTaskTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	tag, err := h.svc.UpdateTaskTag(c.Request.Context(), userID, tagID, business.UpdateTaskTagInput{
		Name:  req.Name,
		Color: req.Color,
	})
	if err != nil {
		h.writeTaskError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, model.TaskTagFromEntity(tag))
}

// DeleteTaskTag handles DELETE /task-tags/:id.
func (h *TaskHandler) DeleteTaskTag(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	tagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid tag id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	if err := h.svc.DeleteTaskTag(c.Request.Context(), userID, tagID); err != nil {
		h.writeTaskError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"deleted": true})
}

func parseUUIDList(raw []string) []uuid.UUID {
	if len(raw) == 0 {
		return nil
	}
	out := make([]uuid.UUID, 0, len(raw))
	for _, s := range raw {
		id, err := uuid.Parse(s)
		if err != nil {
			continue
		}
		out = append(out, id)
	}
	return out
}

func (h *TaskHandler) writeTaskError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not found", constant.ErrorCodeNotFound, err.Error())
	case errors.Is(err, apperr.ErrForbidden):
		response.Error(c, http.StatusForbidden, "forbidden", constant.ErrorCodeForbidden, err.Error())
	case errors.Is(err, apperr.ErrValidation):
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeInvalidRequest, err.Error())
	default:
		h.log.Error(c.Request.Context(), "task journal error", constant.LogAttrError, err)
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "task journal failed")
	}
}
