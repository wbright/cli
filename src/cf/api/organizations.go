package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type OrganizationRepository interface {
	FindAll() (orgs []cf.Organization, apiStatus ApiStatus)
	FindByName(name string) (org cf.Organization, apiStatus ApiStatus)
	Create(name string) (apiStatus ApiStatus)
	Rename(org cf.Organization, name string) (apiStatus ApiStatus)
	Delete(org cf.Organization) (apiStatus ApiStatus)
	FindQuotaByName(name string) (quota cf.Quota, apiStatus ApiStatus)
	UpdateQuota(org cf.Organization, quota cf.Quota) (apiStatus ApiStatus)
}

type CloudControllerOrganizationRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerOrganizationRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerOrganizationRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerOrganizationRepository) FindAll() (orgs []cf.Organization, apiStatus ApiStatus) {
	path := repo.config.Target + "/v2/organizations"
	request, apiStatus := newRequest(repo.gateway, "GET", path, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}
	response := new(OrganizationsApiResponse)

	_, apiStatus = performRequestForJSONResponse(repo.gateway, request, response)

	if apiStatus.NotSuccessful() {
		return
	}

	for _, r := range response.Resources {
		orgs = append(orgs, cf.Organization{
			Name: r.Entity.Name,
			Guid: r.Metadata.Guid,
		},
		)
	}

	return
}

func (repo CloudControllerOrganizationRepository) FindByName(name string) (org cf.Organization, apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/organizations?q=name%s&inline-relations-depth=1", repo.config.Target, "%3A"+strings.ToLower(name))
	request, apiStatus := newRequest(repo.gateway, "GET", path, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}
	response := new(OrganizationsApiResponse)

	_, apiStatus = performRequestForJSONResponse(repo.gateway, request, response)

	if apiStatus.NotSuccessful() {
		return
	}

	if len(response.Resources) == 0 {
		apiStatus = NewNotFoundApiStatus("Org", name)
		return
	}

	r := response.Resources[0]
	spaces := []cf.Space{}

	for _, s := range r.Entity.Spaces {
		spaces = append(spaces, cf.Space{Name: s.Entity.Name, Guid: s.Metadata.Guid})
	}

	domains := []cf.Domain{}

	for _, d := range r.Entity.Domains {
		domains = append(domains, cf.Domain{Name: d.Entity.Name, Guid: d.Metadata.Guid})
	}

	org = cf.Organization{
		Name:    r.Entity.Name,
		Guid:    r.Metadata.Guid,
		Spaces:  spaces,
		Domains: domains,
	}

	return
}

func (repo CloudControllerOrganizationRepository) Create(name string) (apiStatus ApiStatus) {
	path := repo.config.Target + "/v2/organizations"
	data := fmt.Sprintf(
		`{"name":"%s"}`, name,
	)
	request, apiStatus := newRequest(repo.gateway, "POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = performRequest(repo.gateway, request)
	return
}

func (repo CloudControllerOrganizationRepository) Rename(org cf.Organization, name string) (apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/organizations/%s", repo.config.Target, org.Guid)
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	request, apiStatus := newRequest(repo.gateway, "PUT", path, repo.config.AccessToken, strings.NewReader(data))
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = performRequest(repo.gateway, request)
	return
}

func (repo CloudControllerOrganizationRepository) Delete(org cf.Organization) (apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/organizations/%s?recursive=true", repo.config.Target, org.Guid)
	request, apiStatus := newRequest(repo.gateway, "DELETE", path, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = performRequest(repo.gateway, request)
	return
}

func (repo CloudControllerOrganizationRepository) FindQuotaByName(name string) (quota cf.Quota, apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/quota_definitions?q=name%%3A%s", repo.config.Target, name)

	request, apiStatus := newRequest(repo.gateway, "GET", path, repo.config.AccessToken, nil)
	if apiStatus.NotSuccessful() {
		return
	}

	response := new(ApiResponse)

	_, apiStatus = performRequestForJSONResponse(repo.gateway, request, response)
	if apiStatus.NotSuccessful() {
		return
	}

	if len(response.Resources) == 0 {
		apiStatus = NewNotFoundApiStatus("Org", name)
		return
	}

	res := response.Resources[0]
	quota.Guid = res.Metadata.Guid
	quota.Name = res.Entity.Name

	return
}

func (repo CloudControllerOrganizationRepository) UpdateQuota(org cf.Organization, quota cf.Quota) (apiStatus ApiStatus) {
	path := fmt.Sprintf("%s/v2/organizations/%s", repo.config.Target, org.Guid)
	data := fmt.Sprintf(`{"quota_definition_guid":"%s"}`, quota.Guid)
	request, apiStatus := newRequest(repo.gateway, "PUT", path, repo.config.AccessToken, strings.NewReader(data))
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = performRequest(repo.gateway, request)
	return
}
