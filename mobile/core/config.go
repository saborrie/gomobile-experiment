package core

import "github.com/saborrie/go-mobile-experiment/mobile/internal/rpcclient"

// SetEndpoint configures the gRPC server address used by FetchProfile and
// other backend-bound calls. Mobile apps call this once at startup with
// the dev/prod backend URL.
func SetEndpoint(addr string) {
	rpcclient.SetEndpoint(addr)
}
