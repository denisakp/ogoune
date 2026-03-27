package api

import (
	"net/http"

	"github.com/denisakp/ogoune/internal/api/handler"
	"github.com/denisakp/ogoune/internal/api/middleware"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates and configures the main HTTP router with all JSON API routes.
// All endpoints return JSON responses - no HTML rendering.
func NewRouter(
	resourceHandler *handler.ResourceHandler,
	activityHandler *handler.MonitoringActivityHandler,
	tagHandler *handler.TagHandler,
	componentHandler *handler.ComponentHandler,
	statusPageHandler *handler.StatusPageHandler,
	statusPageSettingsHandler *handler.StatusPageSettingsHandler,
	incidentHandler *handler.IncidentHandler,
	notificationHandler *handler.NotificationHandler,
	maintenanceHandler *handler.MaintenanceHandler,
	statsHandler *handler.StatsHandler,
	systemHandler *handler.SystemHandler,
	authHandler *handler.AuthHandler,
	accountHandler *handler.AccountHandler,
	authService *service.AuthService,
	apiKeyService *service.APIKeyService,
) http.Handler {
	r := chi.NewRouter()

	// Standard middleware stack for logging, recovery, and request tracking
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.SetHeader("Content-Type", "application/json"))

	// CORS middleware
	r.Use(corsMiddleware)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// ========================================
	// Public Routes (no authentication required)
	// ========================================

	// Authentication endpoints
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)                            // POST /auth/login - authenticate user
		r.Post("/initialize-password", authHandler.InitializePassword) // POST /auth/initialize-password
		r.Post("/verify-2fa", authHandler.Verify2FA)                   // POST /auth/verify-2fa
		r.Get("/verify", authHandler.Verify)                           // GET /auth/verify - verify JWT token
	})

	// Public status page (returns JSON status data)
	r.Get("/status", statusPageHandler.HandleStatusPage)
	r.Get("/status/{resourceId}", statusPageHandler.HandleResourceDetailStatus) // GET /status/{resourceId} - get detailed status for a specific resource
	r.Get("/system/edition", systemHandler.GetEdition)

	// ========================================
	// Protected Routes (authentication required)
	// All routes below require valid JWT token
	// ========================================
	r.Group(func(r chi.Router) {
		// Apply auth middleware to all routes in this group
		r.Use(middleware.AuthMiddleware(authService, apiKeyService))

		// Account Management API
		r.Route("/account", func(r chi.Router) {
			r.Get("/profile", accountHandler.GetProfile)              // GET /account/profile
			r.Patch("/profile", accountHandler.UpdateProfile)         // PATCH /account/profile
			r.Post("/change-password", accountHandler.ChangePassword) // POST /account/change-password
			r.Post("/reset-password", accountHandler.ResetPassword)   // POST /account/reset-password
			r.Post("/2fa/enable", accountHandler.Enable2FA)           // POST /account/2fa/enable
			r.Post("/2fa/confirm", accountHandler.Confirm2FA)         // POST /account/2fa/confirm
			r.Post("/2fa/disable", accountHandler.Disable2FA)         // POST /account/2fa/disable
			r.With(middleware.RequireJWTOnly).Post("/api-keys", accountHandler.CreateAPIKey)
			r.With(middleware.RequireJWTOnly).Get("/api-keys", accountHandler.ListAPIKeys)
			r.With(middleware.RequireJWTOnly).Delete("/api-keys/{id}", accountHandler.RevokeAPIKey)
		})

		// Resources (Monitors) API
		r.Route("/resources", func(r chi.Router) {
			r.Get("/", resourceHandler.ListResources)                                                                       // GET /resources - list all resources
			r.With(middleware.RequireReadWrite).Post("/", resourceHandler.CreateResource)                                   // POST /resources - create new resource
			r.Get("/{id}", resourceHandler.GetResourceByID)                                                                 // GET /resources/{id} - get resource details
			r.Get("/{id}/live", resourceHandler.GetLive)                                                                    // GET /resources/{id}/live - get live resource snapshot
			r.With(middleware.RequireReadWrite).Patch("/{id}", resourceHandler.UpdateResource)                              // PATCH /resources/{id} - update resource
			r.With(middleware.RequireReadWrite).Delete("/{id}", resourceHandler.DeleteResource)                             // DELETE /resources/{id} - delete resource
			r.With(middleware.RequireReadWrite).Post("/{id}/pause", resourceHandler.PauseResourceMonitoring)                // POST /resources/{id}/pause - pause monitoring
			r.With(middleware.RequireReadWrite).Post("/{id}/resume", resourceHandler.ResumeResourceMonitoring)              // POST /resources/{id}/resume - resume monitoring
			r.With(middleware.RequireReadWrite).Post("/{resourceID}/tags", resourceHandler.AddTagsToResource)               // POST /resources/{resourceID}/tags - add tags
			r.With(middleware.RequireReadWrite).Delete("/{resourceID}/tags/{tagID}", resourceHandler.RemoveTagFromResource) // DELETE /resources/{resourceID}/tags/{tagID} - remove tag
			r.Get("/{resourceId}/uptime-stats", activityHandler.GetUptimeStats)                                             // GET /resources/{resourceId}/uptime-stats - get hourly uptime stats
		})

		// Components API
		r.Route("/components", func(r chi.Router) {
			r.Get("/", componentHandler.ListComponents)
			r.With(middleware.RequireReadWrite).Post("/", componentHandler.CreateComponent)
			r.Get("/{id}", componentHandler.GetComponent)
			r.With(middleware.RequireReadWrite).Patch("/{id}", componentHandler.UpdateComponent)
			r.With(middleware.RequireReadWrite).Delete("/{id}", componentHandler.DeleteComponent)
			r.With(middleware.RequireReadWrite).Post("/{id}/resources/bulk-assign", componentHandler.BulkAssignToComponent) // POST /components/{id}/resources/bulk-assign - assign multiple resources
			r.With(middleware.RequireReadWrite).Post("/resources/bulk-remove", componentHandler.BulkRemoveFromComponent)    // POST /components/resources/bulk-remove - remove resources from components
		})

		// Tags API
		r.Route("/tags", func(r chi.Router) {
			r.Get("/", tagHandler.ListTags)                                           // GET /tags - list all tags
			r.With(middleware.RequireReadWrite).Post("/", tagHandler.CreateTag)       // POST /tags - create new tag
			r.With(middleware.RequireReadWrite).Patch("/{id}", tagHandler.UpdateTag)  // PATCH /tags/{id} - update tag
			r.With(middleware.RequireReadWrite).Delete("/{id}", tagHandler.DeleteTag) // DELETE /tags/{id} - delete tag
		})

		// Monitoring Activities API
		r.Get("/monitoring-activities", activityHandler.ListActivities) // GET /monitoring-activities - list activities (supports ?resource_id=xxx)

		// Incidents API
		r.Route("/incidents", func(r chi.Router) {
			r.Get("/", incidentHandler.ListIncidents)                         // GET /incidents - list all incidents (supports ?unresolved=true, ?limit=x, ?offset=y)
			r.Get("/{id}", incidentHandler.GetIncidentDetail)                 // GET /incidents/{id} - get incident details with event steps
			r.Get("/{id}/event-steps", incidentHandler.GetIncidentEventSteps) // GET /incidents/{id}/event-steps - get event steps for incident
		})

		// Notification Channels API
		r.Route("/notification-channels", func(r chi.Router) {
			r.Get("/", notificationHandler.ListNotificationChannels)                                                   // GET /notification-channels - list all channels
			r.With(middleware.RequireReadWrite).Post("/", notificationHandler.CreateNotificationChannel)               // POST /notification-channels - create new channel
			r.With(middleware.RequireReadWrite).Post("/test-config", notificationHandler.ValidateAndTestChannelConfig) // POST /notification-channels/test-config - test config without saving
			r.Get("/{id}", notificationHandler.GetNotificationChannel)                                                 // GET /notification-channels/{id} - get channel by ID
			r.With(middleware.RequireReadWrite).Patch("/{id}", notificationHandler.UpdateNotificationChannel)          // PATCH /notification-channels/{id} - update channel
			r.With(middleware.RequireReadWrite).Delete("/{id}", notificationHandler.DeleteNotificationChannel)         // DELETE /notification-channels/{id} - delete channel
			r.With(middleware.RequireReadWrite).Post("/{id}/test", notificationHandler.TestNotificationChannelConfig)  // POST /notification-channels/{id}/test - test channel config
		})

		// Maintenances API
		r.Route("/maintenances", func(r chi.Router) {
			r.Get("/", maintenanceHandler.ListMaintenances)                                                // GET /maintenances
			r.With(middleware.RequireReadWrite).Post("/", maintenanceHandler.CreateMaintenance)            // POST /maintenances
			r.With(middleware.RequireReadWrite).Patch("/{id}", maintenanceHandler.UpdateMaintenance)       // PATCH /maintenances/{id}
			r.With(middleware.RequireReadWrite).Delete("/{id}", maintenanceHandler.DeleteMaintenance)      // DELETE /maintenances/{id}
			r.With(middleware.RequireReadWrite).Post("/{id}/finish", maintenanceHandler.FinishMaintenance) // POST /maintenances/{id}/finish
		})

		// Stats API
		r.Route("/stats", func(r chi.Router) {
			r.Get("/summary", statsHandler.GetSummary) // GET /stats/summary?range=24h - get aggregated statistics for all monitors
		})

		// Status Page Settings API
		statusPageSettingsHandler.RegisterRoutes(r)
	})

	return r
}

// corsMiddleware handles Cross-Origin Resource Sharing (CORS) headers
// Allows frontend running on different port to communicate with the API
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from any origin during development
		// In production, set this to specific domain
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
