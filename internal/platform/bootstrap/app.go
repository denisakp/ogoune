// Package bootstrap contains the application composition root.
// It organizes initialization by concern without introducing new abstractions.
package bootstrap

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/denisakp/ogoune/internal/config"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/ee/license"
	"github.com/denisakp/ogoune/internal/maintenance"
	"github.com/denisakp/ogoune/internal/monitoring"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/store"
	"github.com/denisakp/ogoune/internal/scheduler"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/denisakp/ogoune/internal/worker"
	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

const AppVersion = "1.0.0"

// App holds all initialized application components.
// Each Init* function populates a subset of fields.
type App struct {
	// Config phase
	Cfg *config.Config

	// Database phase
	DB                      *gorm.DB
	ResourceRepo            repository.ResourceRepository
	IncidentRepo            repository.IncidentRepository
	IncidentEventStepRepo   repository.IncidentEventStepRepository
	IncidentDiagnosticsRepo repository.IncidentDiagnosticsRepository
	NotificationRepo        repository.NotificationRepository
	MaintenanceRepo         *store.MaintenanceRepository
	NotificationChannelRepo *store.NotificationChannelRepository
	MonitoringActivityRepo  repository.MonitoringActivityRepository
	TagsRepo                repository.TagsRepository
	StatusPageSettingsRepo  *store.StatusPageSettingsRepository
	ComponentRepo           repository.ComponentRepository
	UserRepo                *store.UserRepository
	APIKeyRepo              repository.APIKeyRepository

	// Metrics phase
	MetricsRecorder domain.MetricsRecorder
	MetricsRegistry *prometheus.Registry

	// Scheduler phase
	SchedulerCfg          *scheduler.Config
	RuntimeScheduler      scheduler.Scheduler
	SchedulerAdapter      repository.Scheduler
	ConfirmationScheduler interface {
		ScheduleWithInterval(ctx context.Context, resource *domain.Resource, interval time.Duration) error
	}
	AsynqClient          *asynq.Client
	AsynqInspector       *asynq.Inspector
	AsynqScheduler       *asynq.Scheduler
	RedisOpt             asynq.RedisClientOpt
	MaintenanceScheduler *maintenance.SchedulerService

	// Worker phase
	Processor           *worker.Processor
	DetectorIncidentSvc *monitoring.IncidentService

	// Services phase
	ResourceService  *service.ResourceService
	ComponentService *service.ComponentService
	AuthService      *service.AuthService
	JWTManager       *service.JWTManager
	APIKeyService    *service.APIKeyService

	// Router phase
	RootRouter *chi.Mux
	Server     *http.Server
}

func LogStartupEdition() {
	if license.IsEnterprise() {
		slog.Info("Ogoune Enterprise Edition")
	} else {
		slog.Info("Ogoune Community Edition")
	}
}
