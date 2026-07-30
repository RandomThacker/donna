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

	PathAuthGoogle            = "/auth/google"
	PathAuthGoogleCallback    = "/auth/google/callback"
	PathAuthMicrosoft         = "/auth/microsoft"
	PathAuthMicrosoftCallback = "/auth/microsoft/callback"
	PathAuthLogout            = "/auth/logout"
	PathMe                    = "/me"

	PathCalendarSync       = "/calendar/sync"
	PathCalendarSyncEnsure = "/calendar/sync/ensure"
	PathCalendarSources    = "/calendar/sources"
	PathCalendarEventsSync = "/calendar/events/sync"
	PathCalendarEvents     = "/calendar/events"

	PathIntegrations                  = "/integrations"
	PathIntegrationsGoogle            = "/integrations/google"
	PathIntegrationsGoogleCallback    = "/integrations/google/callback"
	PathIntegrationsMicrosoft         = "/integrations/microsoft"
	PathIntegrationsMicrosoftCallback = "/integrations/microsoft/callback"
	PathIntegrationsICS               = "/integrations/ics"
	PathIntegrationsICSByID           = "/integrations/ics/:id"
	PathIntegrationsICSSync           = "/integrations/ics/:id/sync"
	PathIntegrationsByID              = "/integrations/:id"

	PathTasks                  = "/tasks"
	PathTasksDay               = "/tasks/day/:date"
	PathTasksHistory           = "/tasks/history"
	PathTasksCarryForward      = "/tasks/carry-forward"
	PathTaskByID               = "/tasks/:id"
	PathTaskOccurrences        = "/task-occurrences/:id"
	PathTaskOccurrencesReorder = "/task-occurrences/reorder"
	PathDailyNotesDay          = "/daily-notes/:date"

	PathTaskTags    = "/task-tags"
	PathTaskTagByID = "/task-tags/:id"

	PathNotes    = "/notes"
	PathNoteByID = "/notes/:id"

	PathTimeline = "/timeline"

	PathDonnaEvents       = "/donna/events"
	PathDonnaEventByID    = "/donna/events/:id"
	PathDonnaReminders    = "/donna/reminders"
	PathDonnaReminderByID = "/donna/reminders/:id"

	PathNotifications           = "/notifications"
	PathNotificationRead        = "/notifications/:id/read"
	PathNotificationDismiss     = "/notifications/:id/dismiss"

	PathPushSubscribe         = "/push/subscribe"
	PathPushUnsubscribe       = "/push/unsubscribe"
	PathPushVAPIDPublicKey    = "/push/vapid-public-key"

	PathChatCommand  = "/chat/command"
	PathChatMessages = "/chat/messages"
	PathChatSummary  = "/chat/summary"
)

// Full paths (prefix + relative). Useful for docs, clients, and tests.
const (
	EndpointHealth                        = APIPrefixV1 + PathHealth
	EndpointReady                         = APIPrefixV1 + PathReady
	EndpointVersion                       = APIPrefixV1 + PathVersion
	EndpointUsers                         = APIPrefixV1 + PathUsers
	EndpointAuthGoogle                    = APIPrefixV1 + PathAuthGoogle
	EndpointAuthGoogleCallback            = APIPrefixV1 + PathAuthGoogleCallback
	EndpointAuthMicrosoft                 = APIPrefixV1 + PathAuthMicrosoft
	EndpointAuthMicrosoftCallback         = APIPrefixV1 + PathAuthMicrosoftCallback
	EndpointIntegrations                  = APIPrefixV1 + PathIntegrations
	EndpointIntegrationsGoogle            = APIPrefixV1 + PathIntegrationsGoogle
	EndpointIntegrationsGoogleCallback    = APIPrefixV1 + PathIntegrationsGoogleCallback
	EndpointIntegrationsMicrosoft         = APIPrefixV1 + PathIntegrationsMicrosoft
	EndpointIntegrationsMicrosoftCallback = APIPrefixV1 + PathIntegrationsMicrosoftCallback
	EndpointIntegrationsICS               = APIPrefixV1 + PathIntegrationsICS
	EndpointIntegrationsICSByID           = APIPrefixV1 + PathIntegrationsICSByID
	EndpointIntegrationsICSSync           = APIPrefixV1 + PathIntegrationsICSSync
)
