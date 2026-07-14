# CLAUDE.md

Reusable Go backend boilerplate: auth (email/OTP + Google/GitHub OAuth), user profile, and health check, wired up so a new project can start from working authentication instead of rebuilding it.

## Quick Reference

**Stack:** Go, Echo, GORM, PostgreSQL

**Layer structure:**
```
domain/   → database queries (GORM)
service/  → business logic
handler/  → HTTP request/response
route/    → dependency injection and route registration
```

**Run the server:**
```bash
go run main.go
```

## What's included

- Email/password registration with OTP verification, login, logout, session validation (`domain|service|handler/auth.go`, `models/auth.go`)
- Google and GitHub OAuth login (`handler/auth.go`)
- JWT auth via HTTP-only cookie (`middleware/jwt.go`)
- User profile fetch/update (`domain|service|handler/user.go`)
- Health check endpoint (`domain|service|handler/health.go`)
- Email templates for invite/register/reset-password (`template/`)

See [README.md](README.md) for setup and [QUICK_START.md](QUICK_START.md) for the cookie-based auth flow and frontend integration guides.

## Using this as a starting point for a new project

This repo is meant to be cloned/forked per project. When starting something new from it:
- Update `go.mod` module name if it should differ from `core`
- Add your domain-specific models/services/handlers alongside the existing auth/user/health ones
- Remove OAuth providers you don't need from `config/config.go`, `route/v1.go`, and `handler/auth.go`
