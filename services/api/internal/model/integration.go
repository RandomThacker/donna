package model

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// ConnectedAccountResponse is the HTTP DTO for a connected integration account.
type ConnectedAccountResponse struct {
	ID                 string   `json:"id"`
	PublicID           string   `json:"public_id"`
	Provider           string   `json:"provider"`
	ProviderAccountID  string   `json:"provider_account_id"`
	DisplayName        *string  `json:"display_name,omitempty"`
	Email              *string  `json:"email,omitempty"`
	AvatarURL          *string  `json:"avatar_url,omitempty"`
	Status             string   `json:"status"`
	Scopes             []string `json:"scopes"`
	TokenExpiresAt     *string  `json:"token_expires_at,omitempty"`
	LastSyncedAt       *string  `json:"last_synced_at,omitempty"`
	CalendarSyncStatus string   `json:"calendar_sync_status"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
}

// ConnectICSRequest is the HTTP body for POST /integrations/ics.
type ConnectICSRequest struct {
	Name        string `json:"name"`
	ICSURL      string `json:"ics_url"`
	SyncEnabled *bool  `json:"sync_enabled"`
}

// UpdateICSRequest is the HTTP body for PATCH /integrations/ics/:id.
type UpdateICSRequest struct {
	Name        *string `json:"name"`
	ICSURL      *string `json:"ics_url"`
	SyncEnabled *bool   `json:"sync_enabled"`
}

// ICSIntegrationResponse is the HTTP DTO for an ICS calendar feed (URL never exposed).
type ICSIntegrationResponse struct {
	ID                 string  `json:"id"`
	PublicID           string  `json:"public_id"`
	Provider           string  `json:"provider"`
	Name               string  `json:"name"`
	Status             string  `json:"status"`
	SyncEnabled        bool    `json:"sync_enabled"`
	LastSyncedAt       *string `json:"last_synced_at,omitempty"`
	CalendarSyncStatus string  `json:"calendar_sync_status"`
	EventCount         int64   `json:"event_count"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

// ConnectedAccountFromEntity maps a domain connected account to the transport model.
func ConnectedAccountFromEntity(a entity.ConnectedAccount) ConnectedAccountResponse {
	resp := ConnectedAccountResponse{
		ID:                 a.ID.String(),
		PublicID:           a.PublicID,
		Provider:           a.Provider,
		ProviderAccountID:  a.ProviderAccountID,
		DisplayName:        a.DisplayName,
		Email:              emailFromProviderMetadata(a.ProviderMetadata, a.DisplayName),
		AvatarURL:          avatarURLFromProviderMetadata(a.ProviderMetadata),
		Status:             a.Status,
		Scopes:             a.Scopes,
		CalendarSyncStatus: a.CalendarSyncStatus,
		CreatedAt:          a.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:          a.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if resp.Scopes == nil {
		resp.Scopes = []string{}
	}
	if resp.CalendarSyncStatus == "" {
		resp.CalendarSyncStatus = "idle"
	}
	if a.TokenExpiresAt != nil {
		s := a.TokenExpiresAt.UTC().Format(time.RFC3339Nano)
		resp.TokenExpiresAt = &s
	}
	if a.LastSyncedAt != nil {
		s := a.LastSyncedAt.UTC().Format(time.RFC3339Nano)
		resp.LastSyncedAt = &s
	}
	return resp
}

// ConnectedAccountsFromEntities maps a slice of connected accounts.
func ConnectedAccountsFromEntities(accounts []entity.ConnectedAccount) []ConnectedAccountResponse {
	out := make([]ConnectedAccountResponse, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, ConnectedAccountFromEntity(a))
	}
	return out
}

// ICSIntegrationFromEntity maps an ICS connected account projection to the transport model.
func ICSIntegrationFromEntity(a entity.ConnectedAccount, syncEnabled bool, eventCount int64) ICSIntegrationResponse {
	name := ""
	if a.DisplayName != nil {
		name = strings.TrimSpace(*a.DisplayName)
	}
	if name == "" {
		name = stringFieldFromMetadata(a.ProviderMetadata, "name")
	}
	if name == "" {
		name = "Calendar URL (ICS)"
	}
	resp := ICSIntegrationResponse{
		ID:                 a.ID.String(),
		PublicID:           a.PublicID,
		Provider:           a.Provider,
		Name:               name,
		Status:             a.Status,
		SyncEnabled:        syncEnabled,
		CalendarSyncStatus: a.CalendarSyncStatus,
		EventCount:         eventCount,
		CreatedAt:          a.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:          a.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if resp.CalendarSyncStatus == "" {
		resp.CalendarSyncStatus = "idle"
	}
	if a.LastSyncedAt != nil {
		s := a.LastSyncedAt.UTC().Format(time.RFC3339Nano)
		resp.LastSyncedAt = &s
	}
	return resp
}

func emailFromProviderMetadata(raw []byte, displayName *string) *string {
	if email := stringFieldFromMetadata(raw, "email"); email != "" {
		return &email
	}
	// Legacy rows stored email in display_name before provider_metadata.email existed.
	if displayName != nil {
		if name := strings.TrimSpace(*displayName); strings.Contains(name, "@") {
			return &name
		}
	}
	return nil
}

func avatarURLFromProviderMetadata(raw []byte) *string {
	if avatar := stringFieldFromMetadata(raw, "avatar_url"); avatar != "" {
		return &avatar
	}
	return nil
}

func stringFieldFromMetadata(raw []byte, key string) string {
	if len(raw) == 0 {
		return ""
	}
	var meta map[string]any
	if json.Unmarshal(raw, &meta) != nil {
		return ""
	}
	v, ok := meta[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}
