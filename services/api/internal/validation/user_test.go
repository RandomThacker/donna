package validation_test

import (
	"errors"
	"testing"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/model"
	"github.com/RandomThacker/donna/services/api/internal/validation"
)

func TestCreateUserRequiresEmail(t *testing.T) {
	_, err := validation.CreateUser(model.CreateUserRequest{})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateUserDefaultsTimezone(t *testing.T) {
	got, err := validation.CreateUser(model.CreateUserRequest{Email: "a@b.com"})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if got.Timezone == nil || *got.Timezone != constant.DefaultUserTimezone {
		t.Fatalf("timezone = %#v", got.Timezone)
	}
	if got.Email != "a@b.com" {
		t.Fatalf("email = %q", got.Email)
	}
}

func TestCreateUserNormalizesEmail(t *testing.T) {
	got, err := validation.CreateUser(model.CreateUserRequest{Email: "  Ada@Example.COM "})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if got.Email != "ada@example.com" {
		t.Fatalf("email = %q", got.Email)
	}
}

func TestCreateUserRejectsNamedAddress(t *testing.T) {
	_, err := validation.CreateUser(model.CreateUserRequest{Email: "Ada <ada@example.com>"})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateUserRequiresField(t *testing.T) {
	_, err := validation.UpdateUser(model.UpdateUserRequest{})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateUserRejectsPendingDeletionStatus(t *testing.T) {
	status := constant.UserStatusPendingDeletion
	_, err := validation.UpdateUser(model.UpdateUserRequest{Status: &status})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateUserAllowsDisabled(t *testing.T) {
	status := constant.UserStatusDisabled
	got, err := validation.UpdateUser(model.UpdateUserRequest{Status: &status})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if got.Status == nil || *got.Status != constant.UserStatusDisabled {
		t.Fatalf("status = %#v", got.Status)
	}
}

func TestEmailQuery(t *testing.T) {
	got, err := validation.EmailQuery("  A@B.COM ")
	if err != nil {
		t.Fatalf("EmailQuery: %v", err)
	}
	if got != "a@b.com" {
		t.Fatalf("got = %q", got)
	}
}

func TestEmailQueryRequired(t *testing.T) {
	_, err := validation.EmailQuery("   ")
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}
