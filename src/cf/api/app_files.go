package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
)

type AppFilesRepository interface {
	ListFiles(app cf.Application, path string) (files string, apiStatus ApiStatus)
}

type CloudControllerAppFilesRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerAppFilesRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerAppFilesRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerAppFilesRepository) ListFiles(app cf.Application, path string) (files string, apiStatus ApiStatus) {
	url := fmt.Sprintf("%s/v2/apps/%s/instances/0/files/%s", repo.config.Target, app.Guid, path)
	request, apiStatus := newRequest(repo.gateway, "GET", url, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}

	files, _, apiStatus = performRequestForTextResponse(repo.gateway, request)
	return
}
