// Package constant holds shared application constants.
// Keep values here instead of scattering magic strings across packages.
package constant

// API route prefixes and relative paths registered on the Gin router.
const (
	APIPrefixV1 = "/api/v1"

	PathHealth  = "/health"
	PathReady   = "/ready"
	PathVersion = "/version"

	PathUsers    = "/users"
	PathUserByID = "/users/:id"

	PathAuthGoogle         = "/auth/google"
	PathAuthGoogleCallback = "/auth/google/callback"
	PathAuthLogout         = "/auth/logout"
	PathMe                 = "/me"

	PathCalendarSync       = "/calendar/sync"
	PathCalendarSyncEnsure = "/calendar/sync/ensure"
	PathCalendarSources    = "/calendar/sources"
	PathCalendarEventsSync = "/calendar/events/sync"
	PathCalendarEvents     = "/calendar/events"
)

// Full paths (prefix + relative). Useful for docs, clients, and tests.
const (
	EndpointHealth             = APIPrefixV1 + PathHealth
	EndpointReady              = APIPrefixV1 + PathReady
	EndpointVersion            = APIPrefixV1 + PathVersion
	EndpointUsers              = APIPrefixV1 + PathUsers
	EndpointAuthGoogle         = APIPrefixV1 + PathAuthGoogle
	EndpointAuthGoogleCallback = APIPrefixV1 + PathAuthGoogleCallback
)
