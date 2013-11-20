package api

import (
	"cf"
	"cf/net"
	"net/http"
	"time"
)

type FakeApplicationRepository struct {

	ScaledApp cf.ApplicationFields

	StartAppGuid string
	StartAppErr     bool
	StartUpdatedApp cf.Application

	StopAppGuid  string
	StopAppErr     bool
	StopUpdatedApp cf.Application

	DeletedAppGuid string

	FindAllApps []cf.Application

	FindByNameName      string
	FindByNameApp       cf.Application
	FindByNameErr       bool
	FindByNameAuthErr   bool
	FindByNameNotFound  bool

	SetEnvAppGuid   string
	SetEnvVars  map[string]string
	SetEnvValue string
	SetEnvErr   bool

	CreatedApp  cf.Application

	RenameAppGuid     string
	RenameNewName string

	GetInstancesResponses  [][]cf.ApplicationInstance
	GetInstancesErrorCodes []string
}

func (repo *FakeApplicationRepository) FindByName(name string) (app cf.Application, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	app = repo.FindByNameApp

	if repo.FindByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding app by name.")
	}
	if repo.FindByNameAuthErr {
		apiResponse = net.NewApiResponse("Authentication failed.", "1000", 401)
	}
	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found","App", name)
	}

	return
}

func (repo *FakeApplicationRepository) SetEnv(appGuid string, envVars map[string]string) (apiResponse net.ApiResponse) {
	repo.SetEnvAppGuid = appGuid
	repo.SetEnvVars = envVars

	if repo.SetEnvErr {
		apiResponse = net.NewApiResponseWithMessage("Failed setting env")
	}
	return
}

func (repo *FakeApplicationRepository) Create(name, buildpackUrl, stackGuid, command string, memory uint64, instances int) (resultApp cf.Application, apiResponse net.ApiResponse) {
	resultApp = repo.CreatedApp
	resultApp.Guid = resultApp.Name + "-guid"
	return
}

func (repo *FakeApplicationRepository) Delete(appGuid string) (apiResponse net.ApiResponse) {
	repo.DeletedAppGuid = appGuid
	return
}

func (repo *FakeApplicationRepository) Rename(appGuid, newName string) (apiResponse net.ApiResponse) {
	repo.RenameAppGuid = appGuid
	repo.RenameNewName = newName
	return
}

func (repo *FakeApplicationRepository) Scale(app cf.ApplicationFields) (apiResponse net.ApiResponse) {
	repo.ScaledApp = app
	return
}

func (repo *FakeApplicationRepository) Start(appGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	repo.StartAppGuid = appGuid
	if repo.StartAppErr {
		apiResponse = net.NewApiResponseWithMessage("Error starting application")
	}
	updatedApp = repo.StartUpdatedApp
	return
}

func (repo *FakeApplicationRepository) Stop(appGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	repo.StopAppGuid = appGuid
	if repo.StopAppErr {
		apiResponse = net.NewApiResponseWithMessage("Error stopping application")
	}
	updatedApp = repo.StopUpdatedApp
	return
}

func (repo *FakeApplicationRepository) GetInstances(appGuid string) (instances[]cf.ApplicationInstance, apiResponse net.ApiResponse) {
	time.Sleep(1*time.Millisecond) //needed for Windows only, otherwise it thinks error codes are not assigned
	errorCode := repo.GetInstancesErrorCodes[0]
	repo.GetInstancesErrorCodes = repo.GetInstancesErrorCodes[1:]

	instances = repo.GetInstancesResponses[0]
	repo.GetInstancesResponses = repo.GetInstancesResponses[1:]

	if errorCode != "" {
		apiResponse = net.NewApiResponse("Error staging app", errorCode, http.StatusBadRequest)
		return
	}

	return
}
