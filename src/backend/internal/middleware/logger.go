package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			req := c.Request()
			res := c.Response()

			log.Infof("%s %s %d %s %s",
				req.Method,
				req.URL.Path,
				res.Status,
				time.Since(start).Round(time.Microsecond),
				c.RealIP(),
			)
			return err
		}
	}
}
