package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dillonthompson/dillonthompson.com/internal/handlers"
	"github.com/joho/godotenv"
	"github.com/dillonthompson/dillonthompson.com/internal/middleware"
	"github.com/dillonthompson/dillonthompson.com/internal/repository"
	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	repo, err := repository.Connect(dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer repo.DB.Close()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.StructuredLogger(logger))
	// Security headers are owned by Caddy at the edge (see deploy/Caddyfile.prod).
	// Setting them here too would result in duplicate Content-Security-Policy
	// headers that browsers intersect, leading to subtle policy drift whenever
	// either side is edited. Single source of truth lives in Caddy.

	healthHandler := handlers.NewHealthHandler(repo.DB)
	experienceHandler := handlers.NewExperienceHandler(repo.Queries)
	profileHandler := handlers.NewProfileHandler(repo.Queries)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Check)
		v1.GET("/experience", experienceHandler.List)
		v1.GET("/profile", profileHandler.GetFull)
		v1.GET("/profile/:key", profileHandler.GetSection)
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
