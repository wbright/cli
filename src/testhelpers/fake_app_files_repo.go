package testhelpers

import (
	"cf"
	"cf/api"
)

type FakeAppFilesRepo struct{
	Application cf.Application
	Path string
	FileList string
}


func (repo *FakeAppFilesRepo)ListFiles(app cf.Application, path string) (files string, apiStatus api.ApiStatus) {
	repo.Application = app
	repo.Path = path

	files = repo.FileList

	return
}
