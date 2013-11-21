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

func TestListServiceBrokers(t *testing.T) {
	broker_Auto := cf.ServiceBroker{}
	broker_Auto.Name = "service-broker-to-list-a"
	broker_Auto.Guid = "service-broker-to-list-guid-a"
	broker_Auto.Url = "http://service-a-url.com"
	broker_Auto2 := cf.ServiceBroker{}
	broker_Auto2.Name = "service-broker-to-list-b"
	broker_Auto2.Guid = "service-broker-to-list-guid-b"
	broker_Auto2.Url = "http://service-b-url.com"
	broker_Auto3 := cf.ServiceBroker{}
	broker_Auto3.Name = "service-broker-to-list-c"
	broker_Auto3.Guid = "service-broker-to-list-guid-c"
	broker_Auto3.Url = "http://service-c-url.com"
	serviceBrokers := []cf.ServiceBroker{broker_Auto, broker_Auto2, broker_Auto3}

	repo := &testapi.FakeServiceBrokerRepo{
		ServiceBrokers: serviceBrokers,
	}

	ui := callListServiceBrokers(t, []string{}, repo)

	assert.Contains(t, ui.Outputs[0], "Getting service brokers as")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Contains(t, ui.Outputs[1], "name")
	assert.Contains(t, ui.Outputs[1], "url")

	assert.Contains(t, ui.Outputs[2], "service-broker-to-list-a")
	assert.Contains(t, ui.Outputs[2], "http://service-a-url.com")

	assert.Contains(t, ui.Outputs[3], "service-broker-to-list-b")
	assert.Contains(t, ui.Outputs[3], "http://service-b-url.com")

	assert.Contains(t, ui.Outputs[4], "service-broker-to-list-c")
	assert.Contains(t, ui.Outputs[4], "http://service-c-url.com")
}

func TestListingServiceBrokersWhenNoneExist(t *testing.T) {
	repo := &testapi.FakeServiceBrokerRepo{
		ServiceBrokers: []cf.ServiceBroker{},
	}

	ui := callListServiceBrokers(t, []string{}, repo)

	assert.Contains(t, ui.Outputs[0], "Getting service brokers as")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "No service brokers found")
}

func TestListingServiceBrokersWhenFindFails(t *testing.T) {
	repo := &testapi.FakeServiceBrokerRepo{ListErr: true}

	ui := callListServiceBrokers(t, []string{}, repo)

	assert.Contains(t, ui.Outputs[0], "Getting service brokers as")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "FAILED")
}

func callListServiceBrokers(t *testing.T, args []string, serviceBrokerRepo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	space_Auto := cf.SpaceFields{}
	space_Auto.Name = "my-space"
	org_Auto := cf.OrganizationFields{}
	org_Auto.Name = "my-org"
	config := &configuration.Configuration{
		Space:        space_Auto,
		Organization: org_Auto,
		AccessToken:  token,
	}

	ctxt := testcmd.NewContext("service-brokers", args)
	cmd := NewListServiceBrokers(ui, config, serviceBrokerRepo)
	testcmd.RunCommand(cmd, ctxt, &testreq.FakeReqFactory{})

	return
}
