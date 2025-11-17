package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"pocketeer/internal/app/services"
	"pocketeer/internal/delivery/http/middleware"

	"github.com/labstack/echo/v4"
)

// NOTE(noatu): spaghetti with services but it is http so should be here.
// WARN(noatu): Keep those in the same order as the errors, OR ELSE •̀ᴖ•́
func ServiceErrToHttp(err error) *echo.HTTPError {
	switch {
	case errors.Is(err, services.ErrUnauthenticated):
		return echo.NewHTTPError(http.StatusUnauthorized, err)

	case errors.Is(err, services.ErrForeignKeyViolated) ||
		errors.Is(err, services.ErrTagAlreadyExists):
		return echo.NewHTTPError(http.StatusConflict, err)

	case errors.Is(err, services.ErrItemTypeNotFound) ||
		errors.Is(err, services.ErrItemNotFound) ||
		errors.Is(err, services.ErrTagNotFound) ||
		errors.Is(err, services.ErrPictureNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err)

	case errors.Is(err, services.ErrInvalidImageFormat):
		return echo.NewHTTPError(http.StatusBadRequest, err)

	case errors.Is(err, services.ErrImageTooLarge):
		return echo.NewHTTPError(http.StatusRequestEntityTooLarge, err)

	case errors.Is(err, services.ErrNotImplemented):
		return echo.NewHTTPError(http.StatusInternalServerError, err)

	default: // NOTE: internal errors, the error message is not passed
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
}

func GetAuth0ID(c echo.Context) (string, error) {
	auth0ID, ok := middleware.GetAuth0IDFromRequest(c)
	if !ok || auth0ID == "" {
		return "", ServiceErrToHttp(services.ErrUnauthenticated)
	}
	return auth0ID, nil
}

func ParsePayload(c echo.Context, payload any) error {
	if err := c.Bind(payload); err != nil {
		log.Printf("ERROR: parsing request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	return nil
}

func GetIDParam(c echo.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest)
	}
	return uint(id), nil
}
