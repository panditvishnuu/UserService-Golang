package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/panditvishnuu/userservice/internal/contextkeys"
	"github.com/panditvishnuu/userservice/internal/domain"
	httpx "github.com/panditvishnuu/userservice/internal/httpx"
	"github.com/panditvishnuu/userservice/internal/service"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *registerRequest) validate() error {
	if r.Name == "" || r.Email == "" || r.Password == "" {
		return fmt.Errorf("name, email and password are required")
	}

	// password length check
	if len(r.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// simple email validation
	at := strings.Index(r.Email, "@")
	if at <= 0 || at >= len(r.Email)-1 {
		return fmt.Errorf("invalid email format")
	}

	domain := r.Email[at+1:]
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type authResponse struct {
	Token string `json:"token"`
}

func writeError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	msg := "internal server error"

	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		code = http.StatusUnauthorized
		msg = "invalid credentials"
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		code = http.StatusConflict
		msg = "email already exists"
	case errors.Is(err, domain.ErrUserNotFound):
		code = http.StatusNotFound
		msg = "user not found"
	}

	slog.Error("request error", "error", err, "status", code)
	httpx.WriteError(w, code, msg)
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.validate(); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.svc.Register(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, userResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	})

}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, httpx.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, httpx.ErrorResponse{Error: "email and password are required"})
		return
	}

	token, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, authResponse{Token: token})

}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// extract userID from context
	userID, ok := r.Context().Value(contextkeys.UserIDKey).(string)
	if !ok || userID == "" {
		httpx.WriteJSON(w, http.StatusUnauthorized, httpx.ErrorResponse{Error: "unauthorized"})
		return
	}

	user, err := h.svc.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, userResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	})
}
