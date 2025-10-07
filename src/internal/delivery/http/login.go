package http

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"pocketeer/internal/platform/authenticator"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func LoginHandler(auth *authenticator.Authenticator) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("session", c)

		state, err := generateRandomState()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		verifier, err := generateCodeVerifier()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		sess.Values["state"] = state
		sess.Values["code_verifier"] = verifier
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
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

	state := base64.StdEncoding.EncodeToString(b)

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
