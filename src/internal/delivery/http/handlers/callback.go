package handlers

import (
	"fmt"
	"net/http"
	"pocketeer/internal/platform/authenticator"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func CallbackHandler(auth *authenticator.Authenticator) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("session", c)

		if c.QueryParam("state") != sess.Values["state"] {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid state parameter."})
		}

		verifier, ok := sess.Values["code_verifier"].(string)
		if !ok {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "code verifier not found in session."})
		}

		token, err := auth.ExchangeWithPKCE(c.Request().Context(), c.QueryParam("code"), verifier)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "failed to exchange an authorization code for a token."})
		}

		idToken, err := auth.VerifyIDToken(c.Request().Context(), token)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to verify ID Token."})
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		sess.Values["access_token"] = token.AccessToken
		sess.Values["refresh_token"] = token.RefreshToken
		sess.Values["profile"] = profile
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		fmt.Printf("session saved: %+v\n", sess.Values["refresh_token"])

		return c.Redirect(http.StatusTemporaryRedirect, "/api/users/me")
	}
}
