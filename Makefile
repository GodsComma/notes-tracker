.PHONY: all dev test build build-native dist run clean help

BIN  := bin/notes-server
DIST := dist

# Build used by `dist`: `build` (linux/arm64) or `build-native` (this machine).
#   make dist BUILD_TARGET=build-native
BUILD_TARGET ?= build

# Default target
all: build

# ─── Development ─────────────────────────────────────────────────────────────
# Workflow targets delegate to script/ (see "Scripts to Rule Them All").

# Go backend + Vite dev server. Ctrl-C kills both cleanly.
dev:
	@script/server

# go vet, go test, client type-check.
test:
	@script/test

# ─── Production build ────────────────────────────────────────────────────────

# Build the Vue client into client/dist, then compile the Go binary.
# The binary is run from the project root so it can find both paths:
#   client/dist/             (Vue SPA)
#   server/db/migrations/    (SQL migrations)
build:
	cd client && npm run build
	@mkdir -p bin
	GOOS=linux GOARCH=arm64 go build -tags fts5 -ldflags "-s -w" -o $(BIN) ./server

# Build for the current machine's OS/architecture (useful when deploying
# directly from the host or for testing the production binary locally).
build-native:
	cd client && npm run build
	@mkdir -p bin
	go build -tags fts5 -ldflags "-s -w" -o $(BIN) ./server

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
# Deploy:  script/deploy
dist: $(BUILD_TARGET)
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
	@printf "\nUsage: make <target>   (or use script/* directly)\n\n"
	@printf "  %-14s  %s\n" "dev"          "Go backend + Vite dev server (script/server)"
	@printf "  %-14s  %s\n" "test"         "go vet, go test, client type-check (script/test)"
	@printf "  %-14s  %s\n" "build"        "Production build → client/dist/ + bin/notes-server (linux/arm64)"
	@printf "  %-14s  %s\n" "build-native" "Same build for the current machine's OS/arch"
	@printf "  %-14s  %s\n" "dist"         "Self-contained deployable bundle in dist/ (BUILD_TARGET=build|build-native)"
	@printf "  %-14s  %s\n" "run"          "Build then start the production server"
	@printf "  %-14s  %s\n" "clean"        "Remove bin/, client/dist/, dist/"
	@printf "\n"
