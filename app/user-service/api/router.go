package api

import (
	"github.com/go-chi/chi/v5"
)

// NewRouter creates a new chi router.
func NewRouter() *chi.Mux {
	// create router
	r := chi.NewRouter()
	return r
}
