// Package svc holds the gRPC service implementations registered by main.
package svc

import (
	"context"
	"fmt"

	api "github.com/saborrie/go-mobile-experiment/api/gen/go"
)

type ProfileServer struct {
	api.UnimplementedProfileServer
}

func NewProfileServer() *ProfileServer {
	return &ProfileServer{}
}

func (s *ProfileServer) Fetch(ctx context.Context, req *api.FetchRequest) (*api.FetchResponse, error) {
	// Stub data for now — real lookup arrives when we add storage.
	return &api.FetchResponse{
		Id:   req.Id,
		Name: fmt.Sprintf("Server user %s", req.Id),
	}, nil
}
