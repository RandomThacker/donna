package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// Load reads configs/{appconfig,database,api}.json, expands ${ENV} placeholders,
// and validates the result. Config directory comes from CONFIG_DIR or defaults to "configs".
func Load() (*Config, error) {
	dir := firstNonEmpty(os.Getenv(constant.EnvVarConfigDir), constant.DefaultConfigDir)
	return LoadFromDir(dir)
}

// LoadFromDir loads and validates JSON config files from dir.
func LoadFromDir(dir string) (*Config, error) {
	appRaw, err := readExpandedJSON(filepath.Join(dir, constant.ConfigFileApp))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", constant.ConfigFileApp, err)
	}
	dbRaw, err := readExpandedJSON(filepath.Join(dir, constant.ConfigFileDatabase))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", constant.ConfigFileDatabase, err)
	}
	apiRaw, err := readExpandedJSON(filepath.Join(dir, constant.ConfigFileAPI))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", constant.ConfigFileAPI, err)
	}

	var appFile appconfigFile
	if err := json.Unmarshal(appRaw, &appFile); err != nil {
		return nil, fmt.Errorf("parse %s: %w", constant.ConfigFileApp, err)
	}
	var dbFile databaseFile
	if err := json.Unmarshal(dbRaw, &dbFile); err != nil {
		return nil, fmt.Errorf("parse %s: %w", constant.ConfigFileDatabase, err)
	}
	var apisFile apiFile
	if err := json.Unmarshal(apiRaw, &apisFile); err != nil {
		return nil, fmt.Errorf("parse %s: %w", constant.ConfigFileAPI, err)
	}

	cfg, err := assemble(appFile, dbFile, apisFile)
	if err != nil {
		return nil, err
	}
	if err := validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func readExpandedJSON(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return []byte(expandEnv(string(raw))), nil
}

func assemble(appFile appconfigFile, dbFile databaseFile, apisFile apiFile) (*Config, error) {
	var errs []string

	addr := strings.TrimSpace(appFile.Addr)
	if addr == "" {
		addr = deriveAddrFromPort(os.Getenv(constant.EnvVarPort))
	}

	shutdownTimeout, err := parseDuration(appFile.ShutdownTimeout, constant.DefaultShutdownTimeout)
	if err != nil {
		errs = append(errs, fmt.Sprintf("shutdown_timeout: %v", err))
	}

	maxConns, err := parseInt32(dbFile.MaxConns, constant.DefaultDBMaxConns)
	if err != nil {
		errs = append(errs, fmt.Sprintf("max_conns: %v", err))
	}
	minConns, err := parseInt32(dbFile.MinConns, constant.DefaultDBMinConns)
	if err != nil {
		errs = append(errs, fmt.Sprintf("min_conns: %v", err))
	}
	maxLifetime, err := parseDuration(dbFile.MaxConnLifetime, constant.DBMaxConnLifetime)
	if err != nil {
		errs = append(errs, fmt.Sprintf("max_conn_lifetime: %v", err))
	}
	maxIdle, err := parseDuration(dbFile.MaxConnIdleTime, constant.DBMaxConnIdleTime)
	if err != nil {
		errs = append(errs, fmt.Sprintf("max_conn_idle_time: %v", err))
	}
	pingTimeout, err := parseDuration(dbFile.ConnectPingTimeout, constant.DBConnectPingTimeout)
	if err != nil {
		errs = append(errs, fmt.Sprintf("connect_ping_timeout: %v", err))
	}

	openai, err := mapExternalAPI(apisFile.OpenAI)
	if err != nil {
		errs = append(errs, fmt.Sprintf("openai: %v", err))
	}
	google, err := mapGoogleOAuth(apisFile.GoogleOAuth)
	if err != nil {
		errs = append(errs, fmt.Sprintf("google_oauth: %v", err))
	}
	aiService, err := mapExternalAPI(apisFile.AIService)
	if err != nil {
		errs = append(errs, fmt.Sprintf("ai_service: %v", err))
	}

	jwtExpiry, err := parseDuration(appFile.JWTExpiry, constant.DefaultJWTExpiry)
	if err != nil {
		errs = append(errs, fmt.Sprintf("jwt_expiry: %v", err))
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid configuration:\n  - %s", strings.Join(errs, "\n  - "))
	}

	jwtSecret := firstNonEmpty(appFile.JWTSecret, appFile.SessionSecret)
	credentialsKey := firstNonEmpty(appFile.CredentialsKey, jwtSecret)
	frontendURL := strings.TrimSpace(appFile.FrontendSuccessURL)
	if frontendURL == "" {
		frontendURL = "http://localhost:3000/auth/callback"
	}

	cookieSecure := appFile.Environment == constant.EnvProduction || appFile.Environment == constant.EnvStaging
	if v := strings.TrimSpace(strings.ToLower(appFile.CookieSecure)); v != "" {
		cookieSecure = v == "true" || v == "1" || v == "yes"
	}

	return &Config{
		App: AppConfig{
			Addr:               addr,
			Environment:        strings.ToLower(strings.TrimSpace(appFile.Environment)),
			LogLevel:           strings.ToLower(strings.TrimSpace(appFile.LogLevel)),
			CORSOrigins:        splitCSV(appFile.CORSOrigins),
			JWTSecret:          jwtSecret,
			JWTExpiry:          jwtExpiry,
			CredentialsKey:     credentialsKey,
			FrontendSuccessURL: frontendURL,
			CookieSecure:       cookieSecure,
			ShutdownTimeout:    shutdownTimeout,
		},
		Database: DatabaseConfig{
			URL:                strings.TrimSpace(dbFile.URL),
			MaxConns:           maxConns,
			MinConns:           minConns,
			MaxConnLifetime:    maxLifetime,
			MaxConnIdleTime:    maxIdle,
			ConnectPingTimeout: pingTimeout,
			MigrationsPath:     firstNonEmpty(strings.TrimSpace(dbFile.MigrationsPath), constant.DefaultMigrationsPath),
		},
		API: ExternalAPIConfig{
			OpenAI:      openai,
			GoogleOAuth: google,
			AIService:   aiService,
		},
	}, nil
}

func mapExternalAPI(in externalAPIFile) (ExternalAPI, error) {
	timeout, err := parseDuration(in.Timeout, 30*time.Second)
	if err != nil {
		return ExternalAPI{}, err
	}
	headers := in.Headers
	if headers == nil {
		headers = map[string]string{}
	}
	return ExternalAPI{
		Name:         in.Name,
		BaseURL:      strings.TrimSpace(in.BaseURL),
		Path:         strings.TrimSpace(in.Path),
		Method:       strings.ToUpper(strings.TrimSpace(in.Method)),
		Timeout:      timeout,
		APIKey:       in.APIKey,
		ClientID:     in.ClientID,
		ClientSecret: in.ClientSecret,
		Headers:      headers,
	}, nil
}

func mapGoogleOAuth(in googleOAuthFile) (GoogleOAuthConfig, error) {
	timeout, err := parseDuration(in.Timeout, 15*time.Second)
	if err != nil {
		return GoogleOAuthConfig{}, err
	}
	headers := in.Headers
	if headers == nil {
		headers = map[string]string{}
	}
	tokenURL := strings.TrimSpace(in.BaseURL)
	if tokenURL != "" && strings.TrimSpace(in.Path) != "" {
		tokenURL = strings.TrimRight(tokenURL, "/") + in.Path
	} else if tokenURL == "" {
		tokenURL = "https://oauth2.googleapis.com/token"
	}
	scopes := splitCSV(strings.ReplaceAll(in.Scopes, " ", ","))
	if len(scopes) == 0 {
		scopes = []string{
			"openid",
			"email",
			"profile",
			"https://www.googleapis.com/auth/calendar",
		}
	}
	authURL := strings.TrimSpace(in.AuthURL)
	if authURL == "" {
		authURL = "https://accounts.google.com/o/oauth2/v2/auth"
	}
	userInfoURL := strings.TrimSpace(in.UserInfoURL)
	if userInfoURL == "" {
		userInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"
	}
	redirect := strings.TrimSpace(in.RedirectURL)
	if redirect == "" {
		redirect = "http://localhost:8080/api/v1/auth/google/callback"
	}
	return GoogleOAuthConfig{
		Name:         in.Name,
		ClientID:     strings.TrimSpace(in.ClientID),
		ClientSecret: strings.TrimSpace(in.ClientSecret),
		RedirectURL:  redirect,
		AuthURL:      authURL,
		TokenURL:     tokenURL,
		UserInfoURL:  userInfoURL,
		Scopes:       scopes,
		Timeout:      timeout,
		Headers:      headers,
	}, nil
}

func validate(cfg *Config) error {
	var errs []string

	if cfg.App.Addr == "" {
		errs = append(errs, "appconfig.addr (or PORT) is required")
	}

	switch cfg.App.Environment {
	case constant.EnvDevelopment, constant.EnvStaging, constant.EnvProduction:
	default:
		errs = append(errs, "appconfig.environment must be one of development, staging, production")
	}

	switch cfg.App.LogLevel {
	case constant.LogLevelDebug, constant.LogLevelInfo, constant.LogLevelWarn, constant.LogLevelError:
	default:
		errs = append(errs, "appconfig.log_level must be one of debug, info, warn, error")
	}

	if cfg.Database.URL == "" {
		errs = append(errs, "database.url is required (set DATABASE_URL)")
	}
	if cfg.App.JWTSecret == "" {
		errs = append(errs, "appconfig.jwt_secret is required (set JWT_SECRET or SESSION_SECRET)")
	}
	if cfg.App.Environment == constant.EnvProduction && cfg.App.JWTSecret == constant.DefaultJWTPlaceholder {
		errs = append(errs, "jwt_secret must not use the development placeholder in production")
	}
	if cfg.Database.MaxConns < 1 {
		errs = append(errs, "database.max_conns must be >= 1")
	}
	if cfg.Database.MinConns < 0 {
		errs = append(errs, "database.min_conns must be >= 0")
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid configuration:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

func parseDuration(raw string, fallback time.Duration) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, err
	}
	return d, nil
}

func parseInt32(raw string, fallback int32) (int32, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(n), nil
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func deriveAddrFromPort(port string) string {
	if port == "" {
		return ""
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
