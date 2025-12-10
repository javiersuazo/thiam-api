// Package http implements routing paths. Each services in own file.
package http

import (
	"net/http"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/evrone/go-clean-template/config"
	_ "github.com/evrone/go-clean-template/docs" // Swagger docs.
	"github.com/evrone/go-clean-template/internal/controller/http/middleware"
	v1 "github.com/evrone/go-clean-template/internal/controller/http/v1"
	authuc "github.com/evrone/go-clean-template/internal/usecase/auth"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// AuthUseCases groups all authentication-related use cases.
type AuthUseCases struct {
	Register *authuc.RegisterUseCase
}

// RouterDeps holds dependencies for the HTTP router.
type RouterDeps struct {
	Config   *config.Config
	Postgres *postgres.Postgres
	Logger   logger.Interface
	Auth     *AuthUseCases
}

// NewRouter -.
// Swagger spec:
// @title       Go Clean Template API
// @description Production-grade Go backend API
// @version     1.0
// @host        localhost:8080
// @BasePath    /v1
func NewRouter(app *fiber.App, deps *RouterDeps) {
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger(deps.Logger))
	app.Use(middleware.Recovery(deps.Logger))

	// Prometheus metrics
	if deps.Config.Metrics.Enabled {
		prometheus := fiberprometheus.New("my-service-name")
		prometheus.RegisterAt(app, "/metrics")
		app.Use(prometheus.Middleware)
	}

	// Swagger
	if deps.Config.Swagger.Enabled {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// K8s probes
	app.Get("/healthz", func(ctx *fiber.Ctx) error { return ctx.SendStatus(http.StatusOK) })
	v1.NewHealthRoutes(app, deps.Postgres.Pool)

	// Routers
	v1Group := app.Group("/v1")
	v1.NewAuthRoutes(v1Group, deps.Auth.Register)
}
