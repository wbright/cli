package testhelpers

import (
	"cf"
	"cf/api"
	"net/http"
)

type FakeServiceRepo struct {
	ServiceOfferings []cf.ServiceOffering

	CreateServiceInstanceName string
	CreateServiceInstancePlan cf.ServicePlan
	CreateServiceAlreadyExists bool

	CreateUserProvidedServiceInstanceName string
	CreateUserProvidedServiceInstanceParameters map[string]string

	FindInstanceByNameName string
	FindInstanceByNameServiceInstance cf.ServiceInstance
	FindInstanceByNameErr bool
	FindInstanceByNameNotFound bool

	BindServiceServiceInstance cf.ServiceInstance
	BindServiceApplication cf.Application
	BindServiceErrorCode string

	UnbindServiceServiceInstance cf.ServiceInstance
	UnbindServiceApplication cf.Application
	UnbindServiceBindingNotFound bool

	DeleteServiceServiceInstance cf.ServiceInstance

	RenameServiceServiceInstance cf.ServiceInstance
	RenameServiceNewName string
}

func (repo *FakeServiceRepo) GetServiceOfferings() (offerings []cf.ServiceOffering, apiStatus api.ApiStatus) {
	offerings = repo.ServiceOfferings
	return
}

func (repo *FakeServiceRepo) CreateServiceInstance(name string, plan cf.ServicePlan) (identicalAlreadyExists bool, apiStatus api.ApiStatus) {
	repo.CreateServiceInstanceName = name
	repo.CreateServiceInstancePlan = plan
	identicalAlreadyExists = repo.CreateServiceAlreadyExists

	return
}

func (repo *FakeServiceRepo) CreateUserProvidedServiceInstance(name string, params map[string]string) (apiStatus api.ApiStatus) {
	repo.CreateUserProvidedServiceInstanceName = name
	repo.CreateUserProvidedServiceInstanceParameters = params
	return
}

func (repo *FakeServiceRepo) FindInstanceByName(name string) (instance cf.ServiceInstance, apiStatus api.ApiStatus) {
	repo.FindInstanceByNameName = name
	instance = repo.FindInstanceByNameServiceInstance

	if repo.FindInstanceByNameErr {
		apiStatus = api.NewApiStatusWithMessage("Error finding instance")
	}

	if repo.FindInstanceByNameNotFound {
		apiStatus = api.NewNotFoundApiStatus("Service instance", name)
	}

	return
}

func (repo *FakeServiceRepo) BindService(instance cf.ServiceInstance, app cf.Application) (apiStatus api.ApiStatus) {
	repo.BindServiceServiceInstance = instance
	repo.BindServiceApplication = app

	if repo.BindServiceErrorCode != "" {
		apiStatus = api.NewApiStatus("Error binding service", repo.BindServiceErrorCode, http.StatusBadRequest)
	}

	return
}

func (repo *FakeServiceRepo) UnbindService(instance cf.ServiceInstance, app cf.Application) (found bool, apiStatus api.ApiStatus) {
	repo.UnbindServiceServiceInstance = instance
	repo.UnbindServiceApplication = app
	found = !repo.UnbindServiceBindingNotFound
	return
}

func (repo *FakeServiceRepo) DeleteService(instance cf.ServiceInstance) (apiStatus api.ApiStatus) {
	repo.DeleteServiceServiceInstance = instance
	return
}

func (repo *FakeServiceRepo) RenameService(instance cf.ServiceInstance, newName string) (apiStatus api.ApiStatus){
	repo.RenameServiceServiceInstance = instance
	repo.RenameServiceNewName = newName
	return
}
