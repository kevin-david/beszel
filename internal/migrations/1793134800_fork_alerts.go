package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// The 0.20 collections snapshot replaces select options. Earlier fork
		// migrations are already applied on existing installations.
		if err := updateAlertNameOption(app, "Reboot", true); err != nil {
			return err
		}
		return updateAlertNameOption(app, "Filesystem", true)
	}, nil)
}
