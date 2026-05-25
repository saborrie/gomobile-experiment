// Package integration holds end-to-end tests that wire mobile/core through
// to the real gRPC server. These aren't usually run in tight inner loops
// (use the per-package tests in mobile/internal/* for that) — they catch
// regressions where the FFI surface, the rpc client, and the server stop
// agreeing on the contract.
package integration_test

import (
	"net"
	"testing"

	"google.golang.org/grpc"

	api "github.com/saborrie/go-mobile-experiment/api/gen/go"
	"github.com/saborrie/go-mobile-experiment/mobile/core"
	"github.com/saborrie/go-mobile-experiment/server/svc"
)

// startTestServer spins up the real ProfileServer on a random port and
// returns its address + a stop function. No test plumbing on the server
// side — same code path as `task server`.
func startTestServer(t *testing.T) (addr string, stop func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	s := grpc.NewServer()
	api.RegisterProfileServer(s, svc.NewProfileServer())
	go func() { _ = s.Serve(lis) }()
	return lis.Addr().String(), func() { s.Stop() }
}

func TestFetchProfile_EndToEnd(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	core.SetEndpoint(addr)

	p, err := core.FetchProfile("alice")
	if err != nil {
		t.Fatalf("FetchProfile: %v", err)
	}
	if p.Id != "alice" {
		t.Errorf("Id = %q, want %q", p.Id, "alice")
	}
	if p.Name != "Server user alice" {
		t.Errorf("Name = %q, want %q", p.Name, "Server user alice")
	}
}
