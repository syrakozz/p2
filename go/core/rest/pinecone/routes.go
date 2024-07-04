// Package pinecone integrates with pinecone.io
package pinecone

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/rest/auth"
)

// Routes maps URL path to functions
func Routes(e *echo.Echo) {
	config := auth.GetJWTConfig()

	g := e.Group("/api/pinecone", echojwt.WithConfig(config))
	g.POST("/stats", postStats)
	g.POST("/documents", postDocuments)
	g.GET("/documents", getDocuments)
	g.PATCH("/documents", patchDocuments)
	g.DELETE("/documents", deleteDocuments)
	g.POST("/query", postQuery)
}
