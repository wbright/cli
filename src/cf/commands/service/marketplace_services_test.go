package service_test

import (
	"cf"
	. "cf/commands/service"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestMarketplaceServices(t *testing.T) {
	plan_Auto := cf.ServicePlan{}
	plan_Auto.Name = "service-plan-a"
	plan_Auto2 := cf.ServicePlan{}
	plan_Auto2.Name = "service-plan-b"
	plan_Auto3 := cf.ServicePlan{}
	plan_Auto3.Name = "service-plan-c"
	plan_Auto4 := cf.ServicePlan{}
	plan_Auto4.Name = "service-plan-d"
	offering_Auto := cf.ServiceOffering{}
	offering_Auto.Label = "my-service-offering-1"
	offering_Auto.Description = "service offering 1 description"
	offering_Auto.Plans = []cf.ServicePlan{plan_Auto, plan_Auto2}
	offering_Auto2 := cf.ServiceOffering{}
	offering_Auto2.Label = "my-service-offering-2"
	offering_Auto2.Description = "service offering 2 description"
	offering_Auto2.Plans = []cf.ServicePlan{plan_Auto3, plan_Auto4}
	serviceOfferings := []cf.ServiceOffering{offering_Auto, offering_Auto2}
	serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org_Auto := cf.Organization{}
	org_Auto.Name = "my-org"
	org_Auto.Guid = "my-org-guid"
	space_Auto := cf.Space{}
	space_Auto.Name = "my-space"
	space_Auto.Guid = "my-space-guid"
	config := &configuration.Configuration{
		Space:        space_Auto,
		Organization: org_Auto,
		AccessToken:  token,
	}

	ui := callMarketplaceServices(t, config, serviceRepo)

	assert.Contains(t, ui.Outputs[0], "Getting services from marketplace in org")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[4], "my-service-offering-1")
	assert.Contains(t, ui.Outputs[4], "service offering 1 description")
	assert.Contains(t, ui.Outputs[4], "service-plan-a, service-plan-b")

	assert.Contains(t, ui.Outputs[5], "my-service-offering-2")
	assert.Contains(t, ui.Outputs[5], "service offering 2 description")
	assert.Contains(t, ui.Outputs[5], "service-plan-c, service-plan-d")
}

func TestMarketplaceServicesWhenNotLoggedIn(t *testing.T) {
	serviceOfferings := []cf.ServiceOffering{}
	serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings}

	config := &configuration.Configuration{}

	ui := callMarketplaceServices(t, config, serviceRepo)

	assert.Contains(t, ui.Outputs[0], "Getting services from marketplace...")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func callMarketplaceServices(t *testing.T, config *configuration.Configuration, serviceRepo *testapi.FakeServiceRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	ctxt := testcmd.NewContext("marketplace", []string{})
	reqFactory := &testreq.FakeReqFactory{}

	cmd := NewMarketplaceServices(ui, config, serviceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
