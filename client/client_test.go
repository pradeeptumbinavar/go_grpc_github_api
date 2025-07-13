package main

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	pb "go_grpc_github_api/pb"
	"google.golang.org/grpc"
)

// Step 1: Create a mock server that implements the gRPC interface
type mockUserServiceServer struct {
	pb.UnimplementedUserServiceServer
}

func (s *mockUserServiceServer) GetUserProfile(ctx context.Context, req *pb.GetUserRequest) (*pb.UserProfile, error) {
	return &pb.UserProfile{
		Login:     "mockedlogin",
		Name:      "Mocked Name",
		Email:     "mock@example.com",
		Bio:       "this is a mocked bio",
		Followers: 77,
		AvatarUrl: "https://mock.com/avatar.png",
	}, nil
}

func startMockGRPCServer(t *testing.T) (pb.UserServiceClient, func()) {
	// Step 2: Start a listener on a random port
	lis, err := net.Listen("tcp", "localhost:0")
	assert.NoError(t, err)

	// Step 3: Create a real gRPC server and register mock service
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, &mockUserServiceServer{})

	// Run the server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Mock server failed: %v", err)
		}
	}()

	// Step 4: Dial the mock server
	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
	assert.NoError(t, err)

	client := pb.NewUserServiceClient(conn)

	// Cleanup function
	cleanup := func() {
		grpcServer.Stop()
		conn.Close()
	}

	return client, cleanup
}

func TestClient_GetUserProfile(t *testing.T) {
	client, cleanup := startMockGRPCServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetUserProfile(ctx, &pb.GetUserRequest{Username: "elie222"})
	assert.NoError(t, err)
	assert.Equal(t, "Mocked Name", resp.Name)
	assert.Equal(t, "mockedlogin", resp.Login)
	assert.Equal(t, int32(77), resp.Followers)
	assert.Contains(t, resp.Bio, "mocked bio")
}
