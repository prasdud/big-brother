package auth

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"

	"github.com/prasdud/big-brother/internal/store"
)

type ctxKey int

const sessionKey ctxKey = 0

// WithSession stores the session in the context.
func WithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionKey, s)
}

// SessionFrom reads the session stored by RequireAuth.
func SessionFrom(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(sessionKey).(Session)
	return s, ok
}

// UserFrom reads the authenticated user.
func UserFrom(ctx context.Context) (store.User, bool) {
	s, ok := SessionFrom(ctx)
	return s.User, ok
}

// RequireAuth rejects requests without a valid session.
func (s *Service) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieSession)
		if err != nil {
			authError(w, http.StatusUnauthorized, "unauthenticated", "sign in required")
			return
		}
		sess, err := s.Authenticate(r.Context(), cookie.Value)
		if err != nil {
			authError(w, http.StatusUnauthorized, "unauthenticated", "sign in required")
			return
		}
		next.ServeHTTP(w, r.WithContext(WithSession(r.Context(), sess)))
	})
}

// CSRF requires a matching CSRF header on state-changing requests.
func (s *Service) CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}
		sess, ok := SessionFrom(r.Context())
		if !ok {
			authError(w, http.StatusUnauthorized, "unauthenticated", "sign in required")
			return
		}
		got := r.Header.Get(HeaderCSRF)
		if subtle.ConstantTimeCompare([]byte(got), []byte(sess.CSRF)) != 1 {
			authError(w, http.StatusForbidden, "csrf", "missing or invalid CSRF token")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole rejects requests whose user is below the required role.
func RequireRole(min string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFrom(r.Context())
			if !ok {
				authError(w, http.StatusUnauthorized, "unauthenticated", "sign in required")
				return
			}
			if !AtLeast(u.Role, min) {
				authError(w, http.StatusForbidden, "forbidden", "insufficient role")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func authError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error":{"code":%q,"message":%q}}`, code, message)
}
