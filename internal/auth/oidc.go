package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCVerifier verifies Google (or any OIDC) sign-ins.
type OIDCVerifier struct {
	config   oauth2.Config
	verifier *oidc.IDTokenVerifier
}

// NewOIDCVerifier discovers the provider and prepares the OAuth2 and ID token
// verifiers.
func NewOIDCVerifier(ctx context.Context, issuer, clientID, clientSecret, redirectURL string) (*OIDCVerifier, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery for %q: %w", issuer, err)
	}
	return &OIDCVerifier{
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
		},
		verifier: provider.Verifier(&oidc.Config{ClientID: clientID}),
	}, nil
}

// AuthCodeURL returns the provider sign-in URL with the given state and nonce.
func (v *OIDCVerifier) AuthCodeURL(state, nonce string) string {
	return v.config.AuthCodeURL(state, oidc.Nonce(nonce))
}

// Exchange swaps an authorization code for a verified identity.
func (v *OIDCVerifier) Exchange(ctx context.Context, code, nonce string) (Identity, error) {
	token, err := v.config.Exchange(ctx, code)
	if err != nil {
		return Identity{}, fmt.Errorf("token exchange: %w", err)
	}
	rawID, ok := token.Extra("id_token").(string)
	if !ok || rawID == "" {
		return Identity{}, errors.New("id_token missing from token response")
	}
	idToken, err := v.verifier.Verify(ctx, rawID)
	if err != nil {
		return Identity{}, fmt.Errorf("verify id_token: %w", err)
	}
	if idToken.Nonce != nonce {
		return Identity{}, errors.New("nonce mismatch")
	}
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return Identity{}, fmt.Errorf("decode claims: %w", err)
	}
	return Identity{
		Subject:       idToken.Subject,
		Email:         claims.Email,
		Name:          claims.Name,
		EmailVerified: claims.EmailVerified,
	}, nil
}
