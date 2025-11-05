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

// NOTE(noatu): spaghetti with services but it is http so should be here
func ServiceErrToHttp(err error) *echo.HTTPError {
	switch {
	case errors.Is(err, services.ErrUnauthenticated):
		return echo.NewHTTPError(http.StatusUnauthorized, err)
	case errors.Is(err, services.ErrItemTypeNotFound) ||
		errors.Is(err, services.ErrItemNotFound) ||
		errors.Is(err, services.ErrTagNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err)
	default:
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
