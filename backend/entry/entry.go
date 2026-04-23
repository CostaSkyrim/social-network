package entry

// Called by main, begins the server start up sequence

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	//NOTE: The file names and structure is still in progress
	//g "social-network/global"
	//"social-network/persistence/database"
	//"social-network/persistence/populate"
	//"social-network/server/core/config"
	//"social-network/server/core/handlers"
)

// server starting sequence
func Start() {
	err := g.Initialize()
	if err != nil {
		log.Fatal("Error with global config initialization:", err.Error())
	}

	// Populate the db with mock data
	//populate.StartProcedure()

	// Setup database
	db, shutDownDb, err := setupDatabase()
	if err != nil {
		log.Fatal(err)
	}

	// Assign configs to handlers
	handlers.Configuration = g.Configs.Handlers

	server := g.Configs.Server
	server.Handler = handlers.SetHandlers(db)

	// Configure TLS
	useHTTPS, certFile, certKey, err := configureTLS(server)
	if err != nil {
		log.Fatal(err)
	}

	config.InitOAuthConfig(g.Configs.OAuth, useHTTPS)

	startServer(server, useHTTPS, certFile, certKey)

	// Wait here for process termination signal to initiate graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	// Graceful shutdown
	gracefulShutdown(db, server, shutDownDb)
}

// setupDatabase initializes the database and context
func setupDatabase() (*sql.DB, context.CancelFunc, error) {
	dbCtx, shutDownDb := context.WithCancel(context.Background())
	g.Configs.Database.Ctx = dbCtx

	log.Println("Starting social-network db")

	db, err := database.Open(g.Configs.Database)
	if err != nil {
		return nil, nil, fmt.Errorf("error initializing database: %w", err)
	}

	return db, shutDownDb, nil
}

// configureTLS validates certificates and configures TLS settings
func configureTLS(server *http.Server) (bool, string, string, error) {
	useHTTPS := g.Configs.Certifications.UseHTTPS
	certFile := filepath.Join(g.Configs.Certifications.File...)
	certKey := filepath.Join(g.Configs.Certifications.Key...)

	// Use HTTP if there's no SSL keys
	_, err := os.Stat(certFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("SSL certs not found in %s. Running in HTTP.\n", certFile)
			useHTTPS = false
		} else {
			return false, "", "", fmt.Errorf("error checking certificates: %w", err)
		}
	}

	if useHTTPS {
		server.TLSConfig = &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
			PreferServerCipherSuites: true,
			InsecureSkipVerify:       true,
		}
	}

	return useHTTPS, certFile, certKey, nil
}

// startServer launches the HTTP or HTTPS server
func startServer(server *http.Server, useHTTPS bool, certFile, certKey string) {
	go func() {
		if useHTTPS {
			log.Printf("Server running on https://localhost%s\n", server.Addr)
			if err := server.ListenAndServeTLS(certFile, certKey); err != nil && err != http.ErrServerClosed {
				log.Fatalf("ListenAndServeTLS failed: %v", err)
			}
		} else {
			log.Printf("Server running on http://localhost%s\n", server.Addr)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("ListenAndServe() failed: %v", err)
			}
		}
	}()
}

// gracefulShutdown handles clean database and server shutdown
func gracefulShutdown(db *sql.DB, server *http.Server, shutDownDb context.CancelFunc) {
	shutDownDb()
	if g.Configs.Database.Wal.AutoTruncate {
		database.ManualTruncate <- struct{}{}
		database.Wg.Wait()
	}
	db.Close()

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Graceful server Shutdown Failed: %v", err)
	}
	log.Println("Server stopped")
}
