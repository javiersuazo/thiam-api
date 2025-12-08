package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type healthRoutes struct {
	pool *pgxpool.Pool
}

type DBHealthResponse struct {
	Status           string `json:"status"`
	MigrationVersion *int64 `json:"migration_version,omitempty"`
	Dirty            *bool  `json:"dirty,omitempty"`
	Error            string `json:"error,omitempty"`
}

func NewHealthRoutes(app *fiber.App, pool *pgxpool.Pool) {
	r := &healthRoutes{pool: pool}

	app.Get("/healthz/db", r.dbHealth)
}

func (r *healthRoutes) dbHealth(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := r.pool.Ping(ctx); err != nil {
		return c.Status(http.StatusServiceUnavailable).JSON(DBHealthResponse{
			Status: "unhealthy",
			Error:  "database connection failed",
		})
	}

	var version int64
	var dirty bool

	err := r.pool.QueryRow(ctx, "SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty)
	if err != nil {
		return c.Status(http.StatusOK).JSON(DBHealthResponse{
			Status: "healthy",
			Error:  "no migrations applied yet",
		})
	}

	return c.Status(http.StatusOK).JSON(DBHealthResponse{
		Status:           "healthy",
		MigrationVersion: &version,
		Dirty:            &dirty,
	})
}
