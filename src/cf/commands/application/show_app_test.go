package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"cf/formatters"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
	"time"
)

func TestAppRequirements(t *testing.T) {
	args := []string{"my-app", "/foo"}
	appSummaryRepo := &testapi.FakeAppSummaryRepo{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true, Application: cf.Application{}}
	callApp(t, args, reqFactory, appSummaryRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false, Application: cf.Application{}}
	callApp(t, args, reqFactory, appSummaryRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	callApp(t, args, reqFactory, appSummaryRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestAppFailsWithUsage(t *testing.T) {
	appSummaryRepo := &testapi.FakeAppSummaryRepo{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	ui := callApp(t, []string{}, reqFactory, appSummaryRepo)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestDisplayingAppSummary(t *testing.T) {
	reqApp := cf.Application{}
	reqApp.Name = "my-app"
	domain_Auto := cf.Domain{}
	domain_Auto.Name = "example.com"
	domain_Auto2 := cf.Domain{}
	domain_Auto2.Name = "example.com"
	routes := []cf.Route{
		{Host: "my-app", Domain: domain_Auto},
		{Host: "foo", Domain: domain_Auto2},
	}
	app := cf.Application{}
	app.State = "started"
	app.Instances = 2
	app.RunningInstances = 2
	app.Memory = 256
	app.Routes = routes

	time1, err := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2012")
	assert.NoError(t, err)

	time2, err := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Apr 1 15:04:05 -0700 MST 2012")
	assert.NoError(t, err)
	appInstance_Auto := cf.ApplicationInstance{}
	appInstance_Auto.State = cf.InstanceRunning
	appInstance_Auto.Since = time1
	appInstance_Auto.CpuUsage = 1.0
	appInstance_Auto.DiskQuota = 1 * formatters.GIGABYTE
	appInstance_Auto.DiskUsage = 32 * formatters.MEGABYTE
	appInstance_Auto.MemQuota = 64 * formatters.MEGABYTE
	appInstance_Auto.MemUsage = 13 * formatters.BYTE
	appInstance_Auto2 := cf.ApplicationInstance{}
	appInstance_Auto2.State = cf.InstanceDown
	appInstance_Auto2.Since = time2
	instances := []cf.ApplicationInstance{appInstance_Auto, appInstance_Auto2}
	appSummary := cf.AppSummary{}
	appSummary.App = app
	appSummary.Instances = instances

	appSummaryRepo := &testapi.FakeAppSummaryRepo{GetSummarySummary: appSummary}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: reqApp}
	ui := callApp(t, []string{"my-app"}, reqFactory, appSummaryRepo)

	assert.Equal(t, appSummaryRepo.GetSummaryApp.Name, "my-app")

	assert.Contains(t, ui.Outputs[0], "Showing health and status")
	assert.Contains(t, ui.Outputs[0], "my-app")

	assert.Contains(t, ui.Outputs[2], "state")
	assert.Contains(t, ui.Outputs[2], "started")

	assert.Contains(t, ui.Outputs[3], "instances")
	assert.Contains(t, ui.Outputs[3], "2/2")

	assert.Contains(t, ui.Outputs[4], "usage")
	assert.Contains(t, ui.Outputs[4], "256M x 2 instances")

	assert.Contains(t, ui.Outputs[5], "urls")
	assert.Contains(t, ui.Outputs[5], "my-app.example.com, foo.example.com")

	assert.Contains(t, ui.Outputs[7], "#0")
	assert.Contains(t, ui.Outputs[7], "running")
	assert.Contains(t, ui.Outputs[7], "2012-01-02 03:04:05 PM")
	assert.Contains(t, ui.Outputs[7], "1.0%")
	assert.Contains(t, ui.Outputs[7], "13 of 64M")
	assert.Contains(t, ui.Outputs[7], "32M of 1G")

	assert.Contains(t, ui.Outputs[8], "#1")
	assert.Contains(t, ui.Outputs[8], "down")
	assert.Contains(t, ui.Outputs[8], "2012-04-01 03:04:05 PM")
	assert.Contains(t, ui.Outputs[8], "0%")
	assert.Contains(t, ui.Outputs[8], "0 of 0")
	assert.Contains(t, ui.Outputs[8], "0 of 0")
}

func TestDisplayingStoppedAppSummary(t *testing.T) {
	testDisplayingAppSummaryWithErrorCode(t, cf.APP_STOPPED)
}

func TestDisplayingNotStagedAppSummary(t *testing.T) {
	testDisplayingAppSummaryWithErrorCode(t, cf.APP_NOT_STAGED)
}

func testDisplayingAppSummaryWithErrorCode(t *testing.T, errorCode string) {
	reqApp := cf.Application{}
	reqApp.Name = "my-app"
	domain_Auto3 := cf.Domain{}
	domain_Auto3.Name = "example.com"
	domain_Auto4 := cf.Domain{}
	domain_Auto4.Name = "example.com"
	routes := []cf.Route{
		{Host: "my-app", Domain: domain_Auto3},
		{Host: "foo", Domain: domain_Auto4},
	}
	app := cf.Application{}
	app.State = "stopped"
	app.Instances = 2
	app.RunningInstances = 0
	app.Memory = 256
	app.Routes = routes
	appSummary := cf.AppSummary{}
	appSummary.App = app

	appSummaryRepo := &testapi.FakeAppSummaryRepo{GetSummarySummary: appSummary, GetSummaryErrorCode: errorCode}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: reqApp}
	ui := callApp(t, []string{"my-app"}, reqFactory, appSummaryRepo)

	assert.Equal(t, appSummaryRepo.GetSummaryApp.Name, "my-app")
	assert.Equal(t, len(ui.Outputs), 6)

	assert.Contains(t, ui.Outputs[0], "Showing health and status")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Contains(t, ui.Outputs[2], "state")
	assert.Contains(t, ui.Outputs[2], "stopped")

	assert.Contains(t, ui.Outputs[3], "instances")
	assert.Contains(t, ui.Outputs[3], "0/2")

	assert.Contains(t, ui.Outputs[4], "usage")
	assert.Contains(t, ui.Outputs[4], "256M x 2 instances")

	assert.Contains(t, ui.Outputs[5], "urls")
	assert.Contains(t, ui.Outputs[5], "my-app.example.com, foo.example.com")
}

func callApp(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, appSummaryRepo *testapi.FakeAppSummaryRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("app", args)

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

	cmd := NewShowApp(ui, config, appSummaryRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
