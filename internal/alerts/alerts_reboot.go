package alerts

import (
	"errors"
	"fmt"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// HandleRebootAlert records and sends a point-in-time reboot event.
func (am *AlertManager) HandleRebootAlert(systemRecord *core.Record) error {
	configuredAlerts := am.alertsCache.GetAlertsByName(systemRecord.Id, "Reboot")
	if len(configuredAlerts) == 0 {
		return nil
	}

	systemName := systemRecord.GetString("name")
	var result error
	for _, alert := range configuredAlerts {
		if err := am.createRebootHistory(alert); err != nil {
			result = errors.Join(result, err)
		}
		if err := am.SendAlert(AlertMessageData{
			UserID:   alert.UserID,
			SystemID: systemRecord.Id,
			Title:    fmt.Sprintf("%s rebooted \U0001F504", systemName),
			Message:  fmt.Sprintf("%s rebooted.", systemName),
			Link:     am.hub.MakeLink("system", systemRecord.Id),
			LinkText: "View " + systemName,
		}); err != nil {
			result = errors.Join(result, err)
		}
	}
	return result
}

func (am *AlertManager) createRebootHistory(alert CachedAlertData) error {
	collection, err := am.hub.FindCachedCollectionByNameOrId("alerts_history")
	if err != nil {
		return err
	}
	record := core.NewRecord(collection)
	record.Set("alert_id", alert.Id)
	record.Set("user", alert.UserID)
	record.Set("system", alert.SystemID)
	record.Set("name", "Reboot")
	record.Set("resolved", time.Now().UTC())
	return am.hub.Save(record)
}
