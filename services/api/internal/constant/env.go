package constant

import "time"

// Runtime environment names.
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// Log levels accepted by config.
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// Environment variable names.
const (
	EnvVarAPIAddr            = "API_ADDR"
	EnvVarPort               = "PORT"
	EnvVarAPIEnv             = "API_ENV"
	EnvVarEnvironment        = "ENVIRONMENT"
	EnvVarLogLevel           = "LOG_LEVEL"
	EnvVarDatabaseURL        = "DATABASE_URL"
	EnvVarJWTSecret          = "JWT_SECRET"
	EnvVarSessionSecret      = "SESSION_SECRET"
	EnvVarCORSOrigins        = "CORS_ORIGINS"
	EnvVarShutdownTimeout    = "SHUTDOWN_TIMEOUT"
	EnvVarDBMaxConns         = "DB_MAX_CONNS"
	EnvVarDBMinConns         = "DB_MIN_CONNS"
	EnvVarMigrationsPath     = "MIGRATIONS_PATH"
	EnvVarOpenAIAPIKey       = "OPENAI_API_KEY"
	EnvVarGoogleClientID     = "GOOGLE_CLIENT_ID"
	EnvVarGoogleClientSecret = "GOOGLE_CLIENT_SECRET"
)

// Config defaults.
const (
	DefaultEnvironment     = EnvDevelopment
	DefaultLogLevel        = LogLevelInfo
	DefaultMigrationsPath  = "migrations"
	DefaultDBMaxConns      = int32(10)
	DefaultDBMinConns      = int32(0)
	DefaultJWTPlaceholder  = "change-me-to-a-long-random-string"
	DefaultShutdownTimeout = 15 * time.Second
	DefaultJWTExpiry       = 24 * time.Hour
)

// Timeouts for infrastructure operations.
const (
	DBConnectPingTimeout = 5 * time.Second
	ReadinessPingTimeout = 2 * time.Second
	DBMaxConnLifetime    = time.Hour
	DBMaxConnIdleTime    = 30 * time.Minute

	HTTPReadHeaderTimeout = 5 * time.Second
	HTTPReadTimeout       = 15 * time.Second
	HTTPWriteTimeout      = 30 * time.Second
	HTTPIdleTimeout       = 60 * time.Second
)

// SQL used by infrastructure repositories.
const (
	SQLPing = "SELECT 1"
)
