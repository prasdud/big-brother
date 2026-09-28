package api

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/prasdud/big-brother/internal/metrics"
	"github.com/prasdud/big-brother/internal/store"
)

// Server holds the dependencies for the HTTP API.
type Server struct {
	q         *store.Queries
	db        *sql.DB
	workspace store.Workspace
	metrics   *metrics.Registry
	web       http.Handler
	requests  *metrics.Counter
	logger    *slog.Logger
}

// New builds a Server.
func New(q *store.Queries, db *sql.DB, workspace store.Workspace, reg *metrics.Registry, web http.Handler, logger *slog.Logger) *Server {
	return &Server{
		q:         q,
		db:        db,
		workspace: workspace,
		metrics:   reg,
		web:       web,
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

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/projects", func(r chi.Router) {
			r.Get("/", s.listProjects)
			r.Post("/", s.createProject)
			r.Route("/{project}", func(r chi.Router) {
				r.Get("/", s.getProject)
				r.Patch("/", s.renameProject)
				r.Delete("/", s.deleteProject)
				r.Route("/services", func(r chi.Router) {
					r.Get("/", s.listServices)
					r.Post("/", s.createService)
					r.Route("/{service}", func(r chi.Router) {
						r.Get("/", s.getService)
						r.Patch("/", s.updateService)
						r.Delete("/", s.deleteService)
						r.Post("/pause", s.setServiceEnabled(false))
						r.Post("/resume", s.setServiceEnabled(true))
						r.Get("/status", s.getStatus)
						r.Get("/checks", s.listChecks)
						r.Get("/uptime", s.getUptime)
					})
				})
			})
		})
	})

	r.NotFound(s.web.ServeHTTP)
	return r
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
