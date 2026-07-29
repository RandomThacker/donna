package constant

// Auth / OAuth providers and public-id prefixes.
const (
	AuthProviderGoogle    = "google"
	AuthProviderMicrosoft = "microsoft"
	AuthProviderICS       = "ics"

	PublicIDPrefixAuthIdentity     = "aid_"
	PublicIDPrefixConnectedAccount = "acct_"
	PublicIDPrefixCredential       = "cred_"
	PublicIDPrefixCalendarSource   = "cal_"
	PublicIDPrefixCalendarEvent    = "evt_"
	PublicIDPrefixCalendarSyncRun  = "csync_"
)

// Google OAuth / Calendar scopes.
const (
	GoogleScopeOpenID        = "openid"
	GoogleScopeEmail         = "email"
	GoogleScopeProfile       = "profile"
	GoogleScopeCalendar      = "https://www.googleapis.com/auth/calendar"
	GoogleCalendarAPIBaseURL = "https://www.googleapis.com/calendar/v3"
)

// Microsoft OAuth / Graph scopes.
const (
	MicrosoftScopeOpenID             = "openid"
	MicrosoftScopeEmail              = "email"
	MicrosoftScopeProfile            = "profile"
	MicrosoftScopeOfflineAccess      = "offline_access"
	MicrosoftScopeUserRead           = "User.Read"
	MicrosoftScopeCalendarsRead      = "Calendars.Read"
	MicrosoftScopeCalendarsReadWrite = "Calendars.ReadWrite"
	MicrosoftGraphAPIBaseURL         = "https://graph.microsoft.com/v1.0"
	// Multi-tenant authority ("Accounts in any org + personal Microsoft accounts").
	// Never derive these from MICROSOFT_TENANT_ID.
	MicrosoftOAuthAuthorizeURL = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	MicrosoftOAuthTokenURL     = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
)

// ICS calendar scopes / defaults.
const (
	ICSScopeCalendar          = "ics.calendar"
	ICSDefaultSyncIntervalMin = 15
	ICSHTTPUserAgent          = "DonnaCalendar/1.0"
)

// Connected account statuses.
const (
	ConnectedAccountStatusActive       = "active"
	ConnectedAccountStatusDisconnected = "disconnected"
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
