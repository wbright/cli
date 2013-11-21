package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestApps(t *testing.T) {
	domain_Auto := cf.DomainFields{}
	domain_Auto.Name = "cfapps.io"
	domain_Auto2 := cf.DomainFields{}
	domain_Auto2.Name = "example.com"

	route1 := cf.RouteSummary{}
	route1.Host = "app1"
	route1.Domain = domain_Auto

	route2 := cf.RouteSummary{}
	route2.Host = "app1"
	route2.Domain = domain_Auto2

	app1Routes := []cf.RouteSummary{route1, route2}

	domain_Auto3 := cf.DomainFields{}
	domain_Auto3.Name = "cfapps.io"

	route3 := cf.RouteSummary{}
	route3.Host = "app2"
	route3.Domain = domain_Auto3

	app2Routes := []cf.RouteSummary{route3}

	app_Auto := cf.AppSummary{}
	app_Auto.Name = "Application-1"
	app_Auto.State = "started"
	app_Auto.RunningInstances = 1
	app_Auto.InstanceCount = 1
	app_Auto.Memory = 512
	app_Auto.DiskQuota = 1024
	app_Auto.RouteSummary = app1Routes

	app_Auto2 := cf.AppSummary{}
	app_Auto2.Name = "Application-2"
	app_Auto2.State = "started"
	app_Auto2.RunningInstances = 1
	app_Auto2.InstanceCount = 2
	app_Auto2.Memory = 256
	app_Auto2.DiskQuota = 1024
	app_Auto2.RouteSummary = app2Routes

	apps := []cf.AppSummary{app_Auto, app_Auto2}

	appSummaryRepo := &testapi.FakeAppSummaryRepo{
		GetSummariesInCurrentSpaceApps: apps,
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

	ui := callApps(t, appSummaryRepo, reqFactory)

	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Contains(t, ui.Outputs[0], "Getting apps in")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "development")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[4], "Application-1")
	assert.Contains(t, ui.Outputs[4], "started")
	assert.Contains(t, ui.Outputs[4], "1/1")
	assert.Contains(t, ui.Outputs[4], "512M")
	assert.Contains(t, ui.Outputs[4], "1G")
	assert.Contains(t, ui.Outputs[4], "app1.cfapps.io, app1.example.com")

	assert.Contains(t, ui.Outputs[5], "Application-2")
	assert.Contains(t, ui.Outputs[5], "started")
	assert.Contains(t, ui.Outputs[5], "1/2")
	assert.Contains(t, ui.Outputs[5], "256M")
	assert.Contains(t, ui.Outputs[5], "1G")
	assert.Contains(t, ui.Outputs[5], "app2.cfapps.io")
}

func TestAppsRequiresLogin(t *testing.T) {
	appSummaryRepo := &testapi.FakeAppSummaryRepo{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}

	callApps(t, appSummaryRepo, reqFactory)

	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestAppsRequiresASelectedSpaceAndOrg(t *testing.T) {
	appSummaryRepo := &testapi.FakeAppSummaryRepo{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}

	callApps(t, appSummaryRepo, reqFactory)

	assert.False(t, testcmd.CommandDidPassRequirements)
}

func callApps(t *testing.T, appSummaryRepo *testapi.FakeAppSummaryRepo, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	space_Auto := cf.SpaceFields{}
	space_Auto.Name = "development"
	org_Auto := cf.OrganizationFields{}
	org_Auto.Name = "my-org"
	config := &configuration.Configuration{
		Space:        space_Auto,
		Organization: org_Auto,
		AccessToken:  token,
	}

	ctxt := testcmd.NewContext("apps", []string{})
	cmd := NewListApps(ui, config, appSummaryRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
