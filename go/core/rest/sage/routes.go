package sage

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/rest/auth"
)

// Routes maps URL path to functions
func Routes(e *echo.Echo) {
	config := auth.GetJWTConfig()

	g := e.Group("/api/sage", echojwt.WithConfig(config))
	g.POST("", postUsers)
	g.GET("", getUser)
	g.PATCH("", patchUser)

	g.POST("/projects", postProjects)
	g.GET("/projects", getProjects)
	g.GET("/projects/:_project", getProject)
	g.PATCH("/projects/:_project", patchProject)
	g.DELETE("/projects/:_project", deleteProject)

	g.POST("/projects/:_project/sessions", postSessions)
	g.GET("/projects/:_project/sessions", getSessions)
	g.GET("/projects/:_project/sessions/:_session", getSession)
	g.PATCH("/projects/:_project/sessions/:_session", patchSession)
	g.DELETE("/projects/:_project/sessions/:_session", deleteSession)

	g.POST("/projects/:_project/sessions/:_session/models", postModels)
	g.GET("/projects/:_project/sessions/:_session/models/:_model", getModel)
	g.PATCH("/projects/:_project/sessions/:_session/models/:_model", patchModel)

	g.POST("/projects/:_project/sessions/:_session/models/:_model/chat", postModelChat)
	g.POST("/projects/:_project/sessions/:_session/models/:_model/chat/stream-finalize", postModelChatStreamFinalize)

	g.POST("/projects/:_project/sessions/:_session/models/:_model/chatpad", postModelChatpad)
	g.POST("/projects/:_project/sessions/:_session/models/:_model/chatpad/stream-finalize", postModelChatpadStreamFinalize)

	g.POST("/projects/:_project/sessions/:_session/models/:_model/entries/:_entry", postModelEntry)
	g.GET("/projects/:_project/sessions/:_session/models/:_model/entries/:_entry", getModelEntry)

	g.POST("/projects/:_project/sessions/:_session/models/:_model/entries/:_entry/versions/:_version", postModelEntryVersion)
	g.PUT("/projects/:_project/sessions/:_session/models/:_model/entries/:_entry/versions/:_version", putModelEntryVersion)
}
