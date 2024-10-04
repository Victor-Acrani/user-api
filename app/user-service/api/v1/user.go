package v1

import (
	"context"
	"log"
	"net/http"

	"github.com/Victor-Acrani/user-api/domain/entity"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

//go:generate moq -stub -pkg mocks -out mocks/get_user_uc.go . UserUseCase
type UserUseCase interface {
	GetUser(ctx context.Context, userID string) (entity.User, error)
}

// GetUserHandler returns a handler for getting users.
func GetUserHandler(uc UserUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := chi.URLParam(r, "user_id")
		user, err := uc.GetUser(r.Context(), userId)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrorResponse{ErrorCode: http.StatusBadRequest, Message: "user not found"})
			return
		}

		log.Println("user: ", user)

		render.Status(r, http.StatusOK)
		render.JSON(w, r, user)
	}
}
