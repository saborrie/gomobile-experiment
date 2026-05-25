// server is the gRPC backend for the mobile apps.
package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	api "github.com/saborrie/go-mobile-experiment/api/gen/go"
	"github.com/saborrie/go-mobile-experiment/server/svc"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "0.0.0.0:7777"
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	s := grpc.NewServer()
	api.RegisterProfileServer(s, svc.NewProfileServer())

	log.Printf("gRPC server listening on %s", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
