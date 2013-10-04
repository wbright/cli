package api

import (
	"cf/configuration"
	"cf/net"
)

type EndpointRepository interface {
	UpdateEndpoint(endpoint string) (apiStatus ApiStatus)
}

type RemoteEndpointRepository struct {
	config     *configuration.Configuration
	gateway    net.Gateway
	configRepo configuration.ConfigurationRepository
}

func NewEndpointRepository(config *configuration.Configuration, gateway net.Gateway, configRepo configuration.ConfigurationRepository) (repo RemoteEndpointRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.configRepo = configRepo
	return
}

func (repo RemoteEndpointRepository) UpdateEndpoint(endpoint string) (apiStatus ApiStatus) {
	request, apiStatus := newRequest(repo.gateway, "GET", endpoint+"/v2/info", "", nil)
	if apiStatus.NotSuccessful() {
		return
	}

	scheme := request.URL.Scheme
	if scheme != "http" && scheme != "https" {
		apiStatus = NewApiStatusWithMessage("API Endpoints should start with https:// or http://")
		return
	}

	type infoResponse struct {
		ApiVersion            string `json:"api_version"`
		AuthorizationEndpoint string `json:"authorization_endpoint"`
	}

	serverResponse := new(infoResponse)
	_, apiStatus = performRequestForJSONResponse(repo.gateway, request, &serverResponse)
	if apiStatus.NotSuccessful() {
		return
	}

	repo.configRepo.ClearSession()
	repo.config.Target = endpoint
	repo.config.ApiVersion = serverResponse.ApiVersion
	repo.config.AuthorizationEndpoint = serverResponse.AuthorizationEndpoint

	err := repo.configRepo.Save()
	if err != nil {
		apiStatus = NewApiStatusWithMessage(err.Error())
	}

	return
}
