# AGENTS.md

## Project Overview

A full-stack social network application. Go backend (standard library HTTP + SQLite with golang-migrate, Redis for caching/presence, WebSocket hub for real-time chat) and a React/TypeScript frontend (Next.js 16 App Router + Tailwind CSS). HTTPS is terminated at the edge by Caddy (Docker). Both sides compile cleanly; auth flow is wired end-to-end.

## Commands

All from repo root unless noted.

### Quick start

```bash
make dev          # kill stale ports, start frontend + backend concurrently
make check        # TypeScript typecheck + go vet populate
```

### Backend

```bash
make backend-run              # go run ./backend/cmd/main.go
make backend-run-reseed       # run with --reseed flag (drops + reseeds DB)
make backend-vet-populate     # go vet ./backend/populate/

go build -o bin/social-network ./backend/cmd
go test ./backend/...
```

### Frontend

```bash
cd frontend

npm install              # or: make frontend-install
npm run dev              # or: make frontend-dev     → Next.js dev on :3000
npm run build            # or: make frontend-build    → next build
npm run check            # or: make frontend-check    → tsc --noEmit
```

### Database

Migrations run automatically on startup via `golang-migrate/migrate`. Manual control:

```bash
migrate -database "sqlite3://backend/db/social-network.db" \
  -path backend/db/migrations up
```

Seed data (JSON at `backend/populate/seed.json`) is auto-loaded on first run if the users table is empty. Pass `--reseed` to drop and re-import.

### Docker (dev environment)

```bash
make docker-build   # build backend + frontend images
make docker-up      # start full stack (Caddy :443, backend :8080, frontend :3000, redis)
make docker-down    # stop stack
make docker-reset   # stop + remove volumes
```

- **Caddy** terminates TLS on `https://localhost` and routes `/api/*` to the backend and everything else to the Next.js dev server. No cert handling exists in the Go backend.
- Dev server config is `backend/configs.docker.json` (used by the backend container).

## Architecture

### Backend startup flow

```
cmd/main.go
  → entry.Start(reseed)
    → global.Initialize()            creates shutdown context
    → config.LoadConfig("backend/configs.json")
    → setupDatabase(cfg)             opens SQLite, applies migrations, starts WAL/session goroutines
    → setupRedis(cfg)                connects to Redis (degrades gracefully if unavailable)
    → populate.SeedFromJSON()        seeds if users table empty
    → setupServer(cfg, db, redis)    handlers.SetHandlers(db, redis) → http.ServeMux
    → startServer()                  ListenAndServe (HTTP only; TLS handled by Caddy)
    → waitForShutdown()              SIGINT/SIGTERM → graceful shutdown
```

### Key packages

| Package | Role |
|---------|------|
| `backend/cmd` | Entry point — parses `--reseed`, calls `entry.Start()` |
| `backend/entry` | Orchestrates startup and shutdown. DB setup, Redis setup, server setup, seed, graceful drain |
| `backend/global` | Lightweight now — holds `ShutDownContext`/`CancelShutdown` plus helper functions for path construction and duration/size parsing. No TLS/cert helpers (Caddy handles HTTPS) |
| `backend/config` | Owns `config.LoadConfig()` which reads `backend/configs.json` into a typed `Config` struct. Thread-safe access via `GetConfig()`. Provides `GetRateLimit()`, `GetFrontendURL()`, `GetUniversalRateLimit()`. Contains OAuth structs (future feature) |
| `backend/cache` | Redis client — presence tracking, JSON caching (sessions/users/posts/groups), sliding-window rate limiting, pub/sub |
| `backend/db/sql` | **Core data layer**. `DataBase` struct wraps `*sql.DB`. `New()` opens SQLite, configures WAL/foreign_keys/busy_timeout pragmas, runs migrations, starts WAL truncate + session cleanup goroutines. All CRUD methods are on `*DataBase` in `methods.go`. No prepared statements — all queries use raw SQL with `ExecContext`/`QueryRowContext` |
| `backend/db/queries` | SQL string constants exported for use by both the `database` package methods and the `populate` seeder |
| `backend/db/migrations` | 19 golang-migrate up/down pairs |
| `backend/server/handlers` | `SetHandlers(db, redis)` builds the mux with registered endpoints. Each handler receives `(http.ResponseWriter, *http.Request, *database.DataBase)`. WebSocket hub at `/api/ws` |
| `backend/server/websocket` | WebSocket hub, client read/write pumps, connection handler, message types |
| `backend/populate` | Reads `seed.json`, inserts users/follows/groups/posts/comments/events/messages/notifications. Skips if users table is non-empty. `Reseed()` drops all first |

### Handler pattern

Handlers follow this signature:

```go
func SomeHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase)
```

Endpoints are registered in `SetHandlers()` via `makeEndpoint(path, requireAuth, handler)`. Rate limits come from `config.GetRateLimit()` and `config.GetUniversalRateLimit()` — no longer configured per-endpoint in code.

**Response helpers** in `handler-utils.go`:
- `RespondJSON(w, status, payload)` — writes JSON
- `RespondError(w, status, msg)` — writes `{"error": "msg"}`
- `RespondSuccess(w, status, msg, data)` — writes `{"message": "msg", "data": ...}`

**Registered endpoints** (current, all prefixed `/api`):
- Auth: `POST /signup`, `POST /login`, `GET /auth/check`, `POST /logout`, `POST /logout-all`
- Posts: `GET /feed`, `POST /posts`, `GET /post/{id}`, `DELETE /posts/{id}`, `POST /posts/{id}/edit`, `GET /user/posts`
- Comments: `POST /comments`, `GET /posts/{id}/comments`, `DELETE /comments/{id}`, `POST /comments/{id}/edit`
- Follows: `POST /follow/request`, `POST /follow/accept`, `POST /follow/decline`, `POST /follow/remove`, `GET /followers`, `GET /following`, `GET /follow/pending`
- Users: `GET /users/{id}`, `POST /users/{id}/edit`, `POST /users/{id}/avatar`
- Groups: `POST /groups`, `GET /groups/{id}`, `POST /groups/{id}/update`, `POST /groups/{id}/delete`, `GET /groups/browse`, `GET /user/groups`, invite/join/accept/reject/leave/members/posts/avatar
- Events: `GET /groups/{id}/events`, `GET /events/{id}`, `POST /events/{id}/rsvp`
- Notifications: `GET /notifications`, `GET /notifications/unread-count`, `POST /notifications/{id}/read`, `POST /notifications/read-all`
- Chat: `GET /chat/dms`, `GET /chat/dms/{id}/messages`, `POST /chat/send/{id}`, `GET /chat/unread-count`
- `GET /health` — health check
- `GET /api/ws` — WebSocket hub (auth required)
- `OPTIONS /api` — CORS preflight

### Middleware

`AuthMiddleware` in `middleware.go`:
1. Sets CORS headers (origin from `config.GetFrontendURL()`, credentials allowed)
2. Short-circuits OPTIONS with 204
3. Universal rate limit by IP
4. Path-specific rate limit by `IP + path`
5. Reads `session_token` cookie, resolves user from DB
6. If `requireAuth` and no session → 401 JSON (API) or redirect to `/login`
7. Injects `userID` into request context via `contextKey("userID")`
8. Recovers from panics

`GetUserIDFromContext(r)` extracts the user ID from context in auth-required handlers.

### Session/auth

- Cookie name: `session_token`
- Session ID: UUID generated via `google/uuid`
- 24-hour expiration, HTTP-only, Lax SameSite
- `GetUserFromCookie()` validates UUID, looks up session in DB, returns `*database.User`
- `CheckAuthHandler` (`GET /api/auth/check`) is used by the frontend on mount to restore sessions

### Config

Config lives at `backend/configs.json` (local) or `backend/configs.docker.json` (Docker). Key sections:

- `frontend.url` — used for CORS (defaults to `http://localhost:3000`; `https://localhost` in Docker)
- `redis` — Redis address, pool, timeouts
- `database_configuration` — path, WAL, session cleanup interval, validation limits, system images
- `server` — Addr (`:8080`)
- `handlers` — rate limits per path, image config, cookie expiration

**Note**: HTTPS is terminated by Caddy, so there is no `certifications` config in the backend.

### Database

SQLite with WAL mode, foreign keys enforced, 5s busy timeout. 19 migration files covering the full schema. `DataBase` exposes `GetDB() *sql.DB` for the populate package and any future advanced operations.

The `methods.go` file provides typed CRUD methods: `AddUser`, `GetUserByEmail`, `GetUserByID`, `GetUserByUUID`, `UpdateUserProfile`, `UpdateUserPrivacy`, `DeleteUser`, `CreateSession`, `GetSession`, `DeleteSession`, `DeleteAllUserSessions`, `CreatePost`, `GetPost`, `GetUserPosts`, `DeletePost`, `CreateGroup`, `GetGroup`, `GetUserGroups`, `CreateMessage`, `CreateOrGetDirectMessage`, `CreateNotification`, `GetUserNotifications`, `MarkNotificationAsRead`.

**Note**: The old `Queries` struct with prepared statements was removed. All DB methods now use direct `ExecContext`/`QueryRowContext` with constants from the `queries` package.

### Frontend

**Stack**: Next.js 16 (App Router, Turbopack), React 19, TypeScript 5.6+, Tailwind CSS 3.4, TanStack React Query 5, Axios, emoji-picker-react, @emoji-mart/data.

**State management**: React Context. `AuthProvider` manages `user`, `is_authenticated`, `is_loading`. `UIProvider` manages modals and toasts. `NotificationProvider` manages real-time notifications. On mount, `AuthProvider.check_session()` calls `GET /api/auth/check` to restore the session.

**Auth hooks** (`src/hooks/useAuth.ts`): `useLogin()`, `useSignup()`, `useLogout()` — TanStack mutations that update auth context on success.

**Directory structure**:

| Path | Purpose |
|------|---------|
| `src/api/client.ts` | Axios instance, `NEXT_PUBLIC_API_URL` env defaulting to `http://localhost:8080`, `withCredentials: true`, 401 → redirect to `/login` |
| `src/api/` | `auth.ts`, `posts.ts`, `comments.ts`, `groups.ts`, `notifications.ts`, `chat.ts` endpoint functions |
| `src/app/` | Next.js App Router: `layout.tsx`, `page.tsx`, `providers.tsx`, `error.tsx`, `not-found.tsx`, route groups `(auth)` and `(main)` |
| `src/context/` | `AuthProvider.tsx`, `UIProvider.tsx`, `NotificationProvider.tsx` |
| `src/hooks/` | `useAuth.ts`, `usePosts.ts`, `useComments.ts`, `useGroups.ts`, `useEmojiAutocomplete.ts`, `useMessageBadge.ts`, `useWebSocket.ts` |
| `src/views/` | Page components (mirrors route structure: auth, home, post, profile, groups, chat, notifications, followers, search) |
| `src/components/ui/` | Reusable primitives (Avatar, Button, Card, Input, Modal, Toast, EmojiPicker, etc.) |
| `src/components/layout/` | `TopBar`, `NavMenu`, `MobileNav` (burger nav — no sidebar) |
| `src/components/common/` | EmptyState, ErrorBoundary, ImageUpload, InfiniteScroll, LoadingScreen, LoadingSkeleton |
| `src/components/post/` `comment/` `group/` | Feature components |
| `src/types/` | TypeScript interfaces (snake_case fields matching JSON) |
| `src/lib/` | `cn.ts` (clsx + tailwind-merge), `format.ts`, `media.ts`, `validators.ts`, `nav-link.tsx`, `nav.ts` |

**Routing**: Next.js App Router with route groups `(auth)` (login/signup — redirects to `/home` if authenticated) and `(main)` (all authenticated routes — redirects to `/login` if unauthenticated). Auth guards in each layout's `layout.tsx`. `not-found.tsx` handles 404, `error.tsx` is the error boundary.

**WebSocket**: `useWebSocket.ts` maintains a singleton connection to `/api/ws` with auto-reconnect. Derives `ws://`/`wss://` from `NEXT_PUBLIC_API_URL`. Used for real-time chat, notifications, presence, and typing indicators.

## Conventions

### Go

- **Naming**: MixedCaps exported, camelCase unexported. snake_case in JSON tags.
- **DB package**: `package database` with import path `social-network/backend/db/sql`.
- **Error handling**: `fmt.Errorf("context: %w", err)` wrapping. Handlers return JSON errors via `RespondError()`.
- **Config**: Path segments are `[]string` in JSON, joined with `filepath.Join` at runtime by `global` helper functions.

### TypeScript/React

- **Naming**: snake_case for data fields (matches backend JSON), snake_case for variables and functions.
- **Imports**: `@/` alias maps to `src/`. `import type` for type-only imports.
- **Components**: Default export for pages, named export for shared components.
- **Styling**: Tailwind exclusively. `cn()` from `src/lib/cn.ts` for conditional class merging.
- **TypeScript**: `strict: true`, `noUnusedLocals: true`, `noUnusedParameters: true`.
- **API responses**: Expect `{"message": "...", "data": ...}` shape from the new `RespondSuccess` helper, or `{"error": "..."}` from `RespondError`.
