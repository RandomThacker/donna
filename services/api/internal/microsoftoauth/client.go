package microsoftoauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/calendarprovider"
	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// Config holds Microsoft OAuth client settings.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	// AuthURL / TokenURL default to the multi-tenant "common" authority.
	// Override only in tests (e.g. httptest). Never point these at a tenant GUID.
	AuthURL    string
	TokenURL   string
	GraphMeURL string
	Scopes     []string
	Timeout    time.Duration
	HTTPClient *http.Client
}

// TokenSet is the Microsoft token response we care about.
type TokenSet struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
	Scope        string
}

// Profile is the Microsoft Graph /me projection Donna uses.
type Profile struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
}

// Client talks to Microsoft identity and Graph profile endpoints.
type Client struct {
	cfg Config
}

// NewClient constructs a Microsoft OAuth client.
// Authorization and token endpoints always default to /common/ so personal and
// work/school accounts across any Entra tenant can authenticate.
func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.ClientID) == "" || strings.TrimSpace(cfg.ClientSecret) == "" {
		return nil, fmt.Errorf("microsoft oauth client_id and client_secret are required")
	}
	if strings.TrimSpace(cfg.RedirectURL) == "" {
		return nil, fmt.Errorf("microsoft oauth redirect_url is required")
	}
	if strings.TrimSpace(cfg.AuthURL) == "" {
		cfg.AuthURL = constant.MicrosoftOAuthAuthorizeURL
	}
	if strings.TrimSpace(cfg.TokenURL) == "" {
		cfg.TokenURL = constant.MicrosoftOAuthTokenURL
	}
	if cfg.GraphMeURL == "" {
		cfg.GraphMeURL = constant.MicrosoftGraphAPIBaseURL + "/me"
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{
			constant.MicrosoftScopeOpenID,
			constant.MicrosoftScopeEmail,
			constant.MicrosoftScopeProfile,
			constant.MicrosoftScopeUserRead,
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

// AuthCodeURL builds the Microsoft consent URL (includes offline_access via scopes).
func (c *Client) AuthCodeURL(state string) string {
	scopes := ensureOfflineAccess(c.cfg.Scopes)
	q := url.Values{}
	q.Set("client_id", c.cfg.ClientID)
	q.Set("redirect_uri", c.cfg.RedirectURL)
	q.Set("response_type", "code")
	q.Set("response_mode", "query")
	q.Set("scope", strings.Join(scopes, " "))
	q.Set("state", state)
	// Microsoft allows only one prompt value (AADSTS90023 if combined).
	q.Set("prompt", "select_account")
	return c.cfg.AuthURL + "?" + q.Encode()
}

// Exchange swaps an authorization code for tokens and loads the Graph profile.
func (c *Client) Exchange(ctx context.Context, code string) (TokenSet, Profile, error) {
	tokens, err := c.ExchangeCode(ctx, code)
	if err != nil {
		return TokenSet{}, Profile{}, err
	}
	profile, err := c.FetchProfile(ctx, tokens.AccessToken)
	if err != nil {
		return TokenSet{}, Profile{}, err
	}
	return tokens, profile, nil
}

// ExchangeCode swaps an authorization code for tokens.
func (c *Client) ExchangeCode(ctx context.Context, code string) (TokenSet, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", c.cfg.ClientID)
	form.Set("client_secret", c.cfg.ClientSecret)
	form.Set("redirect_uri", c.cfg.RedirectURL)
	form.Set("grant_type", "authorization_code")
	form.Set("scope", strings.Join(ensureOfflineAccess(c.cfg.Scopes), " "))

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
	}, nil
}

// FetchProfile loads the Microsoft Graph /me profile with an access token.
func (c *Client) FetchProfile(ctx context.Context, accessToken string) (Profile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.GraphMeURL, nil)
	if err != nil {
		return Profile{}, fmt.Errorf("graph me request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return Profile{}, fmt.Errorf("graph me: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Profile{}, fmt.Errorf("read graph me: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Profile{}, fmt.Errorf("graph me status %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var raw struct {
		ID                string `json:"id"`
		DisplayName       string `json:"displayName"`
		Mail              string `json:"mail"`
		UserPrincipalName string `json:"userPrincipalName"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return Profile{}, fmt.Errorf("decode graph me: %w", err)
	}
	if raw.ID == "" {
		return Profile{}, fmt.Errorf("graph me missing id")
	}
	email := firstNonEmpty(raw.Mail, raw.UserPrincipalName)
	return Profile{
		Subject:       raw.ID,
		Email:         email,
		EmailVerified: email != "",
		Name:          raw.DisplayName,
	}, nil
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
	form.Set("scope", strings.Join(ensureOfflineAccess(c.cfg.Scopes), " "))

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

// RefreshAsProvider refreshes tokens into the calendarprovider.TokenSet shape.
func (c *Client) RefreshAsProvider(ctx context.Context, refreshToken string) (calendarprovider.TokenSet, error) {
	ts, err := c.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		return calendarprovider.TokenSet{}, err
	}
	return calendarprovider.TokenSet{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
		TokenType:    ts.TokenType,
		ExpiresIn:    int(ts.ExpiresIn),
		Scope:        ts.Scope,
	}, nil
}

func ensureOfflineAccess(scopes []string) []string {
	for _, s := range scopes {
		if s == constant.MicrosoftScopeOfflineAccess {
			return scopes
		}
	}
	out := make([]string, 0, len(scopes)+1)
	out = append(out, scopes...)
	out = append(out, constant.MicrosoftScopeOfflineAccess)
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
