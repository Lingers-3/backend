package http

import (
	"net/http"
	"pocketeer/internal/platform/authenticator"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func CallbackHandler(auth *authenticator.Authenticator) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("session", c)

		if c.QueryParam("state") != sess.Values["state"] {
			return c.String(http.StatusBadRequest, "Invalid state parameter.")
		}

		verifier, ok := sess.Values["code_verifier"].(string)
		if !ok {
			return c.String(http.StatusBadRequest, "Code verifier not found in session.")
		}

		token, err := auth.ExchangeWithPKCE(c.Request().Context(), c.QueryParam("code"), verifier)
		if err != nil {
			return c.String(http.StatusUnauthorized, "Failed to exchange an authorization code for a token.")
		}

		idToken, err := auth.VerifyIDToken(c.Request().Context(), token)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to verify ID Token.")
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		sess.Values["access_token"] = token.AccessToken
		sess.Values["profile"] = profile
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.Redirect(http.StatusTemporaryRedirect, "/profile")
	}
}
