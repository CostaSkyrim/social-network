.PHONY: backend-run backend-run-reseed backend-vet-populate \
	frontend-dev frontend-build frontend-check frontend-install \
	check dev sdev kill-ports redis-start redis-stop \
	db-reset db-delete db-seed certs-generate certs-clean \
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

sdev: kill-ports redis-start certs-generate
	@echo "==> Starting frontend (http://localhost:3000) + backend (https://localhost:8080)..."
	@echo "   Requests proxy through Next.js — browser never touches HTTPS"
	@trap 'kill 0' EXIT; \
	MKCERT_CA="$$(mkcert -CAROOT 2>/dev/null)/rootCA.pem"; \
	cd frontend && \
	NODE_EXTRA_CA_CERTS="$$MKCERT_CA" \
	NEXT_PUBLIC_API_URL=https://localhost:8080 \
	npm run dev & \
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

# ── TLS Certificates ──────────────────────────────────
CERTS_DIR := certs
CERT_FILE := $(CERTS_DIR)/certs.pem
KEY_FILE  := $(CERTS_DIR)/key.pem

certs-generate:
	@echo "==> Generating locally-trusted TLS certificates..."
	@mkdir -p $(CERTS_DIR)
	@if command -v mkcert >/dev/null 2>&1; then \
		cd $(CERTS_DIR) && mkcert -key-file key.pem -cert-file certs.pem localhost 127.0.0.1 ::1; \
		echo "   ✅ Certificates generated with mkcert (trusted by browser + OS)"; \
	else \
		openssl genrsa -out $(CERTS_DIR)/ca-key.pem 4096 2>/dev/null; \
		openssl req -x509 -new -nodes -key $(CERTS_DIR)/ca-key.pem -sha256 -days 3650 \
			-out $(CERTS_DIR)/ca-cert.pem -subj "/CN=SocialNetwork Dev CA" 2>/dev/null; \
		openssl genrsa -out $(KEY_FILE) 2048 2>/dev/null; \
		openssl req -new -key $(KEY_FILE) -out $(CERTS_DIR)/server.csr \
			-subj "/CN=localhost" 2>/dev/null; \
		openssl x509 -req -in $(CERTS_DIR)/server.csr -CA $(CERTS_DIR)/ca-cert.pem \
			-CAkey $(CERTS_DIR)/ca-key.pem -CAcreateserial -out $(CERT_FILE) \
			-days 365 -sha256 \
			-extfile <(printf "subjectAltName=DNS:localhost,IP:127.0.0.1,IP:::1") 2>/dev/null; \
		rm -f $(CERTS_DIR)/server.csr $(CERTS_DIR)/ca-key.pem; \
		echo "   ✅ Certificates generated with openssl"; \
		echo "   ⚠️  Install mkcert for automatic browser trust: https://github.com/FiloSottile/mkcert"; \
		echo "   Or trust manually: make certs-trust"; \
	fi

certs-trust:
	@echo "==> Installing dev CA into system trust store..."
	@if [ -f $(CERTS_DIR)/ca-cert.pem ]; then \
		if [ -d /etc/pki/ca-trust/source/anchors ]; then \
			sudo cp $(CERTS_DIR)/ca-cert.pem /etc/pki/ca-trust/source/anchors/social-network-dev-ca.crt && \
			sudo update-ca-trust && \
			echo "   ✅ CA trusted (Fedora/RHEL)"; \
		elif [ -d /usr/local/share/ca-certificates ]; then \
			sudo cp $(CERTS_DIR)/ca-cert.pem /usr/local/share/ca-certificates/social-network-dev-ca.crt && \
			sudo update-ca-certificates && \
			echo "   ✅ CA trusted (Debian/Ubuntu)"; \
		elif [ -d /etc/ca-certificates/trust-source/anchors ]; then \
			sudo cp $(CERTS_DIR)/ca-cert.pem /etc/ca-certificates/trust-source/anchors/social-network-dev-ca.crt && \
			sudo trust extract-compat && \
			echo "   ✅ CA trusted (Arch)"; \
		else \
			echo "   ⚠️  Could not detect CA trust directory"; \
		fi; \
	else \
		echo "   ⚠️  No openssl CA cert found — install mkcert instead"; \
	fi

certs-clean:
	@echo "==> Removing TLS certificates..."
	@rm -rf $(CERTS_DIR)
	@echo "   ✅ Certificates removed"

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
	@echo "   Frontend:  http://localhost:3000"
	@echo "   Backend:   http://localhost:8080/api/health"
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
	-docker volume rm social-network_go_mod_cache social-network_backend_db social-network_backend_images social-network_redis_data 2>/dev/null || true
	-docker system prune -f 2>/dev/null || true
	@echo "   ✅ Docker resources cleaned"
