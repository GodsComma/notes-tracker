# notes-tracker

A self-hosted web app for organizing class notes by semester. Notes are tied
to *sources* (a lecture, a reading, a problem set) so material stays easy to
find by class, date, or topic.

```
Semester
└── Class
    └── Source (Lecture, Reading, Problem Set, …)
        └── Note
```

## Features

- **Notes**: a rich-text editor (TipTap) with math (KaTeX), images, and
  typography. One note per source, and notes can be locked.
- **Sources as a to-do list**: set a priority and completion state, and
  reorder sources by dragging. A semester-wide panel lists outstanding work.
- **Calendar**: class meeting times and source due dates for each class and
  for the whole semester (FullCalendar).
- **Attachments**: upload files to classes or sources and preview them in the
  app.
- **Library / Deep Read**: a PDF reader (pdf.js) that tracks reading progress
  and lets you highlight passages and quote them into notes. Each document
  keeps its own focus settings, and the app logs focus sessions and shows them
  as stats.
- **Grades**: a rubric of weighted categories for each class, used to
  calculate grades.
- **Search**: full-text search over notes using SQLite FTS5.
- **Accounts**: register and log in with JWT auth in an HTTP-only cookie. All
  data is scoped to the user who created it.
- Light and dark themes.

## Stack

| Layer | Tech |
|---|---|
| Backend | Go, [chi](https://github.com/go-chi/chi), SQLite (`mattn/go-sqlite3`, built with `-tags fts5`) |
| Frontend | Vue 3, TypeScript, Vite, Pinia, Vue Router, Tailwind CSS 4 |
| Tests | `go vet` / `go test`, `vue-tsc`, Playwright (e2e) |

In production, one Go binary serves both the JSON API (`/api/v1/...`) and the
built SPA from `client/dist/`. SQL migrations in
[`server/db/migrations/`](server/db/migrations/) run automatically at startup.

## Getting started

Requires Go 1.21+ (the toolchain in `go.mod` is downloaded automatically),
Node.js 18+, and a C compiler for SQLite (on macOS, run
`xcode-select --install`).

```bash
script/setup     # first time: install deps, create .env, prepare data/
script/server    # Go API on :3001 + Vite on http://localhost:5173
script/test      # go vet, go test, client type-check (--e2e adds Playwright)
```

After the server starts, open http://localhost:5173 and register an account.

## Configuration

`script/bootstrap` copies [`.env.example`](.env.example) to `.env` and
generates a `JWT_SECRET`. The `.env` file is gitignored.

| Variable | Default | Purpose |
|---|---|---|
| `JWT_SECRET` | — (required) | Signs auth tokens |
| `PORT` | `3001` | API / server port |
| `DATA_DIR` | `data` | SQLite database (`notes.db`) and uploaded files |
| `CORS_ORIGINS` | `http://localhost:5173` | Comma-separated list of allowed origins |
| `TLS_CERT`, `TLS_KEY` | — | Serve HTTPS and use secure cookies when both are set |
| `HTTP_PORT` | — | With TLS, an extra port that redirects HTTP to HTTPS |
| `DEPLOY_USER`, `DEPLOY_HOST`, `DEPLOY_APP_DIR` | — | Target host for `script/deploy` and `script/console production` |

## Scripts

All tasks go through [`script/`](script/), following GitHub's
[Scripts to Rule Them All](https://github.com/github/scripts-to-rule-them-all) convention.
Every script supports `--help`.

| Script | Purpose |
|---|---|
| `script/bootstrap` | Install or verify dependencies, and create `.env` from `.env.example` (`--e2e` also installs Playwright browsers) |
| `script/setup` | First-time setup, or reset local data (the old data moves to `data.bak-*`) |
| `script/update` | Bring a checkout up to date after `git pull` |
| `script/server` | Run the dev servers (`--prod` builds and runs the native binary) |
| `script/test` | `go vet`, `go test`, and the client type-check against a throwaway `DATA_DIR` (`--e2e` adds Playwright) |
| `script/cibuild` | CI entry point: bootstrap, test, and a linux/arm64 build |
| `script/console` | SQLite console: local, or `production` over SSH (read-only unless `--write`) |
| `script/deploy` | Build and ship to the production host (see [DEPLOY.md](DEPLOY.md)) |
| `script/provision-host` | One-time setup, run on the server |

The [`Makefile`](Makefile) has build targets (`make build`,
`make build-native`, `make dist`). Run `make help` to list them.

## Project layout

```
client/            Vue SPA (src/, Playwright tests in e2e/)
server/
  main.go          Router, static file serving, TLS
  handlers/        API handlers (auth, semesters, classes, sources, notes, …)
  middleware/      JWT auth
  db/              SQLite setup + migrations/
script/            Dev, test, and deploy scripts
data/              Local database and uploads (gitignored)
```

## Deployment

The app targets a Linux ARM64 host (for example, a Raspberry Pi) running
systemd, with a self-signed TLS certificate. Run `script/provision-host` once
on the host, then run `script/deploy` from your machine for each release. See
[DEPLOY.md](DEPLOY.md) for the full guide.
