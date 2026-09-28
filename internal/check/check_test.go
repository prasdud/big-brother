package check

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prasdud/big-brother/internal/store"
)

func svc(url string, timeoutSeconds int64) store.Service {
	return store.Service{Type: "http", Url: url, TimeoutSeconds: timeoutSeconds}
}

func TestHTTPChecker(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/redirect":
			http.Redirect(w, r, "/ok", http.StatusFound)
		case "/fail":
			w.WriteHeader(http.StatusInternalServerError)
		case "/slow":
			<-r.Context().Done()
		}
	}))
	defer srv.Close()

	c := New()

	t.Run("2xx is up", func(t *testing.T) {
		res := c.Check(context.Background(), svc(srv.URL+"/ok", 5))
		if !res.Up || res.StatusCode != 200 {
			t.Fatalf("got %+v", res)
		}
	})

	t.Run("3xx is up", func(t *testing.T) {
		res := c.Check(context.Background(), svc(srv.URL+"/redirect", 5))
		if !res.Up || res.StatusCode != 302 {
			t.Fatalf("got %+v", res)
		}
	})

	t.Run("5xx is down", func(t *testing.T) {
		res := c.Check(context.Background(), svc(srv.URL+"/fail", 5))
		if res.Up || res.StatusCode != 500 || res.Error == "" {
			t.Fatalf("got %+v", res)
		}
	})

	t.Run("timeout is down", func(t *testing.T) {
		res := c.Check(context.Background(), svc(srv.URL+"/slow", 1))
		if res.Up || res.Error == "" {
			t.Fatalf("got %+v", res)
		}
	})

	t.Run("connection refused is down", func(t *testing.T) {
		res := c.Check(context.Background(), svc("http://127.0.0.1:1", 2))
		if res.Up || res.Error == "" {
			t.Fatalf("got %+v", res)
		}
	})

	t.Run("unsupported type is down", func(t *testing.T) {
		s := store.Service{Type: "carrier-pigeon", Url: srv.URL + "/ok", TimeoutSeconds: 2}
		if res := c.Check(context.Background(), s); res.Up || res.Error == "" {
			t.Fatalf("got %+v", res)
		}
	})
}
