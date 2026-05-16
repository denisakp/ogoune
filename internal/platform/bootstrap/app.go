// Package bootstrap contains the application composition root.
// It organizes initialization by concern without introducing new abstractions.
package bootstrap

import (
	"log/slog"
	"net/http"

	"github.com/denisakp/ogoune/internal/config"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/ee/license"
	"github.com/denisakp/ogoune/internal/port"
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
	ResourceRepo            port.ResourceRepository
	IncidentRepo            port.IncidentRepository
	IncidentEventStepRepo   port.IncidentEventStepRepository
	IncidentDiagnosticsRepo port.IncidentDiagnosticsRepository
	NotificationRepo        port.NotificationRepository
	MaintenanceRepo         port.MaintenanceRepository
	NotificationChannelRepo port.NotificationChannelRepository
	MonitoringActivityRepo  port.MonitoringActivityRepository
	TagsRepo                port.TagsRepository
	StatusPageSettingsRepo  port.StatusPageSettingsRepository
	ComponentRepo           port.ComponentRepository
	UserRepo                port.UserRepository
	APIKeyRepo              port.APIKeyRepository

	// Metrics phase
	MetricsRecorder domain.MetricsRecorder
	MetricsRegistry *prometheus.Registry

	// Scheduler phase
	SchedulerCfg          *scheduler.Config
	RuntimeScheduler      scheduler.Scheduler
	SchedulerAdapter      port.ResourceScheduler
	ConfirmationScheduler port.ConfirmationRescheduler
	AsynqClient          *asynq.Client
	AsynqInspector       *asynq.Inspector
	AsynqScheduler       *asynq.Scheduler
	RedisOpt             asynq.RedisClientOpt
	MaintenanceScheduler port.MaintenanceScheduler

	// Worker phase
	Processor           *worker.Processor
	DetectorIncidentSvc port.MonitoringIncidentProcessor

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
