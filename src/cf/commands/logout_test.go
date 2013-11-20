package commands_test

import (
	"cf"
	"cf/commands"
	"github.com/stretchr/testify/assert"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
	"testing"
)

func TestLogoutClearsAccessTokenOrgAndSpace(t *testing.T) {
	org := cf.OrganizationFields{}
	org.Name = "MyOrg"

	space := cf.SpaceFields{}
	space.Name = "MySpace"

	configRepo := &testconfig.FakeConfigRepository{}
	config, _ := configRepo.Get()
	config.AccessToken = "MyAccessToken"
	config.Organization = org
	config.Space = space

	ui := new(testterm.FakeUI)

	l := commands.NewLogout(ui, configRepo)
	l.Run(nil)

	updatedConfig, err := configRepo.Get()
	assert.NoError(t, err)

	assert.Empty(t, updatedConfig.AccessToken)
	assert.Equal(t, updatedConfig.Organization, cf.Organization{})
	assert.Equal(t, updatedConfig.Space, cf.Space{})
}
