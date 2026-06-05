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
	if app.SessionRepo != nil {
		app.SessionService = service.NewSessionService(app.SessionRepo)
		app.AuthService.SetSessionService(app.SessionService)
	}
	if app.TwoFactorResetTokenRepo != nil {
		// Magic-link sender: nil → dev-logger fallback per T051. Wire SMTP
		// when configured by reusing pkg/notifier/smtp.go in a follow-up.
		app.TwoFactorService = service.NewTwoFactorService(
			app.AuthService,
			app.UserRepo,
			app.TwoFactorResetTokenRepo,
			nil,
			cfg.AppBaseURL,
		)
	}

	if app.EscalationRepo != nil {
		app.EscalationService = service.NewEscalationService(app.EscalationRepo, app.ResourceRepo)
	}
	_, _ = app.AuthService.CreateDefaultUser(context.Background(), cfg.AuthEmail, cfg.AuthPassword)
}
