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

// NOTE(noatu): all errors from service layer should be mapped explicitly.
func ServiceErrToHttp(err error) *echo.HTTPError {
	switch { // Remember to sort that!
	case errors.Is(err, services.ErrDatabaseError):
		return echo.NewHTTPError(http.StatusInternalServerError) // NOTE: no err
	case errors.Is(err, services.ErrFileSystemError):
		return echo.NewHTTPError(http.StatusInternalServerError) // NOTE: no err
	case errors.Is(err, services.ErrForeignKeyViolated):
		return echo.NewHTTPError(http.StatusConflict, err)
	case errors.Is(err, services.ErrImageTooLarge):
		return echo.NewHTTPError(http.StatusRequestEntityTooLarge, err)
	case errors.Is(err, services.ErrInsufficientResources):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err)
	case errors.Is(err, services.ErrInvalidImageFormat):
		return echo.NewHTTPError(http.StatusBadRequest, err)
	case errors.Is(err, services.ErrInvalidQuantity):
		return echo.NewHTTPError(http.StatusBadRequest, err)
	case errors.Is(err, services.ErrItemMismatch):
		return echo.NewHTTPError(http.StatusBadRequest, err)
	case errors.Is(err, services.ErrItemNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err)
	case errors.Is(err, services.ErrItemTypeNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err)
	case errors.Is(err, services.ErrNotImplemented):
		return echo.NewHTTPError(http.StatusNotImplemented, err)
	case errors.Is(err, services.ErrPictureNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err)
	case errors.Is(err, services.ErrProjectAlreadyActive):
		return echo.NewHTTPError(http.StatusConflict, err)
	case errors.Is(err, services.ErrProjectAlreadyExists):
		return echo.NewHTTPError(http.StatusConflict, err)
	case errors.Is(err, services.ErrProjectNotActive):
		return echo.NewHTTPError(http.StatusConflict, err)
	case errors.Is(err, services.ErrProjectNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err)
	case errors.Is(err, services.ErrProjectNotPlanning):
		return echo.NewHTTPError(http.StatusConflict, err)
	case errors.Is(err, services.ErrResourceReservationNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err)
	case errors.Is(err, services.ErrResourceSpecificationNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err)
	case errors.Is(err, services.ErrTagAlreadyExists):
		return echo.NewHTTPError(http.StatusConflict, err)
	case errors.Is(err, services.ErrTagNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err)
	case errors.Is(err, services.ErrUnauthenticated):
		return echo.NewHTTPError(http.StatusUnauthorized, err)
	case errors.Is(err, services.ErrUserNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err)

	default: // Just in case
		log.Printf("ERROR: ServiceErrToHttp unknown error: %v", err)
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

func ParseAndValidatePayload(c echo.Context, payload any) error {
	if err := c.Bind(payload); err != nil {
		log.Printf("ERROR: parsing request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	if err := c.Validate(payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
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

func GetIDParamWithName(c echo.Context, param string) (uint, error) {
	id, err := strconv.ParseUint(c.Param(param), 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest)
	}
	return uint(id), nil
}
