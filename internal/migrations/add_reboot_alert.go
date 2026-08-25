package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		return updateAlertNameOption(app, "Reboot", true)
	}, func(app core.App) error {
		return updateAlertNameOption(app, "Reboot", false)
	})
}
