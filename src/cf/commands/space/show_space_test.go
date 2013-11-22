package space_test

import (
	"cf"
	. "cf/commands/space"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestShowSpaceRequirements(t *testing.T) {
	args := []string{"my-space"}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callShowSpace(t, args, reqFactory)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	callShowSpace(t, args, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestShowSpaceInfoSuccess(t *testing.T) {
	org := cf.OrganizationFields{}
	org.Name = "my-org"

	app_Auto := cf.ApplicationFields{}
	app_Auto.Name = "app1"
	app_Auto.Guid = "app1-guid"
	apps := []cf.ApplicationFields{app_Auto}

	domain_Auto := cf.DomainFields{}
	domain_Auto.Name = "domain1"
	domain_Auto.Guid = "domain1-guid"
	domains := []cf.DomainFields{domain_Auto}

	serviceInstance_Auto := cf.ServiceInstanceFields{}
	serviceInstance_Auto.Name = "service1"
	serviceInstance_Auto.Guid = "service1-guid"
	services := []cf.ServiceInstanceFields{serviceInstance_Auto}

	space := cf.Space{}
	space.Name = "space1"
	space.Organization = org
	space.Applications = apps
	space.Domains = domains
	space.ServiceInstances = services

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Space: space}
	ui := callShowSpace(t, []string{"space1"}, reqFactory)
	assert.Contains(t, ui.Outputs[0], "Getting info for space")
	assert.Contains(t, ui.Outputs[0], "space1")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "space1")
	assert.Contains(t, ui.Outputs[3], "Org")
	assert.Contains(t, ui.Outputs[3], "my-org")
	assert.Contains(t, ui.Outputs[4], "Apps")
	assert.Contains(t, ui.Outputs[4], "app1")
	assert.Contains(t, ui.Outputs[5], "Domains")
	assert.Contains(t, ui.Outputs[5], "domain1")
	assert.Contains(t, ui.Outputs[6], "Services")
	assert.Contains(t, ui.Outputs[6], "service1")
}

func callShowSpace(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("space", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org_Auto := cf.OrganizationFields{}
	org_Auto.Name = "my-org"
	config := &configuration.Configuration{
		AccessToken:  token,
		Organization: org_Auto,
	}

	cmd := NewShowSpace(ui, config)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
