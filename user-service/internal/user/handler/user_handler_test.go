package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/domain"
	"github.com/FelipeFelipeRenan/goverse/user-service/internal/user/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)



func TestUserHandler_Register(t *testing.T) {
	mockService := new(service.MockUserService)
	handler := NewUserHandler(mockService)

	input := domain.User{Username: "alice", Email: "alice@example.com", Password: "secret"}

	mockService.On("Register", mock.Anything, input).Return("123", nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var resp map[string]string
	json.NewDecoder(res.Body).Decode(&resp)
	assert.Equal(t, "123", resp["id"])

	mockService.AssertExpectations(t)


}