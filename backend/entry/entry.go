package entry

// Called by main, begins the server start up sequence

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// NOTE: The file names and structure is still in progress
	"social-network/backend/config"
	database "social-network/backend/db/sql"
	"social-network/backend/global"
	"social-network/backend/populate"
	"social-network/backend/server/handlers"
	//"social-network/persistence/populate"
	//"social-network/server/core/handlers"
)

// server starting sequence
func Start(reseed bool) error {
	err := global.Initialize()
	if err != nil {
		log.Fatal("Error with global config initialization:", err.Error())
	}

	cfg, err := config.LoadConfig("configs.json")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Setup database
	db, err := setupDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}
	defer db.Close()

	// Seed database with sample data on first run (or reseed if flag is set)
	if reseed {
		if err := populate.Reseed(db, populate.DefaultPath()); err != nil {
			log.Printf("Warning: reseed failed: %v", err)
		}
	} else {
		if _, err := populate.SeedFromJSON(db, populate.DefaultPath()); err != nil {
			log.Printf("Warning: seed failed: %v", err)
		}
	}

	server := setupServer(cfg, db)

	// Configure TLS
	useHTTPS := global.IsHTTPSEnabled(
		cfg.Certifications.UseHTTPS,
		cfg.Certifications.File,
		cfg.Certifications.Key,
	)

	// cfg.InitOAuthConfig(global.Configs.OAuth, useHTTPS)

	go startServer(server, useHTTPS, cfg)

	return waitForShutdown(server, db)
}

// setupDatabase initializes the database and context
func setupDatabase(cfg *config.Config) (*database.DataBase, error) {
	dbPath := global.GetDatabasePath(cfg.DatabaseConfiguration.Path)

	sessionCleanupDuration := global.ParseDuration(
		cfg.DatabaseConfiguration.CleanupSessions,
		10*time.Minute,
	)

	walTruncateDuration := global.ParseDuration(
		cfg.DatabaseConfiguration.WAL.TruncateInterval,
		5*time.Minute,
	)

	dbConfig := &database.DBConfig{
		Path:                   cfg.DatabaseConfiguration.Path,
		MaxOpenConns:           25,
		SessionCleanupDuration: sessionCleanupDuration,
		WAL: database.WALConfig{
			AutoTruncate:             cfg.DatabaseConfiguration.WAL.AutoTruncate,
			TruncateIntervalDuration: walTruncateDuration,
			CacheSize:                cfg.DatabaseConfiguration.WAL.CacheSize,
			Synchronous:              cfg.DatabaseConfiguration.WAL.Synchronous,
		},
		SystemImages: cfg.DatabaseConfiguration.SystemImages,
	}

	log.Printf("Initializing database at: %s", dbPath)

	db, err := database.New(global.ShutDownContext, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("error initializing database: %w", err)
	}

	return db, nil
}

func setupServer(cfg *config.Config, db *database.DataBase) *http.Server {
	handler := handlers.SetHandlers(db)

	server := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           handler,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	return server
}

// startServer launches the HTTP or HTTPS server
func startServer(server *http.Server, useHTTPS bool, cfg *config.Config) {
	var err error

	if useHTTPS {
		certFile := global.GetCertPath(cfg.Certifications.File)
		keyFile := global.GetKeyPath(cfg.Certifications.Key)

		server.TLSConfig = &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
			PreferServerCipherSuites: true,
		}

		log.Printf("🚀 Server starting on https://localhost%s", server.Addr)
		err = server.ListenAndServeTLS(certFile, keyFile)
	} else {
		log.Printf("🚀 Server starting on http://localhost%s", server.Addr)

		err = server.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func waitForShutdown(server *http.Server, db *database.DataBase) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("Received signal: %v. Starting graceful shutdown...", sig)

	global.CancelShutdown()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("Shutting down HTTP server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Closing database connection...")
	if err := db.Close(); err != nil {
		return fmt.Errorf("database close failed: %w", err)
	}

	log.Println("Server shutdown complete")
	return nil
}
