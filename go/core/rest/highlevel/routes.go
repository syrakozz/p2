// Package highlevel integrates with HighLevel
package highlevel

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/rest/auth"
)

// Routes maps URL path to functions
func Routes(e *echo.Echo) {
	config := auth.GetJWTConfig()

	g := e.Group("/api/highlevel", echojwt.WithConfig(config))
	g.GET("/contacts", getContacts)
}
