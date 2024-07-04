// Package monday integrates with Monday
package monday

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/rest/auth"
)

// Routes maps URL path to functions
func Routes(e *echo.Echo) {
	config := auth.GetJWTConfig()

	g := e.Group("/api/monday", echojwt.WithConfig(config))
	g.GET("/ping", getPing)
	g.POST("/query", postQuery)
}
