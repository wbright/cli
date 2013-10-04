package testhelpers

import (
	"cf"
	"cf/api"
)

type FakeDomainRepository struct {
	FindAllInCurrentSpaceDomains []cf.Domain

	FindAllByOrgOrg cf.Organization
	FindAllByOrgDomains []cf.Domain

	FindByNameInOrgDomain cf.Domain
	FindByNameInOrgApiStatus api.ApiStatus

	FindByNameName string
	FindByNameDomain cf.Domain
	FindByNameNotFound bool
	FindByNameErr bool

	ReserveDomainDomainToCreate cf.Domain
	ReserveDomainOwningOrg cf.Organization

	MapDomainDomain cf.Domain
	MapDomainSpace cf.Space
	MapDomainApiStatus api.ApiStatus

	UnmapDomainDomain cf.Domain
	UnmapDomainSpace cf.Space
	UnmapDomainApiStatus api.ApiStatus
}

func (repo *FakeDomainRepository) FindAllInCurrentSpace() (domains []cf.Domain, apiStatus api.ApiStatus){
	domains = repo.FindAllInCurrentSpaceDomains
	return
}

func (repo *FakeDomainRepository) FindAllByOrg(org cf.Organization)(domains []cf.Domain, apiStatus api.ApiStatus){
	repo.FindAllByOrgOrg = org
	domains = repo.FindAllByOrgDomains

	return
}

func (repo *FakeDomainRepository) FindByNameInCurrentSpace(name string) (domain cf.Domain, apiStatus api.ApiStatus){
	repo.FindByNameName = name
	domain = repo.FindByNameDomain

	if repo.FindByNameNotFound {
		apiStatus = api.NewNotFoundApiStatus("Domain", name)
	}
	if repo.FindByNameErr {
		apiStatus = api.NewApiStatusWithMessage("Error finding domain")
	}

	return
}

func (repo *FakeDomainRepository) Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiStatus api.ApiStatus){
	repo.ReserveDomainDomainToCreate = domainToCreate
	repo.ReserveDomainOwningOrg = owningOrg
	return
}

func (repo *FakeDomainRepository) MapDomain(domain cf.Domain, space cf.Space) (apiStatus api.ApiStatus) {
	repo.MapDomainDomain = domain
	repo.MapDomainSpace = space
	apiStatus = repo.MapDomainApiStatus
	return
}

func (repo *FakeDomainRepository) UnmapDomain(domain cf.Domain, space cf.Space) (apiStatus api.ApiStatus) {
	repo.UnmapDomainDomain = domain
	repo.UnmapDomainSpace = space
	apiStatus = repo.UnmapDomainApiStatus
	return
}

func (repo *FakeDomainRepository) FindByNameInOrg(name string, owningOrg cf.Organization) (domain cf.Domain, apiStatus api.ApiStatus) {

	domain = repo.FindByNameInOrgDomain
	apiStatus = repo.FindByNameInOrgApiStatus
	return
}


