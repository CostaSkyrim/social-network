# AGENTS.md

## Project Overview

A full-stack social network application. Go backend (standard library HTTP + SQLite with golang-migrate) and a React/TypeScript frontend (Vite + Tailwind CSS). Both sides compile cleanly. The backend now starts successfully; auth flow is wired end-to-end.

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
npm run dev              # or: make frontend-dev     → Vite on :5173
npm run build            # or: make frontend-build    → tsc -b && vite build
npx tsc --noEmit         # or: make frontend-check    → typecheck only
```

### Database

Migrations run automatically on startup via `golang-migrate/migrate`. Manual control:

```bash
migrate -database "sqlite3://backend/db/social-network.db" \
  -path backend/db/migrations up
```

Seed data (JSON at `backend/populate/seed.json`) is auto-loaded on first run if the users table is empty. Pass `--reseed` to drop and re-import.

### Dev container (distrobox)

```bash
./setup-dev.sh                   # create Fedora container with Go 1.24, Node 20
distrobox enter social-network-dev
./cleanup.sh                     # remove the container
```

## Architecture

### Backend startup flow

```
cmd/main.go
  → entry.Start(reseed)
    → global.Initialize()            creates shutdown context
    → config.LoadConfig("backend/configs.json")
    → setupDatabase(cfg)             opens SQLite, applies migrations, starts WAL/session goroutines
    → populate.SeedFromJSON()        seeds if users table empty
    → setupServer(cfg, db)           handlers.SetHandlers(db) → http.ServeMux
    → startServer()                  ListenAndServe or ListenAndServeTLS
    → waitForShutdown()              SIGINT/SIGTERM → graceful shutdown
```

### Key packages

| Package | Role |
|---------|------|
| `backend/cmd` | Entry point — parses `--reseed`, calls `entry.Start()` |
| `backend/entry` | Orchestrates startup and shutdown. DB setup, server setup, seed, TLS config, graceful drain |
| `backend/global` | Lightweight now — holds `ShutDownContext`/`CancelShutdown` plus helper functions for path construction and duration/size parsing. No longer a singleton config holder |
| `backend/config` | Owns `config.LoadConfig()` which reads `backend/configs.json` into a typed `Config` struct. Thread-safe access via `GetConfig()`. Provides `GetRateLimit()`, `GetFrontendURL()`, `GetUniversalRateLimit()` |
| `backend/db/sql` | **Core data layer**. `DataBase` struct wraps `*sql.DB`. `New()` opens SQLite, configures WAL/foreign_keys/busy_timeout pragmas, runs migrations, starts WAL truncate + session cleanup goroutines. All CRUD methods are on `*DataBase` in `methods.go`. No prepared statements — all queries use raw SQL with `ExecContext`/`QueryRowContext` |
| `backend/db/queries` | SQL string constants exported for use by both the `database` package methods and the `populate` seeder |
| `backend/db/migrations` | 14 golang-migrate up/down pairs |
| `backend/server/handlers` | `SetHandlers(db)` builds the mux with registered endpoints. Each handler receives `(http.ResponseWriter, *http.Request, *database.DataBase)` |
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

**Registered endpoints** (current):
- `POST /api/signup` (no auth)
- `POST /api/login` (no auth)
- `GET /api/auth/check` (no auth) — validates session cookie, returns user or 401
- `POST /api/logout` (auth required)
- `POST /api/logout-all` (auth required)
- `GET /api/health` — health check
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

Config lives at `backend/configs.json`. Key sections:

- `frontend.url` — used for CORS (defaults to `http://localhost:3000` if absent; overridden to `http://localhost:5173` in dev config)
- `database_configuration` — path, WAL, session cleanup interval, validation limits, system images
- `server` — Addr (`:8080`)
- `certifications` — HTTPS certs/key paths
- `handlers` — rate limits per path, image config, cookie expiration

### Database

SQLite with WAL mode, foreign keys enforced, 5s busy timeout. 14 migration files covering the full schema. `DataBase` exposes `GetDB() *sql.DB` for the populate package and any future advanced operations.

The `methods.go` file provides typed CRUD methods: `AddUser`, `GetUserByEmail`, `GetUserByID`, `GetUserByUUID`, `UpdateUserProfile`, `UpdateUserPrivacy`, `DeleteUser`, `CreateSession`, `GetSession`, `DeleteSession`, `DeleteAllUserSessions`, `CreatePost`, `GetPost`, `GetUserPosts`, `DeletePost`, `CreateGroup`, `GetGroup`, `GetUserGroups`, `CreateMessage`, `CreateOrGetDirectMessage`, `CreateNotification`, `GetUserNotifications`, `MarkNotificationAsRead`.

**Note**: The old `Queries` struct with prepared statements was removed. All DB methods now use direct `ExecContext`/`QueryRowContext` with constants from the `queries` package.

### Frontend

**Stack**: React 18, TypeScript 5.6+, Vite 6, Tailwind CSS 3.4, React Router 6, TanStack React Query 5, Axios.

**State management**: React Context instead of Zustand. `AuthProvider` manages `user`, `is_authenticated`, `is_loading`. `UIProvider` manages sidebar, modals, toasts. On mount, `AuthProvider.check_session()` calls `GET /api/auth/check` to restore the session.

**Auth hooks** (`src/hooks/useAuth.ts`): `useLogin()`, `useSignup()`, `useLogout()` — TanStack mutations that update auth context on success.

**Directory structure**:

| Path | Purpose |
|------|---------|
| `src/api/client.ts` | Axios instance, `VITE_API_URL` env defaulting to `http://localhost:8080`, `withCredentials: true`, 401 → redirect to `/login` |
| `src/api/auth.ts` | `login()`, `signup()`, `logout()`, `checkSession()` functions |
| `src/context/AuthProvider.tsx` | Auth state via React Context + `useState` |
| `src/context/UIProvider.tsx` | UI state (sidebar, modals, toasts) via React Context + `useState` |
| `src/hooks/useAuth.ts` | TanStack mutations wrapping auth API calls |
| `src/types/` | TypeScript interfaces (snake_case fields matching JSON) |
| `src/pages/` | Route-level page components |
| `src/pages/errors/` | `NotFoundPage` (404) and `ErrorPage` (route-level error boundary) |
| `src/components/ui/` | Reusable primitives (Avatar, Button, Card, Input, Modal, Toast, etc.) |
| `src/components/layout/` | `MainLayout` and `AuthLayout` with auth guards built in |
| `src/components/common/` | EmptyState, ErrorBoundary, ImageUpload, InfiniteScroll, LoadingScreen, LoadingSkeleton |
| `src/lib/` | `cn.ts` (clsx + tailwind-merge), `format.ts`, `validators.ts` |

**Routing**: `AuthLayout` wraps `/login` and `/signup` — redirects to `/home` if already authenticated, shows spinner while loading. `MainLayout` wraps all authenticated routes — redirects to `/login` if unauthenticated, shows spinner while loading. `NotFoundPage` catches `/404` and `*`. `ErrorPage` is set as `errorElement` on both layout groups.

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
