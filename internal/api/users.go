package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/prasdud/big-brother/internal/auth"
	"github.com/prasdud/big-brother/internal/store"
)

type userView struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	Disabled    bool   `json:"disabled"`
	CreatedAt   string `json:"created_at"`
	LastLoginAt string `json:"last_login_at"`
}

func newUserView(u store.User) userView {
	v := userView{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		Disabled:  u.Disabled != 0,
		CreatedAt: u.CreatedAt,
	}
	if u.LastLoginAt.Valid {
		v.LastLoginAt = u.LastLoginAt.String
	}
	return v
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.q.ListUsers(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	out := make([]userView, 0, len(users))
	for _, u := range users {
		out = append(out, newUserView(u))
	}
	writeJSON(w, http.StatusOK, out)
}

type createUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if !decode(w, r, &req) {
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if !strings.Contains(email, "@") {
		badRequest(w, "a valid email is required")
		return
	}
	if !auth.ValidRole(req.Role) {
		badRequest(w, "role must be one of admin, member, viewer")
		return
	}

	if _, err := s.q.GetUserByEmail(r.Context(), email); err == nil {
		writeError(w, http.StatusConflict, "conflict", "a user with that email already exists")
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		serverError(w, err)
		return
	}

	u := store.User{
		ID:          uuid.NewString(),
		WorkspaceID: s.workspace.ID,
		Email:       email,
		Name:        req.Name,
		Role:        req.Role,
		CreatedAt:   store.NowUTC(),
	}
	if err := s.q.CreateUser(r.Context(), store.CreateUserParams{
		ID:          u.ID,
		WorkspaceID: u.WorkspaceID,
		Email:       u.Email,
		Name:        u.Name,
		Role:        u.Role,
		GoogleSub:   sql.NullString{},
		Disabled:    0,
		CreatedAt:   u.CreatedAt,
		LastLoginAt: sql.NullString{},
	}); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, newUserView(u))
}

type updateUserRequest struct {
	Role     *string `json:"role"`
	Disabled *bool   `json:"disabled"`
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	u, err := s.q.GetUserByID(r.Context(), chi.URLParam(r, "user"))
	if errors.Is(err, sql.ErrNoRows) {
		notFound(w)
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}

	var req updateUserRequest
	if !decode(w, r, &req) {
		return
	}

	if req.Role != nil {
		if !auth.ValidRole(*req.Role) {
			badRequest(w, "role must be one of admin, member, viewer")
			return
		}
		if u.Role == auth.RoleAdmin && *req.Role != auth.RoleAdmin {
			if last, ok := s.isLastAdmin(r, u.ID); ok {
				badRequest(w, last)
				return
			}
		}
		if err := s.q.UpdateUserRole(r.Context(), store.UpdateUserRoleParams{ID: u.ID, Role: *req.Role}); err != nil {
			serverError(w, err)
			return
		}
		u.Role = *req.Role
	}

	if req.Disabled != nil {
		if *req.Disabled && u.Role == auth.RoleAdmin {
			if last, ok := s.isLastAdmin(r, u.ID); ok {
				badRequest(w, last)
				return
			}
		}
		disabled := int64(0)
		if *req.Disabled {
			disabled = 1
		}
		if err := s.q.SetUserDisabled(r.Context(), store.SetUserDisabledParams{ID: u.ID, Disabled: disabled}); err != nil {
			serverError(w, err)
			return
		}
		u.Disabled = disabled
	}

	writeJSON(w, http.StatusOK, newUserView(u))
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	u, err := s.q.GetUserByID(r.Context(), chi.URLParam(r, "user"))
	if errors.Is(err, sql.ErrNoRows) {
		notFound(w)
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	if u.Role == auth.RoleAdmin {
		if last, ok := s.isLastAdmin(r, u.ID); ok {
			badRequest(w, last)
			return
		}
	}
	if err := s.q.DeleteUser(r.Context(), u.ID); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// isLastAdmin reports whether excluding userID would leave no active admin.
func (s *Server) isLastAdmin(r *http.Request, excludeID string) (string, bool) {
	admins, err := s.q.CountAdmins(r.Context())
	if err != nil {
		return "", false
	}
	target, _ := s.q.GetUserByID(r.Context(), excludeID)
	isActiveAdmin := admins == 1 && target.Role == auth.RoleAdmin && target.Disabled == 0
	if isActiveAdmin {
		return "cannot remove the last active admin", true
	}
	return "", false
}
