// Package web exposes the embedded single-page application.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var dist embed.FS

// Handler serves the built SPA, falling back to index.html for client-side
// routes. When dist has not been built, it serves a placeholder page.
func Handler() http.Handler {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		return placeholder()
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return placeholder()
	}

	files := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := strings.TrimPrefix(r.URL.Path, "/"); p != "" {
			if _, err := fs.Stat(sub, p); err == nil {
				files.ServeHTTP(w, r)
				return
			}
		}
		r2 := new(http.Request)
		*r2 = *r
		r2.URL.Path = "/"
		files.ServeHTTP(w, r2)
	})
}

func placeholder() http.Handler {
	const page = `<!doctype html>
<html><head><meta charset="utf-8"><title>big-brother</title></head>
<body><h1>big-brother</h1><p>Web UI not built. Run <code>make web</code>.</p></body></html>`
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(page))
	})
}
