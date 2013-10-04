package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
)

type StackRepository interface {
	FindByName(name string) (stack cf.Stack, apiStatus ApiStatus)
	FindAll() (stacks []cf.Stack, apiStatus ApiStatus)
}

type CloudControllerStackRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerStackRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerStackRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerStackRepository) FindByName(name string) (stack cf.Stack, apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/stacks?q=name%s", repo.config.Target, "%3A"+name)
	request, apiStatus := newRequest(repo.gateway, "GET", path, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}

	findResponse := new(ApiResponse)
	_, apiStatus = performRequestForJSONResponse(repo.gateway, request, findResponse)
	if apiStatus.NotSuccessful() {
		return
	}

	if len(findResponse.Resources) == 0 {
		apiStatus = NewApiStatusWithMessage("Stack %s not found", name)
		return
	}

	res := findResponse.Resources[0]
	stack.Guid = res.Metadata.Guid
	stack.Name = res.Entity.Name

	return
}

func (repo CloudControllerStackRepository) FindAll() (stacks []cf.Stack, apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/stacks", repo.config.Target)
	request, apiStatus := newRequest(repo.gateway, "GET", path, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}

	listResponse := new(StackApiResponse)
	_, apiStatus = performRequestForJSONResponse(repo.gateway, request, listResponse)
	if apiStatus.NotSuccessful() {
		return
	}

	for _, r := range listResponse.Resources {
		stacks = append(stacks, cf.Stack{Guid: r.Metadata.Guid, Name: r.Entity.Name, Description: r.Entity.Description})
	}

	return
}
