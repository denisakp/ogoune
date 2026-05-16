package bootstrap

import (
	"context"
	"time"

	"github.com/denisakp/ogoune/internal/service"
)

// InitServices creates API-layer services and the default admin user.
func InitServices(app *App) {
	cfg := app.Cfg

	enrichmentService := service.NewEnrichmentService(30 * time.Second)
	app.ResourceService = service.NewResourceService(app.ResourceRepo, app.IncidentRepo, app.TagsRepo, app.SchedulerAdapter, app.MonitoringActivityRepo, enrichmentService, app.ComponentService)

	// Auth service + default user
	jwtManager := service.NewJWTManager(cfg.JWTSecret, "ogoune", 24*time.Hour)
	app.JWTManager = jwtManager
	app.APIKeyService = service.NewAPIKeyService(app.APIKeyRepo, app.UserRepo)
	app.AuthService = service.NewAuthService(app.UserRepo, jwtManager)

	_, _ = app.AuthService.CreateDefaultUser(context.Background(), cfg.AuthEmail, cfg.AuthPassword)
}
