//go:build !demo

package core

import "errors"

func fetchProfile(id string) (*Profile, error) {
	return nil, errors.New("FetchProfile: no production backend wired up yet")
}
