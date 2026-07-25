package validation

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/model"
)

// NormalizeEmail trims and lowercases an email address.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateEmail(email string) (string, error) {
	normalized := NormalizeEmail(email)
	if normalized == "" {
		return "", fmt.Errorf("%w: email is required", apperr.ErrValidation)
	}
	addr, err := mail.ParseAddress(normalized)
	if err != nil || !strings.EqualFold(addr.Address, normalized) {
		return "", fmt.Errorf("%w: email is invalid", apperr.ErrValidation)
	}
	return strings.ToLower(addr.Address), nil
}

func optionalTrimmed(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// CreateUser validates and normalizes a create-user request.
func CreateUser(req model.CreateUserRequest) (model.CreateUserRequest, error) {
	email, err := validateEmail(req.Email)
	if err != nil {
		return req, err
	}

	tz := constant.DefaultUserTimezone
	if req.Timezone != nil {
		t := strings.TrimSpace(*req.Timezone)
		if t == "" {
			return req, fmt.Errorf("%w: timezone must not be empty", apperr.ErrValidation)
		}
		tz = t
	}

	return model.CreateUserRequest{
		Email:       email,
		DisplayName: optionalTrimmed(req.DisplayName),
		AvatarURL:   optionalTrimmed(req.AvatarURL),
		Timezone:    &tz,
		Locale:      optionalTrimmed(req.Locale),
	}, nil
}

// UpdateUser validates a partial update request.
// Status may only be active|disabled; pending_deletion is SoftDelete's job.
func UpdateUser(req model.UpdateUserRequest) (model.UpdateUserRequest, error) {
	if req.DisplayName == nil && req.AvatarURL == nil && req.Timezone == nil &&
		req.Locale == nil && req.Status == nil {
		return req, fmt.Errorf("%w: at least one field is required", apperr.ErrValidation)
	}

	out := model.UpdateUserRequest{}

	if req.Timezone != nil {
		t := strings.TrimSpace(*req.Timezone)
		if t == "" {
			return req, fmt.Errorf("%w: timezone must not be empty", apperr.ErrValidation)
		}
		out.Timezone = &t
	}
	if req.Status != nil {
		s := strings.TrimSpace(*req.Status)
		switch s {
		case constant.UserStatusActive, constant.UserStatusDisabled:
			out.Status = &s
		default:
			return req, fmt.Errorf("%w: status is invalid", apperr.ErrValidation)
		}
	}
	if req.DisplayName != nil {
		if trimmed := optionalTrimmed(req.DisplayName); trimmed != nil {
			out.DisplayName = trimmed
		} else {
			empty := ""
			out.DisplayName = &empty
		}
	}
	if req.AvatarURL != nil {
		if trimmed := optionalTrimmed(req.AvatarURL); trimmed != nil {
			out.AvatarURL = trimmed
		} else {
			empty := ""
			out.AvatarURL = &empty
		}
	}
	if req.Locale != nil {
		if trimmed := optionalTrimmed(req.Locale); trimmed != nil {
			out.Locale = trimmed
		} else {
			empty := ""
			out.Locale = &empty
		}
	}

	return out, nil
}

// EmailQuery validates a required email query parameter.
func EmailQuery(email string) (string, error) {
	return validateEmail(email)
}
