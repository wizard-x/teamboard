package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	domainErrors "teamboard/internal/domain/errors"
	"teamboard/internal/dto/response"
)

func newErrorResponse(code, message string) response.ErrorResponse {
	return response.ErrorResponse{
		Error: response.ErrorBody{
			Code:    code,
			Message: message,
		},
	}
}

func handleError(c echo.Context, err error) error {
	var validationErrs domainErrors.ValidationErrors
	if errors.As(err, &validationErrs) {
		details := make([]response.FieldError, 0, len(validationErrs))
		for _, ve := range validationErrs {
			details = append(details, response.FieldError{
				Field:   ve.Field,
				Message: ve.Message,
			})
		}
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: response.ErrorBody{
				Code:    "VALIDATION_ERROR",
				Message: "Validation failed",
				Details: details,
			},
		})
	}

	switch {
	case errors.Is(err, domainErrors.ErrNotFound):
		return c.JSON(http.StatusNotFound, newErrorResponse("NOT_FOUND", "Resource not found"))
	case errors.Is(err, domainErrors.ErrUnauthorized):
		return c.JSON(http.StatusUnauthorized, newErrorResponse("UNAUTHORIZED", "Invalid or missing API key"))
	case errors.Is(err, domainErrors.ErrForbidden):
		return c.JSON(http.StatusForbidden, newErrorResponse("FORBIDDEN", "Insufficient permissions"))
	case errors.Is(err, domainErrors.ErrConflict):
		return c.JSON(http.StatusConflict, newErrorResponse("CONFLICT", "Resource already exists"))
	case errors.Is(err, domainErrors.ErrBadRequest):
		return c.JSON(http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "Bad request"))
	default:
		return c.JSON(http.StatusInternalServerError, newErrorResponse("INTERNAL_ERROR", "An unexpected error occurred"))
	}
}
