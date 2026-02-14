# Production-Ready Backend Improvements

This guide includes advanced security features and best practices for your authentication system.

---

## 1. Environment-Based Cookie Configuration

### `config/config.go` - Add Cookie Config
```go
package config

import (
	"os"
	"strconv"
)

type CookieConfig struct {
	Secure   bool
	SameSite string
	Domain   string
	MaxAge   int
}

func GetCookieConfig() CookieConfig {
	secure, _ := strconv.ParseBool(os.Getenv("COOKIE_SECURE"))
	maxAge, _ := strconv.Atoi(os.Getenv("COOKIE_MAX_AGE"))
	
	if maxAge == 0 {
		maxAge = 86400 // Default 24 hours
	}

	return CookieConfig{
		Secure:   secure,
		SameSite: os.Getenv("COOKIE_SAMESITE"),
		Domain:   os.Getenv("COOKIE_DOMAIN"),
		MaxAge:   maxAge,
	}
}

func GetSameSiteMode(mode string) int {
	switch mode {
	case "Strict":
		return 3 // http.SameSiteStrictMode
	case "Lax":
		return 2 // http.SameSiteLaxMode
	case "None":
		return 1 // http.SameSiteNoneMode
	default:
		return 2 // Default to Lax
	}
}
```

### Update `handler/auth.go` - Use Config
```go
package handler

import (
	"core/config"
	"core/handler/validation"
	"core/models"
	"core/service"
	"core/utils"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

type AuthHandler struct {
	AuthService service.AuthService
}

func (authHandler *AuthHandler) setCookie(c echo.Context, token string) {
	cookieConfig := config.GetCookieConfig()
	
	cookie := &http.Cookie{
		Name:     "Bearer",
		Value:    token,
		HttpOnly: true,
		Secure:   cookieConfig.Secure,
		Path:     "/",
		MaxAge:   cookieConfig.MaxAge,
		SameSite: http.SameSite(config.GetSameSiteMode(cookieConfig.SameSite)),
	}

	if cookieConfig.Domain != "" {
		cookie.Domain = cookieConfig.Domain
	}

	c.SetCookie(cookie)
}

func (authHandler *AuthHandler) clearCookie(c echo.Context) {
	cookieConfig := config.GetCookieConfig()
	
	cookie := &http.Cookie{
		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cookieConfig.Secure,
		SameSite: http.SameSite(config.GetSameSiteMode(cookieConfig.SameSite)),
	}

	if cookieConfig.Domain != "" {
		cookie.Domain = cookieConfig.Domain
	}

	c.SetCookie(cookie)
}

func (authHandler *AuthHandler) LoginUser(c echo.Context) error {
	var err error
	param := new(models.LogInRequest)

	err = c.Bind(param)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.BasicResp{Message: err.Error()})
	}

	data, err := authHandler.AuthService.LoginUser(*param)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.BasicResp{Message: err.Error()})
	}

	// Set cookie using helper
	authHandler.setCookie(c, data.Token)

	// Don't send token in response body
	resp := models.BasicResp{
		Message: utils.Success,
		Data: map[string]interface{}{
			"email":   data.Email,
			"name":    data.Name,
			"role":    data.Role,
			"user_id": data.UserID,
		},
	}
	return c.JSON(http.StatusOK, resp)
}

func (authHandler *AuthHandler) UserLogOut(c echo.Context) error {
	authHandler.clearCookie(c)
	return c.JSON(http.StatusOK, models.BasicResp{
		Message: utils.Success,
		Data:    "Logged out successfully",
	})
}
```

---

## 2. Rate Limiting for Auth Endpoints

### Install Rate Limiter
```bash
go get golang.org/x/time/rate
```

### `middleware/ratelimit.go`
```go
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.RWMutex
)

// Clean up old visitors every 3 minutes
func init() {
	go cleanupVisitors()
}

func cleanupVisitors() {
	for {
		time.Sleep(3 * time.Minute)
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Every(time.Minute), 5) // 5 requests per minute
		visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func RateLimit() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			limiter := getVisitor(ip)

			if !limiter.Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "Too many requests. Please try again later.")
			}

			return next(c)
		}
	}
}
```

### Update Routes - `route/v1.go`
```go
import (
	"core/middleware"
	"github.com/labstack/echo"
)

func v1Routes(g *echo.Group, h AppModel) {
	g.GET("/health", h.Health.Check)

	auth := g.Group("/auth")
	
	// Apply rate limiting to auth endpoints
	auth.POST("/register", h.Auth.RegisterUser, middleware.RateLimit())
	auth.POST("/resend-otp", h.Auth.ResendOTP, middleware.RateLimit())
	auth.POST("/verify-otp", h.Auth.VerifyOTP, middleware.RateLimit())
	auth.POST("/login", h.Auth.LoginUser, middleware.RateLimit())
	auth.GET("/validate", h.Auth.ValidateSession, middleware.JWTVerify())
	auth.GET("/logout", h.Auth.UserLogOut, middleware.JWTVerify())
	auth.GET("/github/callback", h.Auth.GithubOAuthCallback, middleware.JWTVerify())

	// ... rest of routes
}
```

---

## 3. Enhanced JWT with Token Rotation

### Update `models/auth.go`
```go
package models

import "github.com/golang-jwt/jwt"

type JWTPayload struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Language string `json:"language"`
	TokenID  string `json:"token_id"` // Add unique token ID
	jwt.StandardClaims
}
```

### `service/auth.go` - Enhanced Token Generation
```go
package service

import (
	"core/config"
	"core/models"
	"core/utils"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func GenerateJWT(user models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	
	claims := &models.JWTPayload{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Role:     user.Role,
		Language: user.Language,
		TokenID:  uuid.New().String(), // Unique token ID for tracking
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "book-finder",
			Subject:   user.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.GetConfig().JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Optional: Store active tokens in Redis for revocation
func StoreActiveToken(userID uint, tokenID string, expiry time.Duration) error {
	// Implementation with Redis or database
	// key := fmt.Sprintf("token:%d:%s", userID, tokenID)
	// return redis.Set(key, "active", expiry)
	return nil
}

func RevokeToken(userID uint, tokenID string) error {
	// Implementation to revoke specific token
	// key := fmt.Sprintf("token:%d:%s", userID, tokenID)
	// return redis.Del(key)
	return nil
}
```

---

## 4. Login Attempt Tracking & Account Lockout

### `models/user.go` - Add Fields
```go
type User struct {
	ID                uint       `json:"id" gorm:"primarykey"`
	Email             string     `json:"email" gorm:"unique"`
	Password          *string    `json:"-" gorm:"type:varchar(255)"`
	Name              string     `json:"name"`
	Role              string     `json:"role"`
	Language          string     `json:"language"`
	IsActive          bool       `json:"is_active"`
	Otp               string     `json:"-"`
	OtpExpiry         time.Time  `json:"-"`
	FailedLoginCount  int        `json:"-" gorm:"default:0"`
	LastFailedLogin   *time.Time `json:"-"`
	AccountLockedUntil *time.Time `json:"-"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
```

### `service/auth.go` - Enhanced Login with Lockout
```go
func (c *AuthService) LoginUser(param models.LogInRequest) (models.LogInResp, error) {
	user, err := c.UserDomain.Get(models.GetUserParam{Email: param.Email})
	if err != nil {
		return models.LogInResp{}, err
	}

	if user.ID == 0 {
		return models.LogInResp{}, utils.ErrUserNotExist
	}

	// Check if account is locked
	if user.AccountLockedUntil != nil && time.Now().Before(*user.AccountLockedUntil) {
		remainingTime := time.Until(*user.AccountLockedUntil).Minutes()
		return models.LogInResp{}, fmt.Errorf("account locked. Try again in %.0f minutes", remainingTime)
	}

	// Check if user is active
	if !user.IsActive {
		return models.LogInResp{}, utils.ErrUserNotActive
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(param.Password))
	if err != nil {
		// Increment failed login count
		failedCount := user.FailedLoginCount + 1
		now := time.Now()
		
		updateUser := models.User{
			ID:               user.ID,
			FailedLoginCount: failedCount,
			LastFailedLogin:  &now,
		}

		// Lock account after 5 failed attempts
		if failedCount >= 5 {
			lockUntil := now.Add(30 * time.Minute) // Lock for 30 minutes
			updateUser.AccountLockedUntil = &lockUntil
		}

		c.UserDomain.Update(updateUser)
		
		return models.LogInResp{}, utils.ErrWrongEmailOrPassword
	}

	// Reset failed login count on successful login
	if user.FailedLoginCount > 0 {
		c.UserDomain.Update(models.User{
			ID:               user.ID,
			FailedLoginCount: 0,
			LastFailedLogin:  nil,
			AccountLockedUntil: nil,
		})
	}

	// Generate JWT token
	token, err := GenerateJWT(user)
	if err != nil {
		return models.LogInResp{}, err
	}

	resp := models.LogInResp{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Role:   user.Role,
		Token:  token,
	}

	return resp, nil
}
```

---

## 5. Security Headers Middleware

### `middleware/security.go`
```go
package middleware

import (
	"github.com/labstack/echo"
)

func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Prevent clickjacking
			c.Response().Header().Set("X-Frame-Options", "DENY")
			
			// Prevent MIME type sniffing
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			
			// XSS Protection
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			
			// Referrer Policy
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			
			// Content Security Policy
			c.Response().Header().Set("Content-Security-Policy", "default-src 'self'")
			
			// HSTS (only in production with HTTPS)
			// c.Response().Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			
			return next(c)
		}
	}
}
```

### Apply in `main.go`
```go
package main

import (
	"core/config"
	"core/middleware"
	"core/route"

	"github.com/labstack/echo"
	echoMiddleware "github.com/labstack/echo/middleware"
)

func main() {
	e := echo.New()

	// Global middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.SecurityHeaders())
	
	// CORS
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins:     []string{"https://yourdomain.com", "http://localhost:3000"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Set-Cookie"},
	}))

	// Initialize database
	config.InitDB()

	// Setup routes
	route.Init(e)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
```

---

## 6. Logging & Monitoring

### `utils/logger.go` - Enhanced Logging
```go
package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

type LogLevel string

const (
	INFO    LogLevel = "INFO"
	WARNING LogLevel = "WARNING"
	ERROR   LogLevel = "ERROR"
	DEBUG   LogLevel = "DEBUG"
)

type Logger struct {
	*log.Logger
}

var AppLogger *Logger

func init() {
	AppLogger = &Logger{
		Logger: log.New(os.Stdout, "", 0),
	}
}

func (l *Logger) log(level LogLevel, message string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formattedMessage := fmt.Sprintf(message, args...)
	l.Printf("[%s] [%s] %s", timestamp, level, formattedMessage)
}

func (l *Logger) Info(message string, args ...interface{}) {
	l.log(INFO, message, args...)
}

func (l *Logger) Warning(message string, args ...interface{}) {
	l.log(WARNING, message, args...)
}

func (l *Logger) Error(message string, args ...interface{}) {
	l.log(ERROR, message, args...)
}

func (l *Logger) Debug(message string, args ...interface{}) {
	l.log(DEBUG, message, args...)
}

// Auth-specific logging
func LogAuthEvent(eventType, email, ip string, success bool) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}
	AppLogger.Info("AUTH_EVENT: %s | Email: %s | IP: %s | Status: %s", 
		eventType, email, ip, status)
}
```

### Use in Auth Handler
```go
func (authHandler *AuthHandler) LoginUser(c echo.Context) error {
	var err error
	param := new(models.LogInRequest)

	err = c.Bind(param)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.BasicResp{Message: err.Error()})
	}

	data, err := authHandler.AuthService.LoginUser(*param)
	
	// Log login attempt
	ip := c.RealIP()
	utils.LogAuthEvent("LOGIN", param.Email, ip, err == nil)

	if err != nil {
		return c.JSON(http.StatusBadRequest, models.BasicResp{Message: err.Error()})
	}

	authHandler.setCookie(c, data.Token)

	resp := models.BasicResp{
		Message: utils.Success,
		Data: map[string]interface{}{
			"email":   data.Email,
			"name":    data.Name,
			"role":    data.Role,
			"user_id": data.UserID,
		},
	}
	return c.JSON(http.StatusOK, resp)
}
```

---

## 7. Environment Variables

### `.env` file
```env
# Application
APP_ENV=development
APP_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=bookfinder

# JWT
JWT_SECRET=your-super-secret-jwt-key-minimum-32-characters-long

# Cookie Configuration
COOKIE_SECURE=false
COOKIE_SAMESITE=Lax
COOKIE_DOMAIN=
COOKIE_MAX_AGE=86400

# Frontend
FRONTEND_URL=http://localhost:3000

# Email (if using)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

### Production `.env`
```env
APP_ENV=production
APP_PORT=8080

DB_HOST=your-production-db-host
DB_PORT=5432
DB_USER=prod_user
DB_PASSWORD=strong_password_here
DB_NAME=bookfinder_prod

JWT_SECRET=very-strong-secret-key-for-production-min-64-chars-recommended

COOKIE_SECURE=true
COOKIE_SAMESITE=None
COOKIE_DOMAIN=.yourdomain.com
COOKIE_MAX_AGE=86400

FRONTEND_URL=https://yourdomain.com
```

---

## 8. Optional: Refresh Token Implementation

If you want to implement refresh tokens for longer sessions:

### `models/auth.go`
```go
type RefreshToken struct {
	ID        uint      `gorm:"primarykey"`
	UserID    uint      `gorm:"index"`
	Token     string    `gorm:"uniqueIndex"`
	ExpiresAt time.Time
	CreatedAt time.Time
	IsRevoked bool      `gorm:"default:false"`
}
```

### New endpoints in `route/v1.go`
```go
auth.POST("/refresh", h.Auth.RefreshToken)
```

This would allow you to have:
- Short-lived access tokens (15 mins)
- Long-lived refresh tokens (7 days)
- Better security with token rotation

---

## Deployment Checklist

- [ ] Set strong JWT_SECRET (minimum 64 characters)
- [ ] Enable COOKIE_SECURE=true in production
- [ ] Set proper COOKIE_DOMAIN for your domain
- [ ] Configure CORS with exact frontend origin (not *)
- [ ] Add rate limiting to all auth endpoints
- [ ] Implement account lockout after failed attempts
- [ ] Add security headers middleware
- [ ] Set up structured logging
- [ ] Use environment variables for all config
- [ ] Add database migration system
- [ ] Set up monitoring (Sentry, Datadog, etc.)
- [ ] Implement health check endpoint
- [ ] Add graceful shutdown
- [ ] Set up SSL/TLS certificates
- [ ] Configure firewall rules
- [ ] Regular security audits

---

## Testing Commands

```bash
# Test login with rate limiting
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","password":"wrong"}' \
    -c cookies.txt
done

# Test session validation
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -b cookies.txt

# Test logout
curl -X GET http://localhost:8080/api/v1/auth/logout \
  -b cookies.txt
```

This implementation is production-ready with industry-standard security practices!
