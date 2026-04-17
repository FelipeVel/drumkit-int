// @title           Drumkit Integration API
// @version         1.0
// @description     TMS integration service — proxies freight load operations to the Turvo external API.
// @host            localhost:8080
// @BasePath        /v1
// @schemes         http https

package main

import (
	"net/http"
	"time"

	_ "github.com/FelipeVel/drumkit-int/docs"
	"github.com/FelipeVel/drumkit-int/config"
	"github.com/FelipeVel/drumkit-int/internal/controller"
	"github.com/FelipeVel/drumkit-int/internal/middleware"
	"github.com/FelipeVel/drumkit-int/internal/repository"
	"github.com/FelipeVel/drumkit-int/internal/service"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// 5. Register routes under /v1.
	ctrl.RegisterRoutes(router.Group("/v1"))

	// 6. Swagger UI at /swagger/index.html.
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 7. Start the server.
	router.Run(cfg.ServerPort)
}
