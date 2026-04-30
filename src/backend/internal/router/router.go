package router

import (
	"github.com/labstack/echo/v4"

	"teamboard/internal/handler"
	"teamboard/internal/middleware"
	"teamboard/internal/repository/cache"
	"teamboard/internal/service"
)

func Setup(
	e *echo.Echo,
	healthHandler *handler.HealthHandler,
	teamHandler *handler.TeamHandler,
	memberHandler *handler.MemberHandler,
	boardHandler *handler.BoardHandler,
	taskHandler *handler.TaskHandler,
	authService service.Authenticator,
	appCache *cache.Cache,
	corsOrigins string,
) {
	// Global middleware
	e.Use(middleware.SecurityHeaders())
	e.Use(middleware.CORS(corsOrigins))
	e.Use(middleware.Logger())

	// Health check (no auth)
	e.GET("/health", healthHandler.Health)

	// Team registration (no auth, rate limited)
	e.POST("/api/v1/teams/register", teamHandler.Register, middleware.RegistrationRateLimit(appCache))

	// Authenticated API group
	api := e.Group("/api/v1")
	api.Use(middleware.APIKeyAuth(authService))

	// Boards
	api.POST("/boards", boardHandler.Create)
	api.GET("/boards", boardHandler.List)
	api.GET("/boards/:id", boardHandler.Get)
	api.PUT("/boards/:id", boardHandler.Update)
	api.DELETE("/boards/:id", boardHandler.Delete) // Admin check in handler

	// Columns
	api.POST("/boards/:boardId/columns", boardHandler.AddColumn)
	api.PUT("/columns/:id", boardHandler.RenameColumn)
	api.PUT("/columns/:id/position", boardHandler.ReorderColumn)
	api.DELETE("/columns/:id", boardHandler.DeleteColumn)

	// Tasks
	api.POST("/columns/:columnId/tasks", taskHandler.Create)
	api.GET("/tasks/:id", taskHandler.Get)
	api.PUT("/tasks/:id", taskHandler.Update)
	api.PUT("/tasks/:id/move", taskHandler.Move)
	api.DELETE("/tasks/:id", taskHandler.Delete)

	// Comments
	api.POST("/tasks/:taskId/comments", taskHandler.CreateComment)
	api.GET("/tasks/:taskId/comments", taskHandler.ListComments)
	api.DELETE("/comments/:id", taskHandler.DeleteComment)

	// Members
	api.GET("/members/me", memberHandler.Me)
	api.PUT("/members/me", memberHandler.UpdateMe)
	api.POST("/members/me/regenerate-key", memberHandler.RegenerateMyKey)
	api.GET("/members", memberHandler.List)
	api.POST("/members", memberHandler.Create)              // Admin check in handler
	api.PUT("/members/:id", memberHandler.Update)           // Admin check in handler
	api.DELETE("/members/:id", memberHandler.Delete)        // Admin check in handler
	api.POST("/members/:id/regenerate-key", memberHandler.RegenerateKey) // Admin check in handler
}
