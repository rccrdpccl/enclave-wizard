package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/auth"
)

type AuthHandler struct {
	store  *auth.Store
	noAuth bool
}

func NewAuthHandler(store *auth.Store, noAuth bool) *AuthHandler {
	return &AuthHandler{store: store, noAuth: noAuth}
}

type LoginInput struct {
	Body struct {
		Username string `json:"username" doc:"Username (always 'admin')" minLength:"1"`
		Password string `json:"password" doc:"Password" minLength:"1"`
	}
}

type LoginOutput struct {
	Body struct {
		Token              string `json:"token" doc:"Bearer token for API access"`
		MustChangePassword bool   `json:"mustChangePassword" doc:"True if the initial password has not been changed yet"`
	}
}

type ChangePasswordInput struct {
	Body struct {
		CurrentPassword string `json:"currentPassword" doc:"Current password" minLength:"1"`
		NewPassword     string `json:"newPassword" doc:"New password (min 8 characters)" minLength:"8"`
	}
}

type ChangePasswordOutput struct {
	Body struct {
		Token string `json:"token" doc:"New bearer token (previous tokens are invalidated)"`
	}
}

func (h *AuthHandler) login(_ context.Context, input *LoginInput) (*LoginOutput, error) {
	if input.Body.Username != "admin" {
		return nil, huma.Error401Unauthorized("invalid credentials")
	}

	if !h.store.CheckPassword(input.Body.Password) {
		return nil, huma.Error401Unauthorized("invalid credentials")
	}

	token, err := h.store.CreateToken()
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create token")
	}

	out := &LoginOutput{}
	out.Body.Token = token
	out.Body.MustChangePassword = h.store.MustChangePassword()
	return out, nil
}

func (h *AuthHandler) changePassword(_ context.Context, input *ChangePasswordInput) (*ChangePasswordOutput, error) {
	if err := h.store.ChangePassword(input.Body.CurrentPassword, input.Body.NewPassword); err != nil {
		if strings.Contains(err.Error(), "incorrect") {
			return nil, huma.Error401Unauthorized(err.Error())
		}
		return nil, huma.Error400BadRequest(err.Error())
	}

	token, err := h.store.CreateToken()
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create token")
	}

	out := &ChangePasswordOutput{}
	out.Body.Token = token
	return out, nil
}

type AuthModeOutput struct {
	Body struct {
		NoAuth bool `json:"noAuth" doc:"True if authentication is disabled"`
	}
}

func (h *AuthHandler) authMode(_ context.Context, _ *struct{}) (*AuthModeOutput, error) {
	out := &AuthModeOutput{}
	out.Body.NoAuth = h.noAuth
	return out, nil
}

func (h *AuthHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "authMode",
		Method:      http.MethodGet,
		Path:        "/api/v1/auth/mode",
		Summary:     "Check authentication mode",
		Tags:        []string{"auth"},
	}, h.authMode)

	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/login",
		Summary:     "Authenticate and get a bearer token",
		Tags:        []string{"auth"},
	}, h.login)

	huma.Register(api, huma.Operation{
		OperationID: "changePassword",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/password",
		Summary:     "Change admin password",
		Tags:        []string{"auth"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, h.changePassword)
}

// BearerAuthMiddleware returns an HTTP middleware that requires a valid
// bearer token for all paths except the login endpoint.
// When noAuth is true, all requests are passed through without checking tokens.
func BearerAuthMiddleware(store *auth.Store, noAuth bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if noAuth {
				next.ServeHTTP(w, r)
				return
			}

			if r.URL.Path == "/api/v1/auth/login" || r.URL.Path == "/api/v1/auth/mode" || r.URL.Path == "/api/v1/version" {
				next.ServeHTTP(w, r)
				return
			}

			if !strings.HasPrefix(r.URL.Path, "/api/v1/") {
				next.ServeHTTP(w, r)
				return
			}

			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, `{"title":"Unauthorized","status":401,"detail":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")
			if !store.ValidateToken(token) {
				http.Error(w, `{"title":"Unauthorized","status":401,"detail":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
