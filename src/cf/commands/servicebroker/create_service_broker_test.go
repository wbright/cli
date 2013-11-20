package servicebroker_test

import (
	"cf"
	. "cf/commands/servicebroker"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestCreateServiceBrokerFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}

	ui := callCreateServiceBroker(t, []string{}, reqFactory, serviceBrokerRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceBroker(t, []string{"1arg"}, reqFactory, serviceBrokerRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceBroker(t, []string{"1arg", "2arg"}, reqFactory, serviceBrokerRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceBroker(t, []string{"1arg", "2arg", "3arg"}, reqFactory, serviceBrokerRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceBroker(t, []string{"1arg", "2arg", "3arg", "4arg"}, reqFactory, serviceBrokerRepo)
	assert.False(t, ui.FailedWithUsage)

}
func TestCreateServiceBrokerRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}
	args := []string{"1arg", "2arg", "3arg", "4arg"}

	reqFactory.LoginSuccess = false
	callCreateServiceBroker(t, args, reqFactory, serviceBrokerRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callCreateServiceBroker(t, args, reqFactory, serviceBrokerRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestCreateServiceBroker(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}
	args := []string{"my-broker", "my username", "my password", "http://example.com"}
	ui := callCreateServiceBroker(t, args, reqFactory, serviceBrokerRepo)

	assert.Contains(t, ui.Outputs[0], "Creating service broker ")
	assert.Contains(t, ui.Outputs[0], "my-broker")
	assert.Contains(t, ui.Outputs[0], "my-user")
	expectedServiceBroker := cf.ServiceBroker{}
	expectedServiceBroker.Name = "my-broker"
	expectedServiceBroker.Username = "my username"
	expectedServiceBroker.Password = "my password"
	expectedServiceBroker.Url = "http://example.com"

	assert.Equal(t, serviceBrokerRepo.CreatedServiceBroker, expectedServiceBroker)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callCreateServiceBroker(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, serviceBrokerRepo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("create-service-broker", args)

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

	cmd := NewCreateServiceBroker(ui, config, serviceBrokerRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
