// Package auth handles sign-in, sessions, CSRF, and role checks.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/prasdud/big-brother/internal/store"
)

// Cookie and header names used by the session and CSRF scheme.
const (
	CookieSession = "bb_session"
	CookieCSRF    = "bb_csrf"
	CookieState   = "bb_oauth_state"
	HeaderCSRF    = "X-CSRF-Token"
)

var (
	// ErrUnauthenticated means no valid session.
	ErrUnauthenticated = errors.New("unauthenticated")
	// ErrForbidden means the identity or role may not proceed.
	ErrForbidden = errors.New("forbidden")
)

// Identity is the verified result of an OIDC sign-in.
type Identity struct {
	Subject       string
	Email         string
	Name          string
	EmailVerified bool
}

// Verifier is the OIDC provider abstraction, faked in tests.
type Verifier interface {
	AuthCodeURL(state, nonce string) string
	Exchange(ctx context.Context, code, nonce string) (Identity, error)
}

// Config configures a Service.
type Config struct {
	WorkspaceID         string
	AllowedDomains      []string
	BootstrapAdminEmail string
	SessionTTL          time.Duration
	CookieSecure        bool
}

// Session is an authenticated request principal.
type Session struct {
	ID   string
	User store.User
	CSRF string
}

// Service owns sign-in and sessions.
type Service struct {
	q            *store.Queries
	verifier     Verifier
	workspaceID  string
	domains      []string
	bootstrap    string
	sessionTTL   time.Duration
	cookieSecure bool
	now          func() time.Time
}

// New builds a Service.
func New(q *store.Queries, v Verifier, cfg Config) *Service {
	ttl := cfg.SessionTTL
	if ttl <= 0 {
		ttl = 168 * time.Hour
	}
	return &Service{
		q:            q,
		verifier:     v,
		workspaceID:  cfg.WorkspaceID,
		domains:      cfg.AllowedDomains,
		bootstrap:    strings.ToLower(strings.TrimSpace(cfg.BootstrapAdminEmail)),
		sessionTTL:   ttl,
		cookieSecure: cfg.CookieSecure,
		now:          time.Now,
	}
}

// Secure reports whether cookies should be marked Secure.
func (s *Service) Secure() bool { return s.cookieSecure }

// SessionTTL is the lifetime of a new session.
func (s *Service) SessionTTL() time.Duration { return s.sessionTTL }

// LoginURL builds the provider authorization URL.
func (s *Service) LoginURL(state, nonce string) string {
	return s.verifier.AuthCodeURL(state, nonce)
}

// Exchange verifies an authorization code with the provider.
func (s *Service) Exchange(ctx context.Context, code, nonce string) (Identity, error) {
	return s.verifier.Exchange(ctx, code, nonce)
}

// AllowedDomain reports whether the email's domain is allowlisted.
func (s *Service) AllowedDomain(email string) bool {
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return false
	}
	domain := strings.ToLower(email[at+1:])
	for _, d := range s.domains {
		if d == domain {
			return true
		}
	}
	return false
}

// SignIn creates or updates the user for a verified identity.
func (s *Service) SignIn(ctx context.Context, id Identity) (store.User, error) {
	email := strings.ToLower(strings.TrimSpace(id.Email))
	if email == "" {
		return store.User{}, ErrForbidden
	}
	now := s.now().UTC().Format(time.RFC3339)

	u, err := s.q.GetUserByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		role := RoleViewer
		if s.isBootstrap(email) {
			if admins, cerr := s.q.CountAdmins(ctx); cerr == nil && admins == 0 {
				role = RoleAdmin
			}
		}
		uid := uuid.NewString()
		if err := s.q.CreateUser(ctx, store.CreateUserParams{
			ID:          uid,
			WorkspaceID: s.workspaceID,
			Email:       email,
			Name:        id.Name,
			Role:        role,
			GoogleSub:   nullString(id.Subject),
			Disabled:    0,
			CreatedAt:   now,
			LastLoginAt: nullString(now),
		}); err != nil {
			return store.User{}, err
		}
		return s.q.GetUserByID(ctx, uid)
	}
	if err != nil {
		return store.User{}, err
	}
	if u.Disabled != 0 {
		return store.User{}, ErrForbidden
	}

	if s.isBootstrap(email) && u.Role != RoleAdmin {
		if admins, cerr := s.q.CountAdmins(ctx); cerr == nil && admins == 0 {
			if err := s.q.UpdateUserRole(ctx, store.UpdateUserRoleParams{ID: u.ID, Role: RoleAdmin}); err == nil {
				u.Role = RoleAdmin
			}
		}
	}
	if err := s.q.UpdateUserLogin(ctx, store.UpdateUserLoginParams{
		GoogleSub:   nullString(id.Subject),
		Name:        id.Name,
		LastLoginAt: nullString(now),
		ID:          u.ID,
	}); err != nil {
		return store.User{}, err
	}
	return u, nil
}

func (s *Service) isBootstrap(email string) bool {
	return s.bootstrap != "" && email == s.bootstrap
}

// CreateSession issues a new session and returns the raw token and CSRF token.
func (s *Service) CreateSession(ctx context.Context, userID string) (token, csrf string, err error) {
	token, err = RandomToken()
	if err != nil {
		return "", "", err
	}
	csrf, err = RandomToken()
	if err != nil {
		return "", "", err
	}
	now := s.now().UTC()
	if err := s.q.CreateSession(ctx, store.CreateSessionParams{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: hashToken(token),
		CsrfToken: csrf,
		ExpiresAt: now.Add(s.sessionTTL).Format(time.RFC3339),
		CreatedAt: now.Format(time.RFC3339),
	}); err != nil {
		return "", "", err
	}
	return token, csrf, nil
}

// Authenticate resolves a raw session token to a principal, rejecting expired
// sessions and disabled users.
func (s *Service) Authenticate(ctx context.Context, token string) (Session, error) {
	if token == "" {
		return Session{}, ErrUnauthenticated
	}
	sess, err := s.q.GetSessionByTokenHash(ctx, hashToken(token))
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrUnauthenticated
	}
	if err != nil {
		return Session{}, err
	}
	expires, perr := time.Parse(time.RFC3339, sess.ExpiresAt)
	if perr != nil || !s.now().UTC().Before(expires) {
		return Session{}, ErrUnauthenticated
	}
	u, err := s.q.GetUserByID(ctx, sess.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrUnauthenticated
	}
	if err != nil {
		return Session{}, err
	}
	if u.Disabled != 0 {
		return Session{}, ErrUnauthenticated
	}
	return Session{ID: sess.ID, User: u, CSRF: sess.CsrfToken}, nil
}

// Logout deletes a session.
func (s *Service) Logout(ctx context.Context, sessionID string) error {
	return s.q.DeleteSession(ctx, sessionID)
}

// RandomToken returns a 256-bit URL-safe random string.
func RandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
