package testhelpers

import (
	"cf"
	"cf/api"
)

type FakeOrgRepository struct {
	Organizations []cf.Organization

	CreateName string
	CreateOrgExists bool

	FindByNameName         string
	FindByNameErr          bool
	FindByNameNotFound     bool
	FindByNameOrganization cf.Organization

	RenameOrganization cf.Organization
	RenameNewName      string

	DeletedOrganization cf.Organization

	FindQuotaByNameName string
	FindQuotaByNameQuota cf.Quota
	FindQuotaByNameNotFound bool
	FindQuotaByNameErr bool

	UpdateQuotaOrg cf.Organization
	UpdateQuotaQuota cf.Quota
}

func (repo FakeOrgRepository) FindAll() (orgs []cf.Organization, apiStatus api.ApiStatus) {
	orgs = repo.Organizations
	return
}

func (repo *FakeOrgRepository) FindByName(name string) (org cf.Organization, apiStatus api.ApiStatus) {
	repo.FindByNameName = name
	org = repo.FindByNameOrganization

	if repo.FindByNameErr {
		apiStatus = api.NewApiStatusWithMessage("Error finding organization by name.")
	}

	if repo.FindByNameNotFound {
		apiStatus = api.NewNotFoundApiStatus("Org", name)
	}

	return
}

func (repo *FakeOrgRepository) Create(name string) (apiStatus api.ApiStatus) {
	if repo.CreateOrgExists {
		apiStatus = api.NewApiStatus("Space already exists", api.ORG_EXISTS, 400)
		return
	}
	repo.CreateName = name
	return
}

func (repo *FakeOrgRepository) Rename(org cf.Organization, newName string) (apiStatus api.ApiStatus) {
	repo.RenameOrganization = org
	repo.RenameNewName = newName
	return
}

func (repo *FakeOrgRepository) Delete(org cf.Organization) (apiStatus api.ApiStatus) {
	repo.DeletedOrganization = org
	return
}

func (repo *FakeOrgRepository) FindQuotaByName(name string) (quota cf.Quota, apiStatus api.ApiStatus) {
	repo.FindQuotaByNameName = name
	quota = repo.FindQuotaByNameQuota

	if repo.FindQuotaByNameNotFound {
		apiStatus = api.NewNotFoundApiStatus("Org", name)
	}
	if repo.FindQuotaByNameErr {
		apiStatus = api.NewApiStatusWithMessage("Error finding quota")
	}

	return
}

func (repo *FakeOrgRepository) UpdateQuota(org cf.Organization, quota cf.Quota) (apiStatus api.ApiStatus) {
	repo.UpdateQuotaOrg = org
	repo.UpdateQuotaQuota = quota
	return
}
