package slack

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestPostMessage(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		if !strings.HasSuffix(r.URL.Path, "/chat.postMessage") {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := NewAPIClient("xoxb-1")
	c.base = srv.URL
	if err := c.PostMessage(context.Background(), "C1", "hello"); err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer xoxb-1" {
		t.Fatalf("auth = %q", auth)
	}
}

func TestPostMessageSurfacesSlackError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"ok":false,"error":"channel_not_found"}`))
	}))
	defer srv.Close()

	c := NewAPIClient("xoxb-1")
	c.base = srv.URL
	err := c.PostMessage(context.Background(), "C1", "hello")
	if err == nil || !strings.Contains(err.Error(), "channel_not_found") {
		t.Fatalf("err = %v", err)
	}
}

func TestPostMessageRetriesOnRateLimit(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := NewAPIClient("xoxb-1")
	c.base = srv.URL
	if err := c.PostMessage(context.Background(), "C1", "hello"); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestListChannelsPaginates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cursor") == "" {
			w.Write([]byte(`{"ok":true,"channels":[{"id":"C1","name":"general"}],"response_metadata":{"next_cursor":"next"}}`))
			return
		}
		w.Write([]byte(`{"ok":true,"channels":[{"id":"C2","name":"alerts"}],"response_metadata":{"next_cursor":""}}`))
	}))
	defer srv.Close()

	c := NewAPIClient("xoxb-1")
	c.base = srv.URL
	channels, err := c.ListChannels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(channels) != 2 || channels[1].Name != "alerts" {
		t.Fatalf("channels = %+v", channels)
	}
}

func TestAuthorizeURL(t *testing.T) {
	o := NewOAuth("client", "secret", "https://app.test/api/v1/slack/callback")
	got := o.AuthorizeURL("state123")
	for _, want := range []string{"client_id=client", "state=state123", "redirect_uri=", "slack.com/oauth/v2/authorize"} {
		if !strings.Contains(got, want) {
			t.Errorf("authorize URL %q missing %q", got, want)
		}
	}
}
