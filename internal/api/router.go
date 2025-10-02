package api

import (
	"net/http"

	"github.com/denisakp/pulseguard/internal/api/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates and configures the main HTTP router with all application routes.
// Dependencies should be injected here and passed to handlers to maintain clean architecture.
func NewRouter(
	resourceHandler *handler.ResourceHandler,
	activityHandler *handler.MonitoringActivityHandler,
	tagHandler *handler.TagHandler,
	integrationHandler *handler.IntegrationHandler,
) http.Handler {
	r := chi.NewRouter()

	// Standard middleware stack for logging, recovery, and request tracking
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check endpoint (no authentication required)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API routes for resource management
	r.Route("/resources", func(r chi.Router) {
		r.Post("/", resourceHandler.CreateResource)                                   // Create new monitoring resource
		r.Get("/", resourceHandler.ListResources)                                     // List all resources
		r.Patch("/{id}", resourceHandler.UpdateResource)                              // Update resource
		r.Delete("/{id}", resourceHandler.DeleteResource)                             // Delete resource
		r.Post("/{id}/pause", resourceHandler.PauseResourceMonitoring)                // Pause monitoring
		r.Post("/{id}/resume", resourceHandler.ResumeResourceMonitoring)              // Resume monitoring
		r.Post("/{resourceID}/tags", resourceHandler.AddTagsToResource)               // Add tags to resource
		r.Delete("/{resourceID}/tags/{tagID}", resourceHandler.RemoveTagFromResource) // Remove tag from resource
	})

	// API routes for tag management
	r.Route("/tags", func(r chi.Router) {
		r.Post("/", tagHandler.CreateTag)       // Create new tag
		r.Get("/", tagHandler.ListTags)         // List all tags
		r.Patch("/{id}", tagHandler.UpdateTag)  // Update tag
		r.Delete("/{id}", tagHandler.DeleteTag) // Delete tag
	})

	// API routes for integration management
	r.Route("/integrations", func(r chi.Router) {
		r.Post("/", integrationHandler.CreateIntegration)      // Create new integration
		r.Get("/", integrationHandler.ListIntegrations)        // List all integrations
		r.Patch("/{id}", integrationHandler.UpdateIntegration) // Update integration
	})

	// Monitoring activities endpoint
	r.Get("/monitoring-activities", activityHandler.ListActivities)

	return r
}
