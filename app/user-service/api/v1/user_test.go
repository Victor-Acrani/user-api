package v1_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/Victor-Acrani/user-api/app/user-service/api/v1"
	"github.com/Victor-Acrani/user-api/app/user-service/api/v1/mocks"
	"github.com/Victor-Acrani/user-api/domain/entity"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGetUserHandler(t *testing.T) {
	t.Run("should return 200 and an user", func(t *testing.T) {
		uc := mocks.UserUseCaseMock{
			GetUserFunc: func(ctx context.Context, userID string) (entity.User, error) {
				return entity.User{
					Name:     "John Doe",
					Email:    "johndoe@email.com",
					Password: "123456",
					BirthDay: "10/02/1990",
				}, nil
			},
		}

		api := v1.API{
			UserHandler: &uc,
		}

		router := chi.NewRouter()
		api.Routes(router)

		ts := httptest.NewServer(router)
		defer ts.Close()

		serverUrl := fmt.Sprintf("%s/api/v1/users/101", ts.URL)
		req, err := http.NewRequest(http.MethodGet, serverUrl, nil)
		assert.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		expected := entity.User{
			Name:     "John Doe",
			Email:    "johndoe@email.com",
			Password: "123456",
			BirthDay: "10/02/1990",
		}

		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		assert.NoError(t, err)

		var actual entity.User
		err = json.Unmarshal(body, &actual)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("should return 400", func(t *testing.T) {
		uc := mocks.UserUseCaseMock{
			GetUserFunc: func(ctx context.Context, userID string) (entity.User, error) {
				return entity.User{}, errors.New("user not found")
			},
		}

		api := v1.API{
			UserHandler: &uc,
		}

		router := chi.NewRouter()
		api.Routes(router)

		ts := httptest.NewServer(router)
		defer ts.Close()

		serverUrl := fmt.Sprintf("%s/api/v1/users/101", ts.URL)
		req, err := http.NewRequest(http.MethodGet, serverUrl, nil)
		assert.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		expected := v1.ErrorResponse{
			ErrorCode: http.StatusBadRequest,
			Message:   "user not found",
		}

		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		assert.NoError(t, err)

		var actual v1.ErrorResponse
		err = json.Unmarshal(body, &actual)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
