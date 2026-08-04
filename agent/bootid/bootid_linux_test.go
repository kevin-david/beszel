//go:build linux

package bootid

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	originalPath := path
	t.Cleanup(func() { path = originalPath })

	t.Run("returns trimmed boot ID", func(t *testing.T) {
		path = filepath.Join(t.TempDir(), "boot_id")
		require.NoError(t, os.WriteFile(path, []byte("f6b8124f-9f4a-4e83-9985-e4cf7ed2c537\n"), 0o600))

		id, err := Get()

		require.NoError(t, err)
		assert.Equal(t, "f6b8124f-9f4a-4e83-9985-e4cf7ed2c537", id)
	})

	t.Run("rejects empty boot ID", func(t *testing.T) {
		path = filepath.Join(t.TempDir(), "boot_id")
		require.NoError(t, os.WriteFile(path, []byte("\n"), 0o600))

		_, err := Get()

		assert.Error(t, err)
	})
}
