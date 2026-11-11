# shellcheck shell=bash
# common.sh — shared helpers for script/*. Source it; don't execute it.
#
# Callers must `cd` to the repo root before sourcing so relative paths
# (.env, client/, server/) resolve correctly.

# ─── Colours + logging ────────────────────────────────────────────────────────
if [[ -t 1 ]]; then
    RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; BOLD='\033[1m'; NC='\033[0m'
else
    RED=''; GREEN=''; YELLOW=''; BLUE=''; BOLD=''; NC=''
fi
info()    { echo -e "${BLUE}[info]${NC}  $*"; }
success() { echo -e "${GREEN}[ok]${NC}    $*"; }
warn()    { echo -e "${YELLOW}[warn]${NC}  $*"; }
die()     { echo -e "${RED}[error]${NC} $*" >&2; exit 1; }
step()    { echo -e "\n${BOLD}==> $*${NC}"; }

# ─── Help ─────────────────────────────────────────────────────────────────────
# Usage: handle_help "$@"   — prints $USAGE and exits if -h/--help is present.
handle_help() {
    local arg
    for arg in "$@"; do
        case "$arg" in
            -h|--help) echo "${USAGE:-No help available.}"; exit 0 ;;
        esac
    done
}

# ─── Environment ──────────────────────────────────────────────────────────────
# Load .env into the environment if present. Variables already set in the
# environment win, so `JWT_SECRET=x script/server` overrides the file.
load_env() {
    [[ -f .env ]] || return 0
    local key value
    while IFS='=' read -r key value || [[ -n "$key" ]]; do
        key="${key#"${key%%[![:space:]]*}"}"        # trim leading whitespace
        [[ -z "$key" || "$key" == \#* ]] && continue
        [[ "$key" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] || continue
        # Strip one layer of matching surrounding quotes.
        if [[ "$value" =~ ^\"(.*)\"$ || "$value" =~ ^\'(.*)\'$ ]]; then
            value="${BASH_REMATCH[1]}"
        fi
        [[ -n "${!key+x}" ]] || export "$key=$value"
    done < .env
}

# ─── Checks ───────────────────────────────────────────────────────────────────
# require_cmd <name> <install hint>
require_cmd() { command -v "$1" &>/dev/null || die "$1 not found. $2"; }

# confirm <prompt> — returns 0 on y/Y, 1 otherwise. Non-interactive → no.
confirm() {
    local reply
    [[ -t 0 ]] || return 1
    read -rp "$1 [y/N] " reply
    [[ "$reply" == [yY] || "$reply" == [yY][eE][sS] ]]
}
