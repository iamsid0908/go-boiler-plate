package route

import (
	"core/middleware"

	"github.com/labstack/echo"
)

func v1Routes(g *echo.Group, h AppModel) {
	g.GET("/health", h.Health.Check)

	auth := g.Group("/auth")
	auth.POST("/register", h.Auth.RegisterUser)
	auth.POST("/resend-otp", h.Auth.ResendOTP)
	auth.POST("/verify-otp", h.Auth.VerifyOTP)
	auth.POST("/login", h.Auth.LoginUser)
	auth.GET("/validate", h.Auth.ValidateSession, middleware.JWTVerify())
	auth.GET("/logout", h.Auth.UserLogOut, middleware.JWTVerify())
	auth.GET("/github/callback", h.Auth.GithubOAuthCallback, middleware.JWTVerify())
	auth.GET("/google", h.Auth.GoogleAuthURL)
	auth.GET("/google/callback", h.Auth.GoogleOAuthCallback)
	auth.GET("/github", h.Auth.GithubAuthURL)
	auth.GET("/github/callback", h.Auth.GithubAuthCallback)

	user := g.Group("/user", middleware.JWTVerify())
	user.GET("/get-user", h.User.GetUserName)
	user.POST("/update-profile", h.User.UpdateUserProfile)

}
