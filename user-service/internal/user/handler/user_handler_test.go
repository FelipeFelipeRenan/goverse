package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/domain"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserHandler_Register(t *testing.T) {
	t.Parallel()

	mockService := new(service.MockUserService)
	handler := NewUserHandler(mockService)

	input := domain.User{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secret",
	}

	// Espera que o método Register receba qualquer contexto e o input como valor (não ponteiro)
	mockService.
		On("Register", mock.Anything, input).
		Return(&domain.UserResponse{ID: "123"}, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var resp map[string]string
	err := json.NewDecoder(res.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "123", resp["user"])

	mockService.AssertExpectations(t)
}

func TestUserHandler_GetByID(t *testing.T) {
	t.Parallel()
	mockService := new(service.MockUserService)
	handler := NewUserHandler(mockService)

	user := &domain.User{ID: "123", Username: "alice", Email: "alice@example.com"}
	mockService.On("FindByID", mock.Anything, "123").Return(user, nil)

	req := httptest.NewRequest(http.MethodGet, "/user/123", nil)
	req.SetPathValue("id", "123")
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	res := w.Result()

	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var got domain.User
	json.NewDecoder(res.Body).Decode(&got)
	assert.Equal(t, "alice", got.Username)
	mockService.AssertExpectations(t)

}
func TestUserHandler_GetAllUsers(t *testing.T) {
	t.Parallel()
	mockService := new(service.MockUserService)
	handler := NewUserHandler(mockService)

	users := []domain.User{
		{ID: "1", Username: "alice"},
		{ID: "2", Username: "bob"},
	}

	mockService.On("GetAllUsers", mock.Anything).Return(users, nil)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	handler.GetAllUsers(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var got []domain.User

	json.NewDecoder(res.Body).Decode(&got)
	assert.Len(t, got, 2)
	assert.Equal(t, "alice", got[0].Username)
	mockService.AssertExpectations(t)

}

func TestUserHandler_Register_Error(t *testing.T) {
	t.Parallel()
	mockService := new(service.MockUserService)
	h := NewUserHandler(mockService)

	input := domain.User{Username: "alice", Email: "alice@example.com", Password: "secret"}
	mockService.On("Register", mock.Anything, input).Return(nil, errors.New("internal error"))

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	mockService := new(service.MockUserService)
	handler := NewUserHandler(mockService)

	mockService.On("FindByID", mock.Anything, "999").Return(nil, errors.New("not found"))

	req := httptest.NewRequest(http.MethodGet, "/user/999", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	mockService.AssertExpectations(t)

}

func TestUserHandler_GetAllUsers_Error(t *testing.T) {
	t.Parallel()
	mockService := new(service.MockUserService)
	handler := NewUserHandler(mockService)

	mockService.On("GetAllUsers", mock.Anything).Return(nil, errors.New("db failure"))

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	handler.GetAllUsers(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	mockService.AssertExpectations(t)
}
