package space_test

import (
	"cf"
	"cf/api"
	. "cf/commands/space"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestSpacesRequirements(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{}
	config := &configuration.Configuration{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestListingSpaces(t *testing.T) {
	space_Auto := cf.Space{}
	space_Auto.Name = "space1"
	space_Auto2 := cf.Space{}
	space_Auto2.Name = "space2"
	space_Auto3 := cf.Space{}
	space_Auto3.Name = "space3"
	spaceRepo := &testapi.FakeSpaceRepository{
		Spaces: []cf.Space{space_Auto, space_Auto2, space_Auto3},
	}
	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})

	assert.NoError(t, err)
	org_Auto := cf.Organization{}
	org_Auto.Name = "my-org"
	config := &configuration.Configuration{
		Organization: org_Auto,
		AccessToken:  token,
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	ui := callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.Contains(t, ui.Outputs[0], "Getting spaces in org")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[2], "space1")
	assert.Contains(t, ui.Outputs[3], "space2")
	assert.Contains(t, ui.Outputs[4], "space3")
}

func TestListingSpacesWhenNoSpaces(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{
		Spaces: []cf.Space{},
	}
	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})

	assert.NoError(t, err)
	org_Auto2 := cf.Organization{}
	org_Auto2.Name = "my-org"
	config := &configuration.Configuration{
		Organization: org_Auto2,
		AccessToken:  token,
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	ui := callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.Contains(t, ui.Outputs[0], "Getting spaces in org")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "No spaces found")
}

func callSpaces(args []string, reqFactory *testreq.FakeReqFactory, config *configuration.Configuration, spaceRepo api.SpaceRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("spaces", args)

	cmd := NewListSpaces(ui, config, spaceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
