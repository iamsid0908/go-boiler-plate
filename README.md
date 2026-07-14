# Go Boilerplate

A reusable Go backend starter with authentication already built in: email/OTP registration, login/logout with HTTP-only JWT cookies, session validation, and Google/GitHub OAuth login.

## Stack

Go, [Echo](https://echo.labstack.com/), [GORM](https://gorm.io/) + PostgreSQL

## Setup

```bash
git clone <your-repo-url>
cd go-boilerplate
cp .env.example .env   # fill in DB credentials, JWT secret, OAuth keys, etc.
go mod tidy
go run main.go
```

The server starts on the port set by `PORT` in `.env` (defaults used in `.env.example`: `8000`).

## Layout

```
domain/   → database queries (GORM)
service/  → business logic
handler/  → HTTP request/response
route/    → dependency injection and route registration
models/   → request/response and DB models
middleware/ → JWT verification, auth guards
template/ → email templates (invite, register, reset password)
```

## Auth endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Register a new user |
| POST | `/api/v1/auth/resend-otp` | Resend OTP |
| POST | `/api/v1/auth/verify-otp` | Verify email OTP |
| POST | `/api/v1/auth/login` | Login, sets HTTP-only cookie |
| GET | `/api/v1/auth/validate` | Check session validity |
| GET | `/api/v1/auth/logout` | Logout, clears cookie |
| GET | `/api/v1/auth/google` | Start Google OAuth flow |
| GET | `/api/v1/auth/google/callback` | Google OAuth callback |
| GET | `/api/v1/auth/github` | Start GitHub OAuth flow |
| GET | `/api/v1/auth/github/callback` | GitHub OAuth callback |
| GET | `/api/v1/user/get-user` | Get current user |
| POST | `/api/v1/user/update-profile` | Update current user's profile |

See [QUICK_START.md](QUICK_START.md) for the cookie-based auth flow, and [FRONTEND_AUTH_GUIDE.md](FRONTEND_AUTH_GUIDE.md) / [NEXTJS_AUTH_IMPLEMENTATION.md](NEXTJS_AUTH_IMPLEMENTATION.md) for frontend integration examples.
