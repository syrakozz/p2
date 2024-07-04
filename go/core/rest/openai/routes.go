// Package openai integrates with Monday
package openai

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/rest/auth"
)

// Routes maps URL path to functions
func Routes(e *echo.Echo) {
	config := auth.GetJWTConfig()

	g := e.Group("/api/openai", echojwt.WithConfig(config))
	g.GET("/models", getModels)
	g.POST("/chat", postChat)
	g.POST("/embeddings", postEmbeddings)
	g.POST("/embeddings/similarity", postEmbeddingsSimilarity)
}
