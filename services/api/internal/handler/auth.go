package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/httpx"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/model"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/RandomThacker/donna/services/api/internal/session"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// sessionUserLookup verifies that a JWT subject still exists after DB wipes / soft-deletes.
type sessionUserLookup interface {
	GetByID(ctx context.Context, id uuid.UUID) (entity.User, error)
}

// AuthHandler maps multi-provider OAuth login HTTP endpoints to the auth business layer.
type AuthHandler struct {
	svc                *business.AuthService
	users              sessionUserLookup
	tokens             *session.Issuer
	log                *logger.Logger
	frontendSuccessURL string
	allowedOrigins     map[string]struct{}
	cookieSecure       bool
	cookieMaxAge       int
}

// AuthHandlerConfig configures AuthHandler cookie/redirect behavior.
type AuthHandlerConfig struct {
	FrontendSuccessURL string
	AllowedOrigins     []string
	CookieSecure       bool
	CookieMaxAge       time.Duration
	Tokens             *session.Issuer
	Users              sessionUserLookup
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(svc *business.AuthService, log *logger.Logger, cfg AuthHandlerConfig) *AuthHandler {
	maxAge := int(cfg.CookieMaxAge.Seconds())
	if maxAge <= 0 {
		maxAge = int((24 * time.Hour).Seconds())
	}
	allowed := make(map[string]struct{}, len(cfg.AllowedOrigins)+1)
	for _, origin := range cfg.AllowedOrigins {
		origin = strings.TrimRight(strings.TrimSpace(origin), "/")
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}
	if u, err := url.Parse(strings.TrimSpace(cfg.FrontendSuccessURL)); err == nil && u.Scheme != "" && u.Host != "" {
		allowed[u.Scheme+"://"+u.Host] = struct{}{}
	}
	return &AuthHandler{
		svc:                svc,
		users:              cfg.Users,
		tokens:             cfg.Tokens,
		log:                log,
		frontendSuccessURL: cfg.FrontendSuccessURL,
		allowedOrigins:     allowed,
		cookieSecure:       cfg.CookieSecure,
		cookieMaxAge:       maxAge,
	}
}

// BeginGoogle redirects the browser to Google's consent screen.
// If a live Donna session exists, skip Google and send the user back to the app.
// Optional query return_to (e.g. http://localhost:3000/auth/callback) keeps local FE
// on localhost after login when API is proxied through Next.js.
func (h *AuthHandler) BeginGoogle(c *gin.Context) {
	if h.hasLiveSession(c) {
		h.redirectAlreadyAuthenticated(c)
		return
	}

	returnTo, err := h.resolveReturnTo(c.Query("return_to"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid return_to", constant.ErrorCodeValidation, err.Error())
		return
	}

	authURL, _, err := h.svc.BeginGoogleLogin(c.Request.Context(), returnTo)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

// GoogleCallback handles the OAuth redirect from Google.
// Session is stored only in an HttpOnly cookie — never in the redirect URL.
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	if errParam := c.Query("error"); errParam != "" {
		response.Error(
			c,
			http.StatusBadRequest,
			"google oauth denied",
			constant.ErrorCodeOAuthFailed,
			errParam,
		)
		return
	}

	session, err := h.svc.CompleteGoogleLogin(c.Request.Context(), c.Query("code"), c.Query("state"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	h.finishLogin(c, session)
}

// BeginMicrosoft redirects the browser to Microsoft's consent screen.
func (h *AuthHandler) BeginMicrosoft(c *gin.Context) {
	if h.hasLiveSession(c) {
		h.redirectAlreadyAuthenticated(c)
		return
	}

	authURL, _, err := h.svc.BeginMicrosoftLogin(c.Request.Context())
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

// MicrosoftCallback handles the OAuth redirect from Microsoft (login).
func (h *AuthHandler) MicrosoftCallback(c *gin.Context) {
	if errParam := c.Query("error"); errParam != "" {
		response.Error(
			c,
			http.StatusBadRequest,
			"microsoft oauth denied",
			constant.ErrorCodeOAuthFailed,
			errParam,
		)
		return
	}

	session, err := h.svc.CompleteMicrosoftLogin(c.Request.Context(), c.Query("code"), c.Query("state"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	h.finishLogin(c, session)
}

// Logout clears the session cookie.
func (h *AuthHandler) Logout(c *gin.Context) {
	h.clearSessionCookie(c)
	response.OK(c, "logged out", nil)
}

func (h *AuthHandler) finishLogin(c *gin.Context, session business.AuthSession) {
	successURL := h.frontendSuccessURL
	if strings.TrimSpace(session.FrontendReturnURL) != "" {
		successURL = session.FrontendReturnURL
	}

	secure := h.cookieSecure
	if u, err := url.Parse(successURL); err == nil && isLocalDevHost(u) {
		// Local FE proxies API on http://localhost — first-party Lax cookie (no Secure).
		secure = false
	}
	http.SetCookie(c.Writer, httpx.SessionCookie(
		constant.CookieSession,
		session.AccessToken,
		h.cookieMaxAge,
		secure,
	))

	// Programmatic clients may request JSON; browsers get a cookie + redirect.
	if c.GetHeader("Accept") == "application/json" || c.Query("format") == "json" {
		response.OK(c, constant.MessageAuthOK, model.NewAuthSessionResponse(
			"", // never return JWT to browser clients via this path either when cookie is set
			session.TokenType,
			session.ExpiresIn,
			session.ExpiresAt,
			session.IsNewUser,
			model.UserFromEntity(session.User),
		))
		return
	}

	redirectTo, err := url.Parse(successURL)
	if err != nil {
		response.OK(c, constant.MessageAuthOK, model.UserFromEntity(session.User))
		return
	}
	q := redirectTo.Query()
	q.Set("status", "ok")
	if session.IsNewUser {
		q.Set("new_user", "1")
	}
	redirectTo.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, redirectTo.String())
}

func (h *AuthHandler) setSessionCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, httpx.SessionCookie(
		constant.CookieSession,
		token,
		h.cookieMaxAge,
		h.cookieSecure,
	))
}

func (h *AuthHandler) clearSessionCookie(c *gin.Context) {
	// Clear both Secure and non-Secure variants (prod HTTPS vs local http://localhost proxy).
	http.SetCookie(c.Writer, httpx.SessionCookie(constant.CookieSession, "", -1, true))
	http.SetCookie(c.Writer, httpx.SessionCookie(constant.CookieSession, "", -1, false))
}

// hasLiveSession is true only when the cookie JWT is valid AND the user still exists.
// Stale cookies after a DB wipe must not short-circuit OAuth.
func (h *AuthHandler) hasLiveSession(c *gin.Context) bool {
	if h.tokens == nil {
		return false
	}
	raw, err := c.Cookie(constant.CookieSession)
	if err != nil || strings.TrimSpace(raw) == "" {
		return false
	}
	claims, err := h.tokens.Parse(raw)
	if err != nil {
		h.clearSessionCookie(c)
		return false
	}
	if h.users == nil {
		return true
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		h.clearSessionCookie(c)
		return false
	}
	if _, err := h.users.GetByID(c.Request.Context(), userID); err != nil {
		h.clearSessionCookie(c)
		return false
	}
	return true
}

func (h *AuthHandler) redirectAlreadyAuthenticated(c *gin.Context) {
	successURL := h.frontendSuccessURL
	if returnTo, err := h.resolveReturnTo(c.Query("return_to")); err == nil && returnTo != "" {
		successURL = returnTo
	}
	redirectTo, err := url.Parse(successURL)
	if err != nil {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}
	q := redirectTo.Query()
	q.Set("status", "ok")
	redirectTo.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, redirectTo.String())
}

// resolveReturnTo validates an optional frontend callback URL against allowlisted origins.
// Empty input means "use the default AUTH_FRONTEND_SUCCESS_URL".
func (h *AuthHandler) resolveReturnTo(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", errors.New("return_to must be an absolute URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("return_to must be http(s)")
	}
	origin := u.Scheme + "://" + u.Host
	if _, ok := h.allowedOrigins[origin]; !ok {
		return "", errors.New("return_to origin is not allowed")
	}
	// Normalize to the auth callback path so arbitrary deep-links can't be injected.
	return origin + "/auth/callback", nil
}

func isLocalDevHost(u *url.URL) bool {
	if u == nil || u.Scheme != "http" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func (h *AuthHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrValidation):
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, err.Error())
	case errors.Is(err, apperr.ErrInvalid):
		response.Error(c, http.StatusBadRequest, "invalid oauth request", constant.ErrorCodeInvalidRequest, err.Error())
	case errors.Is(err, apperr.ErrNotFound):
		response.Error(c, http.StatusNotFound, "user not found", constant.ErrorCodeNotFound, err.Error())
	case errors.Is(err, apperr.ErrConflict):
		response.Error(c, http.StatusConflict, "account conflict", constant.ErrorCodeConflict, err.Error())
	default:
		h.log.Error(c.Request.Context(), "auth request failed", constant.LogAttrError, err)
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "unexpected error")
	}
}
