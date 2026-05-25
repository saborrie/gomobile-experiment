// Package rpcclient is the mobile core's gRPC client to the backend.
// Owns the connection lifecycle and a configurable endpoint.
package rpcclient

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	api "github.com/saborrie/go-mobile-experiment/api/gen/go"
)

var (
	mu       sync.Mutex
	endpoint = "localhost:7777"
	conn     *grpc.ClientConn
)

// SetEndpoint changes the backend address. Closes the existing connection
// (if any) so the next call uses the new endpoint.
func SetEndpoint(addr string) {
	mu.Lock()
	defer mu.Unlock()
	endpoint = addr
	if conn != nil {
		_ = conn.Close()
		conn = nil
	}
}

func dial() (api.ProfileClient, error) {
	mu.Lock()
	defer mu.Unlock()
	if conn == nil {
		c, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		conn = c
	}
	return api.NewProfileClient(conn), nil
}

// FetchProfile calls Profile.Fetch with a short timeout.
func FetchProfile(id string) (*api.FetchResponse, error) {
	c, err := dial()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.Fetch(ctx, &api.FetchRequest{Id: id})
}
