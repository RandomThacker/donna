package constant

// Auth / OAuth providers and public-id prefixes.
const (
	AuthProviderGoogle = "google"

	PublicIDPrefixAuthIdentity     = "aid_"
	PublicIDPrefixConnectedAccount = "acct_"
	PublicIDPrefixCredential       = "cred_"
)

// Connected account statuses.
const (
	ConnectedAccountStatusActive = "active"
)

// Session cookie.
const (
	CookieSession = "donna_session"
)

// Auth error / messages.
const (
	ErrorCodeUnauthorized  = "UNAUTHORIZED"
	ErrorCodeNotConfigured = "NOT_CONFIGURED"
	ErrorCodeOAuthFailed   = "OAUTH_FAILED"
	MessageAuthOK          = "authenticated"
)
