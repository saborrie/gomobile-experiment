//go:build !demo

package profile

import "github.com/saborrie/go-mobile-experiment/mobile/internal/rpcclient"

func fetch(id string) (*Profile, error) {
	resp, err := rpcclient.FetchProfile(id)
	if err != nil {
		return nil, err
	}
	return &Profile{Id: resp.Id, Name: resp.Name}, nil
}
