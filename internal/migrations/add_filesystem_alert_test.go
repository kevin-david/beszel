package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilesystemAlertValues(t *testing.T) {
	original := []string{"Status", "Filesystem", "CPU"}

	removed := filesystemAlertValues(original, false)
	assert.Equal(t, []string{"Status", "CPU"}, removed)

	added := filesystemAlertValues(removed, true)
	assert.Equal(t, []string{"Status", "CPU", "Filesystem"}, added)
	assert.Equal(t, added, filesystemAlertValues(added, true), "adding twice should remain idempotent")
}
