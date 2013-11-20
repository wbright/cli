package requirements

import (
	"cf"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testterm "testhelpers/terminal"
	"testing"
)

func TestSpaceRequirement(t *testing.T) {
	ui := new(testterm.FakeUI)
	org_Auto := cf.Organization{}
	org_Auto.Name = "my-org"
	org_Auto.Guid = "my-org-guid"
	space_Auto := cf.Space{}
	space_Auto.Name = "my-space"
	space_Auto.Guid = "my-space-guid"
	config := &configuration.Configuration{
		Organization: org_Auto,

		Space: space_Auto,
	}

	req := newTargetedSpaceRequirement(ui, config)
	success := req.Execute()
	assert.True(t, success)

	config.Space = cf.Space{}

	req = newTargetedSpaceRequirement(ui, config)
	success = req.Execute()
	assert.False(t, success)
	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "No space targeted")

	ui.ClearOutputs()
	config.Organization = cf.Organization{}

	req = newTargetedSpaceRequirement(ui, config)
	success = req.Execute()
	assert.False(t, success)
	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "No org and space targeted")
}
