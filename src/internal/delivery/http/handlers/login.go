package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"

	"pocketeer/internal/platform/authenticator"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// LoginHandler initiates Auth0 login flow
// @Summary      Initiate login
// @Description  Redirect to Auth0 login page with PKCE flow
// @Tags         auth
// @Produce      json
// @Param        redirect_uri  query  string  false  "Post-login redirect URI"
// @Success      307           "Temporary Redirect to Auth0"
// @Failure      500           {object}  echo.HTTPError
// @Router       /auth/login [get]
func LoginHandler(auth *authenticator.Authenticator) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("session", c)

		redirectUri := c.QueryParam("redirect_uri")
		if redirectUri != "" {
			sess.Values["redirect_uri"] = redirectUri
		} else {
			sess.Values["redirect_uri"] = "https://pocketeer.linerds.us/"
		}

		state, err := generateRandomState()
		if err != nil {
			log.Printf("failed to generate state: %v", err)
			return redirectWithError(c, http.StatusInternalServerError, "internal server error")
		}
		verifier, err := generateCodeVerifier()
		if err != nil {
			log.Printf("failed to generate code verifier: %v", err)
			return redirectWithError(c, http.StatusInternalServerError, "internal server error")
		}

		sess.Values["state"] = state
		sess.Values["code_verifier"] = verifier
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			log.Printf("failed to save session: %v", err)
			return redirectWithError(c, http.StatusInternalServerError, "internal server error")
		}

		return c.Redirect(http.StatusTemporaryRedirect, auth.AuthCodeURLWithPKCE(state, verifier))
	}
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.RawURLEncoding.EncodeToString(b)

	return state, nil
}

func generateCodeVerifier() (string, error) {
	b := make([]byte, 64)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
