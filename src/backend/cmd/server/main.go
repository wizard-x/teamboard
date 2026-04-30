package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"teamboard/internal/config"
	"teamboard/internal/handler"
	"teamboard/internal/repository/cache"
	"teamboard/internal/repository/postgres"
	"teamboard/internal/router"
	"teamboard/internal/service"
)

//go:embed static
var staticFiles embed.FS

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	// Database
	db, err := postgres.NewDB(ctx, cfg.Database.DSN(), cfg.Database.MaxConns, cfg.Database.MinConns)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()

	// Cache
	appCache := cache.New(rdb, cfg.Auth.CacheTTLMinutes)

	// Repositories
	teamRepo := postgres.NewTeamRepo(db)
	memberRepo := postgres.NewMemberRepo(db)
	boardRepo := postgres.NewBoardRepo(db)
	columnRepo := postgres.NewColumnRepo(db)
	taskRepo := postgres.NewTaskRepo(db)
	commentRepo := postgres.NewCommentRepo(db)

	// Services
	authService := service.NewAuthService(memberRepo, appCache)
	teamService := service.NewTeamService(teamRepo, memberRepo)
	memberService := service.NewMemberService(memberRepo, appCache)
	boardService := service.NewBoardService(boardRepo, columnRepo, taskRepo, memberRepo, commentRepo, appCache)
	taskService := service.NewTaskService(taskRepo, columnRepo, boardRepo, memberRepo, commentRepo, appCache)
	commentService := service.NewCommentService(commentRepo, taskRepo, boardRepo, memberRepo, appCache)

	// Handlers
	healthHandler := handler.NewHealthHandler(db, rdb)
	teamHandler := handler.NewTeamHandler(teamService)
	memberHandler := handler.NewMemberHandler(memberService)
	boardHandler := handler.NewBoardHandler(boardService)
	taskHandler := handler.NewTaskHandler(taskService, commentService)

	// Echo server
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		_ = c.JSON(code, map[string]string{"error": err.Error()})
	}

	router.Setup(e, healthHandler, teamHandler, memberHandler, boardHandler, taskHandler, authService, appCache, cfg.CORS.AllowedOrigins)

	// Serve embedded frontend static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Printf("Warning: could not load embedded static files: %v", err)
	} else {
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// Try static file first, fall through to API routes
				path := c.Request().URL.Path
				if path == "/" {
					path = "/index.html"
				}
				// Don't intercept API or health routes
				if len(path) >= 5 && path[:5] == "/api/" || path == "/health" {
					return next(c)
				}
				// Try to serve static file
				f, err := staticFS.Open(path[1:]) // strip leading /
				if err == nil {
					f.Close()
					http.FileServer(http.FS(staticFS)).ServeHTTP(c.Response(), c.Request())
					return nil
				}
				// SPA fallback: serve index.html for client-side routing
				if len(path) > 1 && !containsDot(path) {
					c.Request().URL.Path = "/"
					http.FileServer(http.FS(staticFS)).ServeHTTP(c.Response(), c.Request())
					return nil
				}
				return next(c)
			}
		})
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)

	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

func containsDot(path string) bool {
	for _, c := range path {
		if c == '.' {
			return true
		}
	}
	return false
}
