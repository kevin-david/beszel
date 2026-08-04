package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		return updateAlertNameOption(app, "Filesystem", true)
	}, func(app core.App) error {
		return updateAlertNameOption(app, "Filesystem", false)
	})
}
