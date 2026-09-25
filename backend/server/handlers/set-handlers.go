package handlers

import (
	"fmt"
	"net/http"
	"path"

	"social-network/backend/cache"
	"social-network/backend/config"
	database "social-network/backend/db/sql"
	"social-network/backend/global"
	ws "social-network/backend/server/websocket"
)

var GlobalHub *ws.Hub

type endpoint struct {
	path        string
	requireAuth bool // this will be true if the endpoint needs you to be logged in
	nextHandler func(http.ResponseWriter, *http.Request, *database.DataBase)
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

	hub := ws.NewHub(db, redisClient)
	GlobalHub = hub
	go hub.Run()

	go syncMapCleaner()

	mux := http.NewServeMux()

	staticDirCss := path.Join("web", "static")
	staticDirImg := global.GetImagePath(config.GetConfig().Handlers.Image.PathPrefix)
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

		// OAuth endpoints
		makeEndpoint("/api/auth/{provider}", false, OAuthLoginHandler),
		makeEndpoint("/api/auth/{provider}/callback", false, OAuthCallbackHandler),

		// Post endpoints
		makeEndpoint("/api/feed", true, GetFeedHandler),
		makeEndpoint("/api/feed/following", true, GetFollowingPostsHandler),
		makeEndpoint("/api/feed/explore", true, GetExplorePostsHandler),
		makeEndpoint("/api/posts", true, CreatePostHandler),
		makeEndpoint("/api/post/{id}", true, GetPostHandler),
		makeEndpoint("/api/posts/{id}", true, DeletePostHandler),
		makeEndpoint("/api/posts/{id}/edit", true, EditPostHandler),
		makeEndpoint("/api/user/posts", true, GetUserPostsHandler),

		// Comment endpoints
		makeEndpoint("/api/comments", true, CreateCommentHandler),
		makeEndpoint("/api/posts/{id}/comments", true, GetPostCommentsHandler),
		makeEndpoint("/api/comments/{id}", true, DeleteCommentHandler),
		makeEndpoint("/api/comments/{id}/edit", true, EditCommentHandler),

		// Follow endpoints
		makeEndpoint("/api/follow/request", true, FollowRequestHandler),
		makeEndpoint("/api/follow/accept", true, AcceptFollowHandler),
		makeEndpoint("/api/follow/decline", true, DeclineFollowHandler),
		makeEndpoint("/api/follow/remove", true, UnfollowHandler),
		makeEndpoint("/api/followers", false, GetFollowersHandler),
		makeEndpoint("/api/following", false, GetFollowingHandler),
		makeEndpoint("/api/follow/pending", true, GetPendingFollowsHandler),

		// User profile endpoints
		makeEndpoint("/api/users/{id}", false, GetUserProfileHandler),
		makeEndpoint("/api/users/{id}/edit", true, UpdateUserProfileHandler),
		makeEndpoint("/api/users/{id}/avatar", true, UpdateUserAvatarHandler),
		makeEndpoint("/api/users/{id}/posts", true, GetUserPostsForViewerHandler),
		makeEndpoint("/api/users/search", true, SearchUsersHandler),

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
		makeEndpoint("/api/groups/{id}/posts", true, GetGroupPostsHandler),
		makeEndpoint("/api/groups/{id}/avatar", true, UpdateGroupAvatarHandler),
		makeEndpoint("/api/groups/{id}/messages", true, GetGroupMessagesHandler),
		makeEndpoint("/api/groups/{id}/messages/send", true, SendGroupMessageHandler),

		// Events endpoint
		makeEndpoint("/api/events", true, GetUserGroupEventsHandler),

		// Notification endpoints
		makeEndpoint("/api/notifications", true, GetNotificationsHandler),
		makeEndpoint("/api/notifications/unread-count", true, GetUnreadNotificationCountHandler),
		makeEndpoint("/api/notifications/{id}/read", true, MarkNotificationReadHandler),
		makeEndpoint("/api/notifications/read-all", true, MarkAllNotificationsReadHandler),

		// Chat/DM endpoints
		makeEndpoint("/api/chat/dms", true, GetDMsHandler),
		makeEndpoint("/api/chat/dms/{id}/messages", true, GetMessagesHandler),
		makeEndpoint("/api/chat/send/{id}", true, SendMessageHandler),
		makeEndpoint("/api/chat/unread-count", true, GetUnreadMessageCountHandler),

		// Event endpoints
		makeEndpoint("/api/groups/{id}/events", true, GroupEventsHandler),
		makeEndpoint("/api/events/{id}", true, GetEventHandler),
		makeEndpoint("/api/events/{id}/rsvp", true, EventRSVPHandler),
		makeEndpoint("/api/events/{id}/edit", true, UpdateEventHandler),
		makeEndpoint("/api/events/{id}/delete", true, DeleteEventHandler),
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

	// WebSocket endpoint
	mux.HandleFunc("/api/ws", func(w http.ResponseWriter, r *http.Request) {
		AuthMiddleware(
			true,
			w, r, db, redisClient,
			func(w http.ResponseWriter, r *http.Request, db *database.DataBase) {
				userID, ok := GetUserIDFromContext(r)
				if !ok {
					RespondError(w, http.StatusUnauthorized, "Authentication required")
					return
				}
				userUUID, _ := GetUserUUIDFromContext(r)
				ws.ServeWS(GlobalHub, w, r, userID, userUUID)
			},
		)
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
