package handlers

import (
	"errors"
	"log"
	"net/http"
	"pocketeer/internal/platform/authenticator"
	"pocketeer/internal/platform/db/models"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func CallbackHandler(auth *authenticator.Authenticator, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("session", c)

		if c.QueryParam("state") != sess.Values["state"] {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid state parameter.")
		}

		verifier, ok := sess.Values["code_verifier"].(string)
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid code verifier.")
		}

		token, err := auth.ExchangeWithPKCE(c.Request().Context(), c.QueryParam("code"), verifier)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "PKCE verification failed.")
		}

		idToken, err := auth.VerifyIDToken(c.Request().Context(), token)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		auth0ID, ok := profile["sub"].(string)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		email, ok := profile["email"].(string)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		emailVerified, _ := profile["email_verified"].(bool)

		if !emailVerified {
			return echo.NewHTTPError(http.StatusForbidden, "email not verified")
		}

		var user models.User
		result := db.First(&user, "auth0_id = ?", auth0ID)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				user = models.User{
					Auth0ID: auth0ID,
					Email:   email,
				}
				if err := db.Create(&user).Error; err != nil {
					log.Printf("failed to create user: %v", err)
					return echo.NewHTTPError(http.StatusInternalServerError)
				}
				log.Printf("New user created: %s (%s)", auth0ID, email)

			} else {
				log.Printf("DB error: %v", result.Error)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
		}

		sess.Values["access_token"] = token.AccessToken

		sess.Values["refresh_token"] = token.RefreshToken
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			log.Printf("failed to save session: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		redirectUri, ok := sess.Values["redirect_uri"].(string)
		if !ok || redirectUri == "" {
			redirectUri = "https://pocketeer.linerds.us/"
		}

		return c.Redirect(http.StatusTemporaryRedirect, redirectUri)
	}
}
