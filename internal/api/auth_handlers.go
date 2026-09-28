package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/prasdud/big-brother/internal/auth"
)

const oauthStateTTL = 600

func (s *Server) authLogin(w http.ResponseWriter, r *http.Request) {
	state, err := auth.RandomToken()
	if err != nil {
		serverError(w, err)
		return
	}
	nonce, err := auth.RandomToken()
	if err != nil {
		serverError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieState,
		Value:    state + "." + nonce,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.auth.Secure(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   oauthStateTTL,
	})
	http.Redirect(w, r, s.auth.LoginURL(state, nonce), http.StatusFound)
}

func (s *Server) authCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(auth.CookieState)
	if err != nil {
		badRequest(w, "missing OAuth state cookie")
		return
	}
	parts := strings.SplitN(stateCookie.Value, ".", 2)
	if len(parts) != 2 {
		badRequest(w, "malformed OAuth state cookie")
		return
	}
	state, nonce := parts[0], parts[1]
	if r.URL.Query().Get("state") != state {
		badRequest(w, "OAuth state mismatch")
		return
	}
	if e := r.URL.Query().Get("error"); e != "" {
		writeError(w, http.StatusUnauthorized, "oauth_error", e)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		badRequest(w, "missing authorization code")
		return
	}

	clearCookie(w, auth.CookieState, s.auth.Secure())

	identity, err := s.auth.Exchange(r.Context(), code, nonce)
	if err != nil {
		s.logger.Warn("oidc exchange failed", "error", err)
		writeError(w, http.StatusUnauthorized, "sign_in_failed", "could not verify sign-in")
		return
	}
	if !s.auth.AllowedDomain(identity.Email) {
		s.logger.Warn("sign-in rejected: domain not allowed", "email", identity.Email)
		writeError(w, http.StatusForbidden, "domain_not_allowed", "email domain is not allowed")
		return
	}

	user, err := s.auth.SignIn(r.Context(), identity)
	if errors.Is(err, auth.ErrForbidden) {
		writeError(w, http.StatusForbidden, "forbidden", "account is disabled")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}

	token, csrf, err := s.auth.CreateSession(r.Context(), user.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	s.setSessionCookies(w, token, csrf)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) authMe(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFrom(r.Context())
	writeJSON(w, http.StatusOK, newUserView(sess.User))
}

func (s *Server) authLogout(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFrom(r.Context())
	if err := s.auth.Logout(r.Context(), sess.ID); err != nil {
		serverError(w, err)
		return
	}
	clearCookie(w, auth.CookieSession, s.auth.Secure())
	clearCookie(w, auth.CookieCSRF, s.auth.Secure())
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) setSessionCookies(w http.ResponseWriter, token, csrf string) {
	maxAge := int(s.auth.SessionTTL().Seconds())
	http.SetCookie(w, &http.Cookie{
		Name: auth.CookieSession, Value: token, Path: "/",
		HttpOnly: true, Secure: s.auth.Secure(),
		SameSite: http.SameSiteLaxMode, MaxAge: maxAge,
	})
	http.SetCookie(w, &http.Cookie{
		Name: auth.CookieCSRF, Value: csrf, Path: "/",
		HttpOnly: false, Secure: s.auth.Secure(),
		SameSite: http.SameSiteLaxMode, MaxAge: maxAge,
	})
}

func clearCookie(w http.ResponseWriter, name string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: "", Path: "/",
		HttpOnly: true, Secure: secure,
		SameSite: http.SameSiteLaxMode, MaxAge: -1,
	})
}
