package api

import (
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ApplicationRepository interface {
	FindByName(name string) (app cf.Application, apiStatus ApiStatus)
	SetEnv(app cf.Application, envVars map[string]string) (apiStatus ApiStatus)
	Create(newApp cf.Application) (createdApp cf.Application, apiStatus ApiStatus)
	Delete(app cf.Application) (apiStatus ApiStatus)
	Rename(app cf.Application, newName string) (apiStatus ApiStatus)
	Scale(app cf.Application) (apiStatus ApiStatus)
	Start(app cf.Application) (updatedApp cf.Application, apiStatus ApiStatus)
	Stop(app cf.Application) (updatedApp cf.Application, apiStatus ApiStatus)
	GetInstances(app cf.Application) (instances []cf.ApplicationInstance, apiStatus ApiStatus)
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

func (repo CloudControllerApplicationRepository) FindByName(name string) (app cf.Application, apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=name%s&inline-relations-depth=1", repo.config.Target, repo.config.Space.Guid, "%3A"+name)
	request, apiStatus := newRequest(repo.gateway, "GET", path, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}

	findResponse := new(ApplicationsApiResponse)
	_, apiStatus = performRequestForJSONResponse(repo.gateway, request, findResponse)
	if apiStatus.NotSuccessful() {
		return
	}

	if len(findResponse.Resources) == 0 {
		apiStatus = NewNotFoundApiStatus("App", name)
		return
	}

	res := findResponse.Resources[0]
	path = fmt.Sprintf("%s/v2/apps/%s/summary", repo.config.Target, res.Metadata.Guid)
	request, apiStatus = newRequest(repo.gateway, "GET", path, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}

	summaryResponse := new(ApplicationSummary)
	_, apiStatus = performRequestForJSONResponse(repo.gateway, request, summaryResponse)
	if apiStatus.NotSuccessful() {
		return
	}

	urls := []string{}
	// This is a little wonky but we made a concious effort
	// to keep the domain very separate from the API repsonses
	// to maintain flexibility.
	domainRoute := cf.Route{}
	for _, route := range summaryResponse.Routes {
		domainRoute.Domain = cf.Domain{Name: route.Domain.Name}
		domainRoute.Host = route.Host
		urls = append(urls, domainRoute.URL())
	}

	app = cf.Application{
		Name:             summaryResponse.Name,
		Guid:             summaryResponse.Guid,
		Instances:        summaryResponse.Instances,
		RunningInstances: summaryResponse.RunningInstances,
		Memory:           summaryResponse.Memory,
		EnvironmentVars:  res.Entity.EnvironmentJson,
		Urls:             urls,
		State:            strings.ToLower(summaryResponse.State),
	}

	return
}

func (repo CloudControllerApplicationRepository) SetEnv(app cf.Application, envVars map[string]string) (apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)

	type setEnvReqBody struct {
		EnvJson map[string]string `json:"environment_json"`
	}

	body := setEnvReqBody{EnvJson: envVars}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		apiStatus = NewApiStatusWithError("Error creating json", err)
		return
	}

	request, apiStatus := newRequest(repo.gateway, "PUT", path, repo.config.AccessToken, bytes.NewReader(jsonBytes))
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = performRequest(repo.gateway, request)
	return
}

func (repo CloudControllerApplicationRepository) Create(newApp cf.Application) (createdApp cf.Application, apiStatus ApiStatus) {
	apiStatus = validateApplication(newApp)
	if apiStatus.NotSuccessful() {
		return
	}

	buildpackUrl := stringOrNull(newApp.BuildpackUrl)
	stackGuid := stringOrNull(newApp.Stack.Guid)
	command := stringOrNull(newApp.Command)

	path := fmt.Sprintf("%s/v2/apps", repo.config.Target)
	data := fmt.Sprintf(
		`{"space_guid":"%s","name":"%s","instances":%d,"buildpack":%s,"command":null,"memory":%d,"stack_guid":%s,"command":%s}`,
		repo.config.Space.Guid, newApp.Name, newApp.Instances, buildpackUrl, newApp.Memory, stackGuid, command,
	)
	request, apiStatus := newRequest(repo.gateway, "POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiStatus.NotSuccessful() {
		return
	}

	resource := new(Resource)
	_, apiStatus = performRequestForJSONResponse(repo.gateway, request, resource)
	if apiStatus.NotSuccessful() {
		return
	}

	createdApp.Guid = resource.Metadata.Guid
	createdApp.Name = resource.Entity.Name
	return
}

func stringOrNull(s string) string {
	if s == "" {
		return "null"
	}

	return fmt.Sprintf(`"%s"`, s)
}

func (repo CloudControllerApplicationRepository) Delete(app cf.Application) (apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", repo.config.Target, app.Guid)
	request, apiStatus := newRequest(repo.gateway, "DELETE", path, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = performRequest(repo.gateway, request)
	return
}

func (repo CloudControllerApplicationRepository) Rename(app cf.Application, newName string) (apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)
	data := fmt.Sprintf(`{"name":"%s"}`, newName)
	request, apiStatus := newRequest(repo.gateway, "PUT", path, repo.config.AccessToken, strings.NewReader(data))
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = performRequest(repo.gateway, request)
	return
}

func (repo CloudControllerApplicationRepository) Scale(app cf.Application) (apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)

	values := map[string]interface{}{}
	if app.DiskQuota > 0 {
		values["disk_quota"] = app.DiskQuota
	}
	if app.Instances > 0 {
		values["instances"] = app.Instances
	}
	if app.Memory > 0 {
		values["memory"] = app.Memory
	}

	bodyBytes, err := json.Marshal(values)
	if err != nil {
		return NewApiStatusWithError("Error generating body", err)
	}

	request, apiStatus := newRequest(repo.gateway, "PUT", path, repo.config.AccessToken, bytes.NewReader(bodyBytes))
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = performRequest(repo.gateway, request)
	return
}

func (repo CloudControllerApplicationRepository) Start(app cf.Application) (updatedApp cf.Application, apiStatus ApiStatus) {
	updates := map[string]interface{}{"state": "STARTED"}
	if app.BuildpackUrl != "" {
		updates["buildpack"] = app.BuildpackUrl
	}
	return repo.updateApplication(app, updates)
}

func (repo CloudControllerApplicationRepository) Stop(app cf.Application) (updatedApp cf.Application, apiStatus ApiStatus) {
	return repo.updateApplication(app, map[string]interface{}{"state": "STOPPED"})
}

type InstancesApiResponse map[string]InstanceApiResponse

type InstanceApiResponse struct {
	State string
	Since float64
}

func (repo CloudControllerApplicationRepository) GetInstances(app cf.Application) (instances []cf.ApplicationInstance, apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/apps/%s/instances", repo.config.Target, app.Guid)
	request, apiStatus := newRequest(repo.gateway, "GET", path, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}

	apiResponse := InstancesApiResponse{}

	_, apiStatus = performRequestForJSONResponse(repo.gateway, request, &apiResponse)
	if apiStatus.NotSuccessful() {
		return
	}

	instances = make([]cf.ApplicationInstance, len(apiResponse), len(apiResponse))
	for k, v := range apiResponse {
		index, err := strconv.Atoi(k)
		if err != nil {
			continue
		}

		instances[index] = cf.ApplicationInstance{
			State: cf.InstanceState(strings.ToLower(v.State)),
			Since: time.Unix(int64(v.Since), 0),
		}
	}
	return
}

func (repo CloudControllerApplicationRepository) updateApplication(app cf.Application, updates map[string]interface{}) (updatedApp cf.Application, apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)

	updates["console"] = true

	body, err := json.Marshal(updates)
	if err != nil {
		apiStatus = NewApiStatusWithError("Could not serialize app updates.", err)
		return
	}

	request, apiStatus := newRequest(repo.gateway, "PUT", path, repo.config.AccessToken, bytes.NewReader(body))

	if apiStatus.NotSuccessful() {
		return
	}

	response := ApplicationResource{}
	_, apiStatus = performRequestForJSONResponse(repo.gateway, request, &response)

	updatedApp = cf.Application{
		Name:  response.Entity.Name,
		Guid:  response.Metadata.Guid,
		State: strings.ToLower(response.Entity.State),
	}

	return
}

func validateApplication(app cf.Application) (apiStatus ApiStatus) {
	reg := regexp.MustCompile("^[0-9a-zA-Z\\-_]*$")
	if !reg.MatchString(app.Name) {
		apiStatus = NewApiStatusWithMessage("Application name is invalid. Name can only contain letters, numbers, underscores and hyphens.")
	}

	return
}
