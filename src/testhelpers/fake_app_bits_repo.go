package testhelpers

import (
	"cf"
	"cf/api"
)

type FakeApplicationBitsRepository struct {
	UploadedApp cf.Application
	UploadedDir string
}

func (repo *FakeApplicationBitsRepository) UploadApp(app cf.Application, dir string) (apiStatus api.ApiStatus) {
	repo.UploadedDir = dir
	repo.UploadedApp = app

	return
}
