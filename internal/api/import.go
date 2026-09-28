package api

import (
	"io"
	"net/http"

	"github.com/prasdud/big-brother/internal/kuma"
)

const maxImportBytes = 10 << 20

func (s *Server) kumaPreview(w http.ResponseWriter, r *http.Request) {
	s.handleKuma(w, r, false)
}

func (s *Server) kumaApply(w http.ResponseWriter, r *http.Request) {
	s.handleKuma(w, r, true)
}

func (s *Server) handleKuma(w http.ResponseWriter, r *http.Request, apply bool) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxImportBytes))
	if err != nil {
		badRequest(w, "could not read request body")
		return
	}
	if len(body) == 0 {
		badRequest(w, "request body is empty")
		return
	}
	if _, err := kuma.Parse(body); err != nil {
		badRequest(w, "invalid Kuma export: "+err.Error())
		return
	}

	importer := kuma.New(s.q, s.db, s.workspace.ID)
	var report kuma.Report
	if apply {
		report, err = importer.Apply(r.Context(), body)
	} else {
		report, err = importer.Preview(r.Context(), body)
	}
	if err != nil {
		serverError(w, err)
		return
	}
	if apply {
		s.logger.Info("kuma import applied",
			"projects", len(report.Projects),
			"services", len(report.Services),
			"skipped", len(report.Skipped),
		)
	}
	writeJSON(w, http.StatusOK, report)
}
