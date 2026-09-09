# Social Network — Plan

## Tech Stack

### Frontend
- Build: **Next.js 16 (App Router, Turbopack)**
- Framework: React 19 with TypeScript
- Server state: TanStack Query v5
- Routing: Next.js App Router (file-based, route groups `(auth)` / `(main)`)
- Client state: React Context (useAuth + useUI, no external dependencies)
- HTTP: Axios with interceptors
- Styling: TailwindCSS
- Emojis: `emoji-picker-react` (picker) + `@emoji-mart/data` (shortcode autocomplete)
- Real-time: Raw WebSocket (shared connection for notifications + chat, planned)

### Backend
- Language: Go 1.24
- Database: SQLite (all persistent data)
- Cache / ephemeral storage: Redis 7 (presence tracking, caching, rate limiting — session migration + cross-instance pub/sub planned)
- Migrations: golang-migrate (17 migrations)
- Real-time: WebSocket hub at `/api/ws` (gorilla/websocket) with Redis presence tracking
- WebSocket: gorilla/websocket

## Project Structure
```
social-network/
├── frontend/
│   ├── next.config.ts          ✅
│   ├── tsconfig.json           ✅
│   ├── tailwind.config.ts      ✅
│   ├── postcss.config.mjs      ✅
│   ├── package.json            ✅
│   ├── .env                    ✅
│   ├── Dockerfile              ✅
│   ├── nginx.conf              # not yet created
│   ├── src/
│   │   ├── app/                ✅ (Next.js App Router)
│   │   │   ├── layout.tsx      ✅ (root layout + <Providers>)
│   │   │   ├── providers.tsx   ✅ (QueryClient + Auth + UI providers)
│   │   │   ├── page.tsx        ✅ (redirects to /login or /home by auth)
│   │   │   ├── error.tsx       ✅ (global error boundary)
│   │   │   ├── not-found.tsx   ✅ (404 page)
│   │   │   ├── (auth)/         ✅ (login, signup + auth guard)
│   │   │   └── (main)/         ✅ (home, profile, posts, groups, chat, etc. + auth guard)
│   │   ├── types/              ✅ (user, post, comment, group, message, notification, api)
│   │   ├── api/
│   │   │   ├── client.ts       ✅ (Axios instance, withCredentials, 401 interceptor)
│   │   │   ├── auth.ts         ✅ (login, signup, logout, checkSession)
│   │   │   ├── posts.ts        ✅ (getFeed, getPost, createPost, editPost, deletePost)
│   │   │   └── comments.ts     ✅ (getComments, createComment, editComment, deleteComment)
│   │   ├── hooks/
│   │   │   ├── useAuth.ts      ✅ (useLogin, useSignup, useLogout)
│   │   │   ├── usePosts.ts     ✅ (useFeed, usePost, useCreatePost, useEditPost, useDeletePost)
│   │   │   ├── useComments.ts  ✅ (useComments, useCreateComment, useEditComment, useDeleteComment)
│   │   │   └── useEmojiAutocomplete.ts ✅ (:shortcode: → emoji)
│   │   ├── context/
│   │   │   ├── AuthProvider.tsx ✅
│   │   │   └── UIProvider.tsx   ✅
│   │   ├── components/
│   │   │   ├── ui/             ✅ (Avatar, Badge, Button, Card, Dropdown, EmojiPicker, EmojiSuggestions, Input, Modal, Spinner, Tabs, Toast)
│   │   │   ├── layout/         ✅ (NavMenu, TopBar, MobileNav — Sidebar removed)
│   │   │   ├── common/         ✅ (EmptyState, ErrorBoundary, ImageUpload, InfiniteScroll, LoadingScreen, LoadingSkeleton)
│   │   │   ├── post/           ✅ (PostCard, PostForm, PostList, PrivacySelector)
│   │   │   ├── comment/        ✅ (CommentItem, CommentList)
│   │   │   ├── group/          ✅ (EventCard, EventList, EventForm, GroupActions, InviteMember, MemberList)
│   │   │   ├── user/           # empty — Phase 6
│   │   │   ├── chat/           # empty — Phase 9
│   │   │   └── notification/   # empty — Phase 8
│   │   ├── views/              ✅ (page components — renamed from pages/ to avoid Next.js conflict)
│   │   │   ├── auth/           ✅ (LoginPage, SignupPage)
│   │   │   ├── home/           ✅ (HomePage — feed + PostForm)
│   │   │   ├── post/           ✅ (PostDetailPage — post + comments)
│   │   │   ├── profile/        ✅ (ProfilePage, EditProfilePage)
│   │   │   ├── groups/         ✅ (GroupsPage, GroupDetailPage, CreateGroupPage)
│   │   │   ├── chat/           ✅ (ChatPage)
│   │   │   ├── notifications/  ✅ (NotificationsPage)
│   │   │   ├── followers/      ✅ (FollowersPage, FollowingPage)
│   │   │   └── search/         ✅ (SearchPage)
│   │   └── lib/                ✅ (cn, format, validators, nav-link, nav)
│   │   └── ws/                 # empty — Phase 7
├── backend/
│   ├── Dockerfile              ✅
│   ├── cmd/main.go             ✅ (reads --reseed flag, propagates errors)
│   ├── entry/entry.go          ✅ (startup, DB init, seeding, server, graceful shutdown)
│   ├── config/config.go        ✅ (config structs, LoadConfig, rate limits, OAuth)
│   ├── configs.json            ✅ (all settings; frontend.url = http://localhost:3000)
│   ├── global/global.go        ✅ (helper functions for paths, durations, TLS)
│   ├── server/handlers/
│   │   ├── set-handlers.go     ✅ (route registration, CORS, static files)
│   │   ├── middleware.go       ✅ (AuthMiddleware, rate limiter, CORS headers)
│   │   ├── handler-utils.go    ✅ (JSON response helpers, session cookie mgmt)
│   │   ├── event-signup.go     ✅ (POST /api/signup with validation + bcrypt)
│   │   ├── event-login.go      ✅ (POST /api/login, /api/logout, /api/logout-all, GET /api/auth/check)
│   │   ├── event-feed.go       ✅ (GET /api/feed, GET /api/post/{id})
│   │   ├── event-posts.go      ✅ (POST /api/posts, DELETE /api/posts/{id}, edit, user posts)
│   │   ├── event-comments.go   ✅ (create, list, delete, edit comments)
│   │   ├── event-follows.go    ✅ (request, accept, decline, remove, followers, following, pending)
│   │   ├── event-users.go      ✅ (get profile, update profile)
│   │   ├── event-groups.go     ✅ (create, get, update, delete, browse, invite, join, accept, reject, leave, members)
│   │   ├── event-events.go     ✅ (create, list, get event, RSVP — going/not_going with counts)
│   │   ├── event-notifications.go ✅ (list, unread count, mark read, mark all read)
│   │   └── notification-types.go  ✅ (notification type constants)
│   ├── cache/
│   │   └── redis.go            ✅ (Redis client, presence tracking, caching, rate limiting, pub/sub)
│   ├── server/websocket/
│   │   ├── hub.go              ✅ (client registry, presence, message dispatch)
│   │   ├── client.go           ✅ (read/write pumps, ping/pong keepalive)
│   │   ├── handler.go          ✅ (ServeWS — connection upgrade)
│   │   └── types.go            ✅ (WSMessage types + payloads)
│   ├── db/
│   │   ├── migrations/         ✅ (17 migrations up/down)
│   │   ├── queries/            ✅ (all SQL constants)
│   │   ├── sql/
│   │   │   ├── models.go       ✅ (all structs incl. is_deleted flags)
│   │   │   ├── methods.go      ✅ (users, sessions, follows, posts, comments, groups, messages, notifications)
│   │   │   └── sqlite.go       ✅ (DB init, migrations, WAL, session cleanup)
│   │   └── tables/             ✅ (reference schemas)
│   ├── populate/
│   │   ├── seed.json           ✅ (6 users, 28 posts, 35 comments, 2 groups, etc.)
│   │   └── seed.go             ✅ (loader with bcrypt, is_deleted support, first-run check, --reseed)
├── docker-compose.yml          # not yet created
├── Makefile                    ✅
├── setup-dev.sh                ✅
└── PLAN.md                     ✅ (this file)
```

## Current Completion Status

### ✅ Backend — Auth, Feed, Posts, Comments, Follows, Users, Groups

| Component | Status |
|-----------|--------|
| Go module + dependencies (`go.mod`) | ✅ |
| Database migrations (17 up/down pairs) | ✅ |
| All SQL query constants (`internal_queries.go`) | ✅ |
| All data models (`models.go`) | ✅ |
| Database init + migration runner (`sqlite.go`) | ✅ |
| Database methods (`methods.go`) — users, sessions, follows, posts, comments, groups, messages, notifications | ✅ |
| Config loading + types (`config.go`) | ✅ |
| Global helper functions (`global.go`) | ✅ |
| Auth middleware + rate limiter + CORS | ✅ |
| Signup / Login / Logout / Logout-all / Check-auth handlers | ✅ |
| Feed handler (paginated, privacy-filtered) | ✅ |
| Post handlers (create, get, edit, delete, user posts) | ✅ |
| Comment handlers (create, list, edit, delete) | ✅ |
| Follow handlers (request, accept, decline, remove, list) | ✅ |
| User handlers (get profile, edit profile) | ✅ |
| Group handlers (CRUD, browse, invite, join, accept, reject, leave, members) | ✅ |
| Event handlers (create, list, get, RSVP going/not_going with counts) | ✅ |
| Notification handlers (list, unread count, mark read, mark all) | ✅ |
| Actionable notifications (accept/decline group-join outcome) | ✅ |
| Direct message handlers (get DMs, messages, send, unread count) | ✅ |
| Soft delete support (is_deleted on posts + comments) | ✅ |
| Seed data JSON + loader (with deleted entries) | ✅ |
| Redis client + connection (degrades gracefully if unavailable) | ✅ |
| Presence tracking (online/offline via Redis TTL keys + pub/sub) | ✅ |
| Redis caching (sessions, users, posts, groups) + rate limiting | ✅ |
| WebSocket hub (`/api/ws`) — register/unregister, broadcast, typing, singleton per tab | ✅ |
| WS message dispatch (chat, group, notification, presence) | ✅ |
| Group chat REST handlers | ⬜ (WS support exists, no HTTP endpoint yet) |
| Redis session store migration (SQLite → Redis) | ⬜ |
| Docker | ⬜ |

### ✅ Frontend — Auth, Feed, Posts, Comments, Groups, Events, Notifications, Chat

| Component | Status |
|-----------|--------|
| Next.js 16 App Router scaffold (replaces Vite) | ✅ |
| Type definitions (7 files) | ✅ |
| Axios client with 401 interceptor (skips redirect on login/signup) | ✅ |
| Auth context (user, loading, session check with dedupe) | ✅ |
| UI context (sidebar, toasts, modals) | ✅ |
| App Router route groups + auth guards | ✅ |
| 404 page + error boundary | ✅ |
| UI primitives (12 components incl. emoji) | ✅ |
| Login / Signup pages with validation | ✅ |
| HomePage — infinite-scroll feed + create post form | ✅ |
| PostCard — edit/delete for owner, [deleted] placeholder, comments link | ✅ |
| CommentList — nested replies via parent_comment_id, [deleted] placeholders | ✅ |
| Emoji picker button (emoji-picker-react) | ✅ |
| Emoji :shortcode: autocomplete (@emoji-mart/data) | ✅ |
| Burger navigation menu (NavMenu, TopBar, MobileNav — sidebar removed) | ✅ |
| Profile pages (view + edit + follow button) | ✅ |
| Groups pages (browse, detail, membership actions, events) | ✅ |
| Notifications (list + actionable accept/decline buttons) | ✅ |
| Chat (DMs, real-time via WebSocket) | ✅ |
| Search / Following / CreateGroup pages | ⬜ stubs |
| Docker | ⬜ |

## Implementation Plan

### Convention
- **Backend phases** are lettered (A, B, C...)
- **Frontend phases** are numbered (1, 2, 3...)
- ✅ = done, 🔜 = next up, ⬜ = planned

---

## Backend Phases

### Phase A — Auth ✅
Signup (bcrypt + validation), login (session), logout, logout-all, auth/check, middleware (rate limit + CORS), cookie utilities.

### Phase B — DB Methods ✅
All database methods implemented: users, sessions, follows, posts (incl. feed), comments, groups, messages, notifications.

### Phase C — Profile & Follow Handlers ✅
- `GET /api/users/{id}` — profile
- `POST /api/users/{id}/edit` — update profile
- `POST /api/follow/request` / `/accept` / `/decline` / `/remove`
- `GET /api/followers` / `/following` / `/follow/pending`

### Phase D — Post & Comment Handlers ✅
All `{id}` params are **UUIDs** (`GetPostByUUID` / `GetCommentByUUID` lookup). Cross-references in responses (`author_id`, `group_id`, `post_id`, `parent_comment_id`) are UUIDs.
- `GET /api/feed` — paginated, privacy-filtered feed
- `POST /api/posts` — create (optional `group_id` as group UUID)
- `GET /api/post/{id}` — single post with author + comment count
- `POST /api/posts/{id}/edit` — edit (owner)
- `DELETE /api/posts/{id}` — soft delete (owner)
- `GET /api/user/posts` — own posts
- `POST /api/comments` — create (`post_id` / `parent_comment_id` as UUIDs)
- `GET /api/posts/{id}/comments` — list (flat, frontend builds tree)
- `POST /api/comments/{id}/edit` — edit (owner)
- `DELETE /api/comments/{id}` — soft delete (owner)

### Phase E — Group Handlers ✅
Create, get, update, delete, browse, user groups, invite, join, accept, reject, leave, members. All group routes use **UUID** lookup (`GetGroupByUUID`).

### Phase F — Event Handlers ✅
- `POST /api/groups/{id}/events` — create (title 5–300, future datetime required); notifies accepted members via `sendNotification` (`new_event`)
- `GET /api/groups/{id}/events` — list with `{ going, not_going, total, my_response }` per event
- `GET /api/events/{id}` — single event detail (same enrichment)
- `POST /api/events/{id}/rsvp` — upsert response (`going` / `not_going`), returns updated counts
- All endpoints require accepted group membership (or creator)
- All event routes use **UUID** lookup (`GetEventByUUID`); numeric `Event.ID` used internally
- `EventResponse` model

**Frontend (Phase 7b, done):**
- `api/groups.ts` + `hooks/useGroups.ts` — browse, group, events, create event, RSVP (optimistic)
- `components/group/`: `EventCard` (RSVP buttons + counts), `EventList`, `EventForm`
- `GroupsPage` — browse list linking to `/groups/{uuid}`

### Phase G — Notification Handlers ✅
- `GET /api/notifications` — paginated list
- `GET /api/notifications/unread-count`
- `PUT /api/notifications/{id}/read` — mark one read
- `PUT /api/notifications/read-all`
- `sendNotification` helper → creates DB row + pushes real-time over WebSocket (`TypeNotification` → target user)
- Frontend: `useWebSocket`, `NotificationProvider`, TopBar bell, NotificationsPage

### Phase H — Chat / Message Handlers ✅
Conversations (DMs), messages, group chat, read receipts.

### Phase I — WebSocket Hub ✅
- Hub + client lifecycle (`backend/server/websocket/`) at `/api/ws` (auth required)
- Message types: chat_message, group_message, notification, presence_update, typing, ping/pong
- Online/offline presence via Redis (`SetUserOnline`/`SetUserOffline` + pub/sub on `presence:online`/`presence:offline`)
- Graceful shutdown on server stop

### Phase J — Redis Integration ✅ (mostly)
- Redis client + config block (`configs.json`), optional — backend continues without it
- Presence tracking: `presence:{userID}` keys with 30s TTL, `IsUserOnline`, `GetOnlineUsers` (batch)
- Caching: sessions, users, posts, groups (JSON + TTL, invalidation helpers)
- Rate limiting: sliding window via Redis ZSET (`CheckRateLimit`), replaces in-memory limiter when Redis present
- Pub/sub channels defined for future use (`notification:new`, `chat:new_message`, `group:message`)

**Remaining:** Migrate sessions from SQLite to Redis (TTL), wire Redis pub/sub for cross-instance chat/notification fan-out, config toggle.

### Phase K — Docker ✅
Backend Dockerfile (multi-stage Go build), frontend Dockerfile (nginx), docker-compose with Redis.

---

## Frontend Phases

### Phase 1 — Foundation ✅
Next.js App Router scaffold, types, Axios client, auth/UI contexts, route groups with auth guards, UI kit (incl. emoji), 404/error pages.

### Phase 2 — Auth ✅
api/auth.ts + useAuth hooks, LoginPage + SignupPage with validation, session check with dedupe, logout (now in NavMenu).

### Phase 3 — Feed & Posts ✅
api/posts.ts + usePosts hooks, PostCard (edit/delete/[deleted]), PostForm (emoji + privacy), PostList, HomePage with infinite scroll, PostDetailPage.

### Phase 4 — Comments ✅
api/comments.ts + useComments hooks, CommentItem/CommentList with nested replies + [deleted] placeholders, wired into PostDetailPage.

### Phase 5 — Emoji Support ✅
Emoji picker button + `:shortcode:` autocomplete with keyboard navigation (Arrow keys, Enter, Escape).

### Phase 6 — Profiles & Followers ✅
ProfilePage (view + edit + privacy toggle + follow button), FollowersPage (accepted followers + chat entry). Remaining: FollowingPage, SearchPage wiring. Uses Backend Phase C.

### Phase 7 — Groups ✅
GroupDetailPage (header, member count, role-aware GroupActions, MemberList with accept/decline, InviteMember by nickname, EventForm, EventList). GroupsPage browse list. Remaining: CreateGroupPage. Uses Backend Phase E.

### Phase 7b — Events ✅
Event API/hooks/components wired into GroupDetailPage: `api/groups.ts`, `useGroups.ts`, `EventCard` (going/not_going buttons + counts, optimistic RSVP), `EventList`, `EventForm`. Event counts exclude non-members. Uses Backend Phase F.

### Phase 8 — Notifications ✅ + WebSocket frontend
NotificationsPage, TopBar bell, `NotificationProvider`, `useWebSocket` connection to `/api/ws` — all done. Real-time notification push + actionable accept/decline on group-join requests. Outcome text updated on accept/decline.

### Phase 9 — Chat 🔜 (mostly done)
DM chat works end-to-end: ChatPage, send/receive via WebSocket, unread count, message history. **Remaining:** group chat REST endpoints + group chat UI.

### Phase 10 — Navigation ✅
Sidebar removed and replaced with a burger-style `NavMenu` dropdown (user info + all nav items + logout) in the sticky TopBar. Shared `NAV_ITEMS` in `lib/nav.ts`; MobileNav shows all destinations. Logout disconnects the WebSocket.

### Phase 11 — Docker & Polish 🔜
Dockerfile + nginx, production build testing, responsive polish.

## Backend API Endpoints

### Registered (with handler)
| Method | Path | Handler | Auth |
|--------|------|---------|------|
| POST | `/api/signup` | SignupHandler | ❌ |
| POST | `/api/login` | LoginHandler | ❌ |
| GET | `/api/auth/check` | CheckAuthHandler | ❌ |
| POST | `/api/logout` | LogoutHandler | ✅ |
| POST | `/api/logout-all` | LogoutAllHandler | ✅ |
| GET | `/api/health` | inline | ❌ |
| GET | `/api/feed` | GetFeedHandler | ✅ |
| POST | `/api/posts` | CreatePostHandler | ✅ |
| GET | `/api/post/{id}` | GetPostHandler | ✅ |
| POST | `/api/posts/{id}/edit` | EditPostHandler | ✅ |
| DELETE | `/api/posts/{id}` | DeletePostHandler | ✅ |
| GET | `/api/user/posts` | GetUserPostsHandler | ✅ |
| POST | `/api/comments` | CreateCommentHandler | ✅ |
| GET | `/api/posts/{id}/comments` | GetPostCommentsHandler | ❌ |
| POST | `/api/comments/{id}/edit` | EditCommentHandler | ✅ |
| DELETE | `/api/comments/{id}` | DeleteCommentHandler | ✅ |
| POST | `/api/follow/request` | FollowRequestHandler | ✅ |
| POST | `/api/follow/accept` | AcceptFollowHandler | ✅ |
| POST | `/api/follow/decline` | DeclineFollowHandler | ✅ |
| POST | `/api/follow/remove` | UnfollowHandler | ✅ |
| GET | `/api/followers` | GetFollowersHandler | ❌ |
| GET | `/api/following` | GetFollowingHandler | ❌ |
| GET | `/api/follow/pending` | GetPendingFollowsHandler | ✅ |
| GET | `/api/users/{id}` | GetUserProfileHandler | ❌ |
| POST | `/api/users/{id}/edit` | UpdateUserProfileHandler | ✅ |
| POST | `/api/groups` | CreateGroupHandler | ✅ |
| GET | `/api/groups/{id}` | GetGroupHandler | ✅ |
| POST | `/api/groups/{id}/update` | UpdateGroupHandler | ✅ |
| DELETE | `/api/groups/{id}/delete` | DeleteGroupHandler | ✅ |
| GET | `/api/groups/browse` | BrowseGroupsHandler | ❌ |
| GET | `/api/user/groups` | GetUserGroupsHandler | ❌ |
| POST | `/api/groups/{id}/invite` | InviteToGroupHandler | ✅ |
| POST | `/api/groups/{id}/join` | RequestJoinGroupHandler | ✅ |
| POST | `/api/groups/{id}/accept` | AcceptGroupMemberHandler | ✅ |
| POST | `/api/groups/{id}/reject` | RejectGroupMemberHandler | ✅ |
| POST | `/api/groups/{id}/leave` | LeaveGroupHandler | ✅ |
| GET | `/api/groups/{id}/members` | GetGroupMembersHandler | ❌ |
| POST/GET | `/api/groups/{id}/events` | GroupEventsHandler (create/list) | ✅ |
| GET | `/api/events/{id}` | GetEventHandler | ✅ |
| POST | `/api/events/{id}/rsvp` | EventRSVPHandler | ✅ |
| GET | `/api/notifications` | GetNotificationsHandler | ✅ |
| GET | `/api/notifications/unread-count` | GetUnreadNotificationCountHandler | ✅ |
| PUT | `/api/notifications/{id}/read` | MarkNotificationReadHandler | ✅ |
| PUT | `/api/notifications/read-all` | MarkAllNotificationsReadHandler | ✅ |
| WS | `/api/ws` | ServeWS (hub upgrade) | ✅ |

### Planned
Chat (conversations/messages/send), moderation endpoints (from configs.json rate limits).

## Key Conventions

- **Auth:** Session cookie with `withCredentials: true`. Hydrate auth context on mount via `GET /api/auth/check`. 401 interceptor redirects to `/login` unless already there.
- **Privacy levels:**
  - `public` — visible to everyone
  - `followers` — visible only to followers of the post author (author always sees own posts)
  - `private` — visible only to specific followers chosen by the author (`post_visibility`)
- **Soft delete:** Posts and comments set `is_deleted = 1`. They remain visible in feeds/threads but render as `[deleted]` with nested replies intact (Reddit-style). Delete requires ownership.
- **Follows:** Public profile → instant follow. Private profile → request must be sent, recipient accepts or declines.
- **Chat access:** Users can only start a DM if at least one of them follows the other (or the recipient has a public profile).
- **Groups:** Invite/request with statuses `pending`, `accepted`, `declined`, `invited`. Only accepted members see group content and chat. Group events have RSVP options: `going`, `not_going`.
- **Emoji:** Picker button + `:shortcode:` autocomplete backed by `@emoji-mart/data`. Emojis are plain Unicode — stored as TEXT in SQLite.
- **WebSocket:** single shared connection at `/api/ws` (auth required), singleton per tab. Hub dispatches `chat_message`, `group_message`, `notification`, `presence_update`, `typing`, `ping`/`pong`. Presence tracked in Redis with 30s TTL keys. Logout disconnects the socket.
- **Nicknames:** mandatory, auto-generated from the email prefix at signup (sanitized, padded, de-duplicated). Used for group invites (`invite by nickname`).
- **Navigation:** no sidebar — a burger `NavMenu` dropdown in the sticky TopBar holds all nav items + logout on desktop and mobile. Mobile bottom nav mirrors the same destinations.
- **Pagination:** `limit`/`offset` query params. TanStack `useInfiniteQuery` on the frontend.
- **Images:** multipart/form-data, max 20MB, jpg/png/gif (planned).
- **Validation:** Mirror backend limits client-side (in `lib/validators.ts`).
- **CORS:** Backend sets `Access-Control-Allow-Origin` from `configs.json` `frontend.url` (`http://localhost:3000`).
- **Port:** Backend `:8080`, frontend dev `:3000` (Next.js default).
- **snake_case** everywhere to match Go JSON tags.
- **Route IDs:** **All** resources (users, groups, events, posts, comments) use **UUID** in route paths and cross-references (looked up internally by numeric FK). Numeric IDs are never exposed in API responses (`json:"-"`).
- **File naming:** React components are PascalCase; all other files (hooks, api, utils, views) are camelCase.

## Redis Integration (Phase J)

### What's implemented
| Feature | Details |
|---------|---------|
| Presence tracking | `presence:{userID}` keys, 30s TTL, `SetUserOnline/Offline`, `IsUserOnline`, `GetOnlineUsers` (batch MGET) |
| Caching | Sessions, users, posts, groups via JSON + TTL (`SetJSON`/`GetJSON`) with invalidation helpers |
| Rate limiting | Sliding window via Redis ZSET (`CheckRateLimit`) — used by middleware when Redis is available |
| Pub/sub channels | Defined: `presence:online`, `presence:offline`, `notification:new`, `chat:new_message`, `group:message` |

Redis is **optional** — if unavailable, the backend logs a warning and continues (WebSocket presence falls back to in-memory hub state, rate limiting falls back to the sync.Map).

### Remaining (planned)
| Feature | Notes |
|---------|-------|
| Sessions (`session_id → user_id`) | Ephemeral key-value with TTL, auto-expire, removes SQLite cleanup routine |
| Cross-instance pub/sub | Subscribe to `chat:new_message` / `notification:new` to fan out across multiple backend instances |
| Config toggle | `session_storage: "sqlite" | "redis"` |

### Session flow (planned)
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

Plus 28 posts (2 marked deleted), 35 comments (3 marked deleted, some nested under deleted parents), 2 groups, 5 events (with going/not_going RSVPs), DMs, and notifications.

To reset: `make backend-run-reseed` (or `go run ./backend/cmd/main.go --reseed`)
