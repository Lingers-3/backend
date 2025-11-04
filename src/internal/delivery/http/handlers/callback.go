package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
			return redirectWithError(c, 400, "invalid state parameter")
		}

		verifier, ok := sess.Values["code_verifier"].(string)
		if !ok {
			return redirectWithError(c, 400, "invalid code verifier")
		}

		token, err := auth.ExchangeWithPKCE(c.Request().Context(), c.QueryParam("code"), verifier)
		if err != nil {
			return redirectWithError(c, 401, "PKCE verification failed")
		}

		idToken, err := auth.VerifyIDToken(c.Request().Context(), token)
		if err != nil {
			return redirectWithError(c, 500, "internal server error")
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			return redirectWithError(c, 500, "internal server error")
		}

		auth0ID, ok := profile["sub"].(string)
		if !ok {
			return redirectWithError(c, 500, "internal server error")
		}

		email, ok := profile["email"].(string)
		if !ok {
			return redirectWithError(c, 500, "internal server error")
		}

		emailVerified, _ := profile["email_verified"].(bool)

		if !emailVerified {
			return redirectWithError(c, 403, "email not verified")
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
					return redirectWithError(c, 500, "internal server error")
				}
				log.Printf("New user created: %s (%s)", auth0ID, email)

			} else {
				log.Printf("DB error: %v", result.Error)
				return redirectWithError(c, 500, "internal server error")
			}
		}

		sess.Values["access_token"] = token.AccessToken

		sess.Values["refresh_token"] = token.RefreshToken
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			log.Printf("failed to save session: %v", err)
			return redirectWithError(c, 500, "internal server error")
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
