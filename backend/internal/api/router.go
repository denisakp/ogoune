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
	integrationHandler *handler.IntegrationHandler,
	statusPageHandler *handler.StatusPageHandler,
) http.Handler {
	r := chi.NewRouter()

	// Standard middleware stack for logging, recovery, and request tracking
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

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

	// Resources (Monitors) API
	r.Route("/resources", func(r chi.Router) {
		r.Get("/", resourceHandler.ListResources)                                     // GET /resources - list all resources
		r.Post("/", resourceHandler.CreateResource)                                   // POST /resources - create new resource
		r.Patch("/{id}", resourceHandler.UpdateResource)                              // PATCH /resources/{id} - update resource
		r.Delete("/{id}", resourceHandler.DeleteResource)                             // DELETE /resources/{id} - delete resource
		r.Post("/{id}/pause", resourceHandler.PauseResourceMonitoring)                // POST /resources/{id}/pause - pause monitoring
		r.Post("/{id}/resume", resourceHandler.ResumeResourceMonitoring)              // POST /resources/{id}/resume - resume monitoring
		r.Post("/{resourceID}/tags", resourceHandler.AddTagsToResource)               // POST /resources/{resourceID}/tags - add tags
		r.Delete("/{resourceID}/tags/{tagID}", resourceHandler.RemoveTagFromResource) // DELETE /resources/{resourceID}/tags/{tagID} - remove tag
	})

	// Tags API
	r.Route("/tags", func(r chi.Router) {
		r.Get("/", tagHandler.ListTags)         // GET /tags - list all tags
		r.Post("/", tagHandler.CreateTag)       // POST /tags - create new tag
		r.Patch("/{id}", tagHandler.UpdateTag)  // PATCH /tags/{id} - update tag
		r.Delete("/{id}", tagHandler.DeleteTag) // DELETE /tags/{id} - delete tag
	})

	// Integrations API
	r.Route("/integrations", func(r chi.Router) {
		r.Get("/", integrationHandler.ListIntegrations)        // GET /integrations - list all integrations
		r.Post("/", integrationHandler.CreateIntegration)      // POST /integrations - create new integration
		r.Patch("/{id}", integrationHandler.UpdateIntegration) // PATCH /integrations/{id} - update integration
	})

	// Monitoring Activities API
	r.Get("/monitoring-activities", activityHandler.ListActivities) // GET /monitoring-activities - list activities (supports ?resource_id=xxx)

	return r
}
