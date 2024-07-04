// Package auth includes common auth functions
package auth

import (
	"github.com/golang-jwt/jwt/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/config"
	"disruptive/lib/users"
)

// GetJWTConfig returns a new jwt config
func GetJWTConfig() echojwt.Config {
	return echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(users.JWTClaims)
		},
		SigningKey:  []byte(config.VARS.JWTSessionSecret),
		TokenLookup: "header:Authorization:Bearer ,cookie:authorization",
	}
}
