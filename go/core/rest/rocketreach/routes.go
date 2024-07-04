// Package rocketreach integrates with rocketreach.co
package rocketreach

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/rest/auth"
)

// Routes maps URL path to functions
func Routes(e *echo.Echo) {
	config := auth.GetJWTConfig()

	g := e.Group("/api/rocketreach", echojwt.WithConfig(config))
	g.GET("/about", getAbout)
	g.GET("/lookup", getLookup)
	g.POST("/search", postSearch)
}
