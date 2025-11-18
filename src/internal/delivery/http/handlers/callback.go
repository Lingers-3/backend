package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"pocketeer/internal/platform/authenticator"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func CallbackHandler(auth *authenticator.Authenticator) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("session", c)

		if c.QueryParam("state") != sess.Values["state"] {
			return redirectWithError(c, http.StatusBadRequest, "invalid state parameter")
		}

		verifier, ok := sess.Values["code_verifier"].(string)
		if !ok {
			return redirectWithError(c, http.StatusBadRequest, "invalid code verifier")
		}

		if errParam := c.QueryParam("error"); errParam != "" {
			errDesc := c.QueryParam("error_description")
			log.Printf("Auth0 callback error: %s - %s", errParam, errDesc)

			var parsed struct {
				Error   string `json:"error"`
				Message string `json:"message"`
			}
			if jErr := json.Unmarshal([]byte(errDesc), &parsed); jErr == nil {
				return redirectWithError(c, http.StatusForbidden, parsed.Message)
			}

			return redirectWithError(c, http.StatusUnauthorized, errDesc)
		}

		code := c.QueryParam("code")
		if code == "" {
			return redirectWithError(c, http.StatusBadRequest, "Missing authorization code")
		}

		token, err := auth.ExchangeWithPKCE(c.Request().Context(), code, verifier)
		if err != nil {
			log.Printf("PKCE exchange failed: %v", err)
			return redirectWithError(c, http.StatusUnauthorized, "PKCE verification failed")
		}

		idToken, err := auth.VerifyIDToken(c.Request().Context(), token)
		if err != nil {
			return redirectWithError(c, http.StatusInternalServerError, "internal server error")
		}

		var profile map[string]any
		if err := idToken.Claims(&profile); err != nil {
			return redirectWithError(c, http.StatusInternalServerError, "internal server error")
		}

		emailVerified, _ := profile["email_verified"].(bool)

		if !emailVerified {
			return redirectWithError(c, http.StatusForbidden, "email not verified")
		}

		sess.Values["access_token"] = token.AccessToken
		sess.Values["refresh_token"] = token.RefreshToken
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			log.Printf("failed to save session: %v", err)
			return redirectWithError(c, http.StatusInternalServerError, "internal server error")
		}

		redirectUri, ok := sess.Values["redirect_uri"].(string)
		if !ok || redirectUri == "" {
			redirectUri = "https://pocketeer.linerds.us/"
		}

		return c.Redirect(http.StatusTemporaryRedirect, redirectUri)
	}
}

func redirectWithError(c echo.Context, code int, message string) error {
	redirectURL := fmt.Sprintf("https://pocketeer.linerds.us/error?code=%d&message=%s",
		code, url.QueryEscape(message))
	return c.Redirect(http.StatusFound, redirectURL)
}
