// Package api contains REST APIs.
package api

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/rest/auth"
)

// Routes maps URL path to functions
func Routes(e *echo.Echo) {
	config := auth.GetJWTConfig()

	g := e.Group("/api")
	g.POST("/login", postLogin)
	g.POST("/logout", postLogout, echojwt.WithConfig(config))
	g.GET("/ping", getPing)
	g.GET("/info", getInfo)

	g = e.Group("/api/users", echojwt.WithConfig(config))
	g.POST("", postUsers)
	g.GET("", getUsers)

	g.GET("/me", getUserMe)
	g.GET("/:username", getUser)
	g.PATCH("/:username", patchUser)
	g.DELETE("/:username", deleteUser)

	g = e.Group("/api/configs", auth.SetMiddleware)
	g.GET("/:config", getConfig)
}
