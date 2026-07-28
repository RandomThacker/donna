package googleoauth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RandomThacker/donna/services/api/internal/googleoauth"
)

func TestExchangeAndProfile(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("code") != "abc" {
			http.Error(w, "bad code", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "access",
			"refresh_token": "refresh",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"scope":         "openid email profile",
		})
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"sub":            "google-sub",
			"email":          "user@example.com",
			"email_verified": true,
			"name":           "User",
			"picture":        "https://example.com/a.png",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := googleoauth.NewClient(googleoauth.Config{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost/callback",
		TokenURL:     srv.URL + "/token",
		UserInfoURL:  srv.URL + "/userinfo",
		HTTPClient:   srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}

	tokens, err := client.ExchangeCode(context.Background(), "abc")
	if err != nil {
		t.Fatalf("ExchangeCode: %v", err)
	}
	if tokens.AccessToken != "access" || tokens.RefreshToken != "refresh" {
		t.Fatalf("tokens = %#v", tokens)
	}

	profile, err := client.FetchProfile(context.Background(), tokens.AccessToken)
	if err != nil {
		t.Fatalf("FetchProfile: %v", err)
	}
	if profile.Subject != "google-sub" || profile.Email != "user@example.com" || !profile.EmailVerified {
		t.Fatalf("profile = %#v", profile)
	}
}

func TestAuthCodeURL(t *testing.T) {
	client, err := googleoauth.NewClient(googleoauth.Config{
		ClientID:      "id",
		ClientSecret:  "secret",
		RedirectURL:   "http://localhost/callback",
		OfflineAccess: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	url := client.AuthCodeURL("state123")
	if url == "" {
		t.Fatal("empty auth url")
	}
	if !strings.Contains(url, "access_type=offline") {
		t.Fatalf("expected offline access: %s", url)
	}
	if !strings.Contains(url, "prompt=consent") {
		t.Fatalf("expected consent prompt: %s", url)
	}
	if strings.Contains(url, "include_granted_scopes") {
		t.Fatalf("include_granted_scopes omits refresh tokens: %s", url)
	}
}
