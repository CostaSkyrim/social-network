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
┌─────────────┐      ┌──────────────┐      ┌─────────────┐
│  Frontend   │      │   Backend    │      │    Redis    │
│  :5173/:80  │────▶│   :8080      │────▶│   :6379     │
│  React SPA  │      │   Go API     │      │ Sessions +  │
│  nginx(prod)│      │   SQLite     │      │ WS pub/sub  │
└─────────────┘      └──────────────┘      └─────────────┘
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

**1. Start Redis**
```bash
docker run -d -p 6379:6379 redis:7-alpine
```

**2. Start the backend**
```bash
cd backend
go run cmd/main.go
```
Server starts on `http://localhost:8080`.

> **Note:** On first run, the backend automatically populates the database with sample data (6 users, 17 posts, 16 comments, 2 groups, and more). Use `--reseed` to reset:
> ```bash
> go run cmd/main.go --reseed
> ```

**3. Start the frontend**
```bash
cd frontend
npm install
npm run dev
```
App opens on `http://localhost:5173`.

### Development (with Docker Compose)

```bash
docker compose up
```

- Frontend: `http://localhost:80`
- Backend: `http://localhost:8080`
- Redis: `localhost:6379`

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

The seed also includes 17 posts, 16 comments (with replies), 2 groups with events, direct messages, and notifications. To reset and re-seed:
```bash
go run cmd/main.go --reseed
```

## Project Structure

```
social-network/
├── frontend/
│   ├── src/
│   │   ├── api/          # API client + endpoint functions
│   │   ├── components/   # UI primitives + feature components
│   │   ├── context/      # Auth + UI providers
│   │   ├── hooks/        # TanStack Query hooks
│   │   ├── lib/          # Utilities (cn, format, validators)
│   │   ├── pages/        # Route-level page components
│   │   ├── types/        # TypeScript interfaces
│   │   └── ws/           # WebSocket connection manager
│   ├── Dockerfile
│   └── nginx.conf
├── backend/
│   ├── cmd/main.go       # Entry point
│   ├── entry/            # Server startup sequence
│   ├── config/           # OAuth configuration
│   ├── global/           # Global config initialization
│   ├── db/
│   │   ├── migrations/   # SQL migration files (up/down)
│   │   ├── queries/      # SQL query constants
│   │   ├── sql/          # Models + methods + connection
│   │   └── tables/       # Reference table schemas
│   ├── server/handlers/  # HTTP handlers + middleware + WebSocket hub
│   ├── redis/            # Redis session store + pub/sub
│   └── populate/         # Seed data loader (seed.json + seed.go)
├── docker-compose.yml    # Backend + Frontend + Redis
└── PLAN.md               # Detailed implementation plan
```
