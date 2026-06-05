package api

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/denisakp/ogoune/internal/api/handler"
	v1handler "github.com/denisakp/ogoune/internal/api/handler/v1"
	"github.com/denisakp/ogoune/internal/api/middleware"
	"github.com/denisakp/ogoune/internal/config"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/denisakp/ogoune/pkg/logger"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	contentTypeJSON      = "application/json"
	headerContentType    = "Content-Type"
	routeCredentials     = "/{id}/credentials"
	routeCredentialsTest = "/{id}/credentials/test"
)

// NewRouter creates and configures the main HTTP router with all JSON API routes.
// All endpoints return JSON responses - no HTML rendering.
func NewRouter(
	resourceHandler *handler.ResourceHandler,
	pingHandler *handler.PingHandler,
	activityHandler *handler.MonitoringActivityHandler,
	tagHandler *handler.TagHandler,
	componentHandler *handler.ComponentHandler,
	statusPageHandler *handler.StatusPageHandler,
	publicStatusHandler *handler.PublicStatusHandler,
	statusPageSettingsHandler *handler.StatusPageSettingsHandler,
	incidentHandler *handler.IncidentHandler,
	notificationHandler *handler.NotificationHandler,
	maintenanceHandler *handler.MaintenanceHandler,
	statsHandler *handler.StatsHandler,
	systemHandler *handler.SystemHandler,
	runtimeConfigHandler *handler.RuntimeConfigHandler,
	authHandler *handler.AuthHandler,
	accountHandler *handler.AccountHandler,
	authService *service.AuthService,
	apiKeyService *service.APIKeyService,
	sessionService *service.SessionService,
	sessionHandler *handler.SessionHandler,
	twoFactorV1Handler *v1handler.TwoFactorHandler,
	escalationV1Handler *v1handler.EscalationHandler,
	monitorV1Handler *v1handler.MonitorHandler,
	incidentV1Handler *v1handler.IncidentHandler,
	channelV1Handler *v1handler.NotificationChannelHandler,
	componentV1Handler *v1handler.ComponentHandler,
	tagV1Handler *v1handler.TagHandler,
	statusPageV1Handler *v1handler.StatusPageV1Handler,
	heartbeatV1Handler *v1handler.HeartbeatV1Handler,
	credentialV1Handler *v1handler.ResourceCredentialHandler,
	enableSwagger bool,
	cfg *config.Config,
) http.Handler {
	r := chi.NewRouter()

	// Standard middleware stack for logging, recovery, and request tracking
	r.Use(logger.RequestIDMiddleware)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
		AppEnv:            cfg.AppEnv,
		EnableSwagger:     enableSwagger,
		SwaggerPathPrefix: "/v1/docs/",
	}))
	r.Use(logger.RequestLoggerMiddleware)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.SetHeader(headerContentType, contentTypeJSON))

	// CORS middleware — configurable origin allow-list
	corsOpts := cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{headerContentType, "Authorization", "X-API-Key"},
		AllowCredentials: true,
		MaxAge:           3600,
	}
	if len(cfg.CORSAllowedOrigins) > 0 {
		corsOpts.AllowedOrigins = cfg.CORSAllowedOrigins
	}
	// When AllowedOrigins is empty, go-chi/cors defaults to ["*"].
	// To enforce same-origin only, we use a validator that rejects everything.
	if len(cfg.CORSAllowedOrigins) == 0 {
		corsOpts.AllowOriginFunc = func(r *http.Request, origin string) bool {
			return false
		}
	}
	r.Use(cors.Handler(corsOpts))
	r.Use(corsRejectionLogger)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerContentType, "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// ========================================
	// Public Routes (no authentication required)
	// ========================================

	// Authentication endpoints — rate limited per IP
	r.Route("/auth", func(r chi.Router) {
		r.Use(httprate.LimitByIP(cfg.RateLimitAuth, cfg.RateLimitAuthWindow))
		r.Use(rateLimitLogger("auth"))
		r.Post("/login", authHandler.Login)                            // POST /auth/login - authenticate user
		r.Post("/initialize-password", authHandler.InitializePassword) // POST /auth/initialize-password
		r.Post("/verify-2fa", authHandler.Verify2FA)                   // POST /auth/verify-2fa
		r.Get("/verify", authHandler.Verify)                           // GET /auth/verify - verify JWT token
		// Public 2FA reset flow (no auth — user lost their authenticator).
		r.Post("/2fa/reset-request", twoFactorV1Handler.RequestReset)
		r.Post("/2fa/reset", twoFactorV1Handler.ConfirmReset)
	})

	// Public heartbeat ping endpoint (unauthenticated by design)
	r.Get("/ping/{slug}", pingHandler.Ping)
	r.Post("/ping/{slug}", pingHandler.Ping)

	// Public status page (spec 060) — short-cached JSON, no auth.
	r.Group(func(r chi.Router) {
		r.Use(middleware.PublicStatusCache(60, 30))
		r.Get("/status", publicStatusHandler.GetCurrent)
		r.Get("/status/incidents", publicStatusHandler.GetIncidents)
	})
	// Legacy resource detail endpoint — replaced by /status/resource/:id/windows in US3.
	r.Get("/status/{resourceId}", statusPageHandler.HandleResourceDetailStatus)
	r.Get("/system/edition", systemHandler.GetEdition)
	r.Get("/system/capabilities", systemHandler.GetCapabilities)
	r.Get("/config/runtime", runtimeConfigHandler.Get)

	// ========================================
	// Protected Routes (authentication required)
	// All routes below require valid JWT token
	// ========================================
	r.Group(func(r chi.Router) {
		// Apply auth middleware to all routes in this group
		r.Use(middleware.AuthMiddleware(authService, apiKeyService, sessionService))
		// Global rate limiting for authenticated endpoints
		r.Use(httprate.LimitByIP(cfg.RateLimitGlobal, cfg.RateLimitGlobalWindow))
		r.Use(rateLimitLogger("global"))

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

		// Sessions API — spec 059 FR-008/009.
		r.Route("/me/sessions", func(r chi.Router) {
			r.Use(middleware.RequireJWTOnly)
			r.Get("/", sessionHandler.List)
			r.Delete("/others", sessionHandler.RevokeOthers)
			r.Delete("/{id}", sessionHandler.Revoke)
		})

		// 2FA setup / verify / disable — spec 059 FR-010..FR-012.
		r.Route("/me/2fa", func(r chi.Router) {
			r.Use(middleware.RequireJWTOnly)
			r.Post("/setup", twoFactorV1Handler.Setup)
			r.Post("/verify", twoFactorV1Handler.Verify)
			r.Post("/disable", twoFactorV1Handler.Disable)
		})

		// Escalation policies — spec 059 FR-023..FR-026a.
		r.Route("/escalation-policies", func(r chi.Router) {
			r.Use(middleware.RequireJWTOnly)
			r.Get("/", escalationV1Handler.List)
			r.Post("/", escalationV1Handler.Create)
			r.Patch("/reorder", escalationV1Handler.Reorder)
			r.Patch("/{id}", escalationV1Handler.Update)
			r.Delete("/{id}", escalationV1Handler.Delete)
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

			// Resource credentials (feature 028)
			r.Get(routeCredentials, credentialV1Handler.Get)                                        // GET /resources/{id}/credentials - get masked credential
			r.With(middleware.RequireReadWrite).Post(routeCredentials, credentialV1Handler.Set)      // POST /resources/{id}/credentials - create/replace credential
			r.With(middleware.RequireReadWrite).Delete(routeCredentials, credentialV1Handler.Delete) // DELETE /resources/{id}/credentials - remove credential
			r.With(middleware.RequireReadWrite, middleware.PerUserRateLimit(10)).
				Post(routeCredentialsTest, credentialV1Handler.Test) // POST /resources/{id}/credentials/test - live-test (10 req/min/user)
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

	// ========================================
	// Public API v1 Routes
	// ========================================
	// Note: the router is mounted at /api by main.go, so /v1 here becomes /api/v1 publicly.
	r.Route("/v1", func(r chi.Router) {
		// Unauthenticated v1 sub-group (heartbeat ping, OpenAPI spec)
		// Individual routes registered per user story (T021, T036)
		// Heartbeat ping — public, no auth required (T037)
		r.Post("/heartbeat/ping/{slug}", heartbeatV1Handler.Ping)
		// OpenAPI spec — always active
		r.Get("/openapi.json", serveOpenAPISpec)

		// Swagger UI — only when enabled
		if enableSwagger {
			registerSwaggerUI(r)
		}

		// Authenticated v1 sub-group
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(authService, apiKeyService, sessionService))
			// Monitors
			r.Route("/monitors", func(r chi.Router) {
				r.Get("/", monitorV1Handler.List)
				r.With(middleware.RequireReadWrite).Post("/", monitorV1Handler.Create)
				r.Get("/{id}", monitorV1Handler.Get)
				r.With(middleware.RequireReadWrite).Put("/{id}", monitorV1Handler.Update)
				r.With(middleware.RequireReadWrite).Delete("/{id}", monitorV1Handler.Delete)
				r.With(middleware.RequireReadWrite).Post("/{id}/pause", monitorV1Handler.Pause)
				r.With(middleware.RequireReadWrite).Post("/{id}/resume", monitorV1Handler.Resume)
			})
			// Incident routes — registered in T029
			r.Route("/incidents", func(r chi.Router) {
				r.Get("/", incidentV1Handler.List)
				r.Get("/{id}", incidentV1Handler.Get)
			})
			// Notification channel routes — registered in T029
			r.Route("/notification-channels", func(r chi.Router) {
				r.Get("/", channelV1Handler.List)
				r.With(middleware.RequireReadWrite).Post("/", channelV1Handler.Create)
				r.Get("/{id}", channelV1Handler.Get)
				r.With(middleware.RequireReadWrite).Put("/{id}", channelV1Handler.Update)
				r.With(middleware.RequireReadWrite).Delete("/{id}", channelV1Handler.Delete)
			})
			// Component routes — registered in T034
			r.Route("/components", func(r chi.Router) {
				r.Get("/", componentV1Handler.List)
				r.With(middleware.RequireReadWrite).Post("/", componentV1Handler.Create)
				r.Get("/{id}", componentV1Handler.Get)
				r.With(middleware.RequireReadWrite).Put("/{id}", componentV1Handler.Update)
				r.With(middleware.RequireReadWrite).Delete("/{id}", componentV1Handler.Delete)
			})
			// Tag routes — registered in T034
			r.Route("/tags", func(r chi.Router) {
				r.Get("/", tagV1Handler.List)
				r.With(middleware.RequireReadWrite).Post("/", tagV1Handler.Create)
				r.Get("/{id}", tagV1Handler.Get)
				r.With(middleware.RequireReadWrite).Put("/{id}", tagV1Handler.Update)
				r.With(middleware.RequireReadWrite).Delete("/{id}", tagV1Handler.Delete)
			})
			// Status page routes — registered in T034
			r.Route("/status-pages", func(r chi.Router) {
				r.Get("/", statusPageV1Handler.List)
			})
		})
	})

	return r
}

// corsRejectionLogger logs a [security] event when a cross-origin request is rejected.
// It runs after the CORS middleware: if an Origin header was present but no
// Access-Control-Allow-Origin was set in the response, the origin was rejected.
func corsRejectionLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		origin := r.Header.Get("Origin")
		if origin != "" && w.Header().Get("Access-Control-Allow-Origin") == "" {
			slog.Warn("CORS request rejected",
				"event", "cors_reject",
				"source_ip", r.RemoteAddr,
				"origin", origin,
				"endpoint", r.URL.Path,
				"method", r.Method,
			)
		}
	})
}

// rateLimitLogger returns middleware that logs a [security] event when a rate limit
// response (429) is about to be sent.
func rateLimitLogger(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			if ww.Status() == http.StatusTooManyRequests {
				slog.Warn("rate limit exceeded",
					"event", "rate_limit",
					"scope", scope,
					"source_ip", r.RemoteAddr,
					"endpoint", r.URL.Path,
					"method", r.Method,
				)
			}
		})
	}
}

// serveOpenAPISpec serves the generated OpenAPI spec JSON file.
// Returns 503 SPEC_NOT_AVAILABLE if docs/swagger.json has not been generated yet.
func serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	const specPath = "docs/swagger.json"
	data, err := os.ReadFile(specPath)
	if err != nil {
		w.Header().Set(headerContentType, contentTypeJSON)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":{"code":"SPEC_NOT_AVAILABLE","message":"OpenAPI spec not generated; run make swag"}}`))
		return
	}
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// registerSwaggerUI mounts the Swagger UI handler under /api/v1/docs/*.
// Only called when ENABLE_SWAGGER=true.
func registerSwaggerUI(r chi.Router) {
	r.Get("/docs/*", httpSwagger.Handler(httpSwagger.URL("/api/v1/openapi.json")))
}
