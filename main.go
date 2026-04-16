package main

import (
	"net/http"
	"time"

	"github.com/FelipeVel/drumkit-int/config"
	"github.com/FelipeVel/drumkit-int/internal/controller"
	"github.com/FelipeVel/drumkit-int/internal/middleware"
	"github.com/FelipeVel/drumkit-int/internal/repository"
	"github.com/FelipeVel/drumkit-int/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration from .env / environment variables.
	cfg := config.Load()

	// 2. Shared HTTP client with configurable timeout.
	httpClient := &http.Client{
		Timeout: time.Duration(cfg.HTTPTimeoutSec) * time.Second,
	}

	// 3. Wire layers bottom-up (no DI framework — compiler-verified).
	repo := repository.NewTurvoLoadRepository(cfg, httpClient)
	svc := service.NewLoadService(repo)
	ctrl := controller.NewLoadController(svc)

	// 4. Set up Gin with our own structured logger (no default Gin logger).
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())

	// 5. Register routes under /loads.
	ctrl.RegisterRoutes(router.Group("/loads"))

	// 6. Start the server.
	router.Run(cfg.ServerPort)
}
