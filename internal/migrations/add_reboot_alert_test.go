package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRebootAlertValues(t *testing.T) {
	original := []string{"Status", "Reboot", "CPU"}

	removed := rebootAlertValues(original, false)
	assert.Equal(t, []string{"Status", "CPU"}, removed)

	added := rebootAlertValues(removed, true)
	assert.Equal(t, []string{"Status", "CPU", "Reboot"}, added)
	assert.Equal(t, added, rebootAlertValues(added, true), "adding twice should remain idempotent")
}
