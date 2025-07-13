package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	pb "go_grpc_github_api/pb"
)

func TestGetUserProfile(t *testing.T) {
	// Step 1: Set up a mock GitHub API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"login": "mockuser",
			"name": "Mock User",
			"email": "mock@example.com",
			"bio": "testing bio",
			"followers": 123,
			"avatar_url": "https://avatars.githubusercontent.com/u/mockuser"
		}`))
	}))
	defer mockServer.Close()

	// Step 2: Create a new server instance with mocked values
	s := NewServer("test-token", mockServer.URL+"/")

	// Step 3: Call the handler
	resp, err := s.GetUserProfile(context.Background(), &pb.GetUserRequest{Username: "mockuser"})

	// Step 4: Assert response
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "mockuser", resp.Login)
	assert.Equal(t, "Mock User", resp.Name)
	assert.Equal(t, "mock@example.com", resp.Email)
	assert.Equal(t, "testing bio", resp.Bio)
	assert.Equal(t, int32(123), resp.Followers)
	assert.Equal(t, "https://avatars.githubusercontent.com/u/mockuser", resp.AvatarUrl)
}
