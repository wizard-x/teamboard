package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"teamboard/internal/repository/postgres"
)

type HealthHandler struct {
	db    *postgres.DB
	cache *redis.Client
}

func NewHealthHandler(db *postgres.DB, cache *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, cache: cache}
}

func (h *HealthHandler) Health(c echo.Context) error {
	status := "ok"
	pgStatus := "ok"
	redisStatus := "ok"

	ctx := c.Request().Context()
	if err := h.db.Pool.Ping(ctx); err != nil {
		pgStatus = "error"
		status = "degraded"
	}
	if err := h.cache.Ping(ctx).Err(); err != nil {
		redisStatus = "error"
		status = "degraded"
	}

	code := http.StatusOK
	if status == "degraded" {
		code = http.StatusServiceUnavailable
	}

	return c.JSON(code, map[string]string{
		"status":   status,
		"postgres": pgStatus,
		"redis":    redisStatus,
		"version":  "1.0.0",
	})
}
