package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"teamboard/internal/dto/response"
	"teamboard/internal/repository/cache"
	"teamboard/internal/service"
)

type RateLimitConfig struct {
	Limit  int
	Window time.Duration
}

func RateLimit(appCache *cache.Cache, limit int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var key string

			// Try member-based rate limiting first
			authCtx := c.Get("auth_context")
			if authCtx != nil {
				if ac, ok := authCtx.(*service.AuthContext); ok {
					key = fmt.Sprintf("ratelimit:%s:%d", ac.MemberID, int(window.Seconds()))
				}
			}

			// Fallback to IP-based
			if key == "" {
				key = fmt.Sprintf("ratelimit:%s:%d", c.RealIP(), int(window.Seconds()))
			}

			count, err := appCache.IncrementRateLimit(c.Request().Context(), key, window)
			if err != nil {
				// On Redis error, allow request through
				return next(c)
			}

			// Set rate limit headers
			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, limit-count)))

			ttl, _ := appCache.GetRateLimitTTL(c.Request().Context(), key)
			c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix()))

			if count > limit {
				c.Response().Header().Set("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
				return c.JSON(http.StatusTooManyRequests, response.ErrorResponse{
					Error: response.ErrorBody{
						Code:    "RATE_LIMITED",
						Message: "Rate limit exceeded",
					},
				})
			}

			return next(c)
		}
	}
}

func RegistrationRateLimit(appCache *cache.Cache) echo.MiddlewareFunc {
	return RateLimit(appCache, 5, time.Hour)
}

func ReadRateLimit(appCache *cache.Cache) echo.MiddlewareFunc {
	return RateLimit(appCache, 100, time.Minute)
}

func WriteRateLimit(appCache *cache.Cache) echo.MiddlewareFunc {
	return RateLimit(appCache, 30, time.Minute)
}
