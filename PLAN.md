# Social Network — Plan

## Tech Stack

### Frontend
- Build: Vite
- Framework: React 18+ with TypeScript
- Server state: TanStack Query v5
- Routing: React Router v6 (createBrowserRouter)
- Client state: React Context (useAuth + useUI, no external dependencies)
- HTTP: Axios with interceptors
- Styling: TailwindCSS
- Real-time: Raw WebSocket (shared connection for notifications + chat)

### Backend
- Language: Go 1.24
- Database: SQLite (users, posts, comments, groups, messages, etc.)
- Cache / ephemeral storage: Redis 7 (sessions, WebSocket pub/sub)
- Migrations: golang-migrate
- Real-time: WebSocket hub with Redis pub/sub

## Project Structure
```
social-network/
├── frontend/
│   ├── Dockerfile
│   ├── .dockerignore
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── tailwind.config.ts
│   ├── postcss.config.js
│   ├── nginx.conf
│   ├── public/
│   ├── src/
│   │   ├── main.tsx
│   │   ├── App.tsx
│   │   ├── routes.tsx
│   │   ├── types/
│   │   │   ├── user.ts
│   │   │   ├── post.ts
│   │   │   ├── comment.ts
│   │   │   ├── group.ts
│   │   │   ├── message.ts
│   │   │   ├── notification.ts
│   │   │   └── api.ts
│   │   ├── api/
│   │   │   ├── client.ts
│   │   │   ├── auth.ts
│   │   │   ├── users.ts
│   │   │   ├── posts.ts
│   │   │   ├── comments.ts
│   │   │   ├── groups.ts
│   │   │   ├── messages.ts
│   │   │   ├── notifications.ts
│   │   │   └── images.ts
│   │   ├── hooks/
│   │   │   ├── useAuth.ts
│   │   │   ├── useProfile.ts
│   │   │   ├── usePosts.ts
│   │   │   ├── useComments.ts
│   │   │   ├── useGroups.ts
│   │   │   ├── useMessages.ts
│   │   │   ├── useNotifications.ts
│   │   │   └── useFollowers.ts
│   │   ├── ws/
│   │   │   └── websocket.ts
│   │   ├── context/
│   │   │   └── AuthProvider.tsx
│   │   ├── components/
│   │   │   ├── ui/
│   │   │   │   ├── Button.tsx
│   │   │   │   ├── Input.tsx
│   │   │   │   ├── Avatar.tsx
│   │   │   │   ├── Card.tsx
│   │   │   │   ├── Modal.tsx
│   │   │   │   ├── Dropdown.tsx
│   │   │   │   ├── Spinner.tsx
│   │   │   │   ├── Toast.tsx
│   │   │   │   ├── Tabs.tsx
│   │   │   │   └── EmojiPicker.tsx
│   │   │   ├── layout/
│   │   │   │   ├── MainLayout.tsx
│   │   │   │   ├── AuthLayout.tsx
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   ├── TopBar.tsx
│   │   │   │   └── MobileNav.tsx
│   │   │   ├── post/
│   │   │   │   ├── PostCard.tsx
│   │   │   │   ├── PostForm.tsx
│   │   │   │   ├── PostList.tsx
│   │   │   │   ├── PrivacySelector.tsx
│   │   │   │   └── VoteButtons.tsx
│   │   │   ├── comment/
│   │   │   │   ├── CommentItem.tsx
│   │   │   │   ├── CommentForm.tsx
│   │   │   │   └── CommentList.tsx
│   │   │   ├── group/
│   │   │   │   ├── GroupCard.tsx
│   │   │   │   ├── GroupList.tsx
│   │   │   │   ├── GroupForm.tsx
│   │   │   │   ├── MemberList.tsx
│   │   │   │   ├── EventCard.tsx
│   │   │   │   └── EventForm.tsx
│   │   │   ├── chat/
│   │   │   │   ├── ChatList.tsx
│   │   │   │   ├── ChatWindow.tsx
│   │   │   │   ├── MessageBubble.tsx
│   │   │   │   └── MessageInput.tsx
│   │   │   ├── notification/
│   │   │   │   ├── NotificationList.tsx
│   │   │   │   └── NotificationBell.tsx
│   │   │   ├── user/
│   │   │   │   ├── UserCard.tsx
│   │   │   │   ├── UserList.tsx
│   │   │   │   └── FollowButton.tsx
│   │   │   └── common/
│   │   │       ├── LoadingScreen.tsx
│   │   │       ├── LoadingSkeleton.tsx
│   │   │       ├── EmptyState.tsx
│   │   │       ├── ErrorBoundary.tsx
│   │   │       ├── InfiniteScroll.tsx
│   │   │       └── ImageUpload.tsx
│   │   ├── pages/
│   │   │   ├── auth/
│   │   │   │   ├── LoginPage.tsx
│   │   │   │   └── SignupPage.tsx
│   │   │   ├── home/
│   │   │   │   └── HomePage.tsx
│   │   │   ├── profile/
│   │   │   │   ├── ProfilePage.tsx
│   │   │   │   └── EditProfilePage.tsx
│   │   │   ├── post/
│   │   │   │   └── PostDetailPage.tsx
│   │   │   ├── groups/
│   │   │   │   ├── GroupsPage.tsx
│   │   │   │   ├── GroupDetailPage.tsx
│   │   │   │   └── CreateGroupPage.tsx
│   │   │   ├── chat/
│   │   │   │   └── ChatPage.tsx
│   │   │   ├── notifications/
│   │   │   │   └── NotificationsPage.tsx
│   │   │   ├── followers/
│   │   │   │   ├── FollowersPage.tsx
│   │   │   │   └── FollowingPage.tsx
│   │   │   └── search/
│   │   │       └── SearchPage.tsx
│   │   ├── lib/
│   │   │   ├── cn.ts
│   │   │   ├── format.ts
│   │   │   └── validators.ts
│   │   └── styles/
│   │       └── globals.css
│   └── .env
├── backend/
│   ├── Dockerfile
│   ├── redis/
│   │   ├── session_store.go     # Redis-backed session CRUD (replaces SQLite sessions)
│   │   └── ws_pubsub.go         # Redis pub/sub for WebSocket hub
│   ├── server/
│   │   └── handlers/
│   │       └── hub.go           # WebSocket hub (Redis pub/sub aware)
│   └── ...
├── docker-compose.yml           # Orchestrates backend + frontend + Redis
└── PLAN.md
```

## Routing
```
/public                     # AuthLayout
  /login                    # LoginPage
  /signup                   # SignupPage

/                           # MainLayout (authenticated)
  /home                     # HomePage (news feed)
  /profile/:uuid            # ProfilePage
  /profile/me               # EditProfilePage
  /posts/:uuid              # PostDetailPage
  /groups                   # GroupsPage
  /groups/new               # CreateGroupPage
  /groups/:uuid             # GroupDetailPage
  /chat                     # ChatPage
  /chat/:id                 # ChatPage (open DM/group)
  /notifications            # NotificationsPage
  /followers                # FollowersPage
  /following                # FollowingPage
  /search                   # SearchPage
```

## Docker

Three containers orchestrated via `docker-compose.yml`:

```yaml
services:
  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
    volumes: ["redis_data:/data"]

  backend:
    build: ./backend
    ports: ["8080:8080"]
    environment:
      REDIS_ADDR: redis:6379
    depends_on: [redis]

  frontend:
    build: ./frontend
    ports: ["80:80"]
    depends_on: [backend]
```

**Backend container:**
- Go binary serving API on `:8080`
- Connects to Redis at `REDIS_ADDR` for session lookups and WebSocket pub/sub
- Needs CORS middleware allowing frontend origin

**Frontend container:**
- `npm run build` → static files in `frontend/dist/`
- Served by nginx on `:80`

**Redis container:**
- Stores session keys with TTL (auto-expire, no cleanup needed)
- WebSocket pub/sub channels for cross-instance message broadcasting
- Persists to `redis_data` volume for session durability across restarts

**Important — static file serving in dev vs prod:**
- In `backend/server/handlers/set-handlers.go:50-56`, the backend serves static files from `web/static/`, `web/images/`, `web/js/` — useful in dev without Docker
- In production with Docker, the frontend nginx container handles static files; the backend's static file handlers can be removed or kept for dev convenience

## Redis Integration

### What goes into Redis

| Data | Why Redis | Current storage |
|------|-----------|-----------------|
| **Sessions** (`session_id → user_id`) | Ephemeral key-value with TTL, auto-expire, faster than SQLite, removes periodic cleanup | SQLite `sessions` table with manual cleanup every 10min |
| **WebSocket pub/sub** (channels) | Enables broadcasting across multiple backend instances; single instance still benefits from the pattern | Does not exist yet |

### What stays in SQLite

All persistent data: users, posts, comments, groups, messages, followers, notifications, events. Redis never holds the canonical copy of any business data.

### Session flow

```
Login:
  bcrypt(password) → match user in SQLite → generate UUID →
  Redis SETEX session:{uuid} 86400 {user_id} → Set cookie

Request (every handler):
  Read cookie → Redis GET session:{uuid} → if nil, 401 →
  otherwise look up user by ID in SQLite (can be cached later)

Logout:
  Redis DEL session:{uuid} → Clear cookie
```

- TTL = 24h, matches `cookie_expiration_hours` config
- No cleanup routine needed — Redis evicts expired keys automatically
- User data stays in SQLite; Redis only maps `session_id → user_id`

### WebSocket pub/sub flow

```
Client A sends message to Client B:
  Backend receives message →
  Redis PUBLISH ws:messages {channel_id, payload} →
  All backend instances receive the message via Redis SUBSCRIBE →
  Instance holding Client B's WebSocket connection delivers it
```

**Channels:**
- `ws:notifications:{user_id}` — per-user notification events
- `ws:messages:{conversation_id}` — per-conversation chat messages
- `ws:presence` — online/offline status updates

### Go dependency

```go
go get github.com/redis/go-redis/v9
```

### Backend package structure

```
backend/redis/
├── session_store.go      # SessionRedis struct: Create, Get, Delete with TTL
└── ws_pubsub.go          # PubSubRedis struct: Publish, Subscribe channels

backend/server/handlers/
└── hub.go                # WebSocket hub managing connections + Redis pub/sub
```

### Migration path (zero-downtime)

1. Add Redis package with the same interface as current SQLite session methods
2. Wire a config toggle: `session_storage: "sqlite" | "redis"` (default `sqlite`)
3. Test Redis path locally
4. Swap default to `redis`
5. Remove SQLite session table + cleanup routine (cleanup migration)

## Feature Implementation Order

**Phase 1 — Foundation:** ✅
1. Vite + React + TS + Tailwind scaffold
2. Project directory structure
3. All TypeScript type definitions
4. Axios client with auth interceptor
5. Auth context (useAuth — user, login, logout)
6. UI context (useUI — sidebar, toasts, modals)
7. React Router setup with AuthLayout/MainLayout
8. UI primitives (Button, Input, Avatar, Card, Modal, Dropdown, Spinner, Toast, Tabs, Badge, EmojiPicker)
9. Common components (LoadingScreen, LoadingSkeleton, EmptyState, ErrorBoundary, InfiniteScroll, ImageUpload)
10. Layout components (AuthLayout, MainLayout, Sidebar, TopBar, MobileNav)

**Phase 2 — Auth:**
11. api/auth.ts — login, signup, logout
12. hooks/useAuth.ts — mutations + session check query
13. LoginPage + SignupPage with validation
14. OAuth buttons (Google, GitHub)

**Phase 3 — Feed & Posts:**
15. api/posts.ts — getFeed, createPost, getPost, deletePost, votePost
16. hooks/usePosts.ts — useFeed, usePost, useCreatePost, useDeletePost, useVotePost
17. PostCard, PostForm, PostList, VoteButtons, PrivacySelector
18. HomePage with infinite-scroll feed + create post form
19. PostDetailPage

**Phase 4 — Comments:**
20. api/comments.ts — create, list, delete
21. hooks/useComments.ts
22. CommentList, CommentItem (nested replies), CommentForm
23. Wire into PostDetailPage

**Phase 5 — Profiles & Followers:**
24. api/users.ts — getProfile, updateBio, uploadAvatar
25. hooks/useProfile.ts
26. api/followers.ts — follow, unfollow, list, accept/decline
27. hooks/useFollowers.ts
28. ProfilePage with tabs
29. EditProfilePage
30. FollowButton, UserCard, UserList
31. FollowersPage, FollowingPage
32. SearchPage

**Phase 6 — Groups:**
33. api/groups.ts — CRUD, members, events
34. hooks/useGroups.ts
35. GroupsPage, CreateGroupPage, GroupDetailPage
36. GroupForm, EventForm, MemberList, EventCard, GroupCard, GroupList

**Phase 7 — Notifications:**
37. api/notifications.ts — list, mark read, unread count
38. hooks/useNotifications.ts
39. NotificationBell in TopBar
40. NotificationsPage
41. Shared WebSocket connection — receives notification events in real time

**Phase 8 — Chat:**
42. api/messages.ts — conversations, messages, send, mark read
43. hooks/useMessages.ts
44. ChatPage with sidebar (DM list) + active chat window
45. ChatList, ChatWindow, MessageBubble, MessageInput (with emoji picker)
46. Shared WebSocket connection — receives incoming messages in real time
47. Chat access constraint: only allow DM if one user follows the other (or recipient has public profile)

## Backend API Endpoints (from configs.json rate limits)
```
POST   /signup
POST   /login
POST   /logout
GET    /auth/google
GET    /auth/google/callback
GET    /auth/github
GET    /auth/github/callback
GET    /profile/:uuid
POST   /profilepic
POST   /updatebio
GET    /post/:uuid
GET    /postlist
POST   /createpost
POST   /postedit
POST   /postdelete
POST   /postremove
POST   /postapprove
POST   /votepost
GET    /commentlist
POST   /createcomment
POST   /commentedit
POST   /commentdelete
POST   /commentremove
POST   /commentapprove
POST   /votecomment
GET    /image/:path
GET    /notificationfeed
GET    /notificationlist
POST   /notificationseen
WS     /ws/notifications
POST   /addcategory
POST   /removecategory
GET    /removedposts
GET    /removedcomments
```

## Key Conventions

- **Auth:** Session cookie with `withCredentials: true`. Hydrate auth store on mount. Sessions stored in Redis with 24h TTL instead of SQLite.
- **Privacy levels:**
  - `public` — visible to everyone
  - `almost_private` — visible only to followers of the post author
  - `private` — visible only to specific followers chosen by the author
- **Follows:** Public profile → instant follow. Private profile → request must be sent, recipient accepts or declines.
- **Chat access:** Users can only start a DM if at least one of them follows the other (or the recipient has a public profile).
- **Groups:** Invite/request with statuses `pending`, `accepted`, `declined`, `invited`. Only accepted members see group content and chat. Group events have RSVP options: `going`, `not_going`, `maybe`.
- **WebSocket:** single shared connection at `/ws/notifications` — delivers both notification events and incoming chat messages. Backed by Redis pub/sub for cross-instance message delivery.
- **Emojis:** supported in chat messages via an emoji picker component.
- **Pagination:** `page`/`limit` params. TanStack `useInfiniteQuery`.
- **Images:** multipart/form-data, max 20MB, jpg/png/gif.
- **Validation:** Mirror backend limits client-side.
- **CORS:** Backend currently has NO CORS middleware (see `set-handlers.go`). Must be added — set `Access-Control-Allow-Origin`, `Allow-Credentials`, handle OPTIONS preflight.
- **Port:** Backend `:8080`, frontend dev `:5173` (Vite default).
- **snake_case** everywhere to match Go JSON tags.
- **File naming:** React components are PascalCase; all other files (hooks, api, utils, pages) are camelCase.
