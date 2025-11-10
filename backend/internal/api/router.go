package api

import (
	"net/http"

	"github.com/denisakp/pulseguard/internal/api/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates and configures the main HTTP router with all JSON API routes.
// All endpoints return JSON responses - no HTML rendering.
func NewRouter(
	resourceHandler *handler.ResourceHandler,
	activityHandler *handler.MonitoringActivityHandler,
	tagHandler *handler.TagHandler,
	statusPageHandler *handler.StatusPageHandler,
	incidentHandler *handler.IncidentHandler,
	notificationHandler *handler.NotificationHandler,
	statsHandler *handler.StatsHandler,
) http.Handler {
	r := chi.NewRouter()

	// Standard middleware stack for logging, recovery, and request tracking
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	// CORS middleware
	r.Use(corsMiddleware)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// ========================================
	// JSON REST API Routes
	// All endpoints return JSON only
	// ========================================

	// Public status page (returns JSON status data)
	r.Get("/status", statusPageHandler.HandleStatusPage)
	r.Get("/status/{resourceId}", statusPageHandler.HandleResourceDetailStatus) // GET /status/{resourceId} - get detailed status for a specific resource

	// Resources (Monitors) API
	r.Route("/resources", func(r chi.Router) {
		r.Get("/", resourceHandler.ListResources)                                     // GET /resources - list all resources
		r.Post("/", resourceHandler.CreateResource)                                   // POST /resources - create new resource
		r.Get("/{id}", resourceHandler.GetResourceByID)                               // GET /resources/{id} - get resource details
		r.Patch("/{id}", resourceHandler.UpdateResource)                              // PATCH /resources/{id} - update resource
		r.Delete("/{id}", resourceHandler.DeleteResource)                             // DELETE /resources/{id} - delete resource
		r.Post("/{id}/pause", resourceHandler.PauseResourceMonitoring)                // POST /resources/{id}/pause - pause monitoring
		r.Post("/{id}/resume", resourceHandler.ResumeResourceMonitoring)              // POST /resources/{id}/resume - resume monitoring
		r.Post("/{resourceID}/tags", resourceHandler.AddTagsToResource)               // POST /resources/{resourceID}/tags - add tags
		r.Delete("/{resourceID}/tags/{tagID}", resourceHandler.RemoveTagFromResource) // DELETE /resources/{resourceID}/tags/{tagID} - remove tag
		r.Get("/{resourceId}/uptime-stats", activityHandler.GetUptimeStats)           // GET /resources/{resourceId}/uptime-stats - get hourly uptime stats
	})

	// Tags API
	r.Route("/tags", func(r chi.Router) {
		r.Get("/", tagHandler.ListTags)         // GET /tags - list all tags
		r.Post("/", tagHandler.CreateTag)       // POST /tags - create new tag
		r.Patch("/{id}", tagHandler.UpdateTag)  // PATCH /tags/{id} - update tag
		r.Delete("/{id}", tagHandler.DeleteTag) // DELETE /tags/{id} - delete tag
	})

	// Monitoring Activities API
	r.Get("/monitoring-activities", activityHandler.ListActivities) // GET /monitoring-activities - list activities (supports ?resource_id=xxx)

	// Incidents API
	r.Route("/incidents", func(r chi.Router) {
		r.Get("/", incidentHandler.ListIncidents)                         // GET /incidents - list all incidents (supports ?unresolved=true, ?limit=x, ?offset=y)
		r.Get("/{id}", incidentHandler.GetIncidentDetail)                 // GET /incidents/{id} - get incident details with event steps
		r.Get("/{id}/event-steps", incidentHandler.GetIncidentEventSteps) // GET /incidents/{id}/event-steps - get event steps for incident
	})

	// Notifications API
	r.Route("/notifications", func(r chi.Router) {
		r.Post("/test", notificationHandler.TestNotification) // POST /notifications/test - send test notification for a resource
	})

	// Stats API
	r.Route("/stats", func(r chi.Router) {
		r.Get("/summary", statsHandler.GetSummary) // GET /stats/summary?range=24h - get aggregated statistics for all monitors
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
