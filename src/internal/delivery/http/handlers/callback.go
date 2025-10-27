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
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse ID token."})
		}

		auth0ID, ok := profile["sub"].(string)
		if !ok {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "invalid sub claim"})
		}

		email, ok := profile["email"].(string)
		if !ok {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "invalid email claim"})
		}

		emailVerified, _ := profile["email_verified"].(bool)

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
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
				}
				log.Printf("New user created: %s (%s)", auth0ID, email)

			} else {
				log.Printf("DB error: %v", result.Error)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
			}
		} else {
			if !emailVerified {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "email not verified"})
			}
		}

		sess.Values["user_id"] = user.ID
		sess.Values["auth0_id"] = user.Auth0ID
		sess.Values["email"] = user.Email
		sess.Values["access_token"] = token.AccessToken
		sess.Values["refresh_token"] = token.RefreshToken
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			log.Printf("failed to save session: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save session"})
		}

		redirectUri, ok := sess.Values["redirect_uri"].(string)
		if !ok || redirectUri == "" {
			redirectUri = "https://pocketeer.linerds.us/"
		}

		return c.Redirect(http.StatusTemporaryRedirect, redirectUri)
	}
}
