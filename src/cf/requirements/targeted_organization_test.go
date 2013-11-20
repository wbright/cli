package requirements

import (
	"cf"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testterm "testhelpers/terminal"
	"testing"
)

func TestTargetedOrgRequirement(t *testing.T) {
	ui := new(testterm.FakeUI)
	org_Auto := cf.Organization{}
	org_Auto.Name = "my-org"
	org_Auto.Guid = "my-org-guid"
	config := &configuration.Configuration{
		Organization: org_Auto,
	}

	req := newTargetedOrgRequirement(ui, config)
	success := req.Execute()
	assert.True(t, success)

	config.Organization = cf.Organization{}

	req = newTargetedOrgRequirement(ui, config)
	success = req.Execute()
	assert.False(t, success)
	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "No org targeted")
}
