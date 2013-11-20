package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type ApplicationSummaries struct {
	Apps []ApplicationFromSummary
}

func (resource ApplicationSummaries) toModels() (apps []cf.ApplicationFields) {
	for _, appSummary := range resource.Apps {
		apps = append(apps, appSummary.toModel())
	}
	return
}

type ApplicationFromSummary struct {
	Guid             string
	Name             string
	Routes           []RouteSummary
	RunningInstances int `json:"running_instances"`
	Memory           uint64
	Instances        int
	DiskQuota        uint64 `json:"disk_quota"`
	Urls             []string
	State            string
}

func (resource ApplicationFromSummary) toModel() (app cf.ApplicationFields) {
	app = cf.ApplicationFields{
		State:            strings.ToLower(resource.State),
		Instances:        resource.Instances,
		DiskQuota:        resource.DiskQuota,
		RunningInstances: resource.RunningInstances,
		Memory:           resource.Memory,
	}
	app.Name = resource.Name
	app.Guid = resource.Guid
	return
}

type RouteSummary struct {
	Guid   string
	Host   string
	Domain DomainSummary
}

func (resource RouteSummary) toModel() (route cf.RouteSummary) {
	domain := cf.DomainFields{}
	domain.Guid = resource.Domain.Guid
	domain.Name = resource.Domain.Name
	domain.Shared = resource.Domain.OwningOrganizationGuid != ""

	route.Guid = resource.Guid
	route.Host = resource.Host
	route.Domain = domain
	return
}

type DomainSummary struct {
	Guid                   string
	Name                   string
	OwningOrganizationGuid string
}

type AppSummaryRepository interface {
	GetSummariesInCurrentSpace() (apps []cf.AppSummary, apiResponse net.ApiResponse)
	GetSummary(appGuid string) (summary cf.AppSummary, apiResponse net.ApiResponse)
}

type CloudControllerAppSummaryRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
	appRepo ApplicationRepository
}

func NewCloudControllerAppSummaryRepository(config *configuration.Configuration, gateway net.Gateway, appRepo ApplicationRepository) (repo CloudControllerAppSummaryRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.appRepo = appRepo
	return
}

func (repo CloudControllerAppSummaryRepository) GetSummariesInCurrentSpace() (apps []cf.AppSummary, apiResponse net.ApiResponse) {
	resources := new(ApplicationSummaries)

	path := fmt.Sprintf("%s/v2/spaces/%s/summary", repo.config.Target, repo.config.Space.Guid)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, resource := range resources.Apps {
		var app cf.AppSummary
		app, apiResponse = repo.createSummary(&resource)
		if apiResponse.IsNotSuccessful() {
			return
		}
		apps = append(apps, app)
	}
	return
}

func (repo CloudControllerAppSummaryRepository) GetSummary(appGuid string) (summary cf.AppSummary, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s/summary", repo.config.Target, appGuid)
	summaryResponse := new(ApplicationFromSummary)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, summaryResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	return repo.createSummary(summaryResponse)
}

func (repo CloudControllerAppSummaryRepository) createSummary(resource *ApplicationFromSummary) (summary cf.AppSummary, apiResponse net.ApiResponse) {
	summary = cf.AppSummary{
		App: resource.toModel(),
	}

	instances, apiResponse := repo.appRepo.GetInstances(summary.App.Guid)
	if apiResponse.IsNotSuccessful() {
		return
	}
	summary.Instances = instances

	return
}
