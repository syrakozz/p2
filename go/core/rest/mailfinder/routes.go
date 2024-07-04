// Package mailfinder contains REST APIs.
package mailfinder

import "github.com/labstack/echo/v4"

// Routes maps URL path to functions
func Routes(e *echo.Echo) {
	g := e.Group("/api/mailfinder")
	g.GET("/ping", getPing)
	g.GET("/account", getAccount)
	g.POST("/search/employees", postSearchEmployees)
}
