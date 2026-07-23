.PHONY: backend-run backend-run-reseed backend-vet-populate \
	frontend-dev frontend-build frontend-check frontend-install \
	check dev kill-ports

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
	@echo "==> Starting Vite dev server..."
	cd frontend && npm run dev

frontend-build:
	@echo "==> Building frontend for production..."
	cd frontend && npm run build && echo "Build complete"

frontend-check:
	@echo "==> Running TypeScript type check..."
	cd frontend && npx tsc --noEmit && echo "TypeScript check passed"

# ── Fullstack ──────────────────────────────────────────
check:
	@echo "==> Running all checks..."
	$(MAKE) backend-vet-populate
	$(MAKE) frontend-check
	@echo "All checks passed"

kill-ports:
	@echo "==> Cleaning up stale ports..."
	@fuser -k 5173/tcp 2>/dev/null; fuser -k 5174/tcp 2>/dev/null; fuser -k 5175/tcp 2>/dev/null
	@fuser -k 8080/tcp 2>/dev/null; sleep 1

dev: kill-ports
	@echo "==> Starting frontend and backend..."
	@trap 'kill 0' EXIT; \
	cd frontend && npm run dev & \
	go run ./backend/cmd/main.go & \
	wait
