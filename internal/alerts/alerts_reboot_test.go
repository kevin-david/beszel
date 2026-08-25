//go:build testing

package alerts_test

import (
	"testing"

	beszelTests "github.com/henrygd/beszel/internal/tests"
	"github.com/pocketbase/dbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRebootAlert(t *testing.T) {
	hub, user := beszelTests.GetHubWithUser(t)
	defer hub.Cleanup()

	setStatusAlertEmail(t, hub, user.Id, "test@example.com")
	systems, err := beszelTests.CreateSystems(hub, 1, user.Id, "up")
	require.NoError(t, err)
	system := systems[0]

	alert, err := beszelTests.CreateRecord(hub, "alerts", map[string]any{
		"name":   "Reboot",
		"system": system.Id,
		"user":   user.Id,
	})
	require.NoError(t, err)

	require.NoError(t, hub.HandleRebootAlert(system))

	assert.Equal(t, 1, hub.TestMailer.TotalSend())
	assert.Contains(t, hub.TestMailer.LastMessage().Subject, "test-system-0 rebooted")
	history, err := hub.FindFirstRecordByFilter(
		"alerts_history",
		"alert_id={:alert_id}",
		dbx.Params{"alert_id": alert.Id},
	)
	require.NoError(t, err)
	assert.Equal(t, "Reboot", history.GetString("name"))
	assert.False(t, history.GetDateTime("resolved").IsZero())
	assert.False(t, alert.GetBool("triggered"), "point-in-time alerts should not remain active")
}

func TestRebootAlertRequiresConfiguration(t *testing.T) {
	hub, user := beszelTests.GetHubWithUser(t)
	defer hub.Cleanup()

	setStatusAlertEmail(t, hub, user.Id, "test@example.com")
	systems, err := beszelTests.CreateSystems(hub, 1, user.Id, "up")
	require.NoError(t, err)

	require.NoError(t, hub.HandleRebootAlert(systems[0]))

	assert.Zero(t, hub.TestMailer.TotalSend())
	historyCount, err := hub.CountRecords("alerts_history", nil)
	require.NoError(t, err)
	assert.Zero(t, historyCount)
}
