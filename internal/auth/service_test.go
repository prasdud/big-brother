package auth

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/prasdud/big-brother/internal/db"
	"github.com/prasdud/big-brother/internal/store"
)

type fakeVerifier struct{}

func (fakeVerifier) AuthCodeURL(state, nonce string) string { return state }
func (fakeVerifier) Exchange(context.Context, string, string) (Identity, error) {
	return Identity{}, nil
}

func setup(t *testing.T, cfg Config) (*Service, store.Workspace) {
	t.Helper()
	sqldb, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { sqldb.Close() })
	if err := db.Migrate(sqldb); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	q := store.New(sqldb)
	ws, err := store.EnsureWorkspace(context.Background(), q, "Test")
	if err != nil {
		t.Fatalf("workspace: %v", err)
	}
	cfg.WorkspaceID = ws.ID
	if cfg.SessionTTL == 0 {
		cfg.SessionTTL = time.Hour
	}
	return New(q, fakeVerifier{}, cfg), ws
}

func TestAllowedDomain(t *testing.T) {
	s, _ := setup(t, Config{AllowedDomains: []string{"example.com", "corp.dev"}})
	for _, email := range []string{"a@example.com", "b@corp.dev"} {
		if !s.AllowedDomain(email) {
			t.Errorf("AllowedDomain(%q) = false", email)
		}
	}
	for _, email := range []string{"a@other.com", "example.com", "", "a@sub.example.com"} {
		if s.AllowedDomain(email) {
			t.Errorf("AllowedDomain(%q) = true, want false", email)
		}
	}
}

func TestSignInDefaultsToViewer(t *testing.T) {
	s, _ := setup(t, Config{AllowedDomains: []string{"example.com"}})
	u, err := s.SignIn(context.Background(), Identity{Email: "Alice@Example.com", Name: "Alice", Subject: "sub1"})
	if err != nil {
		t.Fatal(err)
	}
	if u.Role != RoleViewer {
		t.Fatalf("role = %q, want viewer", u.Role)
	}
	if u.Email != "alice@example.com" {
		t.Fatalf("email = %q, want lowercased", u.Email)
	}
}

func TestBootstrapAdmin(t *testing.T) {
	s, _ := setup(t, Config{
		AllowedDomains:      []string{"example.com"},
		BootstrapAdminEmail: "boss@example.com",
	})
	ctx := context.Background()

	boss, err := s.SignIn(ctx, Identity{Email: "boss@example.com", Subject: "s1"})
	if err != nil {
		t.Fatal(err)
	}
	if boss.Role != RoleAdmin {
		t.Fatalf("bootstrap role = %q, want admin", boss.Role)
	}

	other, err := s.SignIn(ctx, Identity{Email: "someone@example.com", Subject: "s2"})
	if err != nil {
		t.Fatal(err)
	}
	if other.Role != RoleViewer {
		t.Fatalf("other role = %q, want viewer", other.Role)
	}

	// A different bootstrap email must not be promoted once an admin exists.
	s2, _ := setup(t, Config{
		AllowedDomains:      []string{"example.com"},
		BootstrapAdminEmail: "boss@example.com",
	})
	_, _ = s2.SignIn(ctx, Identity{Email: "boss@example.com", Subject: "s1"})
	late, _ := s2.SignIn(ctx, Identity{Email: "boss@example.com", Subject: "s1"})
	if late.Role != RoleAdmin {
		t.Fatalf("late bootstrap role = %q, want admin", late.Role)
	}
}

func TestSignInRejectsDisabled(t *testing.T) {
	s, _ := setup(t, Config{AllowedDomains: []string{"example.com"}})
	ctx := context.Background()
	u, err := s.SignIn(ctx, Identity{Email: "gone@example.com", Subject: "s"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.q.SetUserDisabled(ctx, store.SetUserDisabledParams{ID: u.ID, Disabled: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SignIn(ctx, Identity{Email: "gone@example.com", Subject: "s"}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	s, _ := setup(t, Config{AllowedDomains: []string{"example.com"}, SessionTTL: time.Hour})
	ctx := context.Background()

	u, err := s.SignIn(ctx, Identity{Email: "user@example.com", Subject: "s"})
	if err != nil {
		t.Fatal(err)
	}
	token, csrf, err := s.CreateSession(ctx, u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || csrf == "" {
		t.Fatal("empty token or csrf")
	}

	sess, err := s.Authenticate(ctx, token)
	if err != nil {
		t.Fatal(err)
	}
	if sess.User.ID != u.ID || sess.CSRF != csrf {
		t.Fatalf("session = %+v", sess)
	}

	if _, err := s.Authenticate(ctx, "not-a-token"); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("bogus token err = %v", err)
	}

	s.now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	if _, err := s.Authenticate(ctx, token); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expired session err = %v, want ErrUnauthenticated", err)
	}

	if err := s.Logout(ctx, sess.ID); err != nil {
		t.Fatal(err)
	}
}
