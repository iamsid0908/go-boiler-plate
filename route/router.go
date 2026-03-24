package route

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func InitHttp() *echo.Echo {
	app := App()
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://book-finder0908sid.netlify.app", "http://localhost:3000", "https://commitlens.tech"}, // Add your frontend URLs
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OfficeMesh GitHub App is running")
	})

	v1Routes(e.Group("/v1"), app)
	return e
}
