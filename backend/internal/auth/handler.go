package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"omnishield/pkg/errors"
	"omnishield/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		response.Error(w, r, errors.BadRequest(errors.CodeValidationError, "Format payload JSON register tidak valid."))
		return
	}

	user, err := h.service.Register(r.Context(), &req)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Created(w, user)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		response.Error(w, r, errors.BadRequest(errors.CodeValidationError, "Format payload login tidak valid."))
		return
	}

	user, token, expiresAt, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "os_session",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	response.OK(w, AuthUserResponse{
		User:      *user,
		ExpiresAt: expiresAt,
	})
}

func (h *Handler) MeHandler(w http.ResponseWriter, r *http.Request) {
	token := extractSessionToken(r)

	user, err := h.service.GetMe(r.Context(), token)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.OK(w, user)
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	token := extractSessionToken(r)

	_ = h.service.Logout(r.Context(), token)

	http.SetCookie(w, &http.Cookie{
		Name:     "os_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	response.OK(w, map[string]string{
		"message": "Logout berhasil. Sesi telah dicabut.",
	})
}

func extractSessionToken(r *http.Request) string {
	cookie, err := r.Cookie("os_session")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}
