package model

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// CreateUserRequest is the HTTP body for creating a user.
// email_verified is not accepted from clients; verification happens via IdP later.
type CreateUserRequest struct {
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Timezone    *string `json:"timezone"`
	Locale      *string `json:"locale"`
}

// UpdateUserRequest is the HTTP body for updating a user (partial).
type UpdateUserRequest struct {
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Timezone    *string `json:"timezone"`
	Locale      *string `json:"locale"`
	Status      *string `json:"status"`
}

// UserResponse is the HTTP DTO for a user.
type UserResponse struct {
	ID            string  `json:"id"`
	PublicID      string  `json:"public_id"`
	Email         string  `json:"email"`
	EmailVerified bool    `json:"email_verified"`
	DisplayName   *string `json:"display_name,omitempty"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	Timezone      string  `json:"timezone"`
	Locale        *string `json:"locale,omitempty"`
	Status        string  `json:"status"`
	LastLoginAt   *string `json:"last_login_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// UserFromEntity maps a domain user to the transport model.
func UserFromEntity(u entity.User) UserResponse {
	resp := UserResponse{
		ID:            u.ID.String(),
		PublicID:      u.PublicID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		DisplayName:   u.DisplayName,
		AvatarURL:     u.AvatarURL,
		Timezone:      u.Timezone,
		Locale:        u.Locale,
		Status:        u.Status,
		CreatedAt:     u.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:     u.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if u.LastLoginAt != nil {
		s := u.LastLoginAt.UTC().Format(time.RFC3339Nano)
		resp.LastLoginAt = &s
	}
	return resp
}
