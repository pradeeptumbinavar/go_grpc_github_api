
package main

import (
  "context"
  "flag"
  "log"
  "time"

  pb "go_grpc_github_api/pb"
  "google.golang.org/grpc"
)

func main() {
  addr := flag.String("addr", "localhost:50051", "gRPC server address")
  user := flag.String("user", "pradeeptumbinavar", "GitHub username")
  flag.Parse()

  conn, err := grpc.Dial(*addr, grpc.WithInsecure(), grpc.WithBlock())
  if err != nil {
    log.Fatalf("did not connect: %v", err)
  }
  defer conn.Close()

  client := pb.NewUserServiceClient(conn)
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()

  resp, err := client.GetUserProfile(ctx, &pb.GetUserRequest{Username: *user})
  if err != nil {
    log.Fatalf("could not get profile: %v", err)
  }

  log.Printf("User: %s , Email: %s, Followers: %d",
    resp.Name, resp.Email, resp.Followers)
}
