package handlers

import (
	"errors"
	"net/http"
	"pocketeer/internal/app/services"

	"github.com/labstack/echo/v4"
)

// PostLoginHandler handles post-login user creation/restoration
// @Summary      Post-login callback
// @Description  Create new user or restore soft-deleted user after Auth0 authentication
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body  object{sub=string,email=string,email_verified=boolean}  true  "User Auth0 Data"
// @Success      200   "OK"
// @Failure      400   {object}  echo.HTTPError
// @Failure      403   {object}  echo.HTTPError  "Email not verified"
// @Failure      500   {object}  echo.HTTPError
// @Router       /auth/post-login [post]
func PostLoginHandler(userService *services.UserService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var body struct {
			Sub           string `json:"sub"`
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
		}

		if err := c.Bind(&body); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON body")
		}

		if body.Sub == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "missing sub")
		}

		if !body.EmailVerified {
			return echo.NewHTTPError(http.StatusForbidden, "email not verified")
		}

		_, err := userService.EnsureActiveUser(c.Request().Context(), body.Sub, body.Email)

		if err != nil {
			if errors.Is(err, services.ErrDatabaseError) {
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		return c.NoContent(http.StatusOK)
	}
}
