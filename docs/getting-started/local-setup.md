# Local Setup Guide

This guide gets a fully working Team Health Check stack (frontend + backend + PostgreSQL) running on your laptop. Expected time: **under 10 minutes** with Docker available.

> For environment-specific access (dev / int / prod), see [Environment & Access](./environments.md).

---

## Prerequisites

### Required tools

| Tool | Minimum version | Install |
|------|----------------|---------|
| Git | Any recent | `brew install git` |
| Node.js | 22+ | `brew install node` or [nvm](https://github.com/nvm-sh/nvm) |
| Go | 1.25+ | `brew install go` |
| Docker | 24+ | [Docker Desktop](https://www.docker.com/products/docker-desktop/) |

Check versions:

```bash
node --version   # must be >= 22
go version       # must be >= 1.25
docker --version # must be >= 24
```

### Required access

- Read access to `github.com/guidewire-oss/teams360` (public repo — no approval needed)
- TODO: Internal fork / mirror access if behind a corporate proxy

### Optional tools

| Tool | Purpose |
|------|---------|
| `psql` (PostgreSQL CLI) | Inspect the database directly |
| `air` | Hot-reload for Go backend (`go install github.com/cosmtrek/air@latest`) |
| `ginkgo` CLI | Run backend tests directly (`go install github.com/onsi/ginkgo/v2/ginkgo@latest`) |

---

## 1. Clone the Repository

```bash
git clone https://github.com/guidewire-oss/teams360.git
cd teams360
```

---

## 2. Configure Environment Variables

Copy the provided example and fill in values:

```bash
cp .env.example .env
```

The file is read by the Makefile and passed to both services. Below is a realistic local configuration (all values are safe for local dev — never commit real secrets):

```env
# =============================================================================
# Team Health Check Environment Variables — LOCAL DEV
# =============================================================================

# Database connection (REQUIRED)
DATABASE_URL=postgres://postgres:postgres@localhost:5432/teams360?sslmode=disable

# API server
API_PORT=8080
GIN_MODE=debug

# Frontend
FRONTEND_PORT=3000
NEXT_PUBLIC_API_URL=http://localhost:8080

# Email — leave blank to disable notifications locally
AWS_SES_REGION=
AWS_SES_ACCESS_KEY_ID=
AWS_SES_SECRET_ACCESS_KEY=
SES_FROM_ADDRESS=noreply@teams360.example.com

SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=noreply@teams360.example.com

# Docker image version (unused in local dev)
VERSION=latest
```

### Frontend-only env (optional)

If you need to override frontend env variables directly, create `frontend/.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080

# SSO / OIDC — omit to use username/password only
NEXT_PUBLIC_OAUTH_CLIENT_ID=your-client-id
NEXT_PUBLIC_OAUTH_AUTHORIZE_URL=https://your-provider.com/oauth/authorize
NEXT_PUBLIC_OAUTH_REDIRECT_URI=http://localhost:3000/auth/callback
NEXT_PUBLIC_OAUTH_SCOPES=openid email profile
```

### Backend-only env (optional)

If you need to override backend env variables directly, create `backend/.env`:

```env
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/teams360?sslmode=disable
GIN_MODE=debug

# SSO — must match frontend SSO vars if set
OAUTH_CLIENT_ID=your-client-id
OAUTH_TOKEN_URL=https://your-provider.com/oauth/token
OAUTH_REDIRECT_URI=http://localhost:3000/auth/callback
```

---

## 3. Database Setup

### Option A — Docker (recommended)

The Makefile handles this automatically. To start PostgreSQL only:

```bash
make db-start
```

This runs:

```bash
docker run -d --name teams360-db -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:16-alpine
```

### Option B — Existing PostgreSQL

Set `DATABASE_URL` in your `.env` to point to your PostgreSQL instance:

```bash
DATABASE_URL=postgres://myuser:mypassword@myhost:5432/teams360?sslmode=disable
```

The database and user must already exist. The backend runs migrations automatically on startup.

### Run Migrations and Seed Data

```bash
make db-setup
```

This applies all migrations (`000001` through `000020+`) and seeds:
- 11 health dimensions
- Demo hierarchy levels (VP through Team Member, Admin)
- Demo users and teams

To reset the database (WARNING: deletes all data):

```bash
make db-reset
```

Verify the database is reachable:

```bash
psql "postgres://postgres:postgres@localhost:5432/teams360?sslmode=disable" -c "SELECT count(*) FROM health_dimensions;"
# Expected: count = 11
```

---

## 4. Start the Application

### Option A — One command (recommended)

```bash
make run
```

This installs dependencies (if needed), starts PostgreSQL, runs migrations, and starts both servers. Demo credentials are printed to the terminal.

### Option B — Services separately

**Backend:**

```bash
cd backend
go mod download
go run cmd/api/main.go
```

The API starts at http://localhost:8080.

**Frontend (in a separate terminal):**

```bash
cd frontend
npm install
npm run dev
```

The app starts at http://localhost:3000.

### Option C — Hot reload (development)

```bash
make dev
```

Backend uses `air` for hot reload; frontend uses Next.js's built-in fast refresh.

---

## 5. Validate the Setup

### Check services are up

```bash
# Backend health
curl http://localhost:8080/api/v1/health-dimensions | jq '.[0].name'
# Expected: "Mission" (or the first dimension name)

# Frontend
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000
# Expected: 200
```

### Log in and walk through a basic flow

1. Open http://localhost:3000 in your browser.
2. Log in with the **Team Member** demo account: username `demo`, password `demo`.
3. Navigate to `/home` — you should see your team's health summary.
4. Click **Take Survey** and submit a health check across all 11 dimensions.
5. Log out and log back in as **Team Lead**: username `teamlead1`, password `demo`.
6. Navigate to `/dashboard` — you should see the team's aggregated health data and your response in the list.
7. Log in as **Manager**: username `manager1`, password `demo`.
8. Navigate to `/manager` — you should see health cards for all supervised teams.

### Demo credentials

| Role | Username | Password | Route after login |
|------|----------|----------|------------------|
| Vice President | `vp` | `demo` | `/manager` |
| Director | `director1` | `demo` | `/manager` |
| Manager | `manager1` | `demo` | `/manager` |
| Team Lead | `teamlead1` | `demo` | `/dashboard` |
| Team Member | `demo` | `demo` | `/home` |
| Administrator | `admin` | `admin` | `/admin` |

---

## Troubleshooting

### Mac ARM64 (Apple Silicon) — SWC errors

```bash
npm cache clean --force
rm -rf frontend/node_modules frontend/package-lock.json frontend/.next
cd frontend && npm install
npm install --force @next/swc-darwin-arm64
```

If the above does not work, run the fix script:

```bash
bash fix-mac-issues.sh
```

### PostgreSQL not running

```bash
# Check Docker
docker ps | grep postgres

# Restart the container
make db-start

# Test connection
psql "postgres://postgres:postgres@localhost:5432/teams360?sslmode=disable" -c "SELECT 1"
```

### Port already in use

```bash
# Kill frontend (port 3000)
lsof -ti:3000 | xargs kill -9

# Kill backend (port 8080)
lsof -ti:8080 | xargs kill -9
```

### Backend fails to start — "no such table" or migration errors

```bash
# Re-run migrations and seed
make db-setup

# Or full reset (WARNING: deletes all data)
make db-reset
make db-setup
```

### `go: module lookup disabled by GONOSUMCHECK` or proxy errors

```bash
export GONOSUMCHECK=*
export GOFLAGS=-mod=mod
cd backend && go mod download
```

---

## Next Steps

- [Environment & Access](./environments.md) — connect to dev / int / prod environments
- [Architecture & Key Flows](../architecture/overview.md) — understand the system design
- [Onboarding Checklist](./onboarding.md) — complete your Day-1 checklist
