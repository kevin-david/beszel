package migrations

import (
	"github.com/pocketbase/pocketbase/core"
)

// updateAlertNameOption adds or removes one value from the "name" select field
// of the alerts collection. Each alert type registers its own migration so the
// applied-migration history stays stable; they share this helper because the
// only thing that differs between them is the value.
func updateAlertNameOption(app core.App, name string, add bool) error {
	collection, err := app.FindCollectionByNameOrId("alerts")
	if err != nil {
		return err
	}
	field, ok := collection.Fields.GetByName("name").(*core.SelectField)
	if !ok {
		return nil
	}

	field.Values = alertNameValues(field.Values, name, add)
	return app.Save(collection)
}

// alertNameValues returns existing with name removed, then appended when add is
// set, so applying a migration more than once is idempotent.
func alertNameValues(existing []string, name string, add bool) []string {
	values := make([]string, 0, len(existing)+1)
	for _, value := range existing {
		if value != name {
			values = append(values, value)
		}
	}
	if add {
		values = append(values, name)
	}
	return values
}
