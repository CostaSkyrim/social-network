# Social Network

A Facebook-like social network built with Go, TypeScript, Next.js, TanStack Query, and Redis.

## Authors
- **[Konstantinos Petroutsos](https://github.com/CostaSkyrim)**
- **[Augoustinos Andris]**

## Features

- **Followers** — Follow/unfollow users. Public profiles allow instant following. Private profiles require a follow request that the recipient can accept or decline.
- **Profile** — View user profiles with posts, followers, and following lists. Toggle profile visibility between public and private.
- **Posts** — Create text/image posts with privacy levels: public (everyone), followers only, or private (specific followers). Comment on posts with image support. Edit and soft-delete your own posts (deleted posts render as `[deleted]` with comments intact, Reddit-style).
- **Comments** — Nested reply threads. Edit and soft-delete your own comments (deleted comments render as `[deleted]` with replies intact).
- **Emoji support** — Emoji picker button plus `:shortcode:` autocomplete (e.g. type `:cat` → pick 🐱) with keyboard navigation.
- **Groups** — Create groups with title and description. Invite users or accept join requests. Group members can post, comment, create events, and chat in a shared group room.
- **Events** — Group members can create events with title, description, and date/time. RSVP with Going / Not Going / Maybe.
- **Notifications** — Real-time notifications for follow requests, group invitations, group join requests, and new events. Notifications appear across all pages.
- **Chat** — Real-time private messaging between users who follow each other. Group chat rooms for group members. Emoji support. WebSocket-powered instant delivery.
- **Real-time presence** — Online/offline status tracked via Redis (30s TTL keys) and broadcast over the WebSocket hub (`/api/ws`).

## Current Status

**Working:**
- Full auth flow (signup, login, logout, session cookies, CORS, rate limiting)
- Paginated privacy-filtered news feed
- Post + comment CRUD with ownership checks
- Reddit-style soft deletes (`[deleted]` placeholders, nested replies preserved)
- Follow system (request / accept / decline / unfollow)
- User profiles (view + edit)
- Group management (CRUD, browse, invite, join, accept, reject, leave, members)
- Emoji picker + `:shortcode:` autocomplete in the post composer
- WebSocket hub at `/api/ws` (auth required) — chat/group/notification/presence/typing message dispatch, ping/pong keepalive
- Redis integration — presence tracking, JSON caching (sessions/users/posts/groups), sliding-window rate limiting, pub/sub channels

**Planned:** Events, notifications UI, chat UI (WebSocket frontend), Redis session store migration, cross-instance pub/sub, Docker.

See [PLAN.md](./PLAN.md) for the full implementation roadmap.

## Tech Stack

### Frontend
| Tool | Purpose |
|------|---------|
| Next.js 16 | Framework (App Router, Turbopack) |
| React 19 | UI library |
| TypeScript | Type safety |
| TanStack Query v5 | Server state, caching, pagination |
| Axios | HTTP client with session cookie support |
| TailwindCSS | Utility-first styling |
| emoji-picker-react | Emoji picker UI |
| @emoji-mart/data | Emoji shortcode data for autocomplete |

### Backend
| Tool | Purpose |
|------|---------|
| Go 1.24 | Server language |
| SQLite | Persistent storage for all business data |
| Redis 7 | Presence tracking, caching, rate limiting, pub/sub (optional) |
| gorilla/websocket | WebSocket connections (`/api/ws`) |
| golang-migrate | Database migrations |
| golang.org/x/crypto | bcrypt password hashing |

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  Frontend   │     │   Backend    │     │    Redis    │
│  :3000      │────▶│   :8080      │────▶│   :6379     │
│  Next.js    │     │   Go API     │     │  presence,  │
│  App Router │     │   SQLite     │     │  cache, rl  │
│  nginx(prod)│     │   WS :/api/ws│     │  pub/sub    │
└─────────────┘     └──────────────┘     └─────────────┘
```

- Frontend: Next.js App Router with route groups `(auth)` and `(main)` — auth guards redirect unauthenticated users to `/login`.
- Backend: Go REST API with session-cookie auth, rate limiting, and CORS. WebSocket hub on `/api/ws` (auth required).
- SQLite stores all persistent data. Redis is optional — presence tracking, JSON caching, rate limiting, and pub/sub channels. If Redis is down, the backend continues with in-memory fallbacks.

## Database Schema

14 tables across the SQLite database:

| Table | Purpose |
|-------|---------|
| `users` | User accounts with profile info |
| `sessions` | Session tracking (migrating to Redis) |
| `followers` | Follow relationships + request status |
| `posts` | User and group posts (soft-delete via `is_deleted`) |
| `post_visibility` | Privacy controls for private posts |
| `comments` | Post comments with nested replies (soft-delete via `is_deleted`) |
| `groups` | Group metadata |
| `group_members` | Group membership with invite status |
| `events` | Group events |
| `event_responses` | RSVP responses to events |
| `direct_messages` | Private conversation threads |
| `messages` | Individual messages (DM or group) |
| `message_reads` | Read receipts |
| `notifications` | User notifications |

## Getting Started

### Prerequisites

- Go 1.24+
- Node.js 22+ (Next.js 16 requirement)
- Docker and Docker Compose (for Redis or full stack)
- Redis is optional — the backend runs fine without it (in-memory fallbacks)

### Quick Start (from project root)

```bash
make dev
```
- Backend: `http://localhost:8080`
- Frontend: `http://localhost:3000`
- Log in with a seed account (`alice@example.com` / `password123`)

> **Note:** On first run, the backend automatically creates the database, applies migrations, and populates seed data. Use `make backend-run-reseed` to wipe and reseed.

### Individual commands

| Command | What it does |
|---------|-------------|
| `make backend-run` | Start the Go backend |
| `make backend-run-reseed` | Start backend with fresh seed data |
| `make frontend-dev` | Start the Next.js dev server |
| `make frontend-build` | Production build |
| `make frontend-check` | TypeScript type check |
| `make check` | Run `go vet` + TypeScript check |
| `make kill-ports` | Free ports 3000/5173/5174/5175/8080 |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | API base URL (frontend) |
| `REDIS_ADDR` | `localhost:6379` | Redis address (backend, planned) |

### Seed Data

On first launch, 6 users are pre-loaded. All share the same password: `password123`

| Email | Name | Profile |
|-------|------|---------|
| `alice@example.com` | Alice Johnson | Public, active poster |
| `bob@example.com` | Bob Smith | Private profile |
| `carol@example.com` | Carol Williams | Public, group creator |
| `dave@example.com` | Dave Brown | Public, backend dev |
| `eve@example.com` | Eve Davis | Private, lurker |
| `frank@example.com` | Frank Miller | Public, photographer |

The seed also includes 28 posts (2 deleted), 35 comments (3 deleted), 2 groups with events, DMs, and notifications — useful for testing soft-delete rendering and privacy filtering.

## Project Structure

```
social-network/
├── frontend/
│   ├── src/
│   │   ├── app/           # Next.js App Router (layouts, route groups, error/404)
│   │   ├── api/           # Axios client + endpoint functions
│   │   ├── components/    # UI primitives + feature components
│   │   ├── context/       # Auth + UI providers
│   │   ├── hooks/         # TanStack Query hooks + emoji autocomplete
│   │   ├── lib/           # Utilities (cn, format, validators, nav-link)
│   │   ├── types/         # TypeScript interfaces
│   │   └── views/         # Page components (mirrors route structure)
│   ├── next.config.ts
│   ├── Dockerfile         # planned
│   └── nginx.conf         # planned
├── backend/
│   ├── cmd/main.go        # Entry point
│   ├── entry/             # Server startup sequence
│   ├── config/            # Config structs + rate limit helpers
│   ├── configs.json       # All configuration
│   ├── global/            # Path/duration helper functions
│   ├── db/
│   │   ├── migrations/    # SQL migration files (up/down)
│   │   ├── queries/       # SQL query constants
│   │   ├── sql/           # Models + methods + connection
│   │   └── tables/        # Reference table schemas
│   ├── server/handlers/   # HTTP handlers + middleware + CORS
│   ├── server/websocket/  # WebSocket hub, client, handler, types
│   ├── cache/             # Redis client (presence, caching, rate limit)
│   └── populate/          # Seed data (seed.json + seed.go)
├── docker-compose.yml     # planned
├── Makefile               # dev, check, build commands
├── setup-dev.sh           # Distrobox container setup
└── PLAN.md                # Full implementation plan
```
