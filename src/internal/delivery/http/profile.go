package http

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func ProfileHandler(c echo.Context) error {
	sess, _ := session.Get("session", c)
	profile := sess.Values["profile"]
	return c.JSON(http.StatusOK, profile)
}
