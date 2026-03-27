# CLAUDE.md
Codex will review your code so be careful

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the server
go run main.go

# Install/update dependencies
go mod tidy

# Build binary (Linux production)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server .
```

There are no tests in this codebase currently.

## Architecture

**Stack:** Go 1.25, Echo framework, GORM, PostgreSQL (with pgvector), Redis (Asynq queue), Azure OpenAI

**Layer structure (Domain-Driven):**
```
domain/   → database queries (GORM)
service/  → business logic, orchestrates domain calls
handler/  → HTTP request/response, calls services
route/    → wires DI graph (domain → service → handler) and registers routes
```

`route/app.go` is the composition root — all dependency injection is done there.

**Authentication:** JWT stored in HTTP-Only `Bearer` cookie (never in response body). Middleware in `middleware/jwt.go` extracts the token from the cookie and injects user context (id, email, name, role, language) into the request. Role-based authorization is handled by `middleware/author.go`.

Supported auth methods:
- Email/password with OTP email verification
- Google OAuth2 (`/auth/google` → `/auth/google/callback`)

**Background jobs:** Asynq tasks (Redis-backed). Queue client in `queue/client.go` enqueues tasks; `queue/worker.go` processes them. Task handlers registered in `route/task_handlers.go`. Used for GitHub repo fetching and Azure AI embedding generation.

**AI/Embeddings:** Azure OpenAI (`text-embedding-ada-002`) generates vector embeddings for commit files, stored in PostgreSQL via pgvector. Embedding-based semantic search is a core feature of the GitHub repository analysis flow.

**All API routes** are grouped under `/v1`. Active route groups:

| Group | Purpose |
|---|---|
| `/auth` | Register, OTP verify, login, logout, validate session, Google OAuth |
| `/user` | Get current user |
| `/workspace` | Create/get workspaces, repo/commit browsing, members |
| `/channel` | Create channels, add users |
| `/connect-org` | GitHub App installation, OAuth, webhooks |
| `/github-repository` | Repo activity, commit details, semantic search, AI explanations, embedding backfill |

> **Note:** `/books`, `/book-summary`, and `/cart` routes exist in the codebase but are currently commented out — they were part of a separate project and are not active.

**CORS** is configured in `route/router.go` to allow `https://book-finder0908sid.netlify.app` and `http://localhost:3000`.

## Environment

Copy `.env.example` to `.env`. Key variables:

| Variable | Purpose |
|---|---|
| `DB_URL` | PostgreSQL connection string |
| `JWT_SECRET` | JWT signing key |
| `REDIS_ADDR` | Redis address (default: `localhost:6379`) |
| `GITHUB_APP_ID`, `GITHUB_PRIVATE_KEY_PATH` | GitHub App credentials |
| `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET` | GitHub OAuth |
| `githubwebhooksecret` | GitHub webhook validation |
| `AZURE_EMBEDDING_*` | Azure OpenAI embedding service |
| `AI_BACKEND_URL` | AI backend service (default: `http://localhost:9000`) |
| `PRIMARY_EMAIL`, `PRIMARY_EMAIL_PASSWORD` | SMTP for OTP emails |
| `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` | Google OAuth2 credentials |
| `GOOGLE_REDIRECT_URL` | Google OAuth2 redirect (e.g. `http://localhost:8080/v1/auth/google/callback`) |
| `FRONTEND_URL` | Frontend base URL for post-auth redirects |
