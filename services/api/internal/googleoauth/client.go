package googleoauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Config holds Google OAuth client settings.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
	Timeout      time.Duration
	HTTPClient   *http.Client
}

// TokenSet is the Google token response we care about.
type TokenSet struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
	Scope        string
	IDToken      string
}

// Profile is the Google OpenID userinfo payload.
type Profile struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// Client talks to Google's OAuth endpoints.
type Client struct {
	cfg Config
}

// NewClient constructs a Google OAuth client.
func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.ClientID) == "" || strings.TrimSpace(cfg.ClientSecret) == "" {
		return nil, fmt.Errorf("google oauth client_id and client_secret are required")
	}
	if strings.TrimSpace(cfg.RedirectURL) == "" {
		return nil, fmt.Errorf("google oauth redirect_url is required")
	}
	if cfg.AuthURL == "" {
		cfg.AuthURL = "https://accounts.google.com/o/oauth2/v2/auth"
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = "https://oauth2.googleapis.com/token"
	}
	if cfg.UserInfoURL == "" {
		cfg.UserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{
			"openid",
			"email",
			"profile",
			"https://www.googleapis.com/auth/calendar",
		}
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 15 * time.Second
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: cfg.Timeout}
	}
	return &Client{cfg: cfg}, nil
}

// AuthCodeURL builds the Google consent URL.
func (c *Client) AuthCodeURL(state string) string {
	q := url.Values{}
	q.Set("client_id", c.cfg.ClientID)
	q.Set("redirect_uri", c.cfg.RedirectURL)
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(c.cfg.Scopes, " "))
	q.Set("state", state)
	q.Set("access_type", "offline")
	q.Set("prompt", "consent")
	q.Set("include_granted_scopes", "true")
	return c.cfg.AuthURL + "?" + q.Encode()
}

// ExchangeCode swaps an authorization code for tokens.
func (c *Client) ExchangeCode(ctx context.Context, code string) (TokenSet, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", c.cfg.ClientID)
	form.Set("client_secret", c.cfg.ClientSecret)
	form.Set("redirect_uri", c.cfg.RedirectURL)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenSet{}, fmt.Errorf("token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return TokenSet{}, fmt.Errorf("token exchange: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return TokenSet{}, fmt.Errorf("read token response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return TokenSet{}, fmt.Errorf("token exchange status %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var raw struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		Scope        string `json:"scope"`
		IDToken      string `json:"id_token"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return TokenSet{}, fmt.Errorf("decode token response: %w", err)
	}
	if raw.Error != "" {
		return TokenSet{}, fmt.Errorf("token exchange error: %s (%s)", raw.Error, raw.ErrorDesc)
	}
	if raw.AccessToken == "" {
		return TokenSet{}, fmt.Errorf("token exchange missing access_token")
	}

	return TokenSet{
		AccessToken:  raw.AccessToken,
		RefreshToken: raw.RefreshToken,
		TokenType:    raw.TokenType,
		ExpiresIn:    raw.ExpiresIn,
		Scope:        raw.Scope,
		IDToken:      raw.IDToken,
	}, nil
}

// FetchProfile loads the Google user profile with an access token.
func (c *Client) FetchProfile(ctx context.Context, accessToken string) (Profile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.UserInfoURL, nil)
	if err != nil {
		return Profile{}, fmt.Errorf("userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return Profile{}, fmt.Errorf("userinfo: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Profile{}, fmt.Errorf("read userinfo: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Profile{}, fmt.Errorf("userinfo status %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var profile Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return Profile{}, fmt.Errorf("decode userinfo: %w", err)
	}
	if profile.Subject == "" {
		return Profile{}, fmt.Errorf("userinfo missing sub")
	}
	return profile, nil
}

// RefreshAccessToken exchanges a refresh token for a new access token.
func (c *Client) RefreshAccessToken(ctx context.Context, refreshToken string) (TokenSet, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return TokenSet{}, fmt.Errorf("refresh token is required")
	}

	form := url.Values{}
	form.Set("client_id", c.cfg.ClientID)
	form.Set("client_secret", c.cfg.ClientSecret)
	form.Set("refresh_token", refreshToken)
	form.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenSet{}, fmt.Errorf("refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return TokenSet{}, fmt.Errorf("refresh token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return TokenSet{}, fmt.Errorf("read refresh response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return TokenSet{}, fmt.Errorf("refresh token status %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var raw struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		Scope        string `json:"scope"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return TokenSet{}, fmt.Errorf("decode refresh response: %w", err)
	}
	if raw.Error != "" {
		return TokenSet{}, fmt.Errorf("refresh token error: %s (%s)", raw.Error, raw.ErrorDesc)
	}
	if raw.AccessToken == "" {
		return TokenSet{}, fmt.Errorf("refresh token missing access_token")
	}

	return TokenSet{
		AccessToken:  raw.AccessToken,
		RefreshToken: raw.RefreshToken,
		TokenType:    raw.TokenType,
		ExpiresIn:    raw.ExpiresIn,
		Scope:        raw.Scope,
	}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
