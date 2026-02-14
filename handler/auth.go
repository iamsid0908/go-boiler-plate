package handler

import (
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

func (authHandler *AuthHandler) RegisterUser(c echo.Context) error {
	var err error
	param := new(models.RegisterUserRequest)
	param.Role = "customer"
	param.IsActive = false

	err = c.Bind(param)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.BasicResp{Message: err.Error()})
	}
	err = validation.RegisterUser(param)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.BasicResp{Message: err.Error()})
	}

	data, err := authHandler.AuthService.RegisterUser(param)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.BasicResp{Message: err.Error()})

	}
	resp := models.BasicResp{
		Message: utils.Success,
		Data:    data,
	}

	return c.JSON(http.StatusOK, resp)
}

func (authHandler *AuthHandler) ResendOTP(c echo.Context) error {
	var err error
	param := new(models.ResendOTPRequest)

	err = c.Bind(param)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.BasicResp{Message: err.Error()})
	}
	err = validation.ResendOTP(param)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.BasicResp{Message: err.Error()})
	}

	err = authHandler.AuthService.ResendOTP(param)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.BasicResp{Message: err.Error()})

	}
	resp := models.BasicResp{
		Message: utils.Success,
		Data:    "OTP sent successfully",
	}

	return c.JSON(http.StatusOK, resp)
}

func (authHandler *AuthHandler) VerifyOTP(c echo.Context) error {
	var err error
	param := new(models.VerifyOTPRequest)

	err = c.Bind(param)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.BasicResp{Message: err.Error()})
	}
	err = validation.VerifyOTP(param)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.BasicResp{Message: err.Error()})
	}

	err = authHandler.AuthService.VerifyOTP(*param)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.BasicResp{Message: err.Error()})

	}
	resp := models.BasicResp{
		Message: utils.Success,
		Data:    "OTP verified successfully",
	}

	return c.JSON(http.StatusOK, resp)
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
	cookie := &http.Cookie{
		Name:     "Bearer",
		Value:    data.Token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		SameSite: http.SameSiteNoneMode,
	}
	c.SetCookie(cookie)

	// Don't send token in response body when using HTTP-only cookies
	resp := models.BasicResp{
		Message: utils.Success,
		Data: map[string]interface{}{
			"email":   data.Email,
			"name":    data.Name,
			"role":    data.Role,
			"user_id": data.ID,
		},
	}
	return c.JSON(http.StatusOK, resp)
}

func (authHandler *AuthHandler) UserLogOut(c echo.Context) error {
	cookie := &http.Cookie{
		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, models.BasicResp{
		Message: utils.Success,
		Data:    "Logged out successfully",
	})
}

// ValidateSession - Check if user's session is valid
func (authHandler *AuthHandler) ValidateSession(c echo.Context) error {
	// If middleware passed, user is authenticated
	// Extract user data set by middleware
	userData := map[string]interface{}{
		"id":       c.Get("id"),
		"email":    c.Get("email"),
		"name":     c.Get("name"),
		"role":     c.Get("role"),
		"language": c.Get("language"),
	}

	resp := models.BasicResp{
		Message: utils.Success,
		Data:    userData,
	}
	return c.JSON(http.StatusOK, resp)
}

func (authHandler *AuthHandler) GithubOAuthCallback(c echo.Context) error {
	installationID := c.QueryParam("installation_id")
	setupAction := c.QueryParam("setup_action")

	if installationID == "" {
		return c.JSON(400, models.BasicResp{Message: "installation_id is required"})
	}

	// Get user ID from JWT token (set by middleware)
	userID := c.Get("user_id") // Adjust based on your JWT middleware implementation
	fmt.Println("User ID from token:", userID)
	fmt.Println("Installation ID:", installationID)
	fmt.Println("Setup Action:", setupAction)
	// Store installation in DB via service
	// err := connectOrgHandler.ConnectOrgService.StoreInstallation(userID, installationID, setupAction)
	// if err != nil {
	// 	return c.JSON(500, models.BasicResp{Message: "Failed to store installation: " + err.Error()})
	// }

	resp := models.BasicResp{
		Message: "Installation successful",
		Data:    map[string]string{"installation_id": installationID}}
	return c.JSON(200, resp)
}
