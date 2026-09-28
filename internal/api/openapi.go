package api

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.json
var openapiDocument []byte

func (s *Server) openapiSpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(openapiDocument)
}
