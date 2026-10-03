# dillonthompson.com

Personal portfolio + consulting site. Live at [dillonthompson.com](https://dillonthompson.com).

A Go (Gin + sqlc + goose) API serves a small content surface — experiences, profile, skills, education — backed by Postgres with JSONB for flexible fields. The React frontend (Vite + Tailwind + shadcn/ui) renders a deploy-themed intro animation, a `Cmd+K` terminal command bar, and expandable cards for resume content. Caddy fronts both, terminating TLS via Let's Encrypt.

In production all three (Caddy + Go + frontend) run in a single container supervised by s6-overlay. CI/CD via GitHub Actions → GHCR → SSH-deploy by image digest. Infrastructure managed with OpenTofu (Hetzner VPS + Cloudflare DNS + rate limiting); database managed via Neon.

For the full deploy story see [deploy/README.md](deploy/README.md).

## Tech stack

| Layer | Choice |
|---|---|
| API | Go 1.26, Gin, sqlc, goose (embedded migrations) |
| Frontend | React 19, Vite 8, Tailwind CSS 4, shadcn/ui, react-router-dom v7 |
| DB | PostgreSQL 16, JSONB for flexible fields, Neon in prod |
| Reverse proxy | Caddy with automatic HTTPS |
| Process supervisor (in-container) | s6-overlay |
| Infra | OpenTofu — Hetzner Cloud + Cloudflare |
| CI/CD | GitHub Actions → GHCR (public package, private repo) |

## Prerequisites

- **Go 1.26+**
- **Node 24** (via [nvm](https://github.com/nvm-sh/nvm) — there's an `.nvmrc`)
- **pnpm** (via [corepack](https://nodejs.org/api/corepack.html); not npm/yarn)
- **Docker** (for Postgres locally and prod-like compose testing)
- **Air** for Go hot reload — `go install github.com/air-verse/air@latest`

## Quick start (native, recommended)

The fastest dev loop runs the API and frontend on the host directly, with Postgres in a container.

### 1. Start a local Postgres

```fish
docker run -d --name dillonthompson-postgres \
  -e POSTGRES_USER=dillonthompson \
  -e POSTGRES_PASSWORD=changeme \
  -e POSTGRES_DB=dillonthompson \
  -p 5432:5432 \
  postgres:16-alpine
```

### 2. Set environment variables

```fish
cp .env.example .env
# .env should contain:
#   POSTGRES_USER=dillonthompson
#   POSTGRES_PASSWORD=changeme
#   POSTGRES_DB=dillonthompson
#   PORT=8080
#   DATABASE_URL=postgres://dillonthompson:changeme@localhost:5432/dillonthompson?sslmode=disable
```

### 3. Run the API

```fish
air
```

Hot-reloads on `.go` file changes. Migrations run automatically on first connect via embedded goose.

### 4. Run the frontend

In another terminal:

```fish
nvm use
cd frontend
pnpm install
pnpm dev
```

Open [http://localhost:5173](http://localhost:5173). Vite proxies `/api/*` to the Go service on `:8080`.

## Quick start (Docker Compose)

If you'd rather run everything in containers:

```fish
docker compose up --build
```

This brings up Postgres, the Go API (with Air), and the Vite dev server — all linked via the compose network. Same hot-reload experience for both API and frontend.

For a closer-to-production layout (compiled Go binary, Caddy fronting a built React bundle), use:

```fish
docker compose -f docker-compose.prod.yml up --build
```

That hits `http://localhost` directly through Caddy.

## Testing the production single image locally

The real production setup is a single container built from `Dockerfile.prod`. To exercise it locally:

```fish
docker build -f Dockerfile.prod -t dillonthompson:test .
docker run --rm \
  -p 8090:80 \
  -e SITE_ADDRESS=:80 \
  -e DATABASE_URL="postgres://dillonthompson:changeme@host.docker.internal:5432/dillonthompson?sslmode=disable" \
  dillonthompson:test
```

`SITE_ADDRESS=:80` disables Caddy's auto-HTTPS for plain-HTTP local testing. Hit `http://localhost:8090`.

## Project layout

```
.
├── cmd/api/                  Go entry point — composition root
├── internal/
│   ├── handlers/             HTTP handlers (health, experience, profile)
│   ├── middleware/           Structured request logging via slog
│   ├── migrations/           SQL migrations + goose runner (embedded via //go:embed)
│   └── repository/           sqlc-generated code + connection pool
├── frontend/
│   ├── src/
│   │   ├── App.tsx           Homepage with deploy intro + expandable cards
│   │   ├── components/       Layout, theme toggle, terminal bar, deploy intro
│   │   ├── hooks/            useTheme, usePageTitle
│   │   └── pages/            NotFound (the only route file outside App.tsx)
│   ├── public/               Favicon + og-image
│   └── vite.config.ts        Tailwind plugin + /api proxy
├── deploy/                   Production deploy artifacts — see deploy/README.md
├── Dockerfile                Multi-container compose's API image
├── Dockerfile.dev            Dev image with Air + sqlc preinstalled
├── Dockerfile.prod           Single-image production build (Caddy + Go + s6)
├── docker-compose.yml        Dev stack with HMR
├── docker-compose.prod.yml   Multi-container prod-like local test
├── sqlc.yaml                 sqlc config (JSONB → json.RawMessage, etc.)
└── .github/workflows/        CI/CD pipeline
```

## Useful commands

```fish
# Regenerate sqlc-typed query code after editing internal/repository/queries.sql
sqlc generate

# Build the production image locally (just to verify it compiles)
docker build -f Dockerfile.prod -t dillonthompson:test .

# Run a one-off goose migration against a local DB
goose -dir internal/migrations postgres "$DATABASE_URL" status

# Regenerate the og-image PNG from the SVG source
cd frontend && node -e '
  const { Resvg } = require("@resvg/resvg-js");
  const fs = require("fs");
  const svg = fs.readFileSync("public/og-image.svg", "utf-8");
  fs.writeFileSync("public/og-image.png", new Resvg(svg, { fitTo: { mode: "width", value: 1200 } }).render().asPng());
'
```

## API surface

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | DB ping + timestamp |
| GET | `/api/v1/experience` | All published experiences, ordered by `sort_order` then start date |
| GET | `/api/v1/profile` | All profile sections as a single object keyed by section name |
| GET | `/api/v1/profile/:key` | Single profile section (e.g. `about`, `skills`, `education`) |

## Deployment

See [deploy/README.md](deploy/README.md) for the full runbook — provisioning Hetzner with OpenTofu, setting up Neon, configuring GitHub Secrets, the deploy workflow, rollback behavior, etc.

## License

All rights reserved.
