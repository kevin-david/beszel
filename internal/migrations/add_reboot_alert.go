package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		return updateRebootAlertOption(app, true)
	}, func(app core.App) error {
		return updateRebootAlertOption(app, false)
	})
}

func updateRebootAlertOption(app core.App, add bool) error {
	collection, err := app.FindCollectionByNameOrId("alerts")
	if err != nil {
		return err
	}
	field, ok := collection.Fields.GetByName("name").(*core.SelectField)
	if !ok {
		return nil
	}

	field.Values = rebootAlertValues(field.Values, add)
	return app.Save(collection)
}

func rebootAlertValues(existing []string, add bool) []string {
	values := make([]string, 0, len(existing)+1)
	for _, value := range existing {
		if value != "Reboot" {
			values = append(values, value)
		}
	}
	if add {
		values = append(values, "Reboot")
	}
	return values
}
