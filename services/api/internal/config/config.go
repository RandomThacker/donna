package config

import "time"

// Config is the fully resolved runtime configuration loaded from JSON + env.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	API      ExternalAPIConfig
}

// AppConfig holds process/runtime settings from appconfig.json.
type AppConfig struct {
	Addr               string
	Environment        string
	LogLevel           string
	CORSOrigins        []string
	JWTSecret          string
	JWTExpiry          time.Duration
	CredentialsKey     string
	FrontendSuccessURL string
	CookieSecure       bool
	ShutdownTimeout    time.Duration
}

// DatabaseConfig holds Postgres pool settings from database.json.
type DatabaseConfig struct {
	URL                string
	MaxConns           int32
	MinConns           int32
	MaxConnLifetime    time.Duration
	MaxConnIdleTime    time.Duration
	ConnectPingTimeout time.Duration
	MigrationsPath     string
}

// ExternalAPIConfig holds outbound HTTP client definitions from api.json.
type ExternalAPIConfig struct {
	OpenAI      ExternalAPI
	GoogleOAuth GoogleOAuthConfig
	AIService   ExternalAPI
}

// ExternalAPI describes one outbound HTTP integration.
type ExternalAPI struct {
	Name         string
	BaseURL      string
	Path         string
	Method       string
	Timeout      time.Duration
	APIKey       string
	ClientID     string
	ClientSecret string
	Headers      map[string]string
}

// GoogleOAuthConfig holds Google login OAuth settings.
type GoogleOAuthConfig struct {
	Name         string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
	Timeout      time.Duration
	Headers      map[string]string
}
