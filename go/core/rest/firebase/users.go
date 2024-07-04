package firebase

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/lib/firebase"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

func getUserByUID(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.firebase.getUserByUID")

	uid := c.Param("uid")

	user, err := firebase.GetUserByUID(ctx, logCtx, uid)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get user")
	}

	return c.JSON(http.StatusOK, user)
}

func putUserClaimsByUID(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.firebase.putUserClaimsByUID")

	uid := c.Param("uid")

	req := map[string]any{}

	if err := (&echo.DefaultBinder{}).BindBody(c, &req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	user, err := firebase.SetUserClaimsByUID(ctx, logCtx, uid, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to set user claims")
	}

	return c.JSON(http.StatusOK, user)
}

func patchUserClaimsByUID(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.firebase.patchUserClaimsByUID")

	uid := c.Param("uid")

	req := map[string]any{}

	if err := (&echo.DefaultBinder{}).BindBody(c, &req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	user, err := firebase.AddUserClaimsByUID(ctx, logCtx, uid, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to patch user claims")
	}

	return c.JSON(http.StatusOK, user)
}

func deleteUserClaimsAllByUID(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.firebase.deleteUserClaimsAllByUID")

	uid := c.Param("uid")

	user, err := firebase.DeleteAllUserClaimsByUID(ctx, logCtx, uid)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to delete all user claims")
	}

	return c.JSON(http.StatusOK, user)
}

func deleteUserClaimsByUID(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.firebase.deleteUserClaimsByUID")

	uid := c.Param("uid")

	req := struct {
		Claims []string `json:"claims"`
	}{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if len(req.Claims) < 1 {
		return e.ErrBad(logCtx, fid, "missing claims")
	}

	user, err := firebase.DeleteUserClaimsByUID(ctx, logCtx, uid, req.Claims)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to delete user claims")
	}

	return c.JSON(http.StatusOK, user)
}
