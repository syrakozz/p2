package firebase

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func getAccountMe(c echo.Context) error {
	fbuser := c.Get("fbuser")
	return c.JSON(http.StatusOK, fbuser)
}
