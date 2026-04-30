package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"teamboard/internal/dto/response"
	"teamboard/internal/service"
)

func APIKeyAuth(authService service.Authenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apiKey := c.Request().Header.Get("X-API-Key")
			if apiKey == "" {
				return c.JSON(http.StatusUnauthorized, response.ErrorResponse{
					Error: response.ErrorBody{
						Code:    "UNAUTHORIZED",
						Message: "Missing API key",
					},
				})
			}

			authCtx, err := authService.Authenticate(c.Request().Context(), apiKey)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, response.ErrorResponse{
					Error: response.ErrorBody{
						Code:    "UNAUTHORIZED",
						Message: "Invalid API key",
					},
				})
			}

			c.Set("auth_context", authCtx)
			return next(c)
		}
	}
}

func RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authCtx := c.Get("auth_context")
			if authCtx == nil {
				return c.JSON(http.StatusUnauthorized, response.ErrorResponse{
					Error: response.ErrorBody{
						Code:    "UNAUTHORIZED",
						Message: "Authentication required",
					},
				})
			}
			ac, ok := authCtx.(*service.AuthContext)
			if !ok || ac.Role != "admin" {
				return c.JSON(http.StatusForbidden, response.ErrorResponse{
					Error: response.ErrorBody{
						Code:    "FORBIDDEN",
						Message: "Admin access required",
					},
				})
			}
			return next(c)
		}
	}
}
