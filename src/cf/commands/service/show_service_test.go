package service_test

import (
	"cf"
	. "cf/commands/service"
	"github.com/stretchr/testify/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestShowServiceRequirements(t *testing.T) {
	args := []string{"service1"}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	callShowService(args, reqFactory)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
	callShowService(args, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
	callShowService(args, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.ServiceInstanceName, "service1")
}

func TestShowServiceFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

	ui := callShowService([]string{}, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callShowService([]string{"my-service"}, reqFactory)
	assert.False(t, ui.FailedWithUsage)
}

func TestShowServiceOutput(t *testing.T) {
	offering_Auto := cf.ServiceOffering{}
	offering_Auto.Label = "mysql"
	offering_Auto.DocumentationUrl = "http://documentation.url"
	offering_Auto.Description = "the-description"
	plan_Auto := cf.ServicePlan{}
	plan_Auto.Guid = "plan-guid"
	plan_Auto.Name = "plan-name"
	plan_Auto.ServiceOffering = offering_Auto
	serviceInstance_Auto := cf.ServiceInstance{}
	serviceInstance_Auto.Name = "service1"
	serviceInstance_Auto.Guid = "service1-guid"
	serviceInstance_Auto.ServicePlan = plan_Auto
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:         true,
		TargetedSpaceSuccess: true,
		ServiceInstance:      serviceInstance_Auto,
	}
	ui := callShowService([]string{"service1"}, reqFactory)

	assert.Contains(t, ui.Outputs[0], "")
	assert.Contains(t, ui.Outputs[1], "Service instance: ")
	assert.Contains(t, ui.Outputs[1], "service1")
	assert.Contains(t, ui.Outputs[2], "Service: ")
	assert.Contains(t, ui.Outputs[2], "mysql")
	assert.Contains(t, ui.Outputs[3], "Plan: ")
	assert.Contains(t, ui.Outputs[3], "plan-name")
	assert.Contains(t, ui.Outputs[4], "Description: ")
	assert.Contains(t, ui.Outputs[4], "the-description")
	assert.Contains(t, ui.Outputs[5], "Documentation url: ")
	assert.Contains(t, ui.Outputs[5], "http://documentation.url")
}

func TestShowUserProvidedServiceOutput(t *testing.T) {
	serviceInstance_Auto2 := cf.ServiceInstance{}
	serviceInstance_Auto2.Name = "service1"
	serviceInstance_Auto2.Guid = "service1-guid"
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:         true,
		TargetedSpaceSuccess: true,
		ServiceInstance:      serviceInstance_Auto2,
	}
	ui := callShowService([]string{"service1"}, reqFactory)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "")
	assert.Contains(t, ui.Outputs[1], "Service instance: ")
	assert.Contains(t, ui.Outputs[1], "service1")
	assert.Contains(t, ui.Outputs[2], "Service: ")
	assert.Contains(t, ui.Outputs[2], "user-provided")
}

func callShowService(args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("service", args)
	cmd := NewShowService(ui)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
