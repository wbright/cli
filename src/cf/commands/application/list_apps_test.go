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
	domain_Auto := cf.Domain{}
	domain_Auto.Name = "cfapps.io"
	domain_Auto2 := cf.Domain{}
	domain_Auto2.Name = "example.com"
	app1Routes := []cf.Route{
		{Host: "app1", Domain: domain_Auto},
		{Host: "app1", Domain: domain_Auto2},
	}
	domain_Auto3 := cf.Domain{}
	domain_Auto3.Name = "cfapps.io"
	app2Routes := []cf.Route{{Host: "app2", Domain: domain_Auto3}}
	app_Auto := cf.Application{}
	app_Auto.Name = "Application-1"
	app_Auto.State = "started"
	app_Auto.RunningInstances = 1
	app_Auto.Instances = 1
	app_Auto.Memory = 512
	app_Auto.DiskQuota = 1024
	app_Auto.Routes = app1Routes
	app_Auto2 := cf.Application{}
	app_Auto2.Name = "Application-2"
	app_Auto2.State = "started"
	app_Auto2.RunningInstances = 1
	app_Auto2.Instances = 2
	app_Auto2.Memory = 256
	app_Auto2.DiskQuota = 1024
	app_Auto2.Routes = app2Routes
	apps := []cf.Application{app_Auto, app_Auto2}
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
	space_Auto := cf.Space{}
	space_Auto.Name = "development"
	org_Auto := cf.Organization{}
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
