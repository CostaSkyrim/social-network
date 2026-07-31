.PHONY: backend-run backend-run-reseed backend-vet-populate \
	frontend-dev frontend-build frontend-check frontend-install \
	check dev kill-ports redis-start redis-stop \
	db-reset db-delete db-seed

# ── Redis ───────────────────────────────────────────────
redis-start:
	@echo "==> Starting Redis..."
	@if command -v redis-server >/dev/null 2>&1; then \
		redis-server --daemonize yes 2>/dev/null || true; \
		echo "   Redis started via redis-server"; \
	elif command -v podman >/dev/null 2>&1; then \
		podman run -d --name social-redis -p 6379:6379 docker.io/library/redis:7-alpine 2>/dev/null || true; \
		echo "   Redis started via podman"; \
	elif command -v docker >/dev/null 2>&1; then \
		docker run -d --name social-redis -p 6379:6379 docker.io/library/redis:7-alpine 2>/dev/null || true; \
		echo "   Redis started via docker"; \
	else \
		echo "   ⚠️  No Redis found — install redis-server, podman, or docker"; \
	fi

redis-stop:
	@echo "==> Stopping Redis..."
	@redis-cli shutdown 2>/dev/null || true
	@podman stop social-redis 2>/dev/null || true
	@podman rm social-redis 2>/dev/null || true
	@docker stop social-redis 2>/dev/null || true
	@docker rm social-redis 2>/dev/null || true
	@echo "   Redis stopped"

# ── Backend ───────────────────────────────────────────
backend-run:
	@echo "==> Starting backend server..."
	go run ./backend/cmd/main.go

backend-run-reseed:
	@echo "==> Starting backend server with fresh seed data..."
	go run ./backend/cmd/main.go --reseed

backend-vet-populate:
	@echo "==> Running go vet on populate package..."
	go vet ./backend/populate/ && echo "go vet passed"

# ── Frontend ──────────────────────────────────────────
frontend-install:
	@echo "==> Installing frontend dependencies..."
	cd frontend && npm install

frontend-dev:
	@echo "==> Starting Next.js dev server..."
	cd frontend && npm run dev

frontend-build:
	@echo "==> Building frontend for production..."
	cd frontend && npm run build && echo "Build complete"

frontend-check:
	@echo "==> Running TypeScript type check..."
	cd frontend && npm run check && echo "TypeScript check passed"

# ── Fullstack ──────────────────────────────────────────
check:
	@echo "==> Running all checks..."
	$(MAKE) backend-vet-populate
	$(MAKE) frontend-check
	@echo "All checks passed"

kill-ports:
	@echo "==> Cleaning up stale ports..."
	-fuser -k 5173/tcp 2>/dev/null
	-fuser -k 5174/tcp 2>/dev/null
	-fuser -k 5175/tcp 2>/dev/null
	-fuser -k 3000/tcp 2>/dev/null
	-fuser -k 8080/tcp 2>/dev/null
	@sleep 1

dev: kill-ports redis-start
	@echo "==> Starting frontend and backend..."
	@trap 'kill 0' EXIT; \
	cd frontend && npm run dev & \
	go run ./backend/cmd/main.go & \
	wait

# ── Database ──────────────────────────────────────────
DB_FILE := backend/db/social-network.db
DB_SHARED := $(DB_FILE)-shm $(DB_FILE)-wal

db-delete:
	@echo "==> Deleting database files..."
	rm -f $(DB_FILE) $(DB_SHARED)
	@echo "   Database deleted"

db-seed: db-delete
	@echo "==> Starting backend to seed fresh database..."
	@go run ./backend/cmd/main.go &
	@sleep 3
	@echo "   Seed complete — stop the server with Ctrl+C"
	@wait

db-reset: db-delete
	@echo "==> Reseeding database with --reseed flag..."
	@go run ./backend/cmd/main.go --reseed &
	@sleep 3
	@echo "   Reseed complete — stop the server with Ctrl+C"
	@wait
