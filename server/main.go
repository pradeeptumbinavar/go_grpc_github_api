package main

import (
	"context"
	"encoding/json"
	"fmt"
	pb "go_grpc_github_api/pb"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)


type server struct {
  pb.UnimplementedUserServiceServer
  token     string
  url 		string
  httpClient *http.Client
}

func NewServer(token string,url string) *server {
  return &server{
    token: token,
	url: url,
    httpClient: &http.Client{Timeout: 10 * time.Second},
  }
}

func (s *server) GetUserProfile(ctx context.Context, req *pb.GetUserRequest) (*pb.UserProfile, error) {
  url := s.url + req.Username
  // Create HTTP request with authorization header
  httpReq, err := http.NewRequest("GET", url, nil)
  if err != nil {
	return nil, fmt.Errorf("failed to create HTTP request: %w", err)
}

  httpReq.Header.Set("Authorization", "Bearer "+s.token)
  httpReq.Header.Set("Accept", "application/vnd.github.v3+json")

  resp, err := s.httpClient.Do(httpReq)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, fmt.Errorf("GitHub API responded %d", resp.StatusCode)
  }

  // parse JSON into anonymous struct
  var g struct {
    Login      string `json:"login"`
    Name       string `json:"name"`
    Email      string `json:"email"`
    Bio        string `json:"bio"`
    Followers  int    `json:"followers"`
    AvatarURL  string `json:"avatar_url"`
  }
  if err := json.NewDecoder(resp.Body).Decode(&g); err != nil {
    return nil, err
  }

  fmt.Printf(
  "GitHub API response:\n  Login: %s\n  Name: %s\n  Email: %s\n  Bio: %s\n  Followers: %d\n  AvatarURL: %s\n",
  g.Login, g.Name, g.Email, g.Bio, g.Followers, g.AvatarURL,
)

  return &pb.UserProfile{
    Login:      g.Login,
    Name:       g.Name,
    Email:      g.Email,
    Bio:        g.Bio,
    Followers:  int32(g.Followers),
    AvatarUrl:  g.AvatarURL,
  }, nil
}

func main() {
  // Load env vars
	if err := godotenv.Load(".env"); err != nil {
		log.Println("no .env file found—using environment variables")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatalln("GITHUB_TOKEN is required (set it in .env or env)")
	}
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	githubAPI := os.Getenv("GITHUB_URL")
	if githubAPI == "" {
		log.Fatalln("GITHUB_URL is required (set it in .env or env)")
	}

	// Listen on TCP
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, NewServer(token, githubAPI))

	// Graceful shutdown setup
	go func() {
		log.Printf("gRPC server listening on :%s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Wait for CTRL+C / SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Println("Server stopped gracefully.")
}
