package api

import (
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type PaginatedApplicationResources struct {
	Resources []ApplicationResource
}

type ApplicationResource struct {
	Resource
	Entity ApplicationEntity
}

func (resource ApplicationResource) ToFields() (app cf.ApplicationFields) {
	app.Guid = resource.Metadata.Guid
	app.Name = resource.Entity.Name
	app.EnvironmentVars = resource.Entity.EnvironmentJson
	app.State = strings.ToLower(resource.Entity.State)
	app.InstanceCount = resource.Entity.Instances
	app.Memory = uint64(resource.Entity.Memory)

	return
}

func (resource ApplicationResource) ToModel() (app cf.Application) {
	app.ApplicationFields = resource.ToFields()

	for _, routeResource := range resource.Entity.Routes {
		app.Routes = append(app.Routes, routeResource.ToModel())
	}
	return
}

type ApplicationEntity struct {
	Name            string
	State           string
	Instances       int
	Memory          int
	Routes          []AppRouteResource
	EnvironmentJson map[string]string `json:"environment_json"`
}

type AppRouteResource struct {
	Resource
	Entity AppRouteEntity
}

func (resource AppRouteResource) ToFields() (route cf.RouteFields) {
	route.Guid = resource.Metadata.Guid
	route.Host = resource.Entity.Host
	return
}

func (resource AppRouteResource) ToModel() (route cf.RouteSummary) {
	route.RouteFields = resource.ToFields()
	route.Domain.Guid = resource.Entity.Domain.Metadata.Guid
	route.Domain.Name = resource.Entity.Domain.Entity.Name
	return
}

type AppRouteEntity struct {
	Host   string
	Domain Resource
}

type ApplicationRepository interface {
	FindByName(name string) (app cf.Application, apiResponse net.ApiResponse)
	Create(name, buildpackUrl, stackGuid, command string, memory uint64, instances int) (createdApp cf.Application, apiResponse net.ApiResponse)
	Update(appGuid string, params AppParams) (updatedApp cf.Application, apiResponse net.ApiResponse)
	Delete(appGuid string) (apiResponse net.ApiResponse)
}

type CloudControllerApplicationRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerApplicationRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerApplicationRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerApplicationRepository) FindByName(name string) (app cf.Application, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=name%s&inline-relations-depth=1", repo.config.Target, repo.config.SpaceFields.Guid, "%3A" + name)
	appResources := new(PaginatedApplicationResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, appResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(appResources.Resources) == 0 {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "App", name)
		return
	}

	res := appResources.Resources[0]
	app = res.ToModel()
	return
}

//func (repo CloudControllerApplicationRepository) SetEnv(appGuid string, envVars map[string]string) (apiResponse net.ApiResponse) {
//	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, appGuid)
//
//	type setEnvReqBody struct {
//		EnvJson map[string]string `json:"environment_json"`
//	}
//
//	body := setEnvReqBody{EnvJson: envVars}
//
//	jsonBytes, err := json.Marshal(body)
//	if err != nil {
//		apiResponse = net.NewApiResponseWithError("Error creating json", err)
//		return
//	}
//
//	apiResponse = repo.gateway.UpdateResource(path, repo.config.AccessToken, bytes.NewReader(jsonBytes))
//	return
//}

func (repo CloudControllerApplicationRepository) Create(name, buildpackUrl, stackGuid, command string, memory uint64, instances int) (createdApp cf.Application, apiResponse net.ApiResponse) {
	data, apiResponse := repo.formatAppJSON(name, buildpackUrl, stackGuid, repo.config.SpaceFields.Guid, command, memory, instances)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/v2/apps", repo.config.Target)
	resource := new(ApplicationResource)
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(data), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	createdApp = resource.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) Update(appGuid string, params AppParams) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	repo.put()
	return
}

func (repo CloudControllerApplicationRepository) put(appGuid string, params map[string]interface{}) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	data, apiResponse := repo.formatAppJSON(params)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, appGuid)
	resource := new(ApplicationResource)
	apiResponse = repo.gateway.UpdateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(data), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	updatedApp = resource.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) formatAppJSON(params map[string]string) (data string, apiResponse net.ApiResponse) {
	delete(params,"guid")
	name, ok := params["name"]
	if ok {
		reg := regexp.MustCompile("^[0-9a-zA-Z\\-_]*$")
		if !reg.MatchString(name) {
			apiResponse = net.NewApiResponseWithMessage("App name is invalid: name can only contain letters, numbers, underscores and hyphens")
			return
		}
	}

	data, err := json.Marshal(params)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("error creating json:\n%s", err)
		return
	}

	return
}

func (repo CloudControllerApplicationRepository) Delete(appGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", repo.config.Target, appGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}

func (repo CloudControllerApplicationRepository) startOrStopApp(appGuid string, updates map[string]interface{}) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s?inline-relations-depth=1", repo.config.Target, appGuid)

	body, err := json.Marshal(updates)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Could not serialize app updates.", err)
		return
	}

	resource := new(ApplicationResource)
	apiResponse = repo.gateway.UpdateResourceForResponse(path, repo.config.AccessToken, bytes.NewReader(body), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	updatedApp = resource.ToModel()
	return
}
