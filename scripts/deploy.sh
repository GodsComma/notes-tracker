#!/usr/bin/env bash
# deploy.sh — Build and deploy notes-tracker to the production host.
# Run from the project root on your build machine.
#
# First deploy:  ensure setup-host.sh has already been run on the host.
# Subsequent:    just run this script — it builds, copies, and restarts.
#
# Usage:
#   ./scripts/deploy.sh
#   ./scripts/deploy.sh --native   # build for current arch instead of linux/arm64

set -euo pipefail

# ─── Configuration ────────────────────────────────────────────────────────────
# Edit these three lines before running for the first time.

REMOTE_USER="pi"                 # Your SSH login on the deploy host
REMOTE_HOST="192.168.1.X"        # Host IP or hostname
REMOTE_APP_DIR="/opt/notes/app"  # Must match APP_DIR in setup-host.sh

# ─── Colours ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; BOLD='\033[1m'; NC='\033[0m'
info()    { echo -e "${BLUE}[info]${NC}  $*"; }
success() { echo -e "${GREEN}[ok]${NC}    $*"; }
warn()    { echo -e "${YELLOW}[warn]${NC}  $*"; }
die()     { echo -e "${RED}[error]${NC} $*" >&2; exit 1; }
step()    { echo -e "\n${BOLD}── $* ──${NC}"; }

# ─── Parse args ───────────────────────────────────────────────────────────────
BUILD_TARGET="build"          # 'build' = linux/arm64, 'build-native' = current arch
for arg in "$@"; do
    case $arg in
        --native) BUILD_TARGET="build-native" ;;
        *) die "Unknown argument: $arg" ;;
    esac
done

# ─── Sanity checks ────────────────────────────────────────────────────────────
[[ "$REMOTE_HOST" == "192.168.1.X" ]] && die "Edit REMOTE_HOST in $0 before running."

command -v go    &>/dev/null || die "go not found. Install from https://go.dev"
command -v node  &>/dev/null || die "node not found. Install Node.js 18+."
command -v npm   &>/dev/null || die "npm not found."
command -v rsync &>/dev/null || die "rsync not found."
command -v ssh   &>/dev/null || die "ssh not found."

# Check SSH connection early so we fail fast before a slow build
info "Checking SSH connection to ${REMOTE_USER}@${REMOTE_HOST}..."
ssh -o ConnectTimeout=5 -o BatchMode=yes "${REMOTE_USER}@${REMOTE_HOST}" true \
    || die "Cannot reach ${REMOTE_USER}@${REMOTE_HOST} via SSH. Check your connection."
success "SSH connection OK."

# ─── Build ────────────────────────────────────────────────────────────────────
step "Building"
info "Target: make ${BUILD_TARGET}"
make "${BUILD_TARGET}"
success "Build complete."

# ─── Bundle ───────────────────────────────────────────────────────────────────
step "Bundling"
make dist
BUNDLE_SIZE=$(du -sh dist/ | cut -f1)
success "Bundle ready (${BUNDLE_SIZE})."

# ─── Transfer ─────────────────────────────────────────────────────────────────
step "Transferring"
info "rsync → ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_APP_DIR}/"

# Exclude the runtime data directory — never overwrite live DB or uploads
rsync -az --delete \
    --exclude='data/' \
    --info=progress2 \
    dist/ "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_APP_DIR}/"

success "Transfer complete."

# ─── Restart ──────────────────────────────────────────────────────────────────
step "Restarting service"
ssh "${REMOTE_USER}@${REMOTE_HOST}" bash <<'REMOTE'
    set -e
    sudo systemctl restart notes
    sleep 1
    if sudo systemctl is-active --quiet notes; then
        echo "[ok]    Service is running."
        sudo systemctl status notes --no-pager -l | grep -E 'Active:|Main PID:|listening'
    else
        echo "[error] Service failed to start:" >&2
        sudo journalctl -u notes -n 20 --no-pager >&2
        exit 1
    fi
REMOTE

# ─── Done ─────────────────────────────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "  ${GREEN}Deploy complete!${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  Tail logs:  ssh ${REMOTE_USER}@${REMOTE_HOST} 'journalctl -u notes -f'"
echo ""
