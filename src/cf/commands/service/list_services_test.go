package service_test

import (
	"cf"
	. "cf/commands/service"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
	"testing"
)

func TestServices(t *testing.T) {
	plan_Auto := cf.ServicePlan{}
	plan_Auto.Guid = "spark-guid"
	plan_Auto.Name = "spark"
	offering_Auto = cf.ServiceOffering{}
	offering_Auto.Label = "cleardb"
	plan_Auto2 := cf.ServicePlan{}
	plan_Auto2.Guid = "spark-guid"
	plan_Auto2.Name = "spark"
	offering_Auto2 = cf.ServiceOffering{}
	offering_Auto2.Label = "cleardb"
	serviceInstance_Auto := cf.ServiceInstance{}
	serviceInstance_Auto.Name = "my-service-1"
	serviceInstance_Auto.ServicePlan = plan_Auto
	serviceInstance_Auto.ApplicationNames = []string{"cli1", "cli2"}
	serviceInstance_Auto2 := cf.ServiceInstance{}
	serviceInstance_Auto2.Name = "my-service-2"
	serviceInstance_Auto2.ServicePlan = plan_Auto2
	serviceInstance_Auto2.ApplicationNames = []string{"cli1"}
	serviceInstance_Auto3 := cf.ServiceInstance{}
	serviceInstance_Auto3.Name = "my-service-provided-by-user"
	serviceInstances := []cf.ServiceInstance{serviceInstance_Auto, serviceInstance_Auto2, serviceInstance_Auto3}
	serviceSummaryRepo := &testapi.FakeServiceSummaryRepo{
		GetSummariesInCurrentSpaceInstances: serviceInstances,
	}
	ui := &testterm.FakeUI{}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org_Auto := cf.Organization{}
	org_Auto.Name = "my-org"
	space_Auto := cf.Space{}
	space_Auto.Name = "my-space"
	config := &configuration.Configuration{
		Space:        space_Auto,
		Organization: org_Auto,
		AccessToken:  token,
	}

	cmd := NewListServices(ui, config, serviceSummaryRepo)
	cmd.Run(testcmd.NewContext("services", []string{}))

	assert.Contains(t, ui.Outputs[0], "Getting services in org")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[4], "my-service-1")
	assert.Contains(t, ui.Outputs[4], "cleardb")
	assert.Contains(t, ui.Outputs[4], "spark")
	assert.Contains(t, ui.Outputs[4], "cli1, cli2")

	assert.Contains(t, ui.Outputs[5], "my-service-2")
	assert.Contains(t, ui.Outputs[5], "cleardb")
	assert.Contains(t, ui.Outputs[5], "spark")
	assert.Contains(t, ui.Outputs[5], "cli1")

	assert.Contains(t, ui.Outputs[6], "my-service-provided-by-user")
	assert.Contains(t, ui.Outputs[6], "user-provided")
}
