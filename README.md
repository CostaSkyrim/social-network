# Social Network

A full-stack, Facebook-like social network. A Go backend serves a JSON REST API plus a WebSocket hub for real-time chat and notifications; a React/TypeScript frontend (Next.js App Router) consumes it. SQLite is the durable store, Redis provides session storage, caching, rate limiting, and cross-instance pub/sub, and Caddy terminates TLS at the edge.

## Authors

- **[Konstantinos Petroutsos](https://github.com/CostaSkyrim)**
- **[Augoustinos Andris](https://github.com/yukiblob)**

## Features

- **Followers** — Follow/unfollow users. Public profiles allow instant following. Private profiles require a follow request that the recipient can accept or decline.
- **Profile** — View user profiles with posts, followers, and following lists. Toggle profile visibility between public and private.
- **Posts** — Text/image posts with privacy levels: public, followers-only, or private (specific followers). Comment with image support. Edit and soft-delete your own posts (deleted posts render as `[deleted]` with comments intact, Reddit-style).
- **Comments** — Nested reply threads. Edit and soft-delete your own comments.
- **Image support** — Avatars for users and groups, plus images on posts, comments, events, and messages. Stored locally and served from `/images/...` (20MB max; jpg/png/gif).
- **User search** — Find people by name or `@nickname`; results respect profile privacy and include an inline follow button.
- **Emoji support** — Emoji picker plus `:shortcode:` autocomplete with keyboard navigation.
- **OAuth** — Sign in/up with Google or GitHub (GitHub linking uses only verified emails).
- **Groups** — Create groups with a title, description, and optional photo. Invite by nickname or accept join requests. Members post, comment, create events, and chat in a shared room.
- **Events** — Create events with title, description, optional image, and date/time. RSVP Going / Not Going with live counts. A background scheduler sends reminders before events start.
- **Notifications** — Real-time notifications (follow requests, group invites, join requests, events, reminders) across all pages.
- **Chat** — Real-time private messaging between mutual followers and group chat rooms, with emoji support and WebSocket-powered instant delivery.
- **Real-time presence** — Online/offline tracked via Redis (30s TTL keys) and broadcast over the WebSocket hub.

## Technology Stack

Every technology in this project is explained below so someone unfamiliar with each tool can understand what it does and why it's used here.

### Backend

#### Go 1.24

[Go](https://go.dev/) is a statically-typed, compiled language designed at Google for server software. It compiles to a single native binary, has first-class concurrency (goroutines and channels), and a standard library that covers HTTP servers, JSON, and databases without extra frameworks.

**Why it's used here:** the entire backend API is written in Go using only the standard library `net/http` (no web framework). Goroutines run background workers (WAL checkpointing, session cleanup, the event-reminder scheduler) concurrently with the HTTP server.

#### SQLite (via mattn/go-sqlite3)

[SQLite](https://www.sqlite.org/) is a self-contained, file-based SQL database. It needs no separate database server — the entire database lives in a single `.db` file and is accessed through a C library.

**Why it's used here:** all persistent business data (users, posts, comments, groups, events, messages, notifications, sessions) lives in one SQLite file. It runs in [WAL mode](https://www.sqlite.org/wal.html) (Write-Ahead Logging) for better concurrent read/write throughput, with `foreign_keys` and `busy_timeout` pragmas enabled. `mattn/go-sqlite3` is the Go driver that links SQLite's C library into the binary (via cgo).

#### Redis 7 (via go-redis/v9)

[Redis](https://redis.io/) is an in-memory data structure store. It's dramatically faster than disk databases because data lives in RAM, making it ideal for caching, ephemeral state, and pub/sub. It supports strings, hashes, lists, sets, sorted sets, and publish/subscribe channels.

**Why it's used here** (Redis is required — with no Redis there is no in-memory or database fallback for these features):

- **Sessions** — session records are stored in Redis as the source of truth (`sessions.storage = "redis"`). If Redis is unavailable, authentication fails closed (every request 401s); there is no SQLite fallback.
- **Read-through cache** — user, group, and group-member lookups check Redis first, fall back to SQLite on a miss, and populate the cache. Email/nickname → user-ID index lookups are cached too, storing only the numeric ID (never the password hash). Writes invalidate the relevant keys. Gated by `use_cache`.
- **Sliding-window rate limiting** — requests are counted in Redis sorted sets per IP and per IP+path; old entries are pruned and the count is checked atomically in a pipeline.
- **Presence** — "online" is a Redis key with a 30s TTL that the WebSocket heartbeat refreshes; expiration marks a user offline.
- **Pub/sub fan-out** — the `ws:fanout` channel broadcasts WebSocket messages across multiple backend instances so a message received on one instance reaches clients on others.

`go-redis/v9` is the Go client library.

#### gorilla/websocket

[gorilla/websocket](https://github.com/gorilla/websocket) is a widely-used, well-tested Go library implementing the WebSocket protocol (RFC 6455). WebSockets upgrade a single HTTP connection into a persistent, bidirectional channel — unlike normal HTTP, the server can push data to the client at any time without the client first making a request.

**Why it's used here:** the `/api/ws` endpoint upgrades the connection and feeds a Hub that tracks connected clients and dispatches messages. Real-time chat, group chat, notifications, presence, and typing indicators all flow over this one shared connection (one singleton connection per browser tab), with ping/pong keepalive to drop dead connections.

#### golang-migrate

[golang-migrate](https://github.com/golang-migrate/migrate) is a database migration tool. Migrations are numbered SQL files (`000001_create_users_table.up.sql` / `.down.sql`) applied in order to evolve a schema without wiping data. The tool tracks which migrations have run and applies only the new ones.

**Why it's used here:** the 21 migrations defining the full schema (up through `oauth_accounts` and event reminders) run automatically on startup; `up` applies forward changes and `down` rolls them back.

#### golang.org/x/crypto (bcrypt)

[`golang.org/x/crypto`](https://pkg.go.dev/golang.org/x/crypto) is Google's extended crypto library for Go. The `bcrypt` package implements the bcrypt password-hashing function, which salts each password and is deliberately slow (through a configurable cost factor), making brute-force attacks impractical.

**Why it's used here:** every password is hashed with bcrypt before storage; plain-text passwords are never written to the database.

#### google/uuid

[`google/uuid`](https://github.com/google/uuid) generates Universally Unique Identifiers (UUIDs) version 4 (random). UUIDs are 128-bit identifiers guaranteed unique without a central coordinator.

**Why it's used here:** every user, post, comment, group, event, and session is addressed publicly by a UUID rather than its sequential numeric database ID, so internal row numbers are never exposed in URLs or API responses (prevents enumeration and leaks nothing about row counts).

#### Air (backend hot-reload, dev only)

[Air](https://github.com/air-verse/air) is a live-reload development tool for Go. It watches source files and automatically rebuilds and restarts the server when they change, giving a near-instant feedback loop during development.

**Why it's used here:** the backend Docker image's entrypoint is `air`, configured via `.air.toml` to rebuild `./backend/cmd` and restart on any `.go` change. Local development uses `make backend-run` (`go run`), while the containerized dev environment uses Air for hot reload.

### Frontend

#### Next.js 16 (App Router + Turbopack)

[Next.js](https://nextjs.org/) is a React framework that adds file-based routing, server-side rendering, and an optimized build system. The **App Router** is its current routing model, where folders under `app/` define routes and special `layout.tsx`/`page.tsx` files define shared layouts and pages. **Turbopack** is the Rust-based bundler/dev server that replaces Webpack for much faster builds.

**Why it's used here:** the frontend uses the App Router with route groups `(auth)` (login/signup) and `(main)` (all authenticated pages), each with its own `layout.tsx` guard that redirects unauthenticated users to `/login`. `error.tsx` and `not-found.tsx` provide the global error boundary and 404 page.

#### React 19

[React](https://react.dev/) is a JavaScript library for building user interfaces from reusable, stateful components. Each component declares how the UI should look for a given state, and React updates the DOM efficiently when that state changes.

**Why it's used here:** the entire UI is composed of React components (UI primitives — Button, Modal, Input, Toast — plus feature components — PostCard, GroupChat, CommentList). Function components and hooks (`useState`, `useEffect`, custom hooks) are used throughout.

#### TypeScript 5

[TypeScript](https://www.typescriptlang.org/) is a superset of JavaScript that adds static types. Types are checked at compile time, catching errors (typos, wrong argument types, missing properties) before the code runs.

**Why it's used here:** all frontend code is written in TypeScript with `strict: true`, `noUnusedLocals`, and `noUnusedParameters`. Data field names use snake_case to match the backend JSON exactly, and shared interfaces live in `src/types/`.

#### TanStack Query v5

[TanStack Query](https://tanstack.com/query) (formerly React Query) is a server-state library. It caches API responses, deduplicates in-flight requests, handles background refetching, and manages loading/error/success states, so components stay in sync with the server without manual `useEffect`+`fetch` boilerplate.

**Why it's used here:** every API call goes through a TanStack Query hook (`useFeed`, `useGroupPosts`, `useComments`, `useGroups`, `usePost`, etc.), which gives automatic caching, optimistic RSVP updates on events, and invalidation on mutations.

#### Axios

[Axios](https://axios-http.com/) is a promise-based HTTP client for the browser and Node. It provides a consistent API, automatic JSON handling, request/response interceptors, and `withCredentials` support for sending cookies cross-origin.

**Why it's used here:** `src/api/client.ts` creates a single Axios instance with `withCredentials: true` (so the session cookie travels with every request) and a response interceptor that redirects to `/login` on 401.

#### Tailwind CSS

[Tailwind CSS](https://tailwindcss.com/) is a utility-first CSS framework. Instead of writing custom CSS classes, you compose styling from small single-purpose utility classes (e.g. `flex items-center gap-2`) directly in the markup.

**Why it's used here:** all styling is done with Tailwind utility classes; `src/lib/cn.ts` merges conditional classes using `clsx` and `tailwind-merge`.

#### emoji-picker-react

[emoji-picker-react](https://github.com/ealush/emoji-picker-react) is a React component that renders a searchable, categorized emoji picker.

**Why it's used here:** the post composer and chat inputs include an emoji-picker button that opens this picker to insert emoji.

#### @emoji-mart/data

[`@emoji-mart/data`](https://github.com/missive/emoji-mart) provides structured emoji data (names, shortcodes, keywords, categories) used to build custom emoji experiences.

**Why it's used here:** it powers the `:shortcode:` autocomplete — typing `:cat` shows matching emoji suggestions filtered from this dataset, with keyboard navigation.

### Infrastructure / Tooling

#### Caddy 2

[Caddy](https://caddyserver.com/) is a web server and reverse proxy that automatically manages HTTPS certificates. Its config is a simple text file ("Caddyfile"), and it can terminate TLS with automatically-provisioned certificates.

**Why it's used here:** in the Docker dev stack, Caddy terminates HTTPS at `https://localhost` (using an internal CA/self-signed cert for local development) and reverse-proxies `/api/*` and `/images/*` to the Go backend and everything else to the Next.js dev server. The backend itself serves plain HTTP; TLS is handled entirely at the edge.

#### Docker & Docker Compose

[Docker](https://www.docker.com/) packages applications with their dependencies into portable containers; Docker Compose orchestrates multi-container setups described in `docker-compose.yml`.

**Why it's used here:** the dev stack runs four services — `caddy`, `frontend` (Next.js dev server), `backend` (Go with Air hot-reload), and `redis` — connected on a private `social-net` bridge network, with named volumes for the database, uploaded images, and Redis data.

### How it all connects

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  Caddy      │     │   Backend    │     │    Redis    │
│  :443 (TLS) │────▶│   :8080      │────▶│   :6379     │
│  reverse    │     │   Go API     │     │  sessions,  │
│  proxy      │     │   SQLite     │     │  cache,     │
│      │      │     │   WS :/api/ws│     │  rl, pub/sub│
│      ▼      │     └──────────────┘     └─────────────┘
│  Frontend   │
│  :3000      │
│  Next.js    │
└─────────────┘
```

- **Frontend → Backend**: Axios calls the REST API; a shared WebSocket connects to `/api/ws` for real-time events.
- **Backend → SQLite**: the durable store for all business data, accessed through `database/sql`.
- **Backend → Redis**: sessions, read-through caching, rate limiting, presence, and cross-instance pub/sub.
- **Caddy → Backend/Frontend**: terminates TLS and routes `/api/*` and `/images/*` to Go, everything else to Next.js.

## Architecture

### Backend

```
cmd/main.go            Entry point — parses --reseed, calls entry.Start
entry/                 Startup/shutdown orchestration (DB, Redis, server, seed, drain)
config/                Typed config loaded from backend/configs.json
global/                Shutdown context + path/duration/size helpers
db/sql/                Data layer: models, CRUD methods, SQLite connection
db/queries/            SQL string constants
db/migrations/         21 up/down migration pairs
cache/                 Redis client (sessions, caching, rate limiting, presence, pub/sub)
server/handlers/       HTTP handlers, middleware, CORS, helpers
server/websocket/      WebSocket hub, client pumps, message types
populate/              Seed JSON + loader
```

- **Auth**: session cookies (`session_token`, HTTP-only, Lax SameSite, 24h TTL). `AuthMiddleware` sets CORS, rate-limits, resolves the session, injects `userID` into the request context, and recovers panics.
- **Handlers** follow `func SomeHandler(w, r, db)` and are registered via `makeEndpoint(path, requireAuth, handler)`.
- **Background goroutines**: WAL checkpointing, session cleanup, and the event-reminder scheduler, all tied to the shutdown context for graceful drain on SIGINT/SIGTERM.

### Frontend

- **State**: React Context — `AuthProvider` (session), `UIProvider` (toasts), `NotificationProvider` (real-time). Server state via TanStack Query.
- **WebSocket**: `useWebSocket` keeps one singleton connection per tab with auto-reconnect; derives `ws://`/`wss://` from `NEXT_PUBLIC_API_URL`.

## Database Schema

15 tables:

| Table | Purpose |
|-------|---------|
| `users` | Accounts and profile info |
| `sessions` | Session records (SQLite mode) |
| `followers` | Follow relationships + request status |
| `posts` | User and group posts (soft-delete via `is_deleted`) |
| `post_visibility` | Privacy targets for private posts |
| `comments` | Comments with nested replies (soft-delete via `is_deleted`) |
| `groups` | Group metadata |
| `group_members` | Membership with invite status |
| `events` | Group events |
| `event_responses` | RSVP responses |
| `direct_messages` | Private conversation threads |
| `messages` | Individual messages (DM or group) |
| `message_reads` | Read receipts |
| `notifications` | User notifications |
| `oauth_accounts` | OAuth provider links per user |

## Getting Started

### Prerequisites

- Go 1.24+
- Node.js 22+ (Next.js 16 requirement)
- Redis 7 (required at runtime — sessions, caching, rate limiting, presence, and pub/sub all depend on it)
- Docker & Docker Compose (for the full stack)

### Quick Start (from project root)

```bash
make dev
```

- Backend: `http://localhost:8080`
- Frontend: `http://localhost:3000`
- Log in with a seed account (`alice@example.com` / `password123`)

### Docker (dev environment)

```bash
make docker-up
```

- App over HTTPS (Caddy): `https://localhost`
- Backend health: `http://localhost:8080/api/health`

Caddy terminates TLS and routes `/api/*` to the backend and everything else to the Next.js dev server.

> **Note:** on first run the backend creates the database, applies migrations, and loads seed data. Use `make backend-run-reseed` to wipe and reseed.

### Individual commands

| Command | What it does |
|---------|-------------|
| `make backend-run` | Start the Go backend (`go run`) |
| `make backend-run-reseed` | Start backend with fresh seed data |
| `make frontend-dev` | Start the Next.js dev server |
| `make frontend-build` | Production build |
| `make frontend-check` | TypeScript type check |
| `make check` | `go vet` + TypeScript check |
| `make env` | Create `.env` from `.env.example` if missing |
| `make redis-start` / `make redis-stop` | Start/stop Redis |
| `make kill-ports` | Free ports 3000/5173/5174/5175/8080 |
| `make db-reset` / `make db-seed` / `make db-delete` | Manage the database |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | API base URL (frontend) |
| `REDIS_ADDR` | `localhost:6379` | Redis address (backend, required) |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | — | Google OAuth credentials (backend, optional) |
| `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` | — | GitHub OAuth credentials (backend, optional) |

### Configuration

Settings live in `backend/configs.json` (local) / `backend/configs.docker.json` (Docker):

- `scheduler` — event-reminder worker: `enabled`, `tick_interval`, `reminder_lead`.
- `sessions` — `storage` (`"redis"`) and `ttl` (e.g. `"24h"`). Redis is the source of truth; if it is unavailable authentication fails closed.
- `redis` — address, pool size, timeouts, retries.
- `database_configuration` — path, WAL pragmas, `use_cache` (read-through caching), session cleanup interval, and validation `limits` (username/password/name/bio/title/description/post/comment/message/group-title lengths, `rows_limit` pagination cap).
- `handlers` — image constraints and per-path `rate_limits`.
- `oauth` — Google/GitHub provider endpoints, scopes, and redirect URIs. The `client_id`/`client_secret` fields are intentionally blank in the tracked files; supply real values via environment variables (see below).

#### Secrets

OAuth credentials are never committed. Copy the root `.env.example` to `.env` (gitignored) and fill in the four variables above:

```bash
make env            # creates .env from .env.example
# then edit .env and add your Google/GitHub credentials
```

The backend overrides the blank `oauth` fields from those environment variables at load time. Direct `go run` targets via `make` read `.env` automatically, and Docker Compose interpolates the same variables into the backend container. Leave them unset to run without OAuth (the endpoints will be disabled).

### Seed Data

On first launch 7 users are loaded (all with password `password123`): Alice Johnson, Bob Smith, Carol Williams, Dave Brown, Eve Davis, Frank Miller, and Yuki Minakami. The seed also includes posts, comments (some soft-deleted), groups, events, DMs, and notifications for exercising privacy, soft-delete, membership, and event features.

## Project Structure

```
social-network/
├── frontend/
│   ├── src/
│   │   ├── app/           # Next.js App Router (layouts, route groups, error/404)
│   │   ├── api/           # Axios client + endpoint functions
│   │   ├── components/    # UI primitives + feature components
│   │   ├── context/       # Auth + UI + Notification providers
│   │   ├── hooks/         # TanStack Query hooks + WebSocket + emoji autocomplete
│   │   ├── lib/           # Utilities (cn, format, media, validators, nav)
│   │   ├── types/         # TypeScript interfaces (snake_case, matches JSON)
│   │   └── views/         # Page components (mirrors route structure)
│   ├── next.config.ts
│   ├── .env.example
│   └── Dockerfile         # Frontend image (dev, Next.js dev server)
├── backend/
│   ├── cmd/main.go        # Entry point
│   ├── entry/             # Server startup sequence
│   ├── config/            # Config structs + rate-limit/limit helpers
│   ├── configs.json       # All configuration
│   ├── global/            # Path/duration/size helpers
│   ├── db/
│   │   ├── migrations/    # SQL migration files (up/down)
│   │   ├── queries/       # SQL query constants
│   │   └── sql/           # Models, methods, connection
│   ├── server/handlers/   # HTTP handlers + middleware + CORS
│   ├── server/websocket/  # WebSocket hub, client, handler, types
│   ├── cache/             # Redis client (sessions, caching, rl, presence, pub/sub)
│   ├── populate/          # Seed data (seed.json + seed.go)
│   └── Dockerfile         # Backend image (dev, air hot-reload)
├── docker-compose.yml     # dev stack: Caddy + frontend + backend + redis
├── Caddyfile              # TLS termination + reverse proxy (dev)
└── Makefile               # dev, check, build commands
```
