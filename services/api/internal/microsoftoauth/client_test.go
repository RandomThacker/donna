package microsoftoauth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/microsoftoauth"
)

func TestExchangeAndRefresh(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		switch r.Form.Get("grant_type") {
		case "authorization_code":
			if r.Form.Get("code") != "abc" {
				http.Error(w, "bad code", http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "access",
				"refresh_token": "refresh",
				"token_type":    "Bearer",
				"expires_in":    3600,
				"scope":         "openid email profile offline_access Calendars.ReadWrite",
			})
		case "refresh_token":
			if r.Form.Get("refresh_token") != "refresh" {
				http.Error(w, "bad refresh", http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "access-2",
				"refresh_token": "refresh-2",
				"token_type":    "Bearer",
				"expires_in":    7200,
				"scope":         "openid email profile offline_access Calendars.ReadWrite",
			})
		default:
			http.Error(w, "bad grant", http.StatusBadRequest)
		}
	})
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                "ms-sub",
			"displayName":       "User",
			"mail":              "user@example.com",
			"userPrincipalName": "user@contoso.com",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := microsoftoauth.NewClient(microsoftoauth.Config{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost/callback",
		TokenURL:     srv.URL + "/token",
		GraphMeURL:   srv.URL + "/me",
		HTTPClient:   srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}

	tokens, profile, err := client.Exchange(context.Background(), "abc")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if tokens.AccessToken != "access" || tokens.RefreshToken != "refresh" {
		t.Fatalf("tokens = %#v", tokens)
	}
	if profile.Subject != "ms-sub" || profile.Email != "user@example.com" || !profile.EmailVerified || profile.Name != "User" {
		t.Fatalf("profile = %#v", profile)
	}

	refreshed, err := client.RefreshAccessToken(context.Background(), "refresh")
	if err != nil {
		t.Fatalf("RefreshAccessToken: %v", err)
	}
	if refreshed.AccessToken != "access-2" || refreshed.RefreshToken != "refresh-2" || refreshed.ExpiresIn != 7200 {
		t.Fatalf("refreshed = %#v", refreshed)
	}

	providerTokens, err := client.RefreshAsProvider(context.Background(), "refresh")
	if err != nil {
		t.Fatalf("RefreshAsProvider: %v", err)
	}
	if providerTokens.AccessToken != "access-2" || providerTokens.ExpiresIn != 7200 {
		t.Fatalf("providerTokens = %#v", providerTokens)
	}
}

func TestAuthCodeURLIncludesOfflineAccess(t *testing.T) {
	client, err := microsoftoauth.NewClient(microsoftoauth.Config{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost/callback",
		Scopes: []string{
			constant.MicrosoftScopeOpenID,
			constant.MicrosoftScopeEmail,
			constant.MicrosoftScopeProfile,
			constant.MicrosoftScopeOfflineAccess,
			constant.MicrosoftScopeUserRead,
			constant.MicrosoftScopeCalendarsReadWrite,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	authURL := client.AuthCodeURL("state123")
	if authURL == "" {
		t.Fatal("empty auth url")
	}
	if !strings.Contains(authURL, "offline_access") {
		t.Fatalf("expected offline_access in %q", authURL)
	}
	if !strings.Contains(authURL, constant.MicrosoftOAuthAuthorizeURL) {
		t.Fatalf("unexpected auth host/path: %q", authURL)
	}
	if !strings.Contains(authURL, constant.MicrosoftScopeCalendarsReadWrite) {
		t.Fatalf("expected calendars scope in %q", authURL)
	}
}

func TestNewClientDefaultsToCommonAuthority(t *testing.T) {
	client, err := microsoftoauth.NewClient(microsoftoauth.Config{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost:8080/api/v1/auth/microsoft/callback",
	})
	if err != nil {
		t.Fatal(err)
	}
	authURL := client.AuthCodeURL("s")
	if !strings.HasPrefix(authURL, constant.MicrosoftOAuthAuthorizeURL+"?") {
		t.Fatalf("auth URL must use common authority, got %q", authURL)
	}
	if strings.Contains(authURL, "410843a6") {
		t.Fatalf("auth URL must not embed a tenant GUID: %q", authURL)
	}
}

func TestProfileFallsBackToUPN(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "access",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})
	mux.HandleFunc("/me", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                "ms-sub",
			"displayName":       "User",
			"mail":              nil,
			"userPrincipalName": "user@contoso.com",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := microsoftoauth.NewClient(microsoftoauth.Config{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost/callback",
		TokenURL:     srv.URL + "/token",
		GraphMeURL:   srv.URL + "/me",
		HTTPClient:   srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, profile, err := client.Exchange(context.Background(), "abc")
	if err != nil {
		t.Fatal(err)
	}
	if profile.Email != "user@contoso.com" {
		t.Fatalf("profile.Email = %q", profile.Email)
	}
}
