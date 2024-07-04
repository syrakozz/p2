// Package firebase ...
package firebase

import (
	"github.com/labstack/echo/v4"

	"disruptive/rest/auth"
)

// Routes maps URL path to functions
func Routes(e *echo.Echo) {
	// Admin APIs
	g := e.Group("/api/firebase", auth.SetAdminMiddleware)
	g.GET("/users/:uid", getUserByUID)
	g.PUT("/users/:uid", putUserClaimsByUID)
	g.PATCH("/users/:uid", patchUserClaimsByUID)
	g.DELETE("/users/:uid", deleteUserClaimsByUID)
	g.DELETE("/users/:uid/all", deleteUserClaimsAllByUID)

	g = e.Group("/api/firebase", auth.SetMiddleware)
	g.GET("/accounts/me", getAccountMe)
}
