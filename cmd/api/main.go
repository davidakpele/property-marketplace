// @title           Property Marketplace API
// @version         1.0
// @description     REST API for property listings with geospatial search.
// @contact.name    API Support

// @host      localhost:8080
// @BasePath  /api/v1

// @tag.name  listings
// @tag.description  CRUD and search operations for property listings

// @tag.name  agents
// @tag.description  CRUD operations for agents

// @tag.name  health
// @tag.description  Service health check

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-yaml"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/davidakpele/property-marketplace/docs"
	"github.com/davidakpele/property-marketplace/internal/agent"
	agenthandler "github.com/davidakpele/property-marketplace/internal/agent/handler"
	agentrepo "github.com/davidakpele/property-marketplace/internal/agent/repository"
	"github.com/davidakpele/property-marketplace/internal/health"
	listing "github.com/davidakpele/property-marketplace/internal/listing"
	listinghandler "github.com/davidakpele/property-marketplace/internal/listing/handler"
	listingrepo "github.com/davidakpele/property-marketplace/internal/listing/repository"
	"github.com/davidakpele/property-marketplace/internal/search"
	"github.com/davidakpele/property-marketplace/pkg/cache"
	"github.com/davidakpele/property-marketplace/pkg/database"
	"github.com/davidakpele/property-marketplace/pkg/logger"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Host       string `yaml:"host"`
		Port       int    `yaml:"port"`
		Name       string `yaml:"name"`
		User       string `yaml:"user"`
		Password   string `yaml:"password"`
		SSLMode    string `yaml:"ssl_mode"`
		LogQueries bool   `yaml:"log_queries"`
	} `yaml:"database"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
	} `yaml:"redis"`
	App struct {
		Environment string `yaml:"environment"`
	} `yaml:"app"`
	MigrationsPath string `yaml:"migrations_path"`
}

func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	cfg := &Config{}
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if cfg.MigrationsPath == "" {
		cfg.MigrationsPath = "migrations"
	}

	return cfg, nil
}

func main() {
	configPath := "configs/application.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
		logger.SetLevel("warn")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	dbCfg := database.Config{
		Host:       cfg.Database.Host,
		Port:       cfg.Database.Port,
		Name:       cfg.Database.Name,
		User:       cfg.Database.User,
		Password:   cfg.Database.Password,
		SSLMode:    cfg.Database.SSLMode,
		LogQueries: cfg.Database.LogQueries,
	}

	pool, err := connectWithRetry(ctx, dbCfg)
	if err != nil {
		logger.Fatalf("connect database: %v", err)
	}
	defer pool.Close()
	logger.Info("database connected")

	if err := database.RunMigrations(dbCfg.DSN(), cfg.MigrationsPath); err != nil {
		logger.Fatalf("run migrations: %v", err)
	}
	logger.Info("migrations applied")

	var cacheClient *cache.Client
	cacheClient, err = cache.NewClient(cache.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
	})
	if err != nil {
		logger.Warnf("redis unavailable, search caching disabled: %v", err)
		cacheClient = nil
	} else {
		defer cacheClient.Close()
		logger.Info("redis connected")
	}

	listingRepo := listingrepo.NewPostgresListingRepository(pool)
	agentRepo := agentrepo.NewPostgresAgentRepository(pool)

	listingSvc := listing.NewService(listingRepo)
	agentSvc := agent.NewService(agentRepo)
	searchSvc := search.NewService(listingRepo, cacheClient)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api/v1")

	health.NewHandler(pool).RegisterRoutes(r)
	listinghandler.NewHandler(listingSvc, searchSvc).RegisterRoutes(api)
	agenthandler.NewHandler(agentSvc).RegisterRoutes(api)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Infof("server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Errorf("server forced shutdown: %v", err)
	}
	logger.Info("server stopped")
}

func connectWithRetry(ctx context.Context, cfg database.Config) (*pgxpool.Pool, error) {
	var (
		pool *pgxpool.Pool
		err  error
	)
	for attempt := 1; attempt <= 10; attempt++ {
		pool, err = database.NewPool(ctx, cfg)
		if err == nil {
			return pool, nil
		}
		logger.Warnf("database not ready (attempt %d/10): %v", attempt, err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt) * time.Second):
		}
	}
	return nil, err
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.WithFields(map[string]interface{}{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"status":   c.Writer.Status(),
			"duration": time.Since(start).String(),
			"ip":       c.ClientIP(),
		}).Info("request")
	}
}
