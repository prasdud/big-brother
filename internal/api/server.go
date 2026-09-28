package api

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/prasdud/big-brother/internal/alert"
	"github.com/prasdud/big-brother/internal/auth"
	"github.com/prasdud/big-brother/internal/metrics"
	"github.com/prasdud/big-brother/internal/secret"
	"github.com/prasdud/big-brother/internal/slack"
	"github.com/prasdud/big-brother/internal/store"
)

// SlackDeps wires the Slack integration. A nil value disables Slack routes.
type SlackDeps struct {
	Installer slack.Installer
	Box       *secret.Box
	Alerter   *alert.Alerter
}

// Server holds the dependencies for the HTTP API.
type Server struct {
	q         *store.Queries
	db        *sql.DB
	workspace store.Workspace
	metrics   *metrics.Registry
	web       http.Handler
	auth      *auth.Service
	slack     *SlackDeps
	requests  *metrics.Counter
	logger    *slog.Logger
}

// New builds a Server. auth and slack may be nil, which disables those
// features (local development and tests).
func New(q *store.Queries, db *sql.DB, workspace store.Workspace, reg *metrics.Registry, web http.Handler, authSvc *auth.Service, slackDeps *SlackDeps, logger *slog.Logger) *Server {
	return &Server{
		q:         q,
		db:        db,
		workspace: workspace,
		metrics:   reg,
		web:       web,
		auth:      authSvc,
		slack:     slackDeps,
		requests:  reg.Counter("bb_http_requests_total", "Total HTTP requests handled."),
		logger:    logger,
	}
}

// Router returns the fully wired HTTP handler.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(s.accessLog)

	r.Get("/healthz", s.health)
	r.Handle("/metrics", s.metrics)
	r.Get("/api/v1/openapi.json", s.openapiSpec)

	if s.auth != nil {
		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", s.authLogin)
			r.Get("/callback", s.authCallback)
			r.Group(func(r chi.Router) {
				r.Use(s.auth.RequireAuth)
				r.Get("/me", s.authMe)
				r.Group(func(r chi.Router) {
					r.Use(s.auth.CSRF)
					r.Post("/logout", s.authLogout)
				})
			})
		})
	} else {
		// Signal to the web app that authentication is not configured.
		authDisabled := func(w http.ResponseWriter, _ *http.Request) {
			writeError(w, http.StatusNotFound, "auth_disabled", "authentication is not configured")
		}
		r.Get("/auth/login", authDisabled)
		r.Get("/auth/callback", authDisabled)
		r.Get("/auth/me", authDisabled)
	}

	r.Route("/api/v1", func(r chi.Router) {
		if s.auth != nil {
			r.Use(s.auth.RequireAuth)
			r.Use(s.auth.CSRF)
		}
		if s.slack != nil {
			s.mountSlackRoutes(r)
		}
		r.Route("/projects", func(r chi.Router) {
			r.Get("/", s.listProjects)
			r.Group(func(r chi.Router) {
				r.Use(s.requireRole(auth.RoleAdmin))
				r.Post("/", s.createProject)
			})
			r.Route("/{project}", func(r chi.Router) {
				r.Get("/", s.getProject)
				r.Group(func(r chi.Router) {
					r.Use(s.requireRole(auth.RoleAdmin))
					r.Patch("/", s.renameProject)
					r.Delete("/", s.deleteProject)
				})
				if s.slack != nil {
					s.mountProjectSlackRoutes(r)
				}
				r.Route("/services", func(r chi.Router) {
					r.Get("/", s.listServices)
					r.Group(func(r chi.Router) {
						r.Use(s.requireRole(auth.RoleMember))
						r.Post("/", s.createService)
					})
					r.Route("/{service}", func(r chi.Router) {
						r.Get("/", s.getService)
						r.Get("/status", s.getStatus)
						r.Get("/checks", s.listChecks)
						r.Get("/uptime", s.getUptime)
						if s.slack != nil {
							r.Group(func(r chi.Router) {
								r.Use(s.requireRole(auth.RoleMember))
								r.Put("/channel", s.setServiceChannel)
								r.Delete("/channel", s.clearServiceChannel)
							})
						}
						r.Group(func(r chi.Router) {
							r.Use(s.requireRole(auth.RoleMember))
							r.Patch("/", s.updateService)
							r.Delete("/", s.deleteService)
							r.Post("/pause", s.setServiceEnabled(false))
							r.Post("/resume", s.setServiceEnabled(true))
						})
					})
				})
			})
		})
		r.Route("/users", func(r chi.Router) {
			r.Use(s.requireRole(auth.RoleAdmin))
			r.Get("/", s.listUsers)
			r.Post("/", s.createUser)
			r.Patch("/{user}", s.updateUser)
			r.Delete("/{user}", s.deleteUser)
		})
		r.Route("/import/kuma", func(r chi.Router) {
			r.Use(s.requireRole(auth.RoleAdmin))
			r.Post("/preview", s.kumaPreview)
			r.Post("/apply", s.kumaApply)
		})
	})

	r.NotFound(s.web.ServeHTTP)
	return r
}

func (s *Server) mountSlackRoutes(r chi.Router) {
	r.Route("/slack", func(r chi.Router) {
		r.Get("/", s.slackStatus)
		r.Group(func(r chi.Router) {
			r.Use(s.requireRole(auth.RoleMember))
			r.Get("/channels", s.slackChannels)
		})
		r.Group(func(r chi.Router) {
			r.Use(s.requireRole(auth.RoleAdmin))
			r.Get("/install", s.slackInstall)
			r.Get("/callback", s.slackCallback)
		})
	})
}

func (s *Server) mountProjectSlackRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(s.requireRole(auth.RoleMember))
		r.Put("/channel", s.setProjectChannel)
		r.Delete("/channel", s.clearProjectChannel)
		r.Put("/alert-templates/{trigger}", s.putAlertTemplate)
		r.Post("/alert-templates/preview", s.previewAlertTemplate)
		r.Post("/alert-templates/test-send", s.testSendAlert)
	})
	r.Get("/alert-templates", s.listAlertTemplates)
	r.Get("/channels", s.listProjectChannels)
	r.Get("/deliveries", s.listDeliveries)
}

// requireRole returns a role guard, or a pass-through when auth is disabled.
func (s *Server) requireRole(min string) func(http.Handler) http.Handler {
	if s.auth == nil {
		return func(next http.Handler) http.Handler { return next }
	}
	return auth.RequireRole(min)
}

func (s *Server) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		s.requests.Inc()
		s.logger.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", middleware.GetReqID(r.Context()),
		)
	})
}
