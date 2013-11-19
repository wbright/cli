package api

import (
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	app.Instances = resource.Entity.Instances
	app.Memory = uint64(resource.Entity.Memory)

	return
}

func (resource ApplicationResource) ToModel() (app cf.Application) {
	app.Fields = resource.ToFields()

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
	route.Fields = resource.ToFields()
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
	SetEnv(appGuid string, envVars map[string]string) (apiResponse net.ApiResponse)
	Create(name, buildpackUrl, stackGuid, command string, memory uint64, instances int) (createdApp cf.Application, apiResponse net.ApiResponse)
	Delete(appGuid string) (apiResponse net.ApiResponse)
	Rename(appGuid string, newName string) (apiResponse net.ApiResponse)
	Scale(app cf.ApplicationFields) (apiResponse net.ApiResponse)
	Start(appGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse)
	StartWithDifferentBuildpack(appGuid, buildpack string) (updatedApp cf.Application, apiResponse net.ApiResponse)
	Stop(appGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse)
	GetInstances(appGuid string) (instances []cf.ApplicationInstance, apiResponse net.ApiResponse)
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
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=name%s&inline-relations-depth=1", repo.config.Target, repo.config.Space.Guid, "%3A"+name)
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
func (repo CloudControllerApplicationRepository) SetEnv(appGuid string, envVars map[string]string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, appGuid)

	type setEnvReqBody struct {
		EnvJson map[string]string `json:"environment_json"`
	}

	body := setEnvReqBody{EnvJson: envVars}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Error creating json", err)
		return
	}

	apiResponse = repo.gateway.UpdateResource(path, repo.config.AccessToken, bytes.NewReader(jsonBytes))
	return
}

func (repo CloudControllerApplicationRepository) Create(name, buildpackUrl, stackGuid, command string, memory uint64, instances int) (createdApp cf.Application, apiResponse net.ApiResponse) {
	apiResponse = validateApplicationName(name)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/v2/apps", repo.config.Target)
	data := fmt.Sprintf(
		`{"space_guid":"%s","name":"%s","instances":%d,"buildpack":%s,"memory":%d,"stack_guid":%s,"command":%s}`,
		repo.config.Space.Guid,
		name,
		instances,
		stringOrNull(buildpackUrl),
		memory,
		stringOrNull(stackGuid),
		stringOrNull(command),
	)

	resource := new(ApplicationResource)
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(data), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	createdApp = resource.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) Delete(appGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", repo.config.Target, appGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}

func (repo CloudControllerApplicationRepository) Rename(appGuid, newName string) (apiResponse net.ApiResponse) {
	apiResponse = validateApplicationName(newName)
	if apiResponse.IsNotSuccessful() {
		return
	}

	data := fmt.Sprintf(`{"name":"%s"}`, newName)
	apiResponse = repo.updateApp(appGuid, strings.NewReader(data))
	return
}

func (repo CloudControllerApplicationRepository) Scale(app cf.ApplicationFields) (apiResponse net.ApiResponse) {
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
		return net.NewApiResponseWithError("Error generating body", err)
	}

	apiResponse = repo.updateApp(app.Guid, bytes.NewReader(bodyBytes))
	return
}

func (repo CloudControllerApplicationRepository) updateApp(appGuid string, body io.ReadSeeker) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, appGuid)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, body)
}

func validateApplicationName(name string) (apiResponse net.ApiResponse) {
	reg := regexp.MustCompile("^[0-9a-zA-Z\\-_]*$")
	if !reg.MatchString(name) {
		apiResponse = net.NewApiResponseWithMessage("App name is invalid: name can only contain letters, numbers, underscores and hyphens")
	}

	return
}

func (repo CloudControllerApplicationRepository) Start(appGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	return repo.startOrStopApp(appGuid, map[string]interface{}{"state": "STARTED"})
}

func (repo CloudControllerApplicationRepository) StartWithDifferentBuildpack(appGuid, buildpack string) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	updates := map[string]interface{}{
		"state":     "STARTED",
		"buildpack": buildpack,
	}
	return repo.startOrStopApp(appGuid, updates)
}

func (repo CloudControllerApplicationRepository) Stop(appGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	return repo.startOrStopApp(appGuid, map[string]interface{}{"state": "STOPPED"})
}

func (repo CloudControllerApplicationRepository) startOrStopApp(appGuid string, updates map[string]interface{}) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s?inline-relations-depth=2", repo.config.Target, appGuid)

	updates["console"] = true

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

type InstancesApiResponse map[string]InstanceApiResponse

type InstanceApiResponse struct {
	State string
	Since float64
}

func (repo CloudControllerApplicationRepository) GetInstances(appGuid string) (instances []cf.ApplicationInstance, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s/instances", repo.config.Target, appGuid)
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	instancesResponse := InstancesApiResponse{}

	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, &instancesResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	instances = make([]cf.ApplicationInstance, len(instancesResponse), len(instancesResponse))
	for k, v := range instancesResponse {
		index, err := strconv.Atoi(k)
		if err != nil {
			continue
		}

		instances[index] = cf.ApplicationInstance{
			State: cf.InstanceState(strings.ToLower(v.State)),
			Since: time.Unix(int64(v.Since), 0),
		}
	}

	return repo.updateInstancesWithStats(appGuid, instances)
}

type StatsApiResponse map[string]InstanceStatsApiResponse

type InstanceStatsApiResponse struct {
	Stats struct {
		DiskQuota uint64 `json:"disk_quota"`
		MemQuota  uint64 `json:"mem_quota"`
		Usage     struct {
			Cpu  float64
			Disk uint64
			Mem  uint64
		}
	}
}

func (repo CloudControllerApplicationRepository) updateInstancesWithStats(guid string, instances []cf.ApplicationInstance) (updatedInst []cf.ApplicationInstance, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s/stats", repo.config.Target, guid)
	statsResponse := StatsApiResponse{}
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, &statsResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	updatedInst = make([]cf.ApplicationInstance, len(statsResponse), len(statsResponse))
	for k, v := range statsResponse {
		index, err := strconv.Atoi(k)
		if err != nil {
			continue
		}

		instance := instances[index]
		instance.CpuUsage = v.Stats.Usage.Cpu
		instance.DiskQuota = v.Stats.DiskQuota
		instance.DiskUsage = v.Stats.Usage.Disk
		instance.MemQuota = v.Stats.MemQuota
		instance.MemUsage = v.Stats.Usage.Mem

		updatedInst[index] = instance
	}
	return
}
