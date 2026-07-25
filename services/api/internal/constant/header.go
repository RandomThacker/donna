package constant

// HTTP header names.
const (
	HeaderRequestID     = "X-Request-ID"
	HeaderOrigin        = "Origin"
	HeaderAuthorization = "Authorization"
	HeaderAccept        = "Accept"
	HeaderContentType   = "Content-Type"
	HeaderVary          = "Vary"

	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"
)

// CORS policy values.
const (
	CORSAllowOriginAll = "*"
	CORSAllowMethods   = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	CORSMaxAgeSeconds  = "86400"
)

// CORSAllowHeaders is the comma-joined Access-Control-Allow-Headers value.
const CORSAllowHeaders = HeaderAccept + ", " + HeaderAuthorization + ", " + HeaderContentType + ", " + HeaderRequestID
