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
- Database: SQLite (all persistent data)
- Cache / ephemeral storage: Redis 7 (sessions, WebSocket pub/sub)
- Migrations: golang-migrate
- Real-time: WebSocket hub with Redis pub/sub (planned)

## Project Structure
```
social-network/
├── frontend/
│   ├── Dockerfile              # not yet created
│   ├── .dockerignore           # not yet created
│   ├── nginx.conf              # not yet created
│   ├── index.html              ✅
│   ├── vite.config.ts          ✅
│   ├── tsconfig.json           ✅
│   ├── tailwind.config.ts      ✅
│   ├── postcss.config.js       ✅
│   ├── package.json            ✅
│   ├── .env                    ✅
│   ├── public/
│   ├── src/
│   │   ├── main.tsx            ✅
│   │   ├── App.tsx             ✅
│   │   ├── routes.tsx          ✅
│   │   ├── vite-env.d.ts       ✅
│   │   ├── types/              ✅ (7 files)
│   │   ├── api/
│   │   │   └── client.ts       ✅ (Axios instance only — no endpoint functions yet)
│   │   ├── context/
│   │   │   ├── AuthProvider.tsx ✅
│   │   │   └── UIProvider.tsx   ✅
│   │   ├── components/
│   │   │   ├── ui/             ✅ (Button, Input, Avatar, Card, Modal, Dropdown, Spinner, Toast, Tabs, Badge)
│   │   │   ├── layout/         ✅ (AuthLayout, MainLayout, Sidebar, TopBar, MobileNav)
│   │   │   └── common/         ✅ (LoadingScreen, LoadingSkeleton, EmptyState, ErrorBoundary, InfiniteScroll, ImageUpload)
│   │   ├── lib/
│   │   │   ├── cn.ts           ✅
│   │   │   ├── format.ts       ✅
│   │   │   └── validators.ts   ✅
│   │   └── styles/
│   │       └── globals.css     ✅
├── backend/
│   ├── Dockerfile              # not yet created
│   ├── cmd/main.go             ✅ (reads --reseed flag)
│   ├── entry/entry.go          ✅ (startup, DB init, seeding, server)
│   ├── config/config.go        ✅ (config structs, LoadConfig, rate limits, OAuth)
│   ├── configs.json            ✅ (all settings)
│   ├── global/global.go        ✅ (helper functions for paths, durations, TLS)
│   ├── server/handlers/
│   │   ├── set-handlers.go     ✅ (route registration, CORS, static files)
│   │   ├── middleware.go        ✅ (AuthMiddleware, rate limiter, CORS headers)
│   │   ├── handler-utils.go    ✅ (JSON response helpers, session cookie mgmt)
│   │   ├── event-signup.go     ✅ (POST /api/signup with validation)
│   │   └── event-login.go      ✅ (POST /api/login, POST /api/logout, POST /api/logout-all)
│   ├── db/
│   │   ├── migrations/         ✅ (28 migration files — 14 tables)
│   │   ├── queries/            ✅ (all SQL constants)
│   │   ├── sql/
│   │   │   ├── models.go       ✅ (all structs)
│   │   │   ├── methods.go      ⚠️ (partial — see below)
│   │   │   └── sqlite.go       ✅ (DB init, migrations, WAL, session cleanup)
│   │   └── tables/             ✅ (reference schemas)
│   ├── populate/
│   │   ├── seed.json           ✅ (6 users, 17 posts, 16 comments, 2 groups, etc.)
│   │   └── seed.go             ✅ (loader with bcrypt hashing, first-run check, --reseed)
│   └── redis/                  # not yet created
├── docker-compose.yml          # not yet created
├── Makefile                    ✅
├── setup-dev.sh                ✅
└── PLAN.md                     ✅ (this file)
```

## Current Completion Status

### ✅ Backend — Complete

| Component | Status |
|-----------|--------|
| Go module + dependencies (`go.mod`) | ✅ |
| Database migrations (14 tables up/down) | ✅ |
| All SQL query constants (`internal_queries.go`) | ✅ |
| All data models (`models.go`) | ✅ |
| Database init + migration runner (`sqlite.go`) | ✅ |
| Database methods (`methods.go`) | ⚠️ Partial |
| Config loading + types (`config.go`) | ✅ |
| Config JSON (`configs.json`) | ✅ |
| Global helper functions (`global.go`) | ✅ |
| Auth middleware + rate limiter + CORS | ✅ |
| Session cookie utilities | ✅ |
| Signup handler (validation + bcrypt + session) | ✅ |
| Login handler (email/password check + session) | ✅ |
| Logout handler (delete session + clear cookie) | ✅ |
| Logout-all handler (delete all user sessions) | ✅ |
| Route registration (`set-handlers.go`) | ✅ |
| Seed data JSON + loader | ✅ |
| Health endpoint (`/api/health`) | ✅ |
| `main.go` with `--reseed` flag | ✅ |
| `entry.go` startup sequence | ✅ |
| Makefile | ✅ |
| setup-dev.sh (distrobox container) | ✅ |

### ✅ Frontend — Complete

| Component | Status |
|-----------|--------|
| Vite + React + TS + Tailwind scaffold | ✅ |
| Type definitions (7 files) | ✅ |
| Axios client with 401 interceptor | ✅ |
| Auth context (user state + session check) | ✅ |
| UI context (sidebar, toasts, modals) | ✅ |
| React Router with all routes | ✅ |
| UI primitives (10 components) | ✅ |
| Common components (6 components) | ✅ |
| Layout components (5 components) | ✅ |
| Page stubs (14 pages) | ✅ |
| Utility functions (cn, format, validators) | ✅ |

#### Database Methods (methods.go) — Status Detail

| Method | Status |
|--------|--------|
| AddUser, GetUserByEmail, GetUserByID, GetUserByUUID | ✅ |
| UpdateUserProfile, UpdateUserPrivacy, DeleteUser | ✅ |
| CreateSession, GetSession, DeleteSession, DeleteAllUserSessions | ✅ |
| CreatePost, GetPost, GetUserPosts, DeletePost | ✅ |
| CreateGroup, GetGroup, GetUserGroups | ✅ |
| CreateMessage, CreateOrGetDirectMessage | ✅ |
| CreateNotification, GetUserNotifications, MarkNotificationAsRead | ✅ |

| Method | Status |
|--------|--------|
| CreateFollow/UpdateFollowStatus/GetFollowRequest | ❌ Not in methods.go |
| GetFollowers/GetFollowing/CheckFollowing/GetPending | ❌ Not in methods.go |
| GetFeed (paginated feed query) | ❌ Not in methods.go |
| AddPostVisibility/GetPostVisibleUsers/RemoveVisibility | ❌ Not in methods.go |
| AddGroupMember/UpdateMemberStatus/GetGroupMembers/GetAllGroups | ❌ Not in methods.go |
| CreateComment/GetPostComments/DeleteComment | ❌ Not in methods.go |
| CreateEvent/CreateEventRSVP/GetGroupEvents | ❌ Not in methods.go |
| GetGroupMessages/GetAllDMs/GetPrivateMessages | ❌ Not in methods.go |
| MarkMessageRead/UpdateLastMessage/UpdateLastDM | ❌ Not in methods.go |
| GetUnreadCount (notifications) | ❌ Not in methods.go |

## Implementation Plan

### Convention
- **Backend phases** are lettered (A, B, C...)
- **Frontend phases** are numbered (1, 2, 3...)
- ✅ = done, 🔜 = next up, ⬜ = planned

---

## Backend Phases

### Phase A — Auth Handlers ✅
| Task | File | Status |
|------|------|--------|
| Signup handler with validation + bcrypt | `event-signup.go` | ✅ |
| Login handler with session creation | `event-login.go` | ✅ |
| Logout handler with session deletion | `event-login.go` | ✅ |
| Session middleware + rate limiter | `middleware.go` | ✅ |
| Cookie/session utilities | `handler-utils.go` | ✅ |
| Route registration + CORS | `set-handlers.go` | ✅ |
| Config loading + global helpers | `config.go`, `global.go` | ✅ |
| Startup sequence + seeding | `entry.go` | ✅ |

### Phase B — DB Methods 🔜
Implement the missing DB methods in `methods.go`:

| Task | Query constant |
|------|----------------|
| CreateFollowRequest, UpdateFollowStatus, GetFollowRequest | `queries.CreateFollowRequest`, etc. |
| GetFollowers, GetFollowing, CheckFollowing, GetPending | `queries.GetFollowers`, etc. |
| GetFeed (paginated news feed) | `queries.GetFeed` |
| AddPostVisibility, GetPostVisibleUsers, RemoveVisibility | `queries.AddPostVisibility`, etc. |
| AddGroupMember, UpdateMemberStatus, GetGroupMembers, GetAllGroups | `queries.AddGroupMember`, etc. |
| CreateComment, GetPostComments, DeleteComment | `queries.CreateComment`, etc. |
| CreateEvent, CreateEventRSVP, GetGroupEvents | `queries.CreateEvent`, etc. |
| GetGroupMessages, GetAllDMs, GetPrivateMessages | `queries.GetGroupMessages`, etc. |
| MarkMessageRead, UpdateLastMessage, UpdateLastDM | `queries.MarkMessageRead`, etc. |
| GetUnreadCount (notifications) | `queries.GetUnreadCount` |

### Phase C — Profile & Follow Handlers 🔜
| Task | Endpoint |
|------|----------|
| GET /api/profile/:uuid — view user profile | `/profile/:uuid` (from configs.json) |
| POST /api/profilepic — upload avatar | `/profilepic` |
| POST /api/updatebio — update biography | `/updatebio` |
| POST /api/follow — send follow request | `/follow` |
| POST /api/unfollow — unfollow user | `/unfollow` |
| GET /api/followers — list followers | `/followers` |
| GET /api/following — list following | `/following` |
| GET /api/follow-requests — pending requests | `/follow-requests` |
| POST /api/follow-accept — accept request | `/follow-accept` |
| POST /api/follow-decline — decline request | `/follow-decline` |
| GET /api/search?q= — search users | `/search` |

### Phase D — Post & Comment Handlers 🔜
| Task | Endpoint |
|------|----------|
| GET /api/postlist — paginated feed | `/postlist` |
| GET /api/post/:uuid — single post | `/post/:uuid` |
| POST /api/createpost — create post | `/createpost` |
| POST /api/postedit — edit post | `/postedit` |
| POST /api/postdelete — delete post | `/postdelete` |
| POST /api/votepost — upvote/downvote post | `/votepost` |
| GET /api/commentlist — list comments | `/commentlist` |
| POST /api/createcomment — create comment | `/createcomment` |
| POST /api/commentedit — edit comment | `/commentedit` |
| POST /api/commentdelete — delete comment | `/commentdelete` |
| POST /api/votecomment — vote on comment | `/votecomment` |
| POST /api/image — upload image | `/image/` |
| GET /api/image/:path — serve image | `/image/:path` |

### Phase E — Group Handlers 🔜
| Task | Endpoint |
|------|----------|
| POST /api/creategroup — create group | `/creategroup` |
| GET /api/groups — list all groups | `/groups` |
| GET /api/group/:uuid — group detail | `/group/:uuid` |
| POST /api/group-join — request to join | `/group-join` |
| POST /api/group-invite — invite member | `/group-invite` |
| POST /api/group-accept — accept invite | `/group-accept` |
| POST /api/group-decline — decline invite | `/group-decline` |
| GET /api/group-members — list members | `/group-members` |
| POST /api/group-post — create group post | `/group-post` |

### Phase F — Event Handlers 🔜
| Task | Endpoint |
|------|----------|
| POST /api/createevent — create event | `/createevent` |
| GET /api/events — list group events | `/events` |
| POST /api/event-rsvp — respond to event | `/event-rsvp` |

### Phase G — Notification & Chat Handlers 🔜
| Task | Endpoint |
|------|----------|
| GET /api/notificationlist — list notifications | `/notificationlist` |
| POST /api/notificationseen — mark read | `/notificationseen` |
| GET /api/unread-count — unread badge count | `/unread-count` |
| GET /api/conversations — list DMs | `/conversations` |
| GET /api/messages/:id — get messages | `/messages/:id` |
| POST /api/send-message — send DM | `/send-message` |
| GET /api/group-messages/:id — group chat | `/group-messages/:id` |

### Phase H — WebSocket & Real-time 🔜
| Task | File |
|------|------|
| WebSocket hub (connection manager) | `hub.go` |
| Connection upgrade handler | `ws/notifications` |
| Notification push (new follow, group invite, event) | WS message types |
| Chat message delivery | WS message types |
| Online/offline presence | WS message types |

### Phase I — Redis Integration 🔜
| Task | File |
|------|------|
| Redis session store (replace SQLite sessions) | `redis/session_store.go` |
| Redis pub/sub for WebSocket hub | `redis/ws_pubsub.go` |
| Docker Compose with Redis service | `docker-compose.yml` |
| Config toggle: session_storage (sqlite/redis) | `configs.json` + `config.go` |

### Phase J — Docker 🔜
| Task | File |
|------|------|
| Backend Dockerfile (multi-stage Go build) | `backend/Dockerfile` |
| Frontend Dockerfile (nginx static serve) | `frontend/Dockerfile` |
| Frontend nginx config | `frontend/nginx.conf` |
| Docker Compose (backend + frontend + redis) | `docker-compose.yml` |

---

## Frontend Phases

### Phase 1 — Foundation ✅
| Task | Status |
|------|--------|
| Vite + React + TS + Tailwind scaffold | ✅ |
| All TypeScript types (7 files) | ✅ |
| Axios client with auth interceptor | ✅ |
| Auth context (useAuth hook — user, loading, auth guard) | ✅ |
| UI context (useUI hook) | ✅ |
| React Router with all routes + 404 catch-all | ✅ |
| UI primitives (10 components) | ✅ |
| Common components (6 components) | ✅ |
| Layout components (5 components) | ✅ |
| Page stubs (14 pages) | ✅ |
| 404 page (NotFoundPage) | ✅ |
| Error page (ErrorPage — route-level errors) | ✅ |
| Error boundary (ErrorBoundary — render crash recovery) | ✅ |
| Auth guard on MainLayout (redirects to /login) | ✅ |
| Auth guard on AuthLayout (redirects to /home) | ✅ |

### Phase 2 — Wire Auth 🔜
| Task | Files | Depends on |
|------|-------|------------|
| api/auth.ts — login, signup, logout, getProfile | `src/api/auth.ts` | Backend Phase A ✅ |
| hooks/useAuth.ts — mutations + session check | `src/hooks/useAuth.ts` | api/auth.ts |
| LoginPage — form with validation | `src/pages/auth/LoginPage.tsx` | useAuth |
| SignupPage — form with validation | `src/pages/auth/SignupPage.tsx` | useAuth |
| OAuth buttons (Google, GitHub) | LoginPage/SignupPage | Backend OAuth handlers |

### Phase 3 — Feed & Posts 🔜
| Task | Files | Depends on |
|------|-------|------------|
| api/posts.ts — getFeed, createPost, getPost, deletePost, votePost | `src/api/posts.ts` | Backend Phase D |
| hooks/usePosts.ts | `src/hooks/usePosts.ts` | api/posts |
| PostCard, PostForm, PostList, VoteButtons, PrivacySelector | `src/components/post/` | hooks |
| HomePage — infinite-scroll feed + create post | `src/pages/home/HomePage.tsx` | hooks + components |
| PostDetailPage — single post + comments | `src/pages/post/PostDetailPage.tsx` | hooks + components |

### Phase 4 — Comments (if separate from posts) 🔜
| Task | Files | Depends on |
|------|-------|------------|
| api/comments.ts | `src/api/comments.ts` | Backend Phase D |
| hooks/useComments.ts | `src/hooks/useComments.ts` | api/comments |
| CommentList, CommentItem, CommentForm | `src/components/comment/` | hooks |
| Wire comments into PostDetailPage | — | Phase 3 + Phase 4 |

### Phase 5 — Profiles & Followers 🔜
| Task | Files | Depends on |
|------|-------|------------|
| api/users.ts — getProfile, updateBio, uploadAvatar | `src/api/users.ts` | Backend Phase C |
| api/followers.ts — follow, unfollow, requests | `src/api/followers.ts` | Backend Phase C |
| hooks/useProfile.ts, hooks/useFollowers.ts | `src/hooks/` | api |
| ProfilePage with tabs (posts/about/followers) | `src/pages/profile/ProfilePage.tsx` | hooks + components |
| EditProfilePage | `src/pages/profile/EditProfilePage.tsx` | hooks |
| FollowButton, UserCard, UserList | `src/components/user/` | hooks |
| FollowersPage, FollowingPage | `src/pages/followers/` | hooks |
| SearchPage | `src/pages/search/SearchPage.tsx` | hooks |

### Phase 6 — Groups 🔜
| Task | Files | Depends on |
|------|-------|------------|
| api/groups.ts | `src/api/groups.ts` | Backend Phase E |
| hooks/useGroups.ts | `src/hooks/useGroups.ts` | api/groups |
| GroupCard, GroupList, GroupForm | `src/components/group/` | hooks |
| MemberList, EventCard, EventForm | `src/components/group/` | hooks |
| GroupsPage, CreateGroupPage, GroupDetailPage | `src/pages/groups/` | hooks + components |

### Phase 7 — Notifications + WebSocket 🔜
| Task | Files | Depends on |
|------|-------|------------|
| api/notifications.ts | `src/api/notifications.ts` | Backend Phase G |
| hooks/useNotifications.ts | `src/hooks/useNotifications.ts` | api |
| NotificationBell (TopBar) + NotificationsPage | components + pages | hooks |
| ws/websocket.ts — shared connection manager | `src/ws/websocket.ts` | Backend Phase H |
| WebSocket event handling (notifications + chat) | ws/websocket.ts | Phase H |
| Auto-reconnect with exponential backoff | ws/websocket.ts | — |

### Phase 8 — Chat 🔜
| Task | Files | Depends on |
|------|-------|------------|
| api/messages.ts | `src/api/messages.ts` | Backend Phase G |
| hooks/useMessages.ts | `src/hooks/useMessages.ts` | api |
| ChatPage with sidebar (DM list) + chat window | `src/pages/chat/ChatPage.tsx` | hooks + components |
| ChatList, ChatWindow, MessageBubble, MessageInput (emoji picker) | `src/components/chat/` | hooks |
| Real-time message delivery via WebSocket | ws/websocket.ts + chat components | Phase H |
| Chat access constraint (must be following) | hooks/messages.ts | Backend Phase C |

### Phase 9 — Docker & Polish 🔜
| Task | Depends on |
|------|------------|
| Frontend Dockerfile + nginx.conf | All frontend phases |
| Production build testing | — |
| Responsive polish, error states | — |

## Backend API Endpoints

### Registered (with handler)
| Method | Path | Handler | Auth |
|--------|------|---------|------|
| POST | `/api/signup` | SignupHandler | ❌ |
| POST | `/api/login` | LoginHandler | ❌ |
| POST | `/api/logout` | LogoutHandler | ✅ |
| POST | `/api/logout-all` | LogoutAllHandler | ✅ |
| GET | `/api/health` | inline | ❌ |
| OPTIONS | `/api` | inline | ❌ |

### Planned (from `configs.json` rate limits — handlers not yet created)
| Method | Path | Phase |
|--------|------|-------|
| GET | `/profile/:uuid` | C |
| POST | `/profilepic` | C |
| POST | `/updatebio` | C |
| GET | `/postlist` | D |
| GET | `/post/:uuid` | D |
| POST | `/createpost` | D |
| POST | `/postedit` | D |
| POST | `/postdelete` | D |
| POST | `/votepost` | D |
| GET | `/commentlist` | D |
| POST | `/createcomment` | D |
| POST | `/commentedit` | D |
| POST | `/commentdelete` | D |
| POST | `/votecomment` | D |
| GET | `/image/:path` | D |
| GET | `/notificationlist` | G |
| POST | `/notificationseen` | G |
| WS | `/ws/notifications` | H |
| POST | `/addcategory` | Moderation |
| POST | `/removecategory` | Moderation |
| GET | `/removedposts` | Moderation |
| GET | `/removedcomments` | Moderation |
| POST | `/postapprove` | Moderation |
| POST | `/commentapprove` | Moderation |
| POST | `/postremove` | Moderation |
| POST | `/commentremove` | Moderation |

## Key Conventions

- **Auth:** Session cookie with `withCredentials: true`. Hydrate auth context on mount via `GET /api/health` (or profile endpoint).
- **Privacy levels:**
  - `public` — visible to everyone
  - `almost_private` (followers) — visible only to followers of the post author
  - `private` — visible only to specific followers chosen by the author
- **Follows:** Public profile → instant follow. Private profile → request must be sent, recipient accepts or declines.
- **Chat access:** Users can only start a DM if at least one of them follows the other (or the recipient has a public profile).
- **Groups:** Invite/request with statuses `pending`, `accepted`, `declined`, `invited`. Only accepted members see group content and chat. Group events have RSVP options: `going`, `not_going`, `maybe`.
- **WebSocket:** single shared connection at `/ws/notifications` — delivers both notification events and incoming chat messages.
- **Pagination:** `page`/`limit` query params. TanStack `useInfiniteQuery` on the frontend.
- **Images:** multipart/form-data, max 20MB, jpg/png/gif.
- **Validation:** Mirror backend limits client-side (in `lib/validators.ts`).
- **CORS:** Backend sets `Access-Control-Allow-Origin` from `configs.json` `frontend.url` field. Currently set to `http://localhost:3000` → should be `http://localhost:5173` for dev.
- **Port:** Backend `:8080`, frontend dev `:5173` (Vite default).
- **snake_case** everywhere to match Go JSON tags.
- **File naming:** React components are PascalCase; all other files (hooks, api, utils, pages) are camelCase.

## Redis Integration (Phase I)

### What goes into Redis
| Data | Why |
|------|-----|
| Sessions (`session_id → user_id`) | Ephemeral key-value with TTL, auto-expire, removes SQLite cleanup routine |
| WebSocket pub/sub channels | Enables broadcasting across multiple backend instances |

### What stays in SQLite
All persistent data: users, posts, comments, groups, messages, followers, notifications, events.

### Session flow
```
Login → Generate UUID → Redis SETEX session:{uuid} 86400 user_id → Set cookie
Request → Read cookie → Redis GET session:{uuid} → if nil, 401 → DB lookup by ID
Logout → Redis DEL session:{uuid} → Clear cookie
```

## Seed Data

On first launch, 6 users are pre-loaded. All share password: `password123`

| Email | Name | Profile |
|-------|------|---------|
| `alice@example.com` | Alice Johnson | Public, active poster |
| `bob@example.com` | Bob Smith | Private profile |
| `carol@example.com` | Carol Williams | Public, group creator |
| `dave@example.com` | Dave Brown | Public, backend dev |
| `eve@example.com` | Eve Davis | Private, lurker |
| `frank@example.com` | Frank Miller | Public, photographer |

Plus 17 posts, 16 comments (with replies), 2 groups with events, DMs, and notifications.

To reset: `go run ./backend/cmd/main.go --reseed` or `make backend-run-reseed`
