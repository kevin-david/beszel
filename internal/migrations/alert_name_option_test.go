package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlertNameValues(t *testing.T) {
	for _, name := range []string{"Reboot", "Filesystem"} {
		t.Run(name, func(t *testing.T) {
			original := []string{"Status", name, "CPU"}

			removed := alertNameValues(original, name, false)
			assert.Equal(t, []string{"Status", "CPU"}, removed)

			added := alertNameValues(removed, name, true)
			assert.Equal(t, []string{"Status", "CPU", name}, added)
			assert.Equal(t, added, alertNameValues(added, name, true), "adding twice should remain idempotent")
		})
	}

	t.Run("leaves other alert names alone", func(t *testing.T) {
		values := alertNameValues([]string{"Status", "Reboot", "CPU"}, "Filesystem", true)
		assert.Equal(t, []string{"Status", "Reboot", "CPU", "Filesystem"}, values)
	})
}
