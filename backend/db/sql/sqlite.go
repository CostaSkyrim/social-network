package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"

	"social-network/backend/cache"
)

type DataBase struct {
	conn         *sql.DB
	mu           sync.RWMutex
	cfg          *DBConfig
	SystemImages map[string]struct{}
	wg           sync.WaitGroup
	redis        *cache.RedisClient
}

// SetRedis attaches an optional Redis client used for read-through caching of
// users, posts, and groups. It is nil-safe: caching is skipped when unset.
func (db *DataBase) SetRedis(rc *cache.RedisClient) {
	db.redis = rc
}

type DBConfig struct {
	// Basic settings
	Path                   []string      `json:"path"`
	MigrationsPath         string        `json:"migrations_path"`
	MaxOpenConns           int           `json:"max_open_conns"`
	SessionCleanupDuration time.Duration `json:"-"`

	// WAL
	WAL WALConfig `json:"wal"`

	// System Images
	SystemImages map[string]struct{} `json:"system_images"`
}

// WALConfig holds Write-Ahead Log configuration
type WALConfig struct {
	AutoTruncate             bool          `json:"auto_truncate"`
	TruncateIntervalDuration time.Duration `json:"-"`
	CacheSize                string        `json:"cache_size"`
	Synchronous              string        `json:"synchronous"`
}

// New opens a connection to the SQLite database, applies migrations, and returns a DB instance
func New(ctx context.Context, cfg *DBConfig) (*DataBase, error) {
	dbPath := strings.Join(cfg.Path, string([]rune{filepath.Separator}))

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("Failed to create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(cfg.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.MaxOpenConns / 2)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.PingContext(ctx); err != nil {
		return nil, err
	}

	if err := configurePragmas(conn, cfg.WAL); err != nil {
		return nil, fmt.Errorf("Failed to configure SQLite pragmas: %w", err)
	}

	migrationsPath := cfg.MigrationsPath
	if migrationsPath == "" {
		migrationsPath = "file://backend/db/migrations"
	}

	if err := applyMigrations(migrationsPath, dbPath); err != nil {
		return nil, err
	}

	db := &DataBase{
		conn:         conn,
		cfg:          cfg,
		SystemImages: cfg.SystemImages,
	}

	db.wg.Add(1)
	go db.sessionCleanupRoutine(ctx)

	if cfg.WAL.AutoTruncate {
		db.wg.Add(1)
		go db.walTruncateRoutine(ctx)
	}

	log.Printf("✅ Database initialized successfuly")
	return db, nil
}

// configurePragmas sets SQLite PRAGMA statements for optimal performance
func configurePragmas(conn *sql.DB, wal WALConfig) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA busy_timeout=5000;",
	}

	if wal.CacheSize != "" {
		pragmas = append(pragmas, "PRAGMA cache_size="+wal.CacheSize+";")
	}

	if wal.Synchronous != "" {
		pragmas = append(pragmas, "PRAGMA synchronous="+wal.Synchronous+";")
	}

	for _, pragma := range pragmas {
		if _, err := conn.Exec(pragma); err != nil {
			return fmt.Errorf("failed to execute %s: %w", pragma, err)
		}
	}

	return nil
}

func (db *DataBase) walTruncateRoutine(ctx context.Context) {
	defer db.wg.Done()

	if db.cfg.WAL.TruncateIntervalDuration == 0 {
		db.cfg.WAL.TruncateIntervalDuration = 5 * time.Minute
	}

	ticker := time.NewTicker(db.cfg.WAL.TruncateIntervalDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := db.conn.Exec("PRAGMA wal_checkpoint(TRUNCATE);"); err != nil {
				log.Printf("WAL truncate error: %v", err)
			}
		}
	}
}

// sessionCleanupRoutine periodically cleans up expired sessions
func (db *DataBase) sessionCleanupRoutine(ctx context.Context) {
	defer db.wg.Done()

	if db.cfg.SessionCleanupDuration == 0 {
		db.cfg.SessionCleanupDuration = 10 * time.Minute
	}

	ticker := time.NewTicker(db.cfg.SessionCleanupDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := db.conn.ExecContext(ctx,
				`UPDATE sessions SET is_active = 0 WHERE expires_at < CURRENT_TIMESTAMP`); err != nil {
				log.Printf("Session cleanup error: %v", err)
			}
		}
	}
}

// applyMigrations runs database migrations
func applyMigrations(migrationsPath, dbPath string) error {
	if !strings.HasPrefix(migrationsPath, "file://") {
		migrationsPath = "file://" + migrationsPath
	}

	m, err := migrate.New(
		migrationsPath,
		"sqlite3://"+dbPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("✅ Migrations applied successfully")
	return nil
}

// Gracefuly closes the DB connection
func (db *DataBase) Close() error {
	log.Println("Waiting for background DB routines to finish...")
	db.wg.Wait()

	log.Println("Closing database connection...")
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.conn.Close()
}

// GetDB returns the underlying sql.DB for advanced operations
func (db *DataBase) GetDB() *sql.DB {
	return db.conn
}

// IsSystemImage checks if an image is a system image (can't be deleted)
func (db *DataBase) IsSystemImage(imagePath string) bool {
	if db.SystemImages == nil {
		return false
	}

	_, exists := db.SystemImages[imagePath]
	return exists
}
