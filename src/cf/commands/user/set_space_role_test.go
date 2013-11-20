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

func TestSetSpaceRoleFailsWithUsage(t *testing.T) {
	reqFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()

	ui := callSetSpaceRole(t, []string{}, reqFactory, spaceRepo, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetSpaceRole(t, []string{"my-user"}, reqFactory, spaceRepo, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetSpaceRole(t, []string{"my-user", "my-org"}, reqFactory, spaceRepo, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetSpaceRole(t, []string{"my-user", "my-org", "my-space"}, reqFactory, spaceRepo, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetSpaceRole(t, []string{"my-user", "my-org", "my-space", "my-role"}, reqFactory, spaceRepo, userRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestSetSpaceRoleRequirements(t *testing.T) {
	args := []string{"username", "org", "space", "role"}
	reqFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()

	reqFactory.LoginSuccess = false
	callSetSpaceRole(t, args, reqFactory, spaceRepo, userRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callSetSpaceRole(t, args, reqFactory, spaceRepo, userRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.UserUsername, "username")
	assert.Equal(t, reqFactory.OrganizationName, "org")
}

func TestSetSpaceRole(t *testing.T) {
	args := []string{"some-user", "some-org", "some-space", "some-role"}
	reqFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()

	reqFactory.LoginSuccess = true
	user_Auto = cf.User{}
	user_Auto.Guid = "my-user-guid"
	user_Auto.Username = "my-user"
	org_Auto = cf.Organization{}
	org_Auto.Guid = "my-org-guid"
	org_Auto.Name = "my-org"
	space_Auto = cf.Space{}
	space_Auto.Guid = "my-space-guid"
	space_Auto.Name = "my-space"

	ui := callSetSpaceRole(t, args, reqFactory, spaceRepo, userRepo)

	assert.Equal(t, spaceRepo.FindByNameInOrgName, "some-space")
	assert.Equal(t, spaceRepo.FindByNameInOrgOrg, reqFactory.Organization)

	assert.Contains(t, ui.Outputs[0], "Assigning role ")
	assert.Contains(t, ui.Outputs[0], "some-role")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "current-user")

	assert.Equal(t, userRepo.SetSpaceRoleUser, reqFactory.User)
	assert.Equal(t, userRepo.SetSpaceRoleSpace, spaceRepo.FindByNameInOrgSpace)
	assert.Equal(t, userRepo.SetSpaceRoleRole, "some-role")

	assert.Contains(t, ui.Outputs[1], "OK")
}

func getSetSpaceRoleDeps() (reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) {
	reqFactory = &testreq.FakeReqFactory{}
	spaceRepo = &testapi.FakeSpaceRepository{}
	userRepo = &testapi.FakeUserRepository{}
	return
}

func callSetSpaceRole(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-space-role", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "current-user",
	})
	assert.NoError(t, err)
	space_Auto2 := cf.Space{}
	space_Auto2.Name = "my-space"
	org_Auto2 := cf.Organization{}
	org_Auto2.Name = "my-org"
	config := &configuration.Configuration{
		Space:        space_Auto2,
		Organization: org_Auto2,
		AccessToken:  token,
	}

	cmd := NewSetSpaceRole(ui, config, spaceRepo, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
