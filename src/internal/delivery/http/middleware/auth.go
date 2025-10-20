package middleware

import (
	"net/http"
	"pocketeer/internal/platform/authenticator"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// Checks if access_token is present in the session and tries to refresh it if needed
func AuthMiddleware(auth *authenticator.Authenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, _ := session.Get("session", c)

			accessToken, hasAccess := sess.Values["access_token"].(string)
			if !hasAccess || accessToken == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			if auth.IsExpired(accessToken) {
				refreshToken, hasRefresh := sess.Values["refresh_token"].(string)
				if !hasRefresh {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "session expired"})
				}

				newToken, err := auth.RefreshAccessToken(c.Request().Context(), refreshToken)
				if err != nil {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "session expired"})
				}

				sess.Values["access_token"] = newToken.AccessToken
				if err := sess.Save(c.Request(), c.Response()); err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save session"})
				}
			}

			return next(c)
		}
	}
}
