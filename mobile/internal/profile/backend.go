//go:build !demo

package profile

import "errors"

func fetch(id string) (*Profile, error) {
	return nil, errors.New("FetchProfile: no production backend wired up yet")
}
