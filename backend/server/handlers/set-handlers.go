package handlers

import (
	"fmt"
	"net/http"
	"path"

	"social-network/backend/cache"
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

func SetHandlers(db *database.DataBase, redisClient *cache.RedisClient) *http.ServeMux {
	setGlobalRedis(redisClient)
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
		makeEndpoint("/api/auth/check", false, CheckAuthHandler),
		makeEndpoint("/api/logout", true, LogoutHandler),
		makeEndpoint("/api/logout-all", true, LogoutAllHandler),

		// Post endpoints
		makeEndpoint("/api/feed", true, GetFeedHandler),
		makeEndpoint("/api/posts", true, CreatePostHandler),
		makeEndpoint("/api/post/{id}", true, GetPostHandler),
		makeEndpoint("/api/posts/{id}", true, DeletePostHandler),
		makeEndpoint("/api/user/posts", true, GetUserPostsHandler),

		// Comment endpoints
		makeEndpoint("/api/comments", true, CreateCommentHandler),
		makeEndpoint("/api/posts/{id}/comments", false, GetPostCommentsHandler),
		makeEndpoint("/api/comments/{id}", true, DeleteCommentHandler),

		// Follow endpoints
		makeEndpoint("/api/follow/request", true, FollowRequestHandler),
		makeEndpoint("/api/follow/accept", true, AcceptFollowHandler),
		makeEndpoint("/api/follow/decline", true, DeclineFollowHandler),
		makeEndpoint("/api/follow/remove", true, UnfollowHandler),
		makeEndpoint("/api/followers", false, GetFollowersHandler),
		makeEndpoint("/api/following", false, GetFollowingHandler),
		makeEndpoint("/api/follow/pending", true, GetPendingFollowsHandler),

		// Group endpoints
		makeEndpoint("/api/groups", true, CreateGroupHandler),
		makeEndpoint("/api/groups/{id}", true, GetGroupHandler),
		makeEndpoint("/api/groups/{id}/update", true, UpdateGroupHandler),
		makeEndpoint("/api/groups/{id}/delete", true, DeleteGroupHandler),
		makeEndpoint("/api/groups/browse", false, BrowseGroupsHandler),
		makeEndpoint("/api/user/groups", false, GetUserGroupsHandler),
		makeEndpoint("/api/groups/{id}/invite", true, InviteToGroupHandler),
		makeEndpoint("/api/groups/{id}/join", true, RequestJoinGroupHandler),
		makeEndpoint("/api/groups/{id}/accept", true, AcceptGroupMemberHandler),
		makeEndpoint("/api/groups/{id}/reject", true, RejectGroupMemberHandler),
		makeEndpoint("/api/groups/{id}/leave", true, LeaveGroupHandler),
		makeEndpoint("/api/groups/{id}/members", false, GetGroupMembersHandler),
	}

	for _, ep := range endpoints {
		handler := ep.nextHandler
		mux.HandleFunc(ep.path, func(w http.ResponseWriter, r *http.Request) {
			AuthMiddleware(
				ep.requireAuth,
				w, r, db, redisClient,
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
