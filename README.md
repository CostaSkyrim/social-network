# Social Network

A Facebook-like social network built with Go, TypeScript, React + Tanstack Query, and Redis.

## Authors
- **[Konstantinos Petroutsos](https://github.com/CostaSkyrim)**
- **[Augoustinos Andris]**

## Features

- **Followers** — Follow/unfollow users. Public profiles allow instant following. Private profiles require a follow request that the recipient can accept or decline.
- **Profile** — View user profiles with posts, followers, and following lists. Toggle profile visibility between public and private.
- **Posts** — Create text/image posts with privacy levels: public (everyone), almost private (followers only), or private (specific followers). Comment on posts with image support. Vote on posts and comments.
- **Groups** — Create groups with title and description. Invite users or accept join requests. Group members can post, comment, create events, and chat in a shared group room.
- **Events** — Group members can create events with title, description, and date/time. RSVP with Going / Not Going / Maybe.
- **Notifications** — Real-time notifications for follow requests, group invitations, group join requests, and new events. Notifications appear across all pages.
- **Chat** — Real-time private messaging between users who follow each other. Group chat rooms for group members. Emoji support. WebSocket-powered instant delivery.

## Current Status

The backend auth system is fully functional (signup, login, logout with session cookies + CORS + rate limiting). Seed data auto-populates 6 users on first run. The rest of the features (profiles, posts, groups, chat, WebSocket) are implemented in the DB layer with query constants and need handler wiring.

See [PLAN.md](./PLAN.md) for the full implementation roadmap.

## Tech Stack

### Frontend
| Tool | Purpose |
|------|---------|
| Vite | Build tool and dev server |
| React 18 | UI framework |
| TypeScript | Type safety |
| React Router v6 | Client-side routing |
| TanStack Query v5 | Server state, caching, pagination |
| Axios | HTTP client with session cookie support |
| TailwindCSS | Utility-first styling |
| WebSocket | Real-time notifications and chat |

### Backend
| Tool | Purpose |
|------|---------|
| Go 1.24 | Server language |
| SQLite | Persistent storage for all business data |
| Redis 7 | Session storage with TTL + WebSocket pub/sub |
| golang-migrate | Database migrations |
| gorilla/websocket | WebSocket connections |

## Architecture

```
┌─────────────┐     ┌──────────────┐
│  Frontend   │     │   Backend    │
│  :5173      │────▶│   :8080      │
│  React SPA  │     │   Go API     │
│  Vite(HMR)  │     │   SQLite     │
│  nginx(prod)│     │   Redis(opt) │
└─────────────┘     └──────────────┘
```

- Frontend serves static files via nginx in production, Vite dev server in development
- Backend serves REST API + WebSocket endpoint
- Redis stores session mappings (`session_id → user_id`) with automatic TTL expiry
- Redis pub/sub enables WebSocket message broadcasting across instances
- SQLite stores all persistent data (users, posts, comments, groups, messages, etc.)

## Database Schema

14 tables across the SQLite database:

| Table | Purpose |
|-------|---------|
| `users` | User accounts with profile info |
| `sessions` | Session tracking (migrating to Redis) |
| `followers` | Follow relationships + request status |
| `posts` | User and group posts |
| `post_visibility` | Privacy controls for private posts |
| `comments` | Post comments with nested replies |
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
- Node.js 20+
- Docker and Docker Compose (for Redis or full stack)

### Development (without Docker)

**1. Start Redis** (optional — only needed for Redis integration)
```bash
docker run -d -p 6379:6379 redis:7-alpine
```

**2. Start the backend** (from project root)
```bash
make backend-run
```
Server starts on `http://localhost:8080`.

> **Note:** On first run, the backend automatically populates the database with sample data (6 users, 28 posts, 35 comments, 2 groups, and more). Use `--reseed` to reset:
> ```bash
> make backend-run-reseed
> ```

**3. Start the frontend** (from project root)
```bash
make frontend-dev
```
Or start both together:
```bash
make dev
```

### Quick Start (from project root)
```bash
make dev
```
This starts the backend (`:8080`) and frontend (`:5173`) simultaneously.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_URL` | `http://localhost:8080` | API base URL (frontend) |
| `REDIS_ADDR` | `localhost:6379` | Redis address (backend) |

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

The seed also includes 28 posts, 35 comments (with replies), 2 groups with events, DMs, and notifications. To reset:
```bash
make backend-run-reseed
```

## Project Structure

```
social-network/
├── frontend/
│   ├── src/
│   │   ├── api/          # API client + endpoint functions
│   │   ├── components/   # UI primitives + feature components
│   │   ├── context/      # Auth + UI providers
│   │   ├── hooks/        # TanStack Query hooks (planned)
│   │   ├── lib/          # Utilities (cn, format, validators)
│   │   ├── pages/        # Route-level page components
│   │   ├── types/        # TypeScript interfaces
│   │   └── ws/           # WebSocket connection manager (planned)
│   ├── Dockerfile        # planned
│   └── nginx.conf        # planned
├── backend/
│   ├── cmd/main.go       # Entry point
│   ├── entry/            # Server startup sequence
│   ├── config/           # Config structs + rate limit helpers
│   ├── configs.json      # All configuration
│   ├── global/           # Path/duration helper functions
│   ├── db/
│   │   ├── migrations/   # SQL migration files (up/down)
│   │   ├── queries/      # SQL query constants
│   │   ├── sql/          # Models + methods + connection
│   │   └── tables/       # Reference table schemas
│   ├── server/handlers/  # HTTP handlers + middleware + CORS
│   ├── populate/         # Seed data (seed.json + seed.go)
│   └── redis/            # planned
├── docker-compose.yml    # planned
├── Makefile              # dev, check, build commands
├── setup-dev.sh          # Distrobox container setup
└── PLAN.md               # Full implementation plan
```
