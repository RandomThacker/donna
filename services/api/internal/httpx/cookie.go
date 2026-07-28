package httpx

import "net/http"

// SessionCookie builds the donna_session cookie.
// Cross-site frontends (e.g. Vercel → Railway) require SameSite=None + Secure.
// Localhost same-site ports can use Lax without Secure.
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
