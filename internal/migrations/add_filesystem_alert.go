package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		return updateFilesystemAlertOption(app, true)
	}, func(app core.App) error {
		return updateFilesystemAlertOption(app, false)
	})
}

func updateFilesystemAlertOption(app core.App, add bool) error {
	collection, err := app.FindCollectionByNameOrId("alerts")
	if err != nil {
		return err
	}
	field, ok := collection.Fields.GetByName("name").(*core.SelectField)
	if !ok {
		return nil
	}

	field.Values = filesystemAlertValues(field.Values, add)
	return app.Save(collection)
}

func filesystemAlertValues(existing []string, add bool) []string {
	values := make([]string, 0, len(existing)+1)
	for _, value := range existing {
		if value != "Filesystem" {
			values = append(values, value)
		}
	}
	if add {
		values = append(values, "Filesystem")
	}
	return values
}
