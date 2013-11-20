package route_test

import (
	"cf"
	. "cf/commands/route"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestCreateRouteRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	routeRepo := &testapi.FakeRouteRepository{}

	callCreateRoute(t, []string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}

	callCreateRoute(t, []string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestCreateRouteFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	routeRepo := &testapi.FakeRouteRepository{}

	ui := callCreateRoute(t, []string{""}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateRoute(t, []string{"my-space"}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateRoute(t, []string{"my-space", "example.com", "host"}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateRoute(t, []string{"my-space", "example.com", "-n", "host"}, reqFactory, routeRepo)
	assert.False(t, ui.FailedWithUsage)

	ui = callCreateRoute(t, []string{"my-space", "example.com"}, reqFactory, routeRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestCreateRoute(t *testing.T) {
	space := cf.Space{}
	space.Guid = "my-space-guid"
	space.Name = "my-space"
	domain := cf.Domain{}
	domain.Guid = "domain-guid"
	domain.Name = "example.com"
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess: true,
		Domain:       domain,
		Space:        space,
	}
	routeRepo := &testapi.FakeRouteRepository{}

	ui := callCreateRoute(t, []string{"-n", "host", "my-space", "example.com"}, reqFactory, routeRepo)

	assert.Contains(t, ui.Outputs[0], "Creating route")
	assert.Contains(t, ui.Outputs[0], "host.example.com")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")
	route_Auto := cf.Route{}
	route_Auto.Host = "host"
	route_Auto.Domain = domain
	assert.Equal(t, routeRepo.CreateInSpaceRoute, route_Auto)
	assert.Equal(t, routeRepo.CreateInSpaceDomain, domain)
	assert.Equal(t, routeRepo.CreateInSpaceSpace, space)

}

func TestCreateRouteIsIdempotent(t *testing.T) {
	space := cf.Space{}
	space.Guid = "my-space-guid"
	space.Name = "my-space"
	domain := cf.Domain{}
	domain.Guid = "domain-guid"
	domain.Name = "example.com"
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess: true,
		Domain:       domain,
		Space:        space,
	}
	route_Auto2 := cf.Route{}
	route_Auto2.Guid = "my-route-guid"
	route_Auto2.Host = "host"
	route_Auto2.Domain = domain
	route_Auto2.Space = space
	routeRepo := &testapi.FakeRouteRepository{
		CreateInSpaceErr:         true,
		FindByHostAndDomainRoute: route_Auto2,
	}

	ui := callCreateRoute(t, []string{"-n", "host", "my-space", "example.com"}, reqFactory, routeRepo)

	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "host.example.com")
	assert.Contains(t, ui.Outputs[2], "already exists")
	route_Auto := cf.Route{}
	route_Auto.Host = "host"
	route_Auto.Domain = domain
	assert.Equal(t, routeRepo.CreateInSpaceRoute, route_Auto)
	assert.Equal(t, routeRepo.CreateInSpaceDomain, domain)
	assert.Equal(t, routeRepo.CreateInSpaceSpace, space)

}

func TestRouteCreator(t *testing.T) {
	space := cf.Space{}
	space.Guid = "my-space-guid"
	space.Name = "my-space"
	domain := cf.Domain{}
	domain.Guid = "domain-guid"
	domain.Name = "example.com"
	domain_Auto5 := cf.Domain{}
	domain_Auto5.Name = "example.com"
	route_Auto3 := cf.Route{}
	route_Auto3.Host = "my-host"
	route_Auto3.Guid = "my-route-guid"
	route_Auto3.Domain = domain_Auto5
	routeRepo := &testapi.FakeRouteRepository{
		CreateInSpaceCreatedRoute: route_Auto3,
	}

	ui := new(testterm.FakeUI)
	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org_Auto := cf.Organization{}
	org_Auto.Name = "my-org"
	config := &configuration.Configuration{
		Organization: org_Auto,
		AccessToken:  token,
	}

	cmd := NewCreateRoute(ui, config, routeRepo)
	route, apiResponse := cmd.CreateRoute("my-host", domain, space)

	assert.True(t, apiResponse.IsSuccessful())
	assert.Contains(t, ui.Outputs[0], "Creating route")
	assert.Contains(t, ui.Outputs[0], "my-host.example.com")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, routeRepo.CreateInSpaceRoute.Host, "my-host")
	assert.Equal(t, routeRepo.CreateInSpaceDomain, domain)
	assert.Equal(t, routeRepo.CreateInSpaceSpace, space)
	assert.Equal(t, routeRepo.CreateInSpaceCreatedRoute, route)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func callCreateRoute(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-route", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org_Auto2 := cf.Organization{}
	org_Auto2.Name = "my-org"
	space_Auto5 := cf.Space{}
	space_Auto5.Name = "my-space"
	config := &configuration.Configuration{
		Space:        space_Auto5,
		Organization: org_Auto2,
		AccessToken:  token,
	}

	cmd := NewCreateRoute(fakeUI, config, routeRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
