package users

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/golang-jwt/jwt/v4"

	"disruptive/config"
)

// JWTClaims contains the original JWT claims
type JWTClaims struct {
	Username     string   `json:"username"`
	Name         string   `json:"name"`
	Permissions  []string `json:"permissions"`
	LoginSession string   `json:"login_session"`
	jwt.RegisteredClaims
}

// Claims contains processed claims
type Claims struct {
	Username       string
	Name           string
	Permissions    []string
	LoginSession   string
	IsAdmin        bool
	IsServiceAdmin bool
	IsVerified     bool
}

// GetClaims returns unverified JWT claims
func GetClaims(token *jwt.Token) (Claims, bool) {
	jwtclaims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return Claims{}, false
	}

	return Claims{
		Username:     jwtclaims.Username,
		Name:         jwtclaims.Name,
		Permissions:  jwtclaims.Permissions,
		LoginSession: jwtclaims.LoginSession,
	}, true
}

// VerifyAdminClaims returns a Claims and sets the Verified field based on admin permissions.
//
// If the claims permission is "admin", then verified is true.
func VerifyAdminClaims(ctx context.Context, logCtx *slog.Logger, token *jwt.Token) Claims {
	claims, ok := GetClaims(token)
	if !ok {
		return Claims{}
	}

	if !verifyLoginSession(ctx, logCtx, claims.Username, claims.LoginSession) {
		return Claims{}
	}

	// Check claims permissions

	for _, p := range claims.Permissions {
		if p == "admin" {
			claims.IsAdmin = true
			claims.IsVerified = true
			break
		}
	}

	return claims
}

// VerifyServiceAdminClaims returns a Claims and sets the Verified field based on admin permissions.
//
// If the claims permission is "admin" or "[service].admin", then verified is true.
func VerifyServiceAdminClaims(ctx context.Context, logCtx *slog.Logger, token *jwt.Token) Claims {
	claims, ok := GetClaims(token)
	if !ok {
		return Claims{}
	}

	if !verifyLoginSession(ctx, logCtx, claims.Username, claims.LoginSession) {
		return Claims{}
	}

	// Check claims permissions

	for _, p := range claims.Permissions {
		if p == "admin" {
			claims.IsAdmin = true
		}

		if p == fmt.Sprintf("%s.admin", config.VARS.UserAgent) {
			claims.IsServiceAdmin = true
		}
	}

	claims.IsVerified = claims.IsAdmin || claims.IsServiceAdmin

	return claims
}

// VerifyClaims returns a Claims and sets the Verified field based on permissions and/or allowed values.
//
// If the claims permission is "admin" or "[service].admin" or allowed, then verified is true.
// Othewise, permissions must contain a value in allowed.
func VerifyClaims(ctx context.Context, logCtx *slog.Logger, token *jwt.Token, allowed []string) Claims {
	claims, ok := GetClaims(token)
	if !ok {
		return Claims{}
	}

	if !verifyLoginSession(ctx, logCtx, claims.Username, claims.LoginSession) {
		return Claims{}
	}

	// Check claim permissions

	for _, p := range claims.Permissions {
		if p == "admin" {
			claims.IsAdmin = true
		}

		if p == fmt.Sprintf("%s.admin", config.VARS.UserAgent) {
			claims.IsServiceAdmin = true
		}

		for _, a := range allowed {
			if p == a {
				claims.IsVerified = true
			}
		}
	}

	claims.IsVerified = claims.IsAdmin || claims.IsServiceAdmin || claims.IsVerified

	return claims
}
