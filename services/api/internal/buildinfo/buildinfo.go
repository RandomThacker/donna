package buildinfo

// These are set via -ldflags at build time. Defaults suit local `go run`.
var (
	Version   = "dev"
	Commit    = "none"
	BuildTime = "unknown"
)
