package alerts

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// HandleFilesystemAlert sends alerts when explicitly configured filesystems
// remain unavailable for two consecutive agent samples, change, or recover.
func (am *AlertManager) HandleFilesystemAlert(systemRecord *core.Record, previous, current []string, supported bool) error {
	if !supported {
		return nil
	}

	configuredAlerts := am.alertsCache.GetAlertsByName(systemRecord.Id, "Filesystem")
	if len(configuredAlerts) == 0 {
		return nil
	}

	var result error
	for _, alert := range configuredAlerts {
		switch {
		case !alert.Triggered && len(current) > 0:
			if err := am.setAlertTriggered(alert, true); err != nil {
				result = errors.Join(result, err)
				continue
			}
			if err := am.sendFilesystemAlert(systemRecord, alert, current, false); err != nil {
				result = errors.Join(result, err)
			}
		case alert.Triggered && len(current) == 0:
			if err := am.setAlertTriggered(alert, false); err != nil {
				result = errors.Join(result, err)
				continue
			}
			if err := am.sendFilesystemAlert(systemRecord, alert, nil, true); err != nil {
				result = errors.Join(result, err)
			}
		case alert.Triggered && !slices.Equal(previous, current):
			if err := am.sendFilesystemAlert(systemRecord, alert, current, false); err != nil {
				result = errors.Join(result, err)
			}
		}
	}
	return result
}

func (am *AlertManager) sendFilesystemAlert(systemRecord *core.Record, alert CachedAlertData, missing []string, recovered bool) error {
	systemName := systemRecord.GetString("name")
	title := fmt.Sprintf("%s filesystems available again ✅", systemName)
	message := "All configured filesystems are available again."
	if !recovered {
		names := strings.Join(missing, ", ")
		title = fmt.Sprintf("%s filesystem unavailable: %s ⚠️", systemName, names)
		message = fmt.Sprintf("%s unavailable for two consecutive checks.", names)
	}
	return am.SendAlert(AlertMessageData{
		UserID:   alert.UserID,
		SystemID: systemRecord.Id,
		Title:    title,
		Message:  message,
		Link:     am.hub.MakeLink("system", systemRecord.Id),
		LinkText: "View " + systemName,
	})
}
