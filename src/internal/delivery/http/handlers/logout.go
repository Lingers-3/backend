package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"pocketeer/internal/config"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func LogoutHandler(cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("session", c)
		sess.Options.MaxAge = -1
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to clear session"})
		}

		logoutURL := fmt.Sprintf(
			"https://%s/v2/logout?client_id=%s&returnTo=%s",
			cfg.Auth0Domain,
			cfg.Auth0ClientID,
			url.QueryEscape(cfg.AppBaseUrl),
		)

		return c.Redirect(http.StatusTemporaryRedirect, logoutURL)
	}
}
