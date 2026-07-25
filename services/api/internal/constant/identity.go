package constant

// User lifecycle statuses (users.status check constraint).
const (
	UserStatusActive          = "active"
	UserStatusDisabled        = "disabled"
	UserStatusPendingDeletion = "pending_deletion"
)

// Public ID prefixes.
const (
	PublicIDPrefixUser = "usr_"
)

// Default user timezone when omitted on create.
const DefaultUserTimezone = "UTC"
