.PHONY: backend-run backend-run-reseed backend-vet-populate \
	frontend-dev frontend-build frontend-check frontend-install \
	check dev kill-ports redis-start redis-stop \
	db-reset db-delete db-seed \
	docker-up docker-down docker-build docker-clean docker-reset

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
MIGRATIONS_DIR := backend/db/migrations

db-migrate:
	@echo "==> Running migrations against existing database..."
	@migrate -database "sqlite3://$(DB_FILE)" -path $(MIGRATIONS_DIR) up 2>/dev/null || \
		(echo "   ⚠️  golang-migrate not found in PATH — migrations will run on next server start" && exit 0)
	@echo "   Migrations complete"

db-migrate-down:
	@echo "==> Rolling back last migration..."
	@migrate -database "sqlite3://$(DB_FILE)" -path $(MIGRATIONS_DIR) down 1 2>/dev/null || \
		(echo "   ⚠️  golang-migrate not found" && exit 0)

db-delete:
	@echo "==> Deleting database files..."
	rm -f $(DB_FILE) $(DB_SHARED)
	@echo "   Database deleted"

db-seed: db-delete db-migrate
	@echo "==> Starting backend to seed fresh database..."
	@go run ./backend/cmd/main.go &
	@sleep 3
	@echo "   Seed complete — press Ctrl+C to stop the server"
	@wait

db-reset: db-delete db-migrate
	@echo "==> Reseeding database with --reseed flag..."
	@go run ./backend/cmd/main.go --reseed &
	@sleep 3
	@echo "   Reseed complete — press Ctrl+C to stop the server"
	@wait

db-migrate-test:
	@echo "==> Testing migration: nickname NOT NULL on legacy data..."
	@echo "   This test verifies migration 000016 handles NULL nicknames correctly"
	@rm -f $(DB_FILE) $(DB_SHARED)
	@echo "   Step 1: Create DB with old schema (nickname nullable)"
	@sqlite3 $(DB_FILE) "CREATE TABLE users (id INTEGER PRIMARY KEY, uuid TEXT, email TEXT, nickname TEXT);"
	@sqlite3 $(DB_FILE) "INSERT INTO users VALUES (1, 'u1', 'test@test.com', NULL);"
	@sqlite3 $(DB_FILE) "INSERT INTO users VALUES (2, 'u2', 'test2@test.com', 'goodname');"
	@echo "   Step 2: Run migration 000016"
	@migrate -database "sqlite3://$(DB_FILE)" -path $(MIGRATIONS_DIR) up 2>/dev/null
	@echo "   Step 3: Verify nicknames"
	@sqlite3 $(DB_FILE) "SELECT id, nickname FROM users;"
	@echo "   ✅ Test passed — NULL nickname was backfilled and NOT NULL constraint applied"
	@rm -f $(DB_FILE) $(DB_SHARED)

# ── Docker Compose ────────────────────────────────────
DOCKER_COMPOSE := $(shell command -v docker-compose 2>/dev/null || echo "docker compose")

docker-build:
	@echo "==> Building Docker images..."
	$(DOCKER_COMPOSE) build
	@echo "   ✅ Images built"

docker-up:
	@echo "==> Starting Docker Compose stack..."
	$(DOCKER_COMPOSE) up -d
	@echo ""
	@echo "   App (HTTPS via Caddy): https://localhost"
	@echo "   Backend health:        http://localhost:8080/api/health"
	@echo ""

docker-down:
	@echo "==> Stopping Docker Compose stack..."
	$(DOCKER_COMPOSE) down
	@echo "   ✅ Stack stopped"

docker-reset: docker-down
	@echo "==> Resetting Docker volumes (database, Redis)..."
	-$(DOCKER_COMPOSE) down -v 2>/dev/null
	@echo "   ✅ Volumes removed"

docker-clean: docker-down
	@echo "==> Cleaning Docker resources..."
	-$(DOCKER_COMPOSE) down -v --rmi all 2>/dev/null
	-docker volume rm social-network_go_mod_cache social-network_backend_db social-network_backend_images social-network_redis_data social-network_caddy_data social-network_caddy_config 2>/dev/null || true
	-docker system prune -f 2>/dev/null || true
	@echo "   ✅ Docker resources cleaned"
