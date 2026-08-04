package alerts

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// HandleFilesystemAlert sends alerts when explicitly configured filesystems
// remain unavailable for two consecutive agent samples and when they recover.
func (am *AlertManager) HandleFilesystemAlert(systemRecord *core.Record, previous, current map[string]bool) error {
	// A nil map means the agent does not support filesystem availability data.
	if current == nil {
		return nil
	}

	configuredAlerts := am.alertsCache.GetAlertsByName(systemRecord.Id, "Filesystem")
	if len(configuredAlerts) == 0 {
		return nil
	}

	confirmedMissing := confirmedMissingFilesystems(previous, current)
	allAvailable := true
	for _, available := range current {
		if !available {
			allAvailable = false
			break
		}
	}

	var result error
	for _, alert := range configuredAlerts {
		switch {
		case !alert.Triggered && len(confirmedMissing) > 0:
			if err := am.setAlertTriggered(alert, true); err != nil {
				result = errors.Join(result, err)
				continue
			}
			if err := am.sendFilesystemAlert(systemRecord, alert, confirmedMissing, false); err != nil {
				result = errors.Join(result, err)
			}
		case alert.Triggered && allAvailable:
			if err := am.setAlertTriggered(alert, false); err != nil {
				result = errors.Join(result, err)
				continue
			}
			if err := am.sendFilesystemAlert(systemRecord, alert, nil, true); err != nil {
				result = errors.Join(result, err)
			}
		}
	}
	return result
}

func confirmedMissingFilesystems(previous, current map[string]bool) []string {
	if previous == nil || current == nil {
		return nil
	}
	missing := make([]string, 0)
	for name, available := range current {
		wasAvailable, observed := previous[name]
		if !available && observed && !wasAvailable {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	return missing
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
