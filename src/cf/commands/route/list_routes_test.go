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

func TestListingRoutes(t *testing.T) {
	domain_Auto := cf.Domain{}
	domain_Auto.Name = "example.com"
	domain_Auto2 := cf.Domain{}
	domain_Auto2.Name = "cfapps.com"
	domain_Auto3 := cf.Domain{}
	domain_Auto3.Name = "another-example.com"
	route_Auto := cf.Route{}
	route_Auto.Host = "hostname-1"
	route_Auto.Domain = domain_Auto
	route_Auto.AppNames = []string{"dora", "dora2"}
	route_Auto2 := cf.Route{}
	route_Auto2.Host = "hostname-2"
	route_Auto2.Domain = domain_Auto2
	route_Auto2.AppNames = []string{"my-app", "my-app2"}
	route_Auto3 := cf.Route{}
	route_Auto3.Host = "hostname-3"
	route_Auto3.Domain = domain_Auto3
	route_Auto3.AppNames = []string{"july", "june"}
	routes := []cf.Route{route_Auto, route_Auto2, route_Auto3}

	routeRepo := &testapi.FakeRouteRepository{Routes: routes}

	ui := callListRoutes(t, []string{}, &testreq.FakeReqFactory{}, routeRepo)

	assert.Contains(t, ui.Outputs[0], "Getting routes")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Contains(t, ui.Outputs[1], "host")
	assert.Contains(t, ui.Outputs[1], "domain")
	assert.Contains(t, ui.Outputs[1], "apps")

	assert.Contains(t, ui.Outputs[2], "hostname-1")
	assert.Contains(t, ui.Outputs[2], "example.com")
	assert.Contains(t, ui.Outputs[2], "dora, dora2")

	assert.Contains(t, ui.Outputs[3], "hostname-2")
	assert.Contains(t, ui.Outputs[3], "cfapps.com")
	assert.Contains(t, ui.Outputs[3], "my-app, my-app2")

	assert.Contains(t, ui.Outputs[4], "hostname-3")
	assert.Contains(t, ui.Outputs[4], "another-example.com")
	assert.Contains(t, ui.Outputs[4], "july, june")
}

func TestListingRoutesWhenNoneExist(t *testing.T) {
	routes := []cf.Route{}
	routeRepo := &testapi.FakeRouteRepository{Routes: routes}

	ui := callListRoutes(t, []string{}, &testreq.FakeReqFactory{}, routeRepo)

	assert.Contains(t, ui.Outputs[0], "Getting routes")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "No routes found")
}

func TestListingRoutesWhenFindFails(t *testing.T) {
	routeRepo := &testapi.FakeRouteRepository{ListErr: true}

	ui := callListRoutes(t, []string{}, &testreq.FakeReqFactory{}, routeRepo)

	assert.Contains(t, ui.Outputs[0], "Getting routes")
	assert.Contains(t, ui.Outputs[1], "FAILED")
}

func callListRoutes(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (ui *testterm.FakeUI) {

	ui = &testterm.FakeUI{}

	ctxt := testcmd.NewContext("list-routes", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	space_Auto := cf.Space{}
	space_Auto.Name = "my-space"
	org_Auto := cf.Organization{}
	org_Auto.Name = "my-org"
	config := &configuration.Configuration{
		Space:        space_Auto,
		Organization: org_Auto,
		AccessToken:  token,
	}

	cmd := NewListRoutes(ui, config, routeRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
