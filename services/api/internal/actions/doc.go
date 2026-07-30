// Package actions is Donna's interface-agnostic Action Layer.
//
// REST, and later Chat / Telegram / AI, call Actions. Actions call Services.
// Actions must not import Gin, HTTP, Telegram, or AI packages.
package actions

import "context"

// DomainEvent is a placeholder for future event-bus publishing.
type DomainEvent struct {
	Name    string
	Payload any
}

// DomainEventPublisher publishes domain events after successful workflows.
// Phase 2.4 ships a no-op implementation.
type DomainEventPublisher interface {
	Publish(ctx context.Context, event DomainEvent) error
}

// NoopPublisher discards domain events (placeholder).
type NoopPublisher struct{}

// Publish implements DomainEventPublisher.
func (NoopPublisher) Publish(context.Context, DomainEvent) error { return nil }
