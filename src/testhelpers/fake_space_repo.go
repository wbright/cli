package testhelpers

import (
	"cf"
	"cf/api"
)

type FakeSpaceRepository struct {
	CurrentSpace cf.Space

	Spaces []cf.Space

	FindByNameName string
	FindByNameSpace cf.Space
	FindByNameErr bool
	FindByNameNotFound bool

	SummarySpace cf.Space

	CreateSpaceName string
	CreateSpaceExists bool

	RenameSpace cf.Space
	RenameNewName string

	DeletedSpace cf.Space
}

func (repo FakeSpaceRepository) GetCurrentSpace() (space cf.Space) {
	return repo.CurrentSpace
}

func (repo FakeSpaceRepository) FindAll() (spaces []cf.Space, apiStatus api.ApiStatus) {
	spaces = repo.Spaces
	return
}

func (repo *FakeSpaceRepository) FindByName(name string) (space cf.Space, apiStatus api.ApiStatus) {
	repo.FindByNameName = name
	space = repo.FindByNameSpace

	if repo.FindByNameErr {
		apiStatus = api.NewApiStatusWithMessage("Error finding space by name.")
	}

	if repo.FindByNameNotFound {
		apiStatus = api.NewNotFoundApiStatus("Space", name)
	}

	return
}

func (repo *FakeSpaceRepository) GetSummary() (space cf.Space, apiStatus api.ApiStatus) {
	space = repo.SummarySpace
	return
}

func (repo *FakeSpaceRepository) Create(name string) (apiStatus api.ApiStatus) {
	if repo.CreateSpaceExists {
		apiStatus = api.NewApiStatus("Space already exists", api.SPACE_EXISTS, 400)
		return
	}
	repo.CreateSpaceName = name
	return
}

func (repo *FakeSpaceRepository) Rename(space cf.Space, newName string) (apiStatus api.ApiStatus) {
	repo.RenameSpace = space
	repo.RenameNewName = newName
	return
}

func (repo *FakeSpaceRepository) Delete(space cf.Space) (apiStatus api.ApiStatus) {
	repo.DeletedSpace = space
	return
}
