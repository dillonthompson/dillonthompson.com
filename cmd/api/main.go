package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dillonthompson/dillonthompson.com/internal/handlers"
	"github.com/dillonthompson/dillonthompson.com/internal/middleware"
	"github.com/dillonthompson/dillonthompson.com/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

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

	siteURL := os.Getenv("SITE_URL")
	if siteURL == "" {
		siteURL = "https://dillonthompson.com"
	}
	blogHandler, err := handlers.NewBlogHandler(repo.Queries, siteURL)
	if err != nil {
		slog.Error("failed to initialize blog handler", "error", err)
		os.Exit(1)
	}

	// Crawlers and link checkers probe with HEAD; Gin doesn't derive it from GET.
	readMethods := []string{http.MethodGet, http.MethodHead}

	// Public, server-rendered blog. Caddy routes these paths to the API; see
	// deploy/Caddyfile.prod. Asset routes are registered explicitly so they
	// can't be shadowed by the :slug param.
	router.Match(readMethods, "/blog", blogHandler.Index)
	for _, name := range []string{"blog.css", "syntax.css", "theme.js", "terminal.js", "geist-latin.woff2"} {
		router.Match(readMethods, "/blog/assets/"+name, blogHandler.Asset(name))
	}
	router.Match(readMethods, "/blog/:slug", blogHandler.Post)
	router.Match(readMethods, "/rss.xml", blogHandler.RSS)
	router.Match(readMethods, "/sitemap.xml", blogHandler.Sitemap)

	// Admin API. Authentication is Cloudflare Access (JWT verified here, not just
	// trusted from the edge) and fails closed when unconfigured; see
	// internal/middleware/access.go. Env: CF_ACCESS_TEAM_DOMAIN, CF_ACCESS_AUD,
	// ADMIN_EMAILS (comma-separated), and ADMIN_DEV_BYPASS=true for local dev only.
	accessCfg := middleware.AccessConfig{
		TeamDomain:    os.Getenv("CF_ACCESS_TEAM_DOMAIN"),
		Audience:      os.Getenv("CF_ACCESS_AUD"),
		AllowedEmails: splitCSV(os.Getenv("ADMIN_EMAILS")),
		DevBypass:     os.Getenv("ADMIN_DEV_BYPASS") == "true",
	}
	accessAuth, err := middleware.NewAccessAuth(context.Background(), accessCfg)
	if err != nil {
		// A bad admin config must only take the admin offline. Exiting here
		// would crash-loop the whole API (and the public blog) and trip the
		// deploy rollback.
		slog.Error("admin auth misconfigured; admin endpoints disabled", "error", err)
		accessAuth = func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "admin is not configured"})
		}
	}
	sameOrigin := middleware.RequireSameOrigin([]string{siteURL}, accessCfg.BypassActive())
	adminPostsHandler := handlers.NewAdminPostsHandler(repo.Queries)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Check)
		v1.GET("/experience", experienceHandler.List)
		v1.GET("/profile", profileHandler.GetFull)
		v1.GET("/profile/:key", profileHandler.GetSection)

		admin := v1.Group("/admin", accessAuth, sameOrigin)
		admin.GET("/me", adminPostsHandler.Me)
		admin.GET("/posts", adminPostsHandler.List)
		admin.POST("/posts", adminPostsHandler.Create)
		admin.GET("/posts/:id", adminPostsHandler.Get)
		admin.PUT("/posts/:id", adminPostsHandler.Update)
		admin.DELETE("/posts/:id", adminPostsHandler.Delete)
		admin.POST("/preview", adminPostsHandler.Preview)
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
