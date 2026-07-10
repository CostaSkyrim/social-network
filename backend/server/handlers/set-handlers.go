package handlers

import (
	"fmt"
	"net/http"
	"path"

	"social-network/backend/config"
	database "social-network/backend/db/sql"
)

type endpoint struct {
	path                 string
	rateLimitMaxRequests int
	rateLimitInterval    float64
	requireAuth          bool // this will be true if the endpoint needs you to be logged in
	nextHandler          func(http.ResponseWriter, *http.Request, *database.DataBase)
}

func makeEndpoint(path string, requireAuth bool, nextHandler func(http.ResponseWriter, *http.Request, *database.DataBase)) endpoint {
	return endpoint{
		path:        path,
		requireAuth: requireAuth,
		nextHandler: nextHandler,
	}
}

func SetHandlers(db *database.DataBase) *http.ServeMux {
	// hub := GetWebSocketHub(db)
	// go hub.Run()

	go syncMapCleaner()

	mux := http.NewServeMux()

	staticDirCss := path.Join("web", "static")
	staticDirImg := path.Join("web", "images")
	staticDirJs := path.Join("web", "js")

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDirCss))))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(staticDirImg))))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(staticDirJs))))

	endpoints := []endpoint{
		// Authentication endpoints
		makeEndpoint("/api/signup", false, SignupHandler),
		makeEndpoint("/api/login", false, LoginHandler),
		makeEndpoint("/api/logout", true, LogoutHandler),
		makeEndpoint("/api/logout-all", true, LogoutAllHandler),
	}

	for _, ep := range endpoints {
		handler := ep.nextHandler
		mux.HandleFunc(ep.path, func(w http.ResponseWriter, r *http.Request) {
			AuthMiddleware(
				ep.requireAuth,
				w, r, db,
				handler,
			)
		})
	}

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		RespondSuccess(w, http.StatusOK, "Server is running", map[string]string{
			"version": "1.0.0",
			"status":  "healthy",
		})
	})

	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			frontendURL := config.GetFrontendURL()
			w.Header().Set("Access-Control-Allow-Origin", frontendURL)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		RespondError(w, http.StatusNotFound, "Endpoint not found")
	})

	fmt.Println("✅ Routes registered successfully")
	return mux
}
