package entity

import (
	"time"

	"github.com/google/uuid"
)

// AuthIdentity binds a Donna user to an IdP subject (login).
type AuthIdentity struct {
	ID              uuid.UUID
	PublicID        string
	UserID          uuid.UUID
	Provider        string
	ProviderSubject string
	Email           *string
	EmailVerified   bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

// ConnectedAccount is an integration OAuth account (credentials boundary).
type ConnectedAccount struct {
	ID                uuid.UUID
	PublicID          string
	UserID            uuid.UUID
	Provider          string
	ProviderAccountID string
	DisplayName       *string
	Status            string
	Scopes            []string
	CredentialsRef    string
	TokenExpiresAt    *time.Time
	LastSyncedAt      *time.Time
	ProviderMetadata  []byte
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

// CredentialSecret is a sealed provider token blob.
type CredentialSecret struct {
	ID         uuid.UUID
	Ref        string
	Ciphertext []byte
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}
