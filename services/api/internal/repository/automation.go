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

const automationColumns = `
	id, public_id, user_id, name, description, enabled,
	trigger_type, to_char(trigger_time, 'HH24:MI'), trigger_days, timezone,
	commands, delivery_channels, template_id,
	last_run_at, next_run_at, created_at, updated_at, deleted_at`

const (
	sqlInsertAutomation = `
INSERT INTO automations (
	id, public_id, user_id, name, description, enabled,
	trigger_type, trigger_time, trigger_days, timezone, commands, delivery_channels,
	template_id, last_run_at, next_run_at, created_at, updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8::time,$9,$10,$11,$12,$13,$14,$15,$16,$17
)
RETURNING` + automationColumns

	sqlSelectAutomationByID = `
SELECT` + automationColumns + `
FROM automations
WHERE id = $1 AND deleted_at IS NULL`

	sqlListAutomationsByUser = `
SELECT` + automationColumns + `
FROM automations
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY trigger_time ASC, name ASC`

	sqlListEnabledAutomations = `
SELECT` + automationColumns + `
FROM automations
WHERE deleted_at IS NULL AND enabled = true
ORDER BY user_id, trigger_time`

	sqlUpdateAutomation = `
UPDATE automations SET
	name = COALESCE($2, name),
	description = COALESCE($3, description),
	enabled = COALESCE($4, enabled),
	trigger_type = COALESCE($5, trigger_type),
	trigger_time = COALESCE($6::time, trigger_time),
	trigger_days = COALESCE($7, trigger_days),
	timezone = COALESCE($8, timezone),
	commands = COALESCE($9, commands),
	delivery_channels = COALESCE($10, delivery_channels),
	template_id = COALESCE($11, template_id),
	next_run_at = COALESCE($12, next_run_at),
	updated_at = $13
WHERE id = $1 AND user_id = $14 AND deleted_at IS NULL
RETURNING` + automationColumns

	sqlMarkAutomationRun = `
UPDATE automations SET
	last_run_at = $2,
	next_run_at = $3,
	updated_at = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING` + automationColumns

	sqlSoftDeleteAutomation = `
UPDATE automations SET deleted_at = $3, updated_at = $3
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
)

// AutomationRepository persists automations.
type AutomationRepository interface {
	Create(ctx context.Context, auto entity.Automation) (entity.Automation, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.Automation, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Automation, error)
	ListEnabled(ctx context.Context) ([]entity.Automation, error)
	Update(ctx context.Context, id, userID uuid.UUID, fields AutomationUpdateFields, updatedAt time.Time) (entity.Automation, error)
	MarkRun(ctx context.Context, id uuid.UUID, ranAt time.Time, nextRunAt *time.Time) (entity.Automation, error)
	SoftDelete(ctx context.Context, id, userID uuid.UUID, deletedAt time.Time) error
	WithTx(tx pgx.Tx) AutomationRepository
}

// AutomationUpdateFields are optional patches for an automation.
type AutomationUpdateFields struct {
	Name             *string
	Description      *string
	Enabled          *bool
	TriggerType      *string
	TriggerTime      *string
	TriggerDays      []string // nil = leave unchanged; non-nil replaces (empty OK for daily)
	Timezone         *string
	Commands         []entity.AutomationCommand
	DeliveryChannels []string
	TemplateID       *string
	NextRunAt        *time.Time
	// TriggerDaysSet marks that TriggerDays should be written (including empty).
	TriggerDaysSet bool
}

type automationRepository struct {
	q Querier
}

// NewAutomationRepository constructs an AutomationRepository.
func NewAutomationRepository(pool *pgxpool.Pool) AutomationRepository {
	return &automationRepository{q: pool}
}

func (r *automationRepository) WithTx(tx pgx.Tx) AutomationRepository {
	return &automationRepository{q: tx}
}

func (r *automationRepository) Create(ctx context.Context, auto entity.Automation) (entity.Automation, error) {
	commandsJSON, err := json.Marshal(auto.Commands)
	if err != nil {
		return entity.Automation{}, fmt.Errorf("marshal commands: %w", err)
	}
	deliveryJSON, err := json.Marshal(auto.DeliveryChannels)
	if err != nil {
		return entity.Automation{}, fmt.Errorf("marshal delivery: %w", err)
	}
	days := auto.TriggerDays
	if days == nil {
		days = []string{}
	}
	return scanAutomation(r.q.QueryRow(ctx, sqlInsertAutomation,
		auto.ID, auto.PublicID, auto.UserID, auto.Name, auto.Description, auto.Enabled,
		auto.TriggerType, auto.TriggerTime, days, auto.Timezone, commandsJSON, deliveryJSON,
		auto.TemplateID, auto.LastRunAt, auto.NextRunAt, auto.CreatedAt, auto.UpdatedAt,
	))
}

func (r *automationRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.Automation, error) {
	return scanAutomation(r.q.QueryRow(ctx, sqlSelectAutomationByID, id))
}

func (r *automationRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Automation, error) {
	rows, err := r.q.Query(ctx, sqlListAutomationsByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectAutomations(rows)
}

func (r *automationRepository) ListEnabled(ctx context.Context) ([]entity.Automation, error) {
	rows, err := r.q.Query(ctx, sqlListEnabledAutomations)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectAutomations(rows)
}

func (r *automationRepository) Update(
	ctx context.Context,
	id, userID uuid.UUID,
	fields AutomationUpdateFields,
	updatedAt time.Time,
) (entity.Automation, error) {
	var commandsJSON any
	if fields.Commands != nil {
		b, err := json.Marshal(fields.Commands)
		if err != nil {
			return entity.Automation{}, fmt.Errorf("marshal commands: %w", err)
		}
		commandsJSON = b
	}
	var deliveryJSON any
	if fields.DeliveryChannels != nil {
		b, err := json.Marshal(fields.DeliveryChannels)
		if err != nil {
			return entity.Automation{}, fmt.Errorf("marshal delivery: %w", err)
		}
		deliveryJSON = b
	}
	var days any
	if fields.TriggerDaysSet {
		d := fields.TriggerDays
		if d == nil {
			d = []string{}
		}
		days = d
	}
	return scanAutomation(r.q.QueryRow(ctx, sqlUpdateAutomation,
		id, fields.Name, fields.Description, fields.Enabled, fields.TriggerType,
		fields.TriggerTime, days, fields.Timezone, commandsJSON, deliveryJSON,
		fields.TemplateID, fields.NextRunAt, updatedAt, userID,
	))
}

func (r *automationRepository) MarkRun(ctx context.Context, id uuid.UUID, ranAt time.Time, nextRunAt *time.Time) (entity.Automation, error) {
	return scanAutomation(r.q.QueryRow(ctx, sqlMarkAutomationRun, id, ranAt, nextRunAt))
}

func (r *automationRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID, deletedAt time.Time) error {
	tag, err := r.q.Exec(ctx, sqlSoftDeleteAutomation, id, userID, deletedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func collectAutomations(rows pgx.Rows) ([]entity.Automation, error) {
	out := make([]entity.Automation, 0)
	for rows.Next() {
		auto, err := scanAutomation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, auto)
	}
	return out, rows.Err()
}

func scanAutomation(row pgx.Row) (entity.Automation, error) {
	var (
		auto        entity.Automation
		commandsRaw []byte
		deliveryRaw []byte
	)
	err := row.Scan(
		&auto.ID, &auto.PublicID, &auto.UserID, &auto.Name, &auto.Description, &auto.Enabled,
		&auto.TriggerType, &auto.TriggerTime, &auto.TriggerDays, &auto.Timezone,
		&commandsRaw, &deliveryRaw, &auto.TemplateID,
		&auto.LastRunAt, &auto.NextRunAt, &auto.CreatedAt, &auto.UpdatedAt, &auto.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Automation{}, apperr.ErrNotFound
		}
		return entity.Automation{}, fmt.Errorf("scan automation: %w", err)
	}
	commands, err := entity.UnmarshalAutomationCommands(commandsRaw)
	if err != nil {
		return entity.Automation{}, fmt.Errorf("unmarshal commands: %w", err)
	}
	auto.Commands = commands
	if err := json.Unmarshal(deliveryRaw, &auto.DeliveryChannels); err != nil {
		return entity.Automation{}, fmt.Errorf("unmarshal delivery: %w", err)
	}
	if auto.Commands == nil {
		auto.Commands = []entity.AutomationCommand{}
	}
	if auto.DeliveryChannels == nil {
		auto.DeliveryChannels = []string{}
	}
	if auto.TriggerDays == nil {
		auto.TriggerDays = []string{}
	}
	return auto, nil
}
