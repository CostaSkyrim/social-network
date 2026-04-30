package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"social-network/backend/db/queries"
	"strings"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type DBConfig struct {
	// Basic settings
	Ctx                    context.Context `json:"-"`
	Path                   []string        `json:"path"`
	MigrationsPath         string          `json:"migrations_path"`
	MaxOpenConns           int             `json:"max_open_conns"`
	SessionCleanupInt      string          `json:"clean_up_sessions"`
	SessionCleanupDuration time.Duration   `json:"-"`

	// WAL
	WAL WALConfig `json:"wal"`

	// Cache configuration
	UseCache   bool        `json:"use_cache"`
	CacheSetup CacheConfig `json:"cache_setup"`

	// Validation Limits
	Limits LimitsConfig `json:"limits"`

	// System Images
	SystemImages map[string]struct{} `json:"system_images"`
}

type CacheConfig struct {
	UsersCacheLimit  int `json:"users_cache_limit"`
	PostsCacheLimit  int `json:"posts_cache_limit"`
	GroupsCacheLimit int `json:"groups_cache_limit"`
}

type LimitsConfig struct {
	RowsLimit      int `json:"rows_limit"`
	MaxUserName    int `json:"max_username"`
	MinUserName    int `json:"min_username"`
	MaxPass        int `json:"max_pass"`
	MinPass        int `json:"mins_pass"`
	MaxBio         int `json:"max_bio"`
	MaxFirstName   int `json:"max_first_name"`
	MaxLastName    int `json:"max_last_name"`
	MaxTitle       int `json:"max_title"`
	MinTitle       int `json:"min_title"`
	MaxCommentBody int `json:"max_comment_body"`
	MaxPostBody    int `json:"max_post_body"`
	MinBody        int `json:"min_body"`
	MaxCategories  int `json:"max_categories"`
}

type DataBase struct {
	conn         *sql.DB
	mu           sync.RWMutex
	cfg          *DBConfig
	Queries      *Queries
	useCache     bool
	SystemImages map[string]struct{}
}

type Queries struct {
	Users         UserQueries
	Sessions      SessionQueries
	Follows       FollowQueries
	Posts         PostQueries
	Groups        GroupQueries
	Messages      MessageQueries
	Events        EventQueries
	Notifications NotificationQueries
	Comments      CommentQueries
}

// New opens a connection to the SQLite database, applies migrations, and returns a DB instance
func New(ctx context.Context, cfg *DBConfig) (*DataBase, error) {
	dbPath := strings.Join(cfg.Path, string([]rune{filepath.Separator}))

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
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

	if err := configurePragmas(conn, cfg); err != nil {
		return nil, fmt.Errorf("Failed to configure SQLite pragmas: %w", err)
	}

	migrationsPath := cfg.MigrationsPath
	if migrationsPath == "" {
		migrationsPath = "file://backend/db/migrations/sqlite"
	}

	if err := applyMigrations(migrationsPath, dbPath); err != nil {
		return nil, err
	}

	db := &DataBase{
		conn:         conn,
		cfg:          cfg,
		useCache:     cfg.UseCache,
		SystemImages: cfg.SystemImages,
	}

	if err := db.prepareQueries(ctx); err != nil {
		return nil, err
	}

	go db.sessionCleanupRoutine(ctx)

	if cfg.WAL.AutoTruncate {
		go db.walTruncateRoutine(ctx)
	}

	log.Printf("✅ Database initialized successfuly")
	return db, nil
}

func configurePragmas(conn *sql.DB, cfg *DBConfig) error {
	// Enable WAL
	if _, err := conn.Exec("PRAGMA journal mode=WAL;"); err != nil {
		return err
	}

	// set Cache size
	if cfg.WAL.CacheSize != "" {
		if _, err := conn.Exec("PRAGMA cache_size=" + cfg.WAL.CacheSize + ";"); err != nil {
			return err
		}
	}

	if _, err := conn.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		return err
	}

	if _, err := conn.Exec("PRAGMA busy_timeout=5000;"); err != nil {
		return err
	}

	return nil
}

func (db *DataBase) walTruncateRoutine(ctx context.Context) {
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
			if _, err := db.conn.Exec("Pragma wal_checkpoint(TRUNCATE);"); err != nil {
				log.Printf("WAL truncate error: %v", err)
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
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("✅ Migrations applied successfully")
	return nil
}

func (db *DataBase) prepareQueries(ctx context.Context) error {
	db.Queries = &Queries{}
	var err error

	// User queries
	db.Queries.Users.Create, err = db.conn.PrepareContext(ctx, queries.CreateUser)
	if err != nil {
		return err
	}

	db.Queries.Users.GetByEmail, err = db.conn.PrepareContext(ctx, queries.GetUserByEmail)
	if err != nil {
		return err
	}

	db.Queries.Users.GetByID, err = db.conn.PrepareContext(ctx, queries.GetUserByID)
	if err != nil {
		return err
	}

	db.Queries.Users.GetByUUID, err = db.conn.PrepareContext(ctx, queries.GetUserByUUID)
	if err != nil {
		return err
	}

	db.Queries.Users.UpdatePrivacy, err = db.conn.PrepareContext(ctx, queries.UpdateUserPrivacy)
	if err != nil {
		return err
	}

	db.Queries.Users.UpdateProfile, err = db.conn.PrepareContext(ctx, queries.UpdateUserProfile)
	if err != nil {
		return err
	}

	// Session queries
	db.Queries.Sessions.Create, err = db.conn.PrepareContext(ctx, queries.CreateSession)
	if err != nil {
		return err
	}

	db.Queries.Sessions.Get, err = db.conn.PrepareContext(ctx, queries.GetSession)
	if err != nil {
		return err
	}

	db.Queries.Sessions.Delete, err = db.conn.PrepareContext(ctx, queries.DeleteSession)
	if err != nil {
		return err
	}

	db.Queries.Sessions.Cleanup, err = db.conn.PrepareContext(ctx, queries.CleanupSessions)
	if err != nil {
		return err
	}

	db.Queries.Sessions.GetUserBySession, err = db.conn.PrepareContext(ctx, queries.GetUserBySession)
	if err != nil {
		return err
	}

	// Follow queries
	db.Queries.Follows.Create, err = db.conn.PrepareContext(ctx, queries.CreateFollowRequest)
	if err != nil {
		return err
	}

	db.Queries.Follows.Update, err = db.conn.PrepareContext(ctx, queries.UpdateFollowStatus)
	if err != nil {
		return err
	}

	db.Queries.Follows.GetRequest, err = db.conn.PrepareContext(ctx, queries.GetFollowRequest)
	if err != nil {
		return err
	}

	db.Queries.Follows.GetFollowers, err = db.conn.PrepareContext(ctx, queries.GetFollowers)
	if err != nil {
		return err
	}

	db.Queries.Follows.GetFollowing, err = db.conn.PrepareContext(ctx, queries.GetFollowing)
	if err != nil {
		return err
	}

	db.Queries.Follows.Check, err = db.conn.PrepareContext(ctx, queries.CheckFollowing)
	if err != nil {
		return err
	}

	db.Queries.Follows.GetPending, err = db.conn.PrepareContext(ctx, queries.GetPendingFollowRequests)
	if err != nil {
		return err
	}

	// Post queries
	db.Queries.Posts.Create, err = db.conn.PrepareContext(ctx, queries.CreatePost)
	if err != nil {
		return err
	}

	db.Queries.Posts.GetByID, err = db.conn.PrepareContext(ctx, queries.GetPostByID)
	if err != nil {
		return err
	}

	db.Queries.Posts.GetFeed, err = db.conn.PrepareContext(ctx, queries.GetFeed)
	if err != nil {
		return err
	}

	db.Queries.Posts.Delete, err = db.conn.PrepareContext(ctx, queries.DeletePost)
	if err != nil {
		return err
	}

	db.Queries.Posts.AddVisibility, err = db.conn.PrepareContext(ctx, queries.AddPostVisibility)
	if err != nil {
		return err
	}

	db.Queries.Posts.GetVisible, err = db.conn.PrepareContext(ctx, queries.GetPostVisibleUsers)
	if err != nil {
		return err
	}

	db.Queries.Posts.RemoveVisibility, err = db.conn.PrepareContext(ctx, queries.RemovePostVisibility)
	if err != nil {
		return err
	}

	db.Queries.Posts.GetUserPosts, err = db.conn.PrepareContext(ctx, queries.GetUserPosts)
	if err != nil {
		return err
	}

	// Comments queries
	db.Queries.Comments.Create, err = db.conn.PrepareContext(ctx, queries.CreateComment)
	if err != nil {
		return err
	}

	db.Queries.Comments.GetPostComments, err = db.conn.PrepareContext(ctx, queries.GetPostComments)
	if err != nil {
		return err
	}

	db.Queries.Comments.Delete, err = db.conn.PrepareContext(ctx, queries.DeleteComment)
	if err != nil {
		return err
	}

	// Group queries
	db.Queries.Groups.Create, err = db.conn.PrepareContext(ctx, queries.CreateGroup)
	if err != nil {
		return err
	}

	db.Queries.Groups.GetByID, err = db.conn.PrepareContext(ctx, queries.GetGroupByID)
	if err != nil {
		return err
	}

	db.Queries.Groups.GetUserGroups, err = db.conn.PrepareContext(ctx, queries.GetUserGroups)
	if err != nil {
		return err
	}

	db.Queries.Groups.AddMember, err = db.conn.PrepareContext(ctx, queries.AddGroupMember)
	if err != nil {
		return err
	}

	db.Queries.Groups.UpdateStatus, err = db.conn.PrepareContext(ctx, queries.UpdateMemberStatus)
	if err != nil {
		return err
	}

	db.Queries.Groups.GetMembers, err = db.conn.PrepareContext(ctx, queries.GetGroupMembers)
	if err != nil {
		return err
	}

	db.Queries.Groups.GetAllGroups, err = db.conn.PrepareContext(ctx, queries.GetAllGroups)
	if err != nil {
		return err
	}

	// Message queries
	db.Queries.Messages.GetOrCreate, err = db.conn.PrepareContext(ctx, queries.GetOrCreateDM)
	if err != nil {
		return err
	}

	db.Queries.Messages.GetByID, err = db.conn.PrepareContext(ctx, queries.GetDMid)
	if err != nil {
		return err
	}

	db.Queries.Messages.Create, err = db.conn.PrepareContext(ctx, queries.CreateMessage)
	if err != nil {
		return err
	}

	db.Queries.Messages.GetGroup, err = db.conn.PrepareContext(ctx, queries.GetGroupMessages)
	if err != nil {
		return err
	}

	db.Queries.Messages.GetAll, err = db.conn.PrepareContext(ctx, queries.GetAllDMs)
	if err != nil {
		return err
	}

	db.Queries.Messages.GetPrivate, err = db.conn.PrepareContext(ctx, queries.GetPrivateMessages)
	if err != nil {
		return err
	}

	db.Queries.Messages.MarkRead, err = db.conn.PrepareContext(ctx, queries.MarkMessageRead)
	if err != nil {
		return err
	}

	db.Queries.Messages.UpdateMessage, err = db.conn.PrepareContext(ctx, queries.UpdateLastMessage)
	if err != nil {
		return err
	}

	db.Queries.Messages.UpdateDM, err = db.conn.PrepareContext(ctx, queries.UpdateLastDM)
	if err != nil {
		return err
	}

	// Notification queries
	db.Queries.Notifications.Create, err = db.conn.PrepareContext(ctx, queries.CreateNotification)
	if err != nil {
		return err
	}

	db.Queries.Notifications.Get, err = db.conn.PrepareContext(ctx, queries.GetUserNotifications)
	if err != nil {
		return err
	}

	db.Queries.Notifications.MarkRead, err = db.conn.PrepareContext(ctx, queries.MarkNotificationsAsRead)
	if err != nil {
		return err
	}

	db.Queries.Notifications.GetUnread, err = db.conn.PrepareContext(ctx, queries.GetUnreadCount)
	if err != nil {
		return err
	}

	// Event queries
	db.Queries.Events.Create, err = db.conn.PrepareContext(ctx, queries.CreateEvent)
	if err != nil {
		return err
	}

	db.Queries.Events.GetGroupEvents, err = db.conn.PrepareContext(ctx, queries.GetGroupEvents)
	if err != nil {
		return err
	}

	db.Queries.Events.RSVP, err = db.conn.PrepareContext(ctx, queries.CreateEventRSVP)
	if err != nil {
		return err
	}

	return nil
}

func (db *DataBase) sessionCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(db.cfg.SessionCleanupDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := db.Queries.Sessions.Cleanup.ExecContext(ctx); err != nil {
				log.Printf("Session cleanup error: %v", err)
			}
		}
	}
}

// Gracefuly closes the DB connection
func (db *DataBase) Close() error {
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
