package global

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	ShutDownContext context.Context
	CancelShutdown  context.CancelFunc
)

func Initialize() error {
	ShutDownContext, CancelShutdown = context.WithCancel(context.Background())
	log.Println("✅ Global context initialized")
	return nil
}

// Helper functions for commonly accessed configs

// GetDatabasePath returns the full database path
func GetDatabasePath(cfg DatabasePathConfig) string {
	if len(cfg.Path) == 0 {
		return filepath.Join("backend", "db", "social-network.db")
	}
	return filepath.Join(cfg.Path...)
}

// GetMigrationsPath returns the migrations path
func GetMigrationsPath(migrationsPath string) string {
	if migrationsPath == "" {
		return "file://backend/db/migrations/sqlite"
	}
	if !filepath.IsAbs(migrationsPath) && !hasPrefix(migrationsPath, "file://") {
		return "file://" + migrationsPath
	}
	return migrationsPath
}

// GetImagePath returns the full image storage path
func GetImagePath(pathPrefix []string) string {
	if len(pathPrefix) == 0 {
		return filepath.Join("backend", "data", "images")
	}
	return filepath.Join(pathPrefix...)
}

// GetCertPath returns the certificate file path
func GetCertPath(certFiles []string) string {
	if len(certFiles) == 0 {
		return ""
	}
	return filepath.Join(certFiles...)
}

// GetKeyPath returns the certificate key path
func GetKeyPath(keyfiles []string) string {
	if len(keyfiles) == 0 {
		return ""
	}
	return filepath.Join(keyfiles...)
}

// IsHTTPSEnabled returns whether HTTPS is configured and enabled
func IsHTTPSEnabled(useHTTPS bool, certFiles, keyfiles []string) bool {
	if !useHTTPS {
		return false
	}

	certPath := GetCertPath(certFiles)
	keyPath := GetKeyPath(keyfiles)

	if certPath == "" || keyPath == "" {
		return false
	}

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		log.Printf("Certificate file not found: %s", certPath)
		return false
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		log.Printf("Key file not found: %s", keyPath)
		return false
	}

	return true
}

// ParseDuration safely parses a duration string with a default fallback
func ParseDuration(durationStr string, defaultDuration time.Duration) time.Duration {
	if durationStr == "" {
		return defaultDuration
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Printf("Warning: failed to parse duration '%s', using default %v: %v",
			durationStr, defaultDuration, err)
		return defaultDuration
	}

	return duration
}

// ParseSize parses a size string like "20MB" to bytes
func ParseSize(sizeStr string) (int64, error) {
	var value float64
	var unit string

	_, err := fmt.Sscanf(sizeStr, "%f%s", &value, &unit)
	if err != nil {
		// Try parsing as just a number (bytes)
		var bytes int64
		_, err = fmt.Sscanf(sizeStr, "%d", &bytes)
		if err != nil {
			return 0, fmt.Errorf("invalid size format: %s", sizeStr)
		}
		return bytes, nil
	}

	switch unit {
	case "B":
		return int64(value), nil
	case "KB", "kB":
		return int64(value * 1024), nil
	case "MB", "mB":
		return int64(value * 1024 * 1024), nil
	case "GB", "gB":
		return int64(value * 1024 * 1024 * 1024), nil
	default:
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}
}

// DatabasePathConfig interface for getting database path
type DatabasePathConfig interface {
	GetPath() []string
}

// hasPrefix checks if string starts with prefix (to avoid importing strings)
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
