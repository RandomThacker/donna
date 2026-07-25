package config

// Raw JSON shapes (all strings so ${ENV} placeholders expand cleanly).

type appconfigFile struct {
	Environment     string `json:"environment"`
	Addr            string `json:"addr"`
	LogLevel        string `json:"log_level"`
	CORSOrigins     string `json:"cors_origins"`
	ShutdownTimeout string `json:"shutdown_timeout"`
	JWTSecret       string `json:"jwt_secret"`
	SessionSecret   string `json:"session_secret"`
}

type databaseFile struct {
	URL                string `json:"url"`
	MaxConns           string `json:"max_conns"`
	MinConns           string `json:"min_conns"`
	MaxConnLifetime    string `json:"max_conn_lifetime"`
	MaxConnIdleTime    string `json:"max_conn_idle_time"`
	ConnectPingTimeout string `json:"connect_ping_timeout"`
	MigrationsPath     string `json:"migrations_path"`
}

type apiFile struct {
	OpenAI      externalAPIFile `json:"openai"`
	GoogleOAuth externalAPIFile `json:"google_oauth"`
	AIService   externalAPIFile `json:"ai_service"`
}

type externalAPIFile struct {
	Name         string            `json:"name"`
	BaseURL      string            `json:"base_url"`
	Path         string            `json:"path"`
	Method       string            `json:"method"`
	Timeout      string            `json:"timeout"`
	APIKey       string            `json:"api_key"`
	ClientID     string            `json:"client_id"`
	ClientSecret string            `json:"client_secret"`
	Headers      map[string]string `json:"headers"`
}
