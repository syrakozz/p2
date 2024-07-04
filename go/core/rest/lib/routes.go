// Package lib manages the low-level lib APIs
package lib

import (
	"github.com/labstack/echo/v4"

	"disruptive/rest/auth"
	"disruptive/rest/lib/deepgram"
	"disruptive/rest/lib/elevenlabs"
	"disruptive/rest/lib/openai"
	"disruptive/rest/lib/stabilityai"
)

// Routes maps URL path to functions
func Routes(e *echo.Echo) {
	// Defaults
	g := e.Group("/api/lib/defaults", auth.SetMiddleware)
	g.POST("/stt/file", deepgram.PostSTTFile)
	g.POST("/stt/stream", deepgram.PostSTTStream)
	g.POST("/tts", elevenlabs.PostTTS)

	// Deepgram
	g = e.Group("/api/lib/deepgram", auth.SetMiddleware)
	g.POST("/stt/file", deepgram.PostSTTFile)
	g.POST("/stt/stream", deepgram.PostSTTStream)

	// Elevenlabs
	g = e.Group("/api/lib/elevenlabs", auth.SetMiddleware)
	g.POST("/tts", elevenlabs.PostTTS)

	// OpenAI
	g = e.Group("/api/lib/openai", auth.SetMiddleware)
	g.POST("/stt/file", openai.PostSTTFile)
	g.POST("/stt/stream", openai.PostSTTStream)

	// StabilityAI
	g = e.Group("/api/lib/stabilityai", auth.SetMiddleware)
	g.POST("/tti", stabilityai.PostTTI)

}
