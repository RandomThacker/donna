package config

// Raw JSON shapes (all strings so ${ENV} placeholders expand cleanly).

type appconfigFile struct {
	Environment        string `json:"environment"`
	Addr               string `json:"addr"`
	LogLevel           string `json:"log_level"`
	CORSOrigins        string `json:"cors_origins"`
	ShutdownTimeout    string `json:"shutdown_timeout"`
	JWTSecret          string `json:"jwt_secret"`
	SessionSecret      string `json:"session_secret"`
	JWTExpiry          string `json:"jwt_expiry"`
	CredentialsKey     string `json:"credentials_encryption_key"`
	FrontendSuccessURL            string `json:"frontend_success_url"`
	IntegrationFrontendSuccessURL string `json:"integration_frontend_success_url"`
	CookieSecure                  string `json:"cookie_secure"`
	VAPIDPublicKey                string `json:"vapid_public_key"`
	VAPIDPrivateKey               string `json:"vapid_private_key"`
	VAPIDSubject                  string `json:"vapid_subject"`
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
	OpenAI         externalAPIFile    `json:"openai"`
	GoogleOAuth    googleOAuthFile    `json:"google_oauth"`
	MicrosoftOAuth microsoftOAuthFile `json:"microsoft_oauth"`
	AIService      externalAPIFile    `json:"ai_service"`
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

type googleOAuthFile struct {
	Name                   string            `json:"name"`
	BaseURL                string            `json:"base_url"`
	Path                   string            `json:"path"`
	Method                 string            `json:"method"`
	Timeout                string            `json:"timeout"`
	ClientID               string            `json:"client_id"`
	ClientSecret           string            `json:"client_secret"`
	RedirectURL            string            `json:"redirect_url"`
	IntegrationRedirectURL string            `json:"integration_redirect_url"`
	AuthURL                string            `json:"auth_url"`
	UserInfoURL            string            `json:"userinfo_url"`
	Scopes                 string            `json:"scopes"`
	IntegrationScopes      string            `json:"integration_scopes"`
	Headers                map[string]string `json:"headers"`
}

type microsoftOAuthFile struct {
	Name                   string            `json:"name"`
	BaseURL                string            `json:"base_url"`
	Path                   string            `json:"path"`
	Method                 string            `json:"method"`
	Timeout                string            `json:"timeout"`
	ClientID               string            `json:"client_id"`
	ClientSecret           string            `json:"client_secret"`
	RedirectURL            string            `json:"redirect_url"`
	IntegrationRedirectURL string            `json:"integration_redirect_url"`
	Tenant                 string            `json:"tenant"`
	AuthURL                string            `json:"auth_url"`
	TokenURL               string            `json:"token_url"`
	GraphMeURL             string            `json:"graph_me_url"`
	Scopes                 string            `json:"scopes"`
	IntegrationScopes      string            `json:"integration_scopes"`
	Headers                map[string]string `json:"headers"`
}
