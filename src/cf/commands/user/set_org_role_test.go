package user_test

import (
	"cf"
	. "cf/commands/user"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestSetOrgRoleFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	userRepo := &testapi.FakeUserRepository{}

	ui := callSetOrgRole(t, []string{"my-user", "my-org", "my-role"}, reqFactory, userRepo)
	assert.False(t, ui.FailedWithUsage)

	ui = callSetOrgRole(t, []string{"my-user", "my-org"}, reqFactory, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetOrgRole(t, []string{"my-user"}, reqFactory, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetOrgRole(t, []string{}, reqFactory, userRepo)
	assert.True(t, ui.FailedWithUsage)
}

func TestSetOrgRoleRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	userRepo := &testapi.FakeUserRepository{}

	reqFactory.LoginSuccess = false
	callSetOrgRole(t, []string{"my-user", "my-org", "my-role"}, reqFactory, userRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callSetOrgRole(t, []string{"my-user", "my-org", "my-role"}, reqFactory, userRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.UserUsername, "my-user")
	assert.Equal(t, reqFactory.OrganizationName, "my-org")
}

func TestSetOrgRole(t *testing.T) {
	org_Auto := cf.Organization{}
	org_Auto.Guid = "my-org-guid"
	org_Auto.Name = "my-org"
	user_Auto := cf.User{}
	user_Auto.Guid = "my-user-guid"
	user_Auto.Username = "my-user"
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess: true,
		User:         user_Auto,
		Organization: org_Auto,
	}
	userRepo := &testapi.FakeUserRepository{}

	ui := callSetOrgRole(t, []string{"some-user", "some-org", "some-role"}, reqFactory, userRepo)

	assert.Contains(t, ui.Outputs[0], "Assigning role ")
	assert.Contains(t, ui.Outputs[0], "some-role")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "current-user")

	assert.Equal(t, userRepo.SetOrgRoleUser, reqFactory.User)
	assert.Equal(t, userRepo.SetOrgRoleOrganization, reqFactory.Organization)
	assert.Equal(t, userRepo.SetOrgRoleRole, "some-role")

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callSetOrgRole(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-org-role", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "current-user",
	})
	assert.NoError(t, err)
	org_Auto2 := cf.Organization{}
	org_Auto2.Name = "my-org"
	space_Auto := cf.Space{}
	space_Auto.Name = "my-space"
	config := &configuration.Configuration{
		Space:        space_Auto,
		Organization: org_Auto2,
		AccessToken:  token,
	}

	cmd := NewSetOrgRole(ui, config, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
