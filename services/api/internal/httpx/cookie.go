package httpx

import "net/http"

// SessionCookie builds the donna_session cookie.
// Cross-site frontends (e.g. Vercel → Railway) require SameSite=None + Secure.
// Prefer same-origin proxying in production so the cookie is first-party on the
// frontend host (see apps/web next.config.ts API_PROXY_TARGET).
func SessionCookie(name, value string, maxAge int, secure bool) *http.Cookie {
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	}
}
