package api

import (
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"encoding/json"
	"fmt"
	"strings"
)

type ServiceRepository interface {
	GetServiceOfferings() (offerings []cf.ServiceOffering, apiStatus ApiStatus)
	CreateServiceInstance(name string, plan cf.ServicePlan) (identicalAlreadyExists bool, apiStatus ApiStatus)
	CreateUserProvidedServiceInstance(name string, params map[string]string) (apiStatus ApiStatus)
	FindInstanceByName(name string) (instance cf.ServiceInstance, apiStatus ApiStatus)
	BindService(instance cf.ServiceInstance, app cf.Application) (apiStatus ApiStatus)
	UnbindService(instance cf.ServiceInstance, app cf.Application) (found bool, apiStatus ApiStatus)
	DeleteService(instance cf.ServiceInstance) (apiStatus ApiStatus)
	RenameService(instance cf.ServiceInstance, newName string) (apiStatus ApiStatus)
}

type CloudControllerServiceRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerServiceRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerServiceRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerServiceRepository) GetServiceOfferings() (offerings []cf.ServiceOffering, apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/services?inline-relations-depth=1", repo.config.Target)
	request, apiStatus := newRequest(repo.gateway, "GET", path, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}

	response := new(ServiceOfferingsApiResponse)

	_, apiStatus = performRequestForJSONResponse(repo.gateway, request, response)
	if apiStatus.NotSuccessful() {
		return
	}

	for _, r := range response.Resources {
		plans := []cf.ServicePlan{}
		for _, p := range r.Entity.ServicePlans {
			plans = append(plans, cf.ServicePlan{Name: p.Entity.Name, Guid: p.Metadata.Guid})
		}
		offerings = append(offerings, cf.ServiceOffering{
			Label:       r.Entity.Label,
			Version:     r.Entity.Version,
			Provider:    r.Entity.Provider,
			Description: r.Entity.Description,
			Guid:        r.Metadata.Guid,
			Plans:       plans,
		})
	}

	return
}

func (repo CloudControllerServiceRepository) CreateServiceInstance(name string, plan cf.ServicePlan) (identicalAlreadyExists bool, apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/service_instances", repo.config.Target)

	data := fmt.Sprintf(
		`{"name":"%s","service_plan_guid":"%s","space_guid":"%s"}`,
		name, plan.Guid, repo.config.Space.Guid,
	)
	request, apiStatus := newRequest(repo.gateway, "POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = performRequest(repo.gateway, request)

	if apiStatus.NotSuccessful() && apiStatus.ErrorCode == SERVICE_INSTANCE_NAME_TAKEN {

		serviceInstance, findInstanceApiStatus := repo.FindInstanceByName(name)

		if !findInstanceApiStatus.NotSuccessful() &&
			serviceInstance.ServicePlan.Guid == plan.Guid {
			apiStatus = ApiStatus{}
			identicalAlreadyExists = true
			return
		}
	}

	return
}

func (repo CloudControllerServiceRepository) CreateUserProvidedServiceInstance(name string, params map[string]string) (apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/user_provided_service_instances", repo.config.Target)

	type RequestBody struct {
		Name        string            `json:"name"`
		Credentials map[string]string `json:"credentials"`
		SpaceGuid   string            `json:"space_guid"`
	}

	reqBody := RequestBody{name, params, repo.config.Space.Guid}
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		apiStatus = NewApiStatusWithError("Error parsing response", err)
		return
	}

	request, apiStatus := newRequest(repo.gateway, "POST", path, repo.config.AccessToken, bytes.NewReader(jsonBytes))
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = performRequest(repo.gateway, request)
	return
}

func (repo CloudControllerServiceRepository) FindInstanceByName(name string) (instance cf.ServiceInstance, apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/spaces/%s/service_instances?return_user_provided_service_instances=true&q=name%s&inline-relations-depth=2", repo.config.Target, repo.config.Space.Guid, "%3A"+name)
	request, apiStatus := newRequest(repo.gateway, "GET", path, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}

	response := new(ServiceInstancesApiResponse)
	_, apiStatus = performRequestForJSONResponse(repo.gateway, request, response)
	if apiStatus.NotSuccessful() {
		return
	}

	if len(response.Resources) == 0 {
		apiStatus = NewNotFoundApiStatus("Service instance", name)
		return
	}

	resource := response.Resources[0]
	serviceOfferingEntity := resource.Entity.ServicePlan.Entity.ServiceOffering.Entity
	instance.Guid = resource.Metadata.Guid
	instance.Name = resource.Entity.Name

	instance.ServiceOffering.Label = serviceOfferingEntity.Label
	instance.ServiceOffering.DocumentationUrl = serviceOfferingEntity.DocumentationUrl
	instance.ServiceOffering.Description = serviceOfferingEntity.Description

	instance.ServicePlan = cf.ServicePlan{
		Name: resource.Entity.ServicePlan.Entity.Name,
		Guid: resource.Entity.ServicePlan.Metadata.Guid,
	}
	instance.ServiceBindings = []cf.ServiceBinding{}

	for _, bindingResource := range resource.Entity.ServiceBindings {
		newBinding := cf.ServiceBinding{
			Url:     bindingResource.Metadata.Url,
			Guid:    bindingResource.Metadata.Guid,
			AppGuid: bindingResource.Entity.AppGuid,
		}
		instance.ServiceBindings = append(instance.ServiceBindings, newBinding)
	}

	return
}

func (repo CloudControllerServiceRepository) BindService(instance cf.ServiceInstance, app cf.Application) (apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/service_bindings", repo.config.Target)
	body := fmt.Sprintf(
		`{"app_guid":"%s","service_instance_guid":"%s"}`,
		app.Guid, instance.Guid,
	)
	request, apiStatus := newRequest(repo.gateway, "POST", path, repo.config.AccessToken, strings.NewReader(body))
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = performRequest(repo.gateway, request)
	return
}

func (repo CloudControllerServiceRepository) UnbindService(instance cf.ServiceInstance, app cf.Application) (found bool, apiStatus ApiStatus) {
	var path string

	for _, binding := range instance.ServiceBindings {
		if binding.AppGuid == app.Guid {
			path = repo.config.Target + binding.Url
			break
		}
	}

	if path == "" {
		return
	} else {
		found = true
	}

	request, apiStatus := newRequest(repo.gateway, "DELETE", path, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = performRequest(repo.gateway, request)
	return
}

func (repo CloudControllerServiceRepository) DeleteService(instance cf.ServiceInstance) (apiStatus ApiStatus) {
	if len(instance.ServiceBindings) > 0 {
		return NewApiStatusWithMessage("Cannot delete service instance, apps are still bound to it")
	}

	path := fmt.Sprintf("%s/v2/service_instances/%s", repo.config.Target, instance.Guid)
	request, apiStatus := newRequest(repo.gateway, "DELETE", path, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = performRequest(repo.gateway, request)
	return
}

func (repo CloudControllerServiceRepository) RenameService(instance cf.ServiceInstance, newName string) (apiStatus ApiStatus) {
	body := fmt.Sprintf(`{"name":"%s"}`, newName)
	path := fmt.Sprintf("%s/v2/service_instances/%s", repo.config.Target, instance.Guid)
	request, apiStatus := newRequest(repo.gateway, "PUT", path, repo.config.AccessToken, strings.NewReader(body))
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = performRequest(repo.gateway, request)
	return
}
