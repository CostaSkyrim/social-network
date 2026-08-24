package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"social-network/backend/cache"
)

type Config struct {
	DatabaseConfiguration DatabaseConfig    `json:"database_configuration"`
	Server                ServerConfig      `json:"server"`
	Handlers              HandlersConfig    `json:"handlers"`
	Frontend              FrontendConfig    `json:"frontend"`
	Redis                 cache.RedisConfig `json:"redis"`
}

type FrontendConfig struct {
	URL string `json:"url"`
}

type DatabaseConfig struct {
	Path            []string            `json:"path"`
	WAL             WALConfig           `json:"wal"`
	CleanupSessions string              `json:"clean_up_sessions"`
	Limits          LimitsConfig        `json:"limits"`
	SystemImages    map[string]struct{} `json:"system_images"`
}

type WALConfig struct {
	AutoTruncate     bool   `json:"auto_truncate"`
	TruncateInterval string `json:"truncate_interval"`
	CacheSize        string `json:"cache_size"`
	Synchronous      string `json:"synchronous"`
}

type LimitsConfig struct {
	RowsLimit      int `json:"rows_limit"`
	MaxUsername    int `json:"max_username"`
	MinUsername    int `json:"min_username"`
	MaxPass        int `json:"max_pass"`
	MinPass        int `json:"min_pass"`
	MaxBio         int `json:"max_bio"`
	MaxFirstName   int `json:"max_first_name"`
	MinFirstName   int `json:"min_first_name"`
	MaxTitle       int `json:"max_title"`
	MinTitle       int `json:"min_title"`
	MaxCommentBody int `json:"max_comment_body"`
	MaxPostBody    int `json:"max_post_body"`
	MinBody        int `json:"min_body"`
	MaxCategories  int `json:"max_categories"`
}

type ServerConfig struct {
	Addr string `json:"Addr"`
}

type HandlersConfig struct {
	Image                 ImageConfig                `json:"image"`
	MaxPostSize           string                     `json:"max_post_size"`
	CookieExpirationHours string                     `json:"cookie_expiration_hours"`
	RateLimits            map[string]RateLimitConfig `json:"rate_limits"`
}

type ImageConfig struct {
	MaxSize    string   `json:"max_size"`
	FileTypes  []string `json:"file_types"`
	PathPrefix []string `json:"path_prefix"`
}

type RateLimitConfig struct {
	RateLimitCount    int     `json:"rate_limit_count"`
	RateLimitInterval float64 `json:"rate_limit_second_interval"`
}

type OAuthProvider struct {
	ClientID        string   `json:"client_id"`
	ClientSecret    string   `json:"client_secret"`
	Scopes          []string `json:"scopes"`
	AuthURL         string   `json:"auth_url"`
	TokenURL        string   `json:"token_url"`
	BaseRedirectURI string   `json:"base_redirect_uri"`
}

type OAuthConfig struct {
	Google *OAuthProvider
	Github *OAuthProvider
}

var (
	AppConfig   *Config
	GoogleOAuth *OAuthProvider
	GithubOAuth *OAuthProvider
	configMutex sync.RWMutex
)

func LoadConfig(configPath string) (*Config, error) {
	configMutex.Lock()
	defer configMutex.Unlock()

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed t open config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	if config.Frontend.URL == "" {
		config.Frontend.URL = "http://localhost:3000" // this is a default dev url
	}

	AppConfig = &config
	return &config, nil
}

func GetConfig() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return AppConfig
}

func GetFrontendURL() string {
	config := GetConfig()
	if config != nil && config.Frontend.URL != "" {
		return config.Frontend.URL
	}
	return "http://localhost:3000"
}

func GetRateLimit(path string) (int, float64) {
	config := GetConfig()
	if config == nil {
		return 20, 20.0
	}

	if limit, exists := config.Handlers.RateLimits[path]; exists {
		return limit.RateLimitCount, limit.RateLimitInterval
	}

	for configPath, limit := range config.Handlers.RateLimits {
		if strings.HasPrefix(path, configPath) {
			return limit.RateLimitCount, limit.RateLimitInterval
		}
	}

	if universal, exists := config.Handlers.RateLimits["universal"]; exists {
		return universal.RateLimitCount, universal.RateLimitInterval
	}

	return 20, 20.0
}

func GetUniversalRateLimit() (int, float64) {
	config := GetConfig()
	if config == nil {
		return 40, 2.0
	}

	if universal, exists := config.Handlers.RateLimits["universal"]; exists {
		return universal.RateLimitCount, universal.RateLimitInterval
	}

	return 40, 2.0
}

// InitOAuthConfig initializes the global OAuth configuration
func InitOAuthConfig(oauthConfig *OAuthConfig, useHTTPS bool) {
	if oauthConfig == nil {
		return
	}

	GoogleOAuth = &OAuthProvider{
		ClientID:        oauthConfig.Google.ClientID,
		ClientSecret:    oauthConfig.Google.ClientSecret,
		Scopes:          oauthConfig.Google.Scopes,
		AuthURL:         oauthConfig.Google.AuthURL,
		TokenURL:        oauthConfig.Google.TokenURL,
		BaseRedirectURI: oauthConfig.Google.BaseRedirectURI,
	}

	GithubOAuth = &OAuthProvider{
		ClientID:        oauthConfig.Github.ClientID,
		ClientSecret:    oauthConfig.Github.ClientSecret,
		Scopes:          oauthConfig.Github.Scopes,
		AuthURL:         oauthConfig.Github.AuthURL,
		TokenURL:        oauthConfig.Github.TokenURL,
		BaseRedirectURI: oauthConfig.Github.BaseRedirectURI,
	}
}

// ExchangeCodeForToken exchanges an authorization code for an access token
func (c *OAuthProvider) ExchangeCodeForToken(code string) (string, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("redirect_uri", c.getRedirectURL(true))
	data.Set("grant_type", "authorization_code")

	r, err := http.NewRequest("POST", c.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(r)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return "", fmt.Errorf("oauth: cannot fetch token %d\nresponse: %s", response.StatusCode, body)
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(response.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	if tokenResponse.AccessToken == "" || tokenResponse.TokenType == "" {
		return "", fmt.Errorf("oauth: missing token in response")
	}

	return tokenResponse.AccessToken, nil
}

// GetAuthURL returns the authorization URL for the OAuth provider
func (c *OAuthProvider) GetAuthURL(state string) string {
	u, _ := url.Parse(c.AuthURL)
	q := u.Query()
	q.Set("client_id", c.ClientID)
	q.Set("redirect_uri", c.getRedirectURL(true))
	q.Set("response_type", "code")
	q.Set("state", state)
	q.Set("scope", strings.Join(c.Scopes, " "))
	u.RawQuery = q.Encode()
	return u.String()
}

// getRedirectURL constructs the redirect URL (supports both HTTP and HTTPS)
func (c *OAuthProvider) getRedirectURL(useHTTPS bool) string {
	protocol := "http"
	if useHTTPS {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s", protocol, c.BaseRedirectURI)
}

func GetCookieExpiration() time.Duration {
	config := GetConfig()
	if config == nil {
		return 24 * time.Hour
	}

	duration, err := time.ParseDuration(config.Handlers.CookieExpirationHours)
	if err != nil {
		return 24 * time.Hour
	}
	return duration
}

func GetSessionCleanupInterval() time.Duration {
	config := GetConfig()
	if config == nil {
		return 10 * time.Minute
	}

	duration, err := time.ParseDuration(config.DatabaseConfiguration.CleanupSessions)
	if err != nil {
		return 10 * time.Minute
	}
	return duration
}
