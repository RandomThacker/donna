package model

import "time"

// AuthSessionResponse is returned after a successful OAuth login.
type AuthSessionResponse struct {
	AccessToken string       `json:"access_token,omitempty"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int64        `json:"expires_in"`
	ExpiresAt   string       `json:"expires_at"`
	IsNewUser   bool         `json:"is_new_user"`
	User        UserResponse `json:"user"`
}

// NewAuthSessionResponse builds the transport DTO from session fields.
func NewAuthSessionResponse(accessToken, tokenType string, expiresIn int64, expiresAt time.Time, isNewUser bool, user UserResponse) AuthSessionResponse {
	return AuthSessionResponse{
		AccessToken: accessToken,
		TokenType:   tokenType,
		ExpiresIn:   expiresIn,
		ExpiresAt:   expiresAt.UTC().Format(time.RFC3339Nano),
		IsNewUser:   isNewUser,
		User:        user,
	}
}
