// Package constant holds shared application constants.
// Keep values here instead of scattering magic strings across packages.
package constant

// API route prefixes and relative paths registered on the Gin router.
const (
	APIPrefixV1 = "/api/v1"

	PathHealth  = "/health"
	PathReady   = "/ready"
	PathVersion = "/version"
)

// Full paths (prefix + relative). Useful for docs, clients, and tests.
const (
	EndpointHealth  = APIPrefixV1 + PathHealth
	EndpointReady   = APIPrefixV1 + PathReady
	EndpointVersion = APIPrefixV1 + PathVersion
)
