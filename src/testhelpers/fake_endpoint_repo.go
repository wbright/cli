package testhelpers

import (
	"cf/api"
)

type FakeEndpointRepo struct {
	UpdateEndpointEndpoint string
}

func (repo *FakeEndpointRepo) UpdateEndpoint(endpoint string) (apiStatus api.ApiStatus) {
	repo.UpdateEndpointEndpoint = endpoint
	return
}
