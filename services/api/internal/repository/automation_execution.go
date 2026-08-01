package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const automationExecutionColumns = `
	e.id, e.public_id, e.automation_id, e.user_id, e.started_at, e.completed_at,
	e.status, e.duration_ms, e.commands_total, e.commands_success, e.commands_failed,
	e.trigger_source, e.delivery_channels, e.delivery_status, e.response, e.error,
	e.created_at, e.updated_at`

const automationCommandExecutionColumns = `
	id, public_id, execution_id, order_index, command, command_type,
	started_at, completed_at, status, duration_ms, response, error, created_at`

const (
	sqlInsertAutomationExecution = `
INSERT INTO automation_executions (
	id, public_id, automation_id, user_id, started_at, completed_at, status,
	duration_ms, commands_total, commands_success, commands_failed,
	trigger_source, delivery_channels, delivery_status, response, error,
	created_at, updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18
)
RETURNING` + `
	id, public_id, automation_id, user_id, started_at, completed_at,
	status, duration_ms, commands_total, commands_success, commands_failed,
	trigger_source, delivery_channels, delivery_status, response, error,
	created_at, updated_at`

	sqlUpdateAutomationExecution = `
UPDATE automation_executions SET
	completed_at = $2,
	status = $3,
	duration_ms = $4,
	commands_total = $5,
	commands_success = $6,
	commands_failed = $7,
	delivery_status = $8,
	response = $9,
	error = $10,
	updated_at = $11
WHERE id = $1
RETURNING` + `
	id, public_id, automation_id, user_id, started_at, completed_at,
	status, duration_ms, commands_total, commands_success, commands_failed,
	trigger_source, delivery_channels, delivery_status, response, error,
	created_at, updated_at`

	sqlSelectAutomationExecutionByID = `
SELECT` + automationExecutionColumns + `, a.name
FROM automation_executions e
JOIN automations a ON a.id = e.automation_id
WHERE e.id = $1 AND e.user_id = $2`

	sqlListAutomationExecutionsByAutomation = `
SELECT` + automationExecutionColumns + `, a.name
FROM automation_executions e
JOIN automations a ON a.id = e.automation_id
WHERE e.automation_id = $1 AND e.user_id = $2
ORDER BY e.started_at DESC
LIMIT $3`

	sqlListAutomationExecutionsByUser = `
SELECT` + automationExecutionColumns + `, a.name
FROM automation_executions e
JOIN automations a ON a.id = e.automation_id
WHERE e.user_id = $1
ORDER BY e.started_at DESC
LIMIT $2`

	sqlInsertAutomationCommandExecution = `
INSERT INTO automation_command_executions (
	id, public_id, execution_id, order_index, command, command_type,
	started_at, completed_at, status, duration_ms, response, error, created_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13
)
RETURNING` + automationCommandExecutionColumns

	sqlListAutomationCommandExecutions = `
SELECT` + automationCommandExecutionColumns + `
FROM automation_command_executions
WHERE execution_id = $1
ORDER BY order_index ASC`

	sqlAutomationRunMetrics = `
SELECT
	automation_id,
	COUNT(*)::int AS total_executions,
	COUNT(*) FILTER (WHERE status = 'SUCCESS')::int AS successful_runs,
	AVG(duration_ms) FILTER (WHERE duration_ms IS NOT NULL) AS avg_duration_ms,
	(
		SELECT status FROM automation_executions e2
		WHERE e2.automation_id = automation_executions.automation_id
		  AND e2.user_id = $1
		  AND e2.status <> 'RUNNING'
		ORDER BY e2.started_at DESC
		LIMIT 1
	) AS last_status,
	(
		SELECT commands_total FROM automation_executions e3
		WHERE e3.automation_id = automation_executions.automation_id
		  AND e3.user_id = $1
		  AND e3.status <> 'RUNNING'
		ORDER BY e3.started_at DESC
		LIMIT 1
	) AS last_commands_total,
	(
		SELECT commands_success FROM automation_executions e4
		WHERE e4.automation_id = automation_executions.automation_id
		  AND e4.user_id = $1
		  AND e4.status <> 'RUNNING'
		ORDER BY e4.started_at DESC
		LIMIT 1
	) AS last_commands_success
FROM automation_executions
WHERE user_id = $1
  AND status <> 'RUNNING'
  AND automation_id = ANY($2)
GROUP BY automation_id`

	sqlAutomationAnalytics = `
SELECT
	COUNT(*)::int AS total_executions,
	COUNT(*) FILTER (WHERE status = 'SUCCESS')::int AS successful,
	COUNT(*) FILTER (WHERE status = 'FAILED')::int AS failed,
	COUNT(*) FILTER (WHERE status = 'PARTIAL_SUCCESS')::int AS partial,
	AVG(duration_ms) FILTER (WHERE duration_ms IS NOT NULL) AS avg_duration_ms,
	AVG(commands_total)::float8 AS avg_commands,
	(
		SELECT automation_id FROM automation_executions e2
		WHERE e2.user_id = $1 AND e2.status <> 'RUNNING'
		GROUP BY automation_id
		ORDER BY COUNT(*) DESC
		LIMIT 1
	) AS top_automation_id
FROM automation_executions
WHERE user_id = $1 AND status <> 'RUNNING'`
)

// AutomationExecutionRepository persists automation run history.
type AutomationExecutionRepository interface {
	CreateExecution(ctx context.Context, exec entity.AutomationExecution) (entity.AutomationExecution, error)
	CompleteExecution(ctx context.Context, id uuid.UUID, fields AutomationExecutionCompleteFields) (entity.AutomationExecution, error)
	GetExecutionByID(ctx context.Context, id, userID uuid.UUID) (entity.AutomationExecution, error)
	ListByAutomation(ctx context.Context, automationID, userID uuid.UUID, limit int) ([]entity.AutomationExecution, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]entity.AutomationExecution, error)
	CreateCommandExecution(ctx context.Context, cmd entity.AutomationCommandExecution) (entity.AutomationCommandExecution, error)
	ListCommandExecutions(ctx context.Context, executionID uuid.UUID) ([]entity.AutomationCommandExecution, error)
	MetricsByAutomations(ctx context.Context, userID uuid.UUID, automationIDs []uuid.UUID) (map[uuid.UUID]entity.AutomationRunMetrics, error)
	Analytics(ctx context.Context, userID uuid.UUID) (AutomationAnalyticsRow, error)
	WithTx(tx pgx.Tx) AutomationExecutionRepository
}

// AutomationExecutionCompleteFields patches a running execution to terminal state.
type AutomationExecutionCompleteFields struct {
	CompletedAt     time.Time
	Status          string
	DurationMs      int
	CommandsTotal   int
	CommandsSuccess int
	CommandsFailed  int
	DeliveryStatus  string
	Response        *string
	Error           *string
	UpdatedAt       time.Time
}

// AutomationAnalyticsRow is aggregate stats for a user.
type AutomationAnalyticsRow struct {
	TotalExecutions int
	Successful      int
	Failed          int
	Partial         int
	AvgDurationMs   *float64
	AvgCommands     *float64
	TopAutomationID *uuid.UUID
}

type automationExecutionRepository struct {
	q Querier
}

// NewAutomationExecutionRepository constructs an AutomationExecutionRepository.
func NewAutomationExecutionRepository(pool *pgxpool.Pool) AutomationExecutionRepository {
	return &automationExecutionRepository{q: pool}
}

func (r *automationExecutionRepository) WithTx(tx pgx.Tx) AutomationExecutionRepository {
	return &automationExecutionRepository{q: tx}
}

func (r *automationExecutionRepository) CreateExecution(ctx context.Context, exec entity.AutomationExecution) (entity.AutomationExecution, error) {
	deliveryJSON, err := json.Marshal(exec.DeliveryChannels)
	if err != nil {
		return entity.AutomationExecution{}, fmt.Errorf("marshal delivery: %w", err)
	}
	return scanAutomationExecution(r.q.QueryRow(ctx, sqlInsertAutomationExecution,
		exec.ID, exec.PublicID, exec.AutomationID, exec.UserID, exec.StartedAt, exec.CompletedAt,
		exec.Status, exec.DurationMs, exec.CommandsTotal, exec.CommandsSuccess, exec.CommandsFailed,
		exec.TriggerSource, deliveryJSON, exec.DeliveryStatus, exec.Response, exec.Error,
		exec.CreatedAt, exec.UpdatedAt,
	), nil)
}

func (r *automationExecutionRepository) CompleteExecution(
	ctx context.Context,
	id uuid.UUID,
	fields AutomationExecutionCompleteFields,
) (entity.AutomationExecution, error) {
	return scanAutomationExecution(r.q.QueryRow(ctx, sqlUpdateAutomationExecution,
		id, fields.CompletedAt, fields.Status, fields.DurationMs,
		fields.CommandsTotal, fields.CommandsSuccess, fields.CommandsFailed,
		fields.DeliveryStatus, fields.Response, fields.Error, fields.UpdatedAt,
	), nil)
}

func (r *automationExecutionRepository) GetExecutionByID(ctx context.Context, id, userID uuid.UUID) (entity.AutomationExecution, error) {
	var name string
	exec, err := scanAutomationExecution(r.q.QueryRow(ctx, sqlSelectAutomationExecutionByID, id, userID), &name)
	if err != nil {
		return entity.AutomationExecution{}, err
	}
	exec.AutomationName = &name
	return exec, nil
}

func (r *automationExecutionRepository) ListByAutomation(
	ctx context.Context,
	automationID, userID uuid.UUID,
	limit int,
) ([]entity.AutomationExecution, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.q.Query(ctx, sqlListAutomationExecutionsByAutomation, automationID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectAutomationExecutions(rows)
}

func (r *automationExecutionRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]entity.AutomationExecution, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.q.Query(ctx, sqlListAutomationExecutionsByUser, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectAutomationExecutions(rows)
}

func (r *automationExecutionRepository) CreateCommandExecution(
	ctx context.Context,
	cmd entity.AutomationCommandExecution,
) (entity.AutomationCommandExecution, error) {
	return scanAutomationCommandExecution(r.q.QueryRow(ctx, sqlInsertAutomationCommandExecution,
		cmd.ID, cmd.PublicID, cmd.ExecutionID, cmd.OrderIndex, cmd.Command, cmd.CommandType,
		cmd.StartedAt, cmd.CompletedAt, cmd.Status, cmd.DurationMs, cmd.Response, cmd.Error, cmd.CreatedAt,
	))
}

func (r *automationExecutionRepository) ListCommandExecutions(
	ctx context.Context,
	executionID uuid.UUID,
) ([]entity.AutomationCommandExecution, error) {
	rows, err := r.q.Query(ctx, sqlListAutomationCommandExecutions, executionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]entity.AutomationCommandExecution, 0)
	for rows.Next() {
		cmd, err := scanAutomationCommandExecution(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, cmd)
	}
	return out, rows.Err()
}

func (r *automationExecutionRepository) MetricsByAutomations(
	ctx context.Context,
	userID uuid.UUID,
	automationIDs []uuid.UUID,
) (map[uuid.UUID]entity.AutomationRunMetrics, error) {
	out := make(map[uuid.UUID]entity.AutomationRunMetrics)
	if len(automationIDs) == 0 {
		return out, nil
	}
	rows, err := r.q.Query(ctx, sqlAutomationRunMetrics, userID, automationIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			m                   entity.AutomationRunMetrics
			avg                 *float64
			lastStatus          *string
			lastCommandsTotal   *int
			lastCommandsSuccess *int
		)
		if err := rows.Scan(
			&m.AutomationID, &m.TotalExecutions, &m.SuccessfulRuns,
			&avg, &lastStatus, &lastCommandsTotal, &lastCommandsSuccess,
		); err != nil {
			return nil, fmt.Errorf("scan automation metrics: %w", err)
		}
		m.AverageDurationMs = avg
		m.LastStatus = lastStatus
		m.LastCommandsTotal = lastCommandsTotal
		m.LastCommandsSuccess = lastCommandsSuccess
		if m.TotalExecutions > 0 {
			rate := float64(m.SuccessfulRuns) / float64(m.TotalExecutions)
			m.SuccessRate = &rate
		}
		out[m.AutomationID] = m
	}
	return out, rows.Err()
}

func (r *automationExecutionRepository) Analytics(ctx context.Context, userID uuid.UUID) (AutomationAnalyticsRow, error) {
	var row AutomationAnalyticsRow
	err := r.q.QueryRow(ctx, sqlAutomationAnalytics, userID).Scan(
		&row.TotalExecutions, &row.Successful, &row.Failed, &row.Partial,
		&row.AvgDurationMs, &row.AvgCommands, &row.TopAutomationID,
	)
	if err != nil {
		return AutomationAnalyticsRow{}, fmt.Errorf("scan automation analytics: %w", err)
	}
	return row, nil
}

func collectAutomationExecutions(rows pgx.Rows) ([]entity.AutomationExecution, error) {
	out := make([]entity.AutomationExecution, 0)
	for rows.Next() {
		var name string
		exec, err := scanAutomationExecution(rows, &name)
		if err != nil {
			return nil, err
		}
		exec.AutomationName = &name
		out = append(out, exec)
	}
	return out, rows.Err()
}

func scanAutomationExecution(row pgx.Row, nameOut *string) (entity.AutomationExecution, error) {
	var (
		exec        entity.AutomationExecution
		deliveryRaw []byte
	)
	dest := []any{
		&exec.ID, &exec.PublicID, &exec.AutomationID, &exec.UserID, &exec.StartedAt, &exec.CompletedAt,
		&exec.Status, &exec.DurationMs, &exec.CommandsTotal, &exec.CommandsSuccess, &exec.CommandsFailed,
		&exec.TriggerSource, &deliveryRaw, &exec.DeliveryStatus, &exec.Response, &exec.Error,
		&exec.CreatedAt, &exec.UpdatedAt,
	}
	if nameOut != nil {
		dest = append(dest, nameOut)
	}
	err := row.Scan(dest...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.AutomationExecution{}, apperr.ErrNotFound
		}
		return entity.AutomationExecution{}, fmt.Errorf("scan automation execution: %w", err)
	}
	if err := json.Unmarshal(deliveryRaw, &exec.DeliveryChannels); err != nil {
		return entity.AutomationExecution{}, fmt.Errorf("unmarshal delivery: %w", err)
	}
	if exec.DeliveryChannels == nil {
		exec.DeliveryChannels = []string{}
	}
	return exec, nil
}

func scanAutomationCommandExecution(row pgx.Row) (entity.AutomationCommandExecution, error) {
	var cmd entity.AutomationCommandExecution
	err := row.Scan(
		&cmd.ID, &cmd.PublicID, &cmd.ExecutionID, &cmd.OrderIndex, &cmd.Command, &cmd.CommandType,
		&cmd.StartedAt, &cmd.CompletedAt, &cmd.Status, &cmd.DurationMs, &cmd.Response, &cmd.Error, &cmd.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.AutomationCommandExecution{}, apperr.ErrNotFound
		}
		return entity.AutomationCommandExecution{}, fmt.Errorf("scan automation command execution: %w", err)
	}
	return cmd, nil
}
