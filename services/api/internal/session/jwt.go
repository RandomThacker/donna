package session

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims are Donna session JWT claims.
type Claims struct {
	UserID   string `json:"user_id"`
	PublicID string `json:"public_id"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// Issuer creates and validates HS256 JWTs.
type Issuer struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// NewIssuer constructs a JWT issuer.
func NewIssuer(secret string, ttl time.Duration) (*Issuer, error) {
	if secret == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &Issuer{
		secret: []byte(secret),
		ttl:    ttl,
		now:    time.Now,
	}, nil
}

// IssueResult is a minted session token.
type IssueResult struct {
	AccessToken string
	ExpiresAt   time.Time
	ExpiresIn   int64
}

// Issue mints a session JWT for a user.
func (i *Issuer) Issue(userID uuid.UUID, publicID, email string) (IssueResult, error) {
	now := i.now().UTC()
	exp := now.Add(i.ttl)
	claims := Claims{
		UserID:   userID.String(),
		PublicID: publicID,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        uuid.NewString(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return IssueResult{}, fmt.Errorf("sign jwt: %w", err)
	}
	return IssueResult{
		AccessToken: signed,
		ExpiresAt:   exp,
		ExpiresIn:   int64(i.ttl.Seconds()),
	}, nil
}

// Parse validates a JWT and returns claims.
func (i *Issuer) Parse(tokenString string) (Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return i.secret, nil
	})
	if err != nil {
		return Claims{}, fmt.Errorf("parse jwt: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}
	return *claims, nil
}
