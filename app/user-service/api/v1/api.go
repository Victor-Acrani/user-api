package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// API contains the data required for API v1.
type API struct {
	LivenessHandler  http.HandlerFunc
	ReadinessHandler http.HandlerFunc
	UserHandler      UserUseCase
}

// Routes set routes for chi mux.
func (a *API) Routes(r *chi.Mux) {
	r.Get("/liveness", a.LivenessHandler)
	r.Get("/readness", a.ReadinessHandler)
	r.Get("/api/v1/users/{user_id}", GetUserHandler(a.UserHandler))
}

type ErrorResponse struct {
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
}
