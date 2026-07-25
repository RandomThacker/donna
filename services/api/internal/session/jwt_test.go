package session

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIssueAndParse(t *testing.T) {
	iss, err := NewIssuer("test-jwt-secret-value", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse("01900000-0000-7000-8000-000000000099")
	res, err := iss.Issue(id, "usr_abc", "a@b.com")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := iss.Parse(res.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != id.String() || claims.Email != "a@b.com" || claims.PublicID != "usr_abc" {
		t.Fatalf("claims = %#v", claims)
	}
}
