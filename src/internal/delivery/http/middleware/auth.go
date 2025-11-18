package middleware

import (
	"net/http"
	"strings"

	"pocketeer/internal/platform/authenticator"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// Checks if access_token is present in the session and tries to refresh it if needed
func AuthMiddleware(auth *authenticator.Authenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, _ := session.Get("session", c)

			accessToken, hasAccess := sess.Values["access_token"].(string)
			refreshToken, hasRefresh := sess.Values["refresh_token"].(string)

			// mobile or auth0 m2m: jwt bearer token-based access
			if !hasAccess || accessToken == "" {
				authHeader := c.Request().Header.Get("Authorization")
				if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
					accessToken = strings.TrimSpace(after)
					hasAccess = true
				}

				isValid, err := auth.IsValidAccessToken(c.Request().Context(), accessToken)
				if err != nil || !isValid {
					return echo.NewHTTPError(http.StatusUnauthorized)
				}
			}

			if !hasAccess || accessToken == "" {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			if auth.IsExpired(accessToken) {
				if !hasRefresh {
					return echo.NewHTTPError(http.StatusUnauthorized)
				}

				newToken, err := auth.RefreshAccessToken(c.Request().Context(), refreshToken)
				if err != nil {
					return echo.NewHTTPError(http.StatusUnauthorized)
				}

				sess.Values["access_token"] = newToken.AccessToken
				if err := sess.Save(c.Request(), c.Response()); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError)
				}
			}

			return next(c)
		}
	}
}

func GetAuth0IDFromRequest(c echo.Context) (string, bool) {
	var tokenStr string

	sess, _ := session.Get("session", c)
	if tok, ok := sess.Values["access_token"].(string); ok && tok != "" {
		tokenStr = tok
	} else {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return "", false
		}
		tokenStr = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}

	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return "", false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", false
	}

	return sub, true
}

func GetAccessTokenFromRequest(c echo.Context) (string, bool) {
	var tokenStr string

	sess, _ := session.Get("session", c)
	if tok, ok := sess.Values["access_token"].(string); ok && tok != "" {
		tokenStr = tok
	} else {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return "", false
		}
		tokenStr = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}

	return tokenStr, true
}
