package httpx_test

import (
	"net/http"
	"testing"

	"github.com/RandomThacker/donna/services/api/internal/httpx"
)

func TestSessionCookieLocalUsesLax(t *testing.T) {
	c := httpx.SessionCookie("donna_session", "tok", 3600, false)
	if c.SameSite != http.SameSiteLaxMode {
		t.Fatalf("SameSite = %v, want Lax", c.SameSite)
	}
	if c.Secure {
		t.Fatal("Secure should be false locally")
	}
}

func TestSessionCookieSecureUsesNone(t *testing.T) {
	c := httpx.SessionCookie("donna_session", "tok", 3600, true)
	if c.SameSite != http.SameSiteNoneMode {
		t.Fatalf("SameSite = %v, want None", c.SameSite)
	}
	if !c.Secure {
		t.Fatal("Secure should be true for cross-site cookies")
	}
}
