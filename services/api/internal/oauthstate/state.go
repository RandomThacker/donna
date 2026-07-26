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

// Manager creates and verifies OAuth CSRF state values.
type Manager struct {
	secret []byte
	now    func() time.Time
}

// NewManager constructs a state manager keyed by the app secret.
func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret), now: time.Now}
}

// Create returns a signed state token.
func (m *Manager) Create() (string, error) {
	return m.create("")
}

// CreateWithUser returns a signed state token bound to a Donna user id (integration connect).
func (m *Manager) CreateWithUser(userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", fmt.Errorf("user id is required")
	}
	return m.create(userID)
}

func (m *Manager) create(userID string) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("nonce: %w", err)
	}
	exp := m.now().UTC().Add(maxAge).Unix()
	payload := hex.EncodeToString(nonce) + ":" + strconv.FormatInt(exp, 10)
	if userID != "" {
		payload += ":" + userID
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
	if len(parts) < 3 || strings.TrimSpace(parts[2]) == "" {
		return "", fmt.Errorf("state missing user binding")
	}
	return parts[2], nil
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
