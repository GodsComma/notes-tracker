.PHONY: all dev build dist run clean help

BIN  := bin/notes-server
DIST := dist

# Default target
all: build

# ─── Development ─────────────────────────────────────────────────────────────

# Start both servers. Ctrl-C kills both cleanly.
dev:
	@trap 'kill 0' EXIT; \
	  go run -tags fts5 ./server & \
	  cd client && npm run dev

# ─── Production build ────────────────────────────────────────────────────────

# Build the Vue client into client/dist, then compile the Go binary.
# The binary is run from the project root so it can find both paths:
#   client/dist/             (Vue SPA)
#   server/db/migrations/    (SQL migrations)
build:
	cd client && npm run build
	@mkdir -p bin
	GOOS=linux GOARCH=arm64 go build -tags fts5 -ldflags "-s -w" -o $(BIN) ./server

# ─── Deployable bundle ───────────────────────────────────────────────────────

# Produce a self-contained dist/ directory.  Copy everything the binary needs
# at the correct relative paths so it can be moved and run anywhere:
#
#   dist/
#   ├── notes-server            ← binary
#   ├── client/dist/            ← Vue SPA (served by the binary)
#   ├── server/db/migrations/   ← SQL migrations (read on startup)
#   └── data/                   ← runtime DB + uploads (created automatically)
#
# Deploy:  scp -r dist/ user@host:/opt/notes && ssh user@host 'cd /opt/notes && ./notes-server'
dist: build
	@rm -rf $(DIST)
	@mkdir -p $(DIST)/client $(DIST)/server/db $(DIST)/data
	cp $(BIN) $(DIST)/notes-server
	cp -r client/dist  $(DIST)/client/dist
	cp -r server/db/migrations $(DIST)/server/db/migrations
	@echo ""
	@echo "  Bundle ready in $(DIST)/"
	@echo "  Run with:  cd $(DIST) && ./notes-server"
	@echo ""

# ─── Run ─────────────────────────────────────────────────────────────────────

# Build and run the production binary from the project root.
run: build
	./$(BIN)

# ─── Cleanup ─────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/ client/dist/ $(DIST)/

# ─── Help ────────────────────────────────────────────────────────────────────

help:
	@printf "\nUsage: make <target>\n\n"
	@printf "  %-8s  %s\n" "dev"   "Go backend + Vite dev server (hot reload)"
	@printf "  %-8s  %s\n" "build" "Production build → client/dist/ + bin/notes-server"
	@printf "  %-8s  %s\n" "dist"  "Self-contained deployable bundle in dist/"
	@printf "  %-8s  %s\n" "run"   "Build then start the production server"
	@printf "  %-8s  %s\n" "clean" "Remove bin/, client/dist/, dist/"
	@printf "\n"
