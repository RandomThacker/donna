package oauthstate

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const maxAge = 10 * time.Minute

const loginMetaMarker = "#"

// Manager creates and verifies OAuth CSRF state values.
type Manager struct {
	secret []byte
	now    func() time.Time
}

// LoginMeta is bound into login OAuth state so the callback can restore
// the correct Google redirect_uri and frontend return URL.
type LoginMeta struct {
	ReturnTo    string
	RedirectURI string
}

// NewManager constructs a state manager keyed by the app secret.
func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret), now: time.Now}
}

// Create returns a signed state token.
func (m *Manager) Create() (string, error) {
	return m.create("", LoginMeta{})
}

// CreateWithUser returns a signed state token bound to a Donna user id (integration connect).
func (m *Manager) CreateWithUser(userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", fmt.Errorf("user id is required")
	}
	return m.create(userID, LoginMeta{})
}

// CreateLogin returns a signed state token bound to login return URLs.
func (m *Manager) CreateLogin(meta LoginMeta) (string, error) {
	meta.ReturnTo = strings.TrimSpace(meta.ReturnTo)
	meta.RedirectURI = strings.TrimSpace(meta.RedirectURI)
	if meta.ReturnTo == "" || meta.RedirectURI == "" {
		return "", fmt.Errorf("return_to and redirect_uri are required")
	}
	return m.create("", meta)
}

func (m *Manager) create(userID string, meta LoginMeta) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("nonce: %w", err)
	}
	exp := m.now().UTC().Add(maxAge).Unix()
	payload := hex.EncodeToString(nonce) + ":" + strconv.FormatInt(exp, 10)
	if userID != "" {
		payload += ":" + userID
	} else if meta.ReturnTo != "" {
		payload += ":" + loginMetaMarker + ":" +
			base64.RawURLEncoding.EncodeToString([]byte(meta.ReturnTo)) + ":" +
			base64.RawURLEncoding.EncodeToString([]byte(meta.RedirectURI))
	}
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	raw := payload + "." + sig
	return base64.RawURLEncoding.EncodeToString([]byte(raw)), nil
}

// Verify checks signature and expiry.
func (m *Manager) Verify(state string) error {
	_, err := m.verify(state)
	return err
}

// VerifyWithUser checks signature/expiry and returns the bound user id.
func (m *Manager) VerifyWithUser(state string) (userID string, err error) {
	payload, err := m.verify(state)
	if err != nil {
		return "", err
	}
	parts := strings.Split(payload, ":")
	if len(parts) < 3 || strings.TrimSpace(parts[2]) == "" || parts[2] == loginMetaMarker {
		return "", fmt.Errorf("state missing user binding")
	}
	return parts[2], nil
}

// VerifyLogin checks signature/expiry and returns login return-URL meta when present.
func (m *Manager) VerifyLogin(state string) (LoginMeta, error) {
	payload, err := m.verify(state)
	if err != nil {
		return LoginMeta{}, err
	}
	parts := strings.Split(payload, ":")
	if len(parts) >= 5 && parts[2] == loginMetaMarker {
		returnTo, err := base64.RawURLEncoding.DecodeString(parts[3])
		if err != nil {
			return LoginMeta{}, fmt.Errorf("decode return_to: %w", err)
		}
		redirectURI, err := base64.RawURLEncoding.DecodeString(parts[4])
		if err != nil {
			return LoginMeta{}, fmt.Errorf("decode redirect_uri: %w", err)
		}
		return LoginMeta{
			ReturnTo:    string(returnTo),
			RedirectURI: string(redirectURI),
		}, nil
	}
	return LoginMeta{}, nil
}

func (m *Manager) verify(state string) (payload string, err error) {
	raw, err := base64.RawURLEncoding.DecodeString(state)
	if err != nil {
		return "", fmt.Errorf("decode state: %w", err)
	}
	parts := strings.Split(string(raw), ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("malformed state")
	}
	payload, sigHex := parts[0], parts[1]
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sigHex)) {
		return "", fmt.Errorf("invalid state signature")
	}
	payloadParts := strings.Split(payload, ":")
	if len(payloadParts) < 2 {
		return "", fmt.Errorf("malformed state payload")
	}
	exp, err := strconv.ParseInt(payloadParts[1], 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid state expiry")
	}
	if m.now().UTC().Unix() > exp {
		return "", fmt.Errorf("state expired")
	}
	return payload, nil
}
