# notes-tracker
Feature Full
Production Ready Application

## Getting started

Requires Go, Node.js 18+, and a C compiler (macOS: `xcode-select --install`).

```bash
script/setup     # first time: install deps, create .env, prepare data/
script/server    # run the app at http://localhost:5173
script/test      # run checks
```

## Scripts

All tasks go through [`script/`](script/), following GitHub's
[Scripts to Rule Them All](https://github.com/github/scripts-to-rule-them-all) convention.
Every script supports `--help`.

| Script | Purpose |
|---|---|
| `script/bootstrap` | Install/verify dependencies; create `.env` from `.env.example` |
| `script/setup` | First-time setup, or reset local data (old data is moved to `data.bak-*`) |
| `script/update` | Bring a checkout up to date after `git pull` |
| `script/server` | Run the dev servers (`--prod` runs the built binary) |
| `script/test` | `go vet`, `go test`, client type-check |
| `script/cibuild` | CI entry point: bootstrap, test, linux/arm64 build |
| `script/console` | SQLite console: local, or `production` (read-only) over SSH |
| `script/deploy` | Build and ship to the production host (see [DEPLOY.md](DEPLOY.md)) |
| `script/provision-host` | One-time host setup, run on the server |
