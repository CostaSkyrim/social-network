package global

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"social-network/backend/config"
	database "social-network/backend/db/sql"
	"social-network/backend/handlers"
	"time"
)

//g stands for global, this is a package meant to be accessed globally from the entire project
//should be read only, holding enums and configs

type Configuration struct {
	Database       *database.DBConfig     `json:"database_configuration"`
	Server         *http.Server           `json:"server"`
	Certifications *config.Certifications `json:"certifications"`
	Handlers       *handlers.Config       `json:"handlers"`
	OAuth          *config.OAuthConfig    `json:"oauth"`
}

var Configs *Configuration

var ShutDownContext context.Context
var CancelShutdown context.CancelFunc

func Initialize() error {
	var err error
	Configs, err = setUp()
	if err != nil {
		return fmt.Errorf("start: %w", err)
	}

	ShutDownContext, CancelShutdown = context.WithCancel(context.Background())
	setDerivedConfigs()

	if Configs.OAuth != nil {
		useHTTPS := IsHTTPSEnabled()
		config.InitOAuthConfig(Configs.OAuth, useHTTPS)
	}

	return nil
}

func setUp() (*Configuration, error) {
	errorMsg := "setUp: %w"
	data, err := os.ReadFile(filepath.Join("configs.json"))
	if err != nil {
		return nil, fmt.Errorf(errorMsg, err)
	}

	var cfg *Configuration
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf(errorMsg, err)
	}

	setDefaults(cfg)
	if cfg.Server != nil {
		setTimeouts(cfg.Server)
	}

	if err := validateConfigs(cfg); err != nil {
		return nil, fmt.Errorf(errorMsg, err)
	}

	return cfg, nil
}

func setDefaults(cfg *Configuration) {
	// Database defaults
	if cfg.Database != nil {
		if cfg.Database.MaxOpenConns == 0 {
			cfg.Database.MaxOpenConns = 25
		}
		if cfg.Database.SessionCleanupInt == "" {
			cfg.Database.SessionCleanupInt = "10m"
		}
		if cfg.Database.Limits.RowsLimit == 0 {
			cfg.Database.Limits.RowsLimit = 1000
		}
		if cfg.Database.Limits.MaxUserName == 0 {
			cfg.Database.Limits.MaxUserName = 25
		}
		if cfg.Database.Limits.MinUserName == 0 {
			cfg.Database.Limits.MinUserName = 4
		}
		if cfg.Database.Limits.MaxPass == 0 {
			cfg.Database.Limits.MaxPass = 40
		}
		if cfg.Database.Limits.MinPass == 0 {
			cfg.Database.Limits.MinPass = 7
		}
		if cfg.Database.Limits.MaxBio == 0 {
			cfg.Database.Limits.MaxBio = 1000
		}
		if cfg.Database.Limits.MaxFirstName == 0 {
			cfg.Database.Limits.MaxFirstName = 30
		}
		if cfg.Database.Limits.MaxLastName == 0 {
			cfg.Database.Limits.MaxLastName = 30
		}
		if cfg.Database.Limits.MaxTitle == 0 {
			cfg.Database.Limits.MaxTitle = 300
		}
		if cfg.Database.Limits.MinTitle == 0 {
			cfg.Database.Limits.MinTitle = 5
		}
		if cfg.Database.Limits.MaxCommentBody == 0 {
			cfg.Database.Limits.MaxCommentBody = 3000
		}
		if cfg.Database.Limits.MaxPostBody == 0 {
			cfg.Database.Limits.MaxPostBody = 7000
		}
		if cfg.Database.Limits.MinBody == 0 {
			cfg.Database.Limits.MinBody = 1
		}
		if cfg.Database.Limits.MaxCategories == 0 {
			cfg.Database.Limits.MaxCategories = 5
		}

		// WAL defaults
		if cfg.Database.WAL.CacheSize == "" {
			cfg.Database.WAL.CacheSize = "10000"
		}
		if cfg.Database.WAL.Synchronous == "" {
			cfg.Database.WAL.Synchronous = "NORMAL"
		}
		if cfg.Database.WAL.TruncateInterval == "" {
			cfg.Database.WAL.TruncateInterval = "5m"
		}

		// Cache defaults
		if cfg.Database.CacheSetup.UsersCacheLimit == 0 {
			cfg.Database.CacheSetup.UsersCacheLimit = 1000
		}
		if cfg.Database.CacheSetup.PostsCacheLimit == 0 {
			cfg.Database.CacheSetup.PostsCacheLimit = 500
		}
		if cfg.Database.CacheSetup.GroupsCacheLimit == 0 {
			cfg.Database.CacheSetup.GroupsCacheLimit = 200
		}

		// System images (if not provided)
		if cfg.Database.SystemImages == nil {
			cfg.Database.SystemImages = make(map[string]struct{})
			cfg.Database.SystemImages["default_avatar.png"] = struct{}{}
		}
	}

	// Server defaults
	if cfg.Server != nil {
		if cfg.Server.Addr == "" {
			cfg.Server.Addr = ":8080"
		}
	}

	// Certifications defaults (if nil, create empty struct)
	if cfg.Certifications == nil {
		cfg.Certifications = &config.Certifications{
			UseHTTPS: false,
			File:     []string{},
			Key:      []string{},
		}
	}

	// Handler defaults
	if cfg.Handlers != nil {
		if cfg.Handlers.CookieExpirationHours == "" {
			cfg.Handlers.CookieExpirationHours = "24h"
		}
		if cfg.Handlers.MaxPostSize == "" {
			cfg.Handlers.MaxPostSize = "21MB"
		}
		if cfg.Handlers.ImageConfig.MaxSize == "" {
			cfg.Handlers.ImageConfig.MaxSize = "20MB"
		}
		if len(cfg.Handlers.ImageConfig.FileTypes) == 0 {
			cfg.Handlers.ImageConfig.FileTypes = []string{"image/jpeg", "image/png", "image/gif"}
		}
		if len(cfg.Handlers.ImageConfig.PathPrefix) == 0 {
			cfg.Handlers.ImageConfig.PathPrefix = []string{"data", "images"}
		}
	}

	// OAuth defaults (if nil, create empty configs)
	if cfg.OAuth == nil {
		cfg.OAuth = &config.OAuthConfig{}
	}
}

func setDerivedConfigs() {
	if Configs.Database == nil {
		return
	}

	if Configs.Database.SessionCleanupInt != "" {
		duration, err := time.ParseDuration(Configs.Database.SessionCleanupInt)
		if err == nil {
			Configs.Database.SessionCleanupDuration = duration
		} else {
			Configs.Database.SessionCleanupDuration = 10 * time.Minute
		}
	}

	if Configs.Database.WAL.TruncateInterval != "" {
		duration, err := time.ParseDuration(Configs.Database.WAL.TruncateInterval)
		if err == nil {
			Configs.Database.WAL.TruncateIntervalDuration = duration
		} else {
			Configs.Database.WAL.TruncateIntervalDuration = 5 * time.Minute
		}
	}

	if Configs.Handlers != nil {
		if Configs.Handlers.CookieExpirationHours != "" {
			duration, err := time.ParseDuration(Configs.Handlers.CookieExpirationHours)
			if err == nil {
				Configs.Handlers.CookieExpirationDuration = duration
			} else {
				Configs.Handlers.CookieExpirationDuration = 24 * time.Hour
			}
		}

		if Configs.Handlers.MaxPostSize != "" {
			size, err := parseSize(Configs.Handlers.MaxPostSize)
			if err == nil {
				Configs.Handlers.MaxPostSizeBytes = size
			}
		}

		if Configs.Handlers.ImageConfig.MaxSize != "" {
			size, err := parseSize(Configs.Handlers.ImageConfig.MaxSize)
			if err == nil {
				Configs.Handlers.ImageConfig.MaxSizeBytes = size
			}
		}
	}
}

func validateConfigs(cfg *Configuration) error {
	if cfg.Database == nil || reflect.ValueOf(*cfg.Database).IsZero() {
		return fmt.Errorf("Database configuration is missing or empty")
	}
	if cfg.Server == nil || reflect.ValueOf(*cfg.Server).IsZero() {
		return fmt.Errorf("Server configuration is missing of empty")
	}

	if cfg.Certifications == nil || reflect.ValueOf(*cfg.Certifications).IsZero() {
		log.Printf("Certifications configuration is missing, HTTPS disabled")
		cfg.Certifications = &config.Certifications{UseHTTPS: false}
	}

	if cfg.Handlers == nil || reflect.ValueOf(*cfg.Handlers).IsZero() {
		return fmt.Errorf("Handlers configuration is missing or empty")
	}

	if cfg.OAuth == nil {
		log.Printf("OAuth configuration missing, social login disabled")
		cfg.OAuth = &config.OAuthConfig{}
	}

	return nil
}

func setTimeouts(server *http.Server) {
	server.ReadTimeout = 4 * time.Second
	server.WriteTimeout = 8 * time.Second
	server.IdleTimeout = 10 * time.Second
	server.ReadHeaderTimeout = 2 * time.Second
}

func parseSize(sizeStr string) (int64, error) {
	var multiplier int64 = 1
	var number int64

	_, err := fmt.Sscanf(sizeStr, "%d", &number)
	if err != nil {
		return 0, err
	}

	if len(sizeStr) > 2 {
		suffix := sizeStr[len(sizeStr)-2:]
		switch suffix {
		case "KB":
			multiplier = 1024
		case "MB":
			multiplier = 1024 * 1024
		case "GB":
			multiplier = 1024 * 1024 * 1024
		}
	}

	return number * multiplier, nil
}

// Helper functions for commonly accessed configs

// GetDatabasePath returns the full database path
func GetDatabasePath() string {
	if Configs.Database == nil || len(Configs.Database.Path) == 0 {
		return "backend/db/social-network.db"
	}
	return filepath.Join(Configs.Database.Path...)
}

// GetMigrationsPath returns the migrations path
func GetMigrationsPath() string {
	if Configs.Database == nil || Configs.Database.MigrationsPath == "" {
		return "file://backend/db/migrations/sqlite"
	}
	return Configs.Database.MigrationsPath
}

// GetImagePath returns the full image storage path
func GetImagePath() string {
	if Configs.Handlers == nil || len(Configs.Handlers.ImageConfig.PathPrefix) == 0 {
		return "backend/data/images"
	}
	return filepath.Join(Configs.Handlers.ImageConfig.PathPrefix...)
}

// GetCertPath returns the certificate file path
func GetCertPath() string {
	if Configs.Certifications == nil || len(Configs.Certifications.File) == 0 {
		return ""
	}
	return filepath.Join(Configs.Certifications.File...)
}

// GetKeyPath returns the certificate key path
func GetKeyPath() string {
	if Configs.Certifications == nil || len(Configs.Certifications.Key) == 0 {
		return ""
	}
	return filepath.Join(Configs.Certifications.Key...)
}

// IsHTTPSEnabled returns whether HTTPS is configured and enabled
func IsHTTPSEnabled() bool {
	if Configs.Certifications == nil {
		return false
	}

	// Check if HTTPS is requested and cert files exist
	if !Configs.Certifications.UseHTTPS {
		return false
	}

	certPath := GetCertPath()
	keyPath := GetKeyPath()

	if certPath == "" || keyPath == "" {
		return false
	}

	// Check if files exist
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return false
	}

	return true
}

// GetValidationLimits returns the validation limits for convenience
func GetValidationLimits() database.LimitsConfig {
	if Configs.Database == nil {
		return database.LimitsConfig{}
	}
	return Configs.Database.Limits
}
