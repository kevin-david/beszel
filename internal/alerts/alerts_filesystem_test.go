//go:build testing

package alerts_test

import (
	"testing"

	beszelTests "github.com/henrygd/beszel/internal/tests"
	"github.com/pocketbase/dbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesystemAlertTracksMissingSetAndRecovers(t *testing.T) {
	hub, user := beszelTests.GetHubWithUser(t)
	defer hub.Cleanup()

	setStatusAlertEmail(t, hub, user.Id, "test@example.com")
	systems, err := beszelTests.CreateSystems(hub, 1, user.Id, "up")
	require.NoError(t, err)
	system := systems[0]

	alert, err := beszelTests.CreateRecord(hub, "alerts", map[string]any{
		"name":   "Filesystem",
		"system": system.Id,
		"user":   user.Id,
	})
	require.NoError(t, err)

	require.NoError(t, hub.HandleFilesystemAlert(system, nil, []string{"usb-hdd"}, true))
	require.Equal(t, 1, hub.TestMailer.TotalSend())
	assert.Contains(t, hub.TestMailer.LastMessage().Subject, "usb-hdd")
	require.NoError(t, hub.HandleFilesystemAlert(system, []string{"usb-hdd"}, []string{"usb-hdd"}, true))
	assert.Equal(t, 1, hub.TestMailer.TotalSend(), "an active alert should not repeat")

	alert, err = hub.FindRecordById("alerts", alert.Id)
	require.NoError(t, err)
	assert.True(t, alert.GetBool("triggered"))

	require.NoError(t, hub.HandleFilesystemAlert(system, []string{"usb-hdd"}, nil, true))
	require.Equal(t, 2, hub.TestMailer.TotalSend())
	assert.Contains(t, hub.TestMailer.LastMessage().Subject, "available again")

	history, err := hub.FindFirstRecordByFilter(
		"alerts_history",
		"alert_id={:alert_id}",
		dbx.Params{"alert_id": alert.Id},
	)
	require.NoError(t, err)
	assert.False(t, history.GetDateTime("resolved").IsZero())
}

func TestFilesystemAlertUpdatesWhenMissingSetChanges(t *testing.T) {
	hub, user := beszelTests.GetHubWithUser(t)
	defer hub.Cleanup()

	setStatusAlertEmail(t, hub, user.Id, "test@example.com")
	systems, err := beszelTests.CreateSystems(hub, 1, user.Id, "up")
	require.NoError(t, err)
	alert, err := beszelTests.CreateRecord(hub, "alerts", map[string]any{
		"name":      "Filesystem",
		"system":    systems[0].Id,
		"user":      user.Id,
		"triggered": true,
	})
	require.NoError(t, err)

	require.NoError(t, hub.HandleFilesystemAlert(systems[0], []string{"backup"}, []string{"media"}, true))
	require.Equal(t, 1, hub.TestMailer.TotalSend())
	assert.Contains(t, hub.TestMailer.LastMessage().Subject, "media")
	assert.NotContains(t, hub.TestMailer.LastMessage().Subject, "backup")

	alert, err = hub.FindRecordById("alerts", alert.Id)
	require.NoError(t, err)
	assert.True(t, alert.GetBool("triggered"), "changing the missing set should keep one active history interval")
}

func TestFilesystemAlertIgnoresAgentsWithoutAvailabilityData(t *testing.T) {
	hub, user := beszelTests.GetHubWithUser(t)
	defer hub.Cleanup()

	setStatusAlertEmail(t, hub, user.Id, "test@example.com")
	systems, err := beszelTests.CreateSystems(hub, 1, user.Id, "up")
	require.NoError(t, err)
	alert, err := beszelTests.CreateRecord(hub, "alerts", map[string]any{
		"name":      "Filesystem",
		"system":    systems[0].Id,
		"user":      user.Id,
		"triggered": true,
	})
	require.NoError(t, err)

	require.NoError(t, hub.HandleFilesystemAlert(systems[0], []string{"usb-hdd"}, nil, false))
	assert.Zero(t, hub.TestMailer.TotalSend())
	alert, err = hub.FindRecordById("alerts", alert.Id)
	require.NoError(t, err)
	assert.True(t, alert.GetBool("triggered"), "missing support data must not resolve an active alert")
}
