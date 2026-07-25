package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/model"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/RandomThacker/donna/services/api/internal/session"
	"github.com/gin-gonic/gin"
)

// AuthHandler maps Google OAuth HTTP endpoints to the auth business layer.
type AuthHandler struct {
	svc                *business.AuthService
	tokens             *session.Issuer
	log                *logger.Logger
	frontendSuccessURL string
	cookieSecure       bool
	cookieMaxAge       int
}

// AuthHandlerConfig configures AuthHandler cookie/redirect behavior.
type AuthHandlerConfig struct {
	FrontendSuccessURL string
	CookieSecure       bool
	CookieMaxAge       time.Duration
	Tokens             *session.Issuer
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(svc *business.AuthService, log *logger.Logger, cfg AuthHandlerConfig) *AuthHandler {
	maxAge := int(cfg.CookieMaxAge.Seconds())
	if maxAge <= 0 {
		maxAge = int((24 * time.Hour).Seconds())
	}
	return &AuthHandler{
		svc:                svc,
		tokens:             cfg.Tokens,
		log:                log,
		frontendSuccessURL: cfg.FrontendSuccessURL,
		cookieSecure:       cfg.CookieSecure,
		cookieMaxAge:       maxAge,
	}
}

// BeginGoogle redirects the browser to Google's consent screen.
// If a valid Donna session cookie already exists, skip Google and send the user back to the app.
func (h *AuthHandler) BeginGoogle(c *gin.Context) {
	if h.hasValidSession(c) {
		h.redirectAlreadyAuthenticated(c)
		return
	}

	authURL, _, err := h.svc.BeginGoogleLogin(c.Request.Context())
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

	h.setSessionCookie(c, session.AccessToken)

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

	redirectTo, err := url.Parse(h.frontendSuccessURL)
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

// Logout clears the session cookie.
func (h *AuthHandler) Logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     constant.CookieSession,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	response.OK(c, "logged out", nil)
}

func (h *AuthHandler) setSessionCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     constant.CookieSession,
		Value:    token,
		Path:     "/",
		MaxAge:   h.cookieMaxAge,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) hasValidSession(c *gin.Context) bool {
	if h.tokens == nil {
		return false
	}
	raw, err := c.Cookie(constant.CookieSession)
	if err != nil || strings.TrimSpace(raw) == "" {
		return false
	}
	_, err = h.tokens.Parse(raw)
	return err == nil
}

func (h *AuthHandler) redirectAlreadyAuthenticated(c *gin.Context) {
	redirectTo, err := url.Parse(h.frontendSuccessURL)
	if err != nil {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}
	q := redirectTo.Query()
	q.Set("status", "ok")
	redirectTo.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, redirectTo.String())
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
