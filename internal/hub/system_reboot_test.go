//go:build testing

package hub_test

import (
	"testing"

	"github.com/henrygd/beszel/internal/entities/system"
	beszelTests "github.com/henrygd/beszel/internal/tests"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemRecordsDetectRebootFromUptime(t *testing.T) {
	hub, user := beszelTests.GetHubWithUser(t)
	defer hub.Cleanup()

	systemRecords, err := beszelTests.CreateSystems(hub, 1, user.Id, "up")
	require.NoError(t, err)
	monitoredSystem, err := hub.GetSystemManager().GetSystemFromStore(systemRecords[0].Id)
	require.NoError(t, err)
	monitoredSystem.StopUpdater()
	alertCollection, err := hub.FindCollectionByNameOrId("alerts")
	require.NoError(t, err)
	alert := core.NewRecord(alertCollection)
	alert.Set("system", systemRecords[0].Id)
	alert.Set("user", user.Id)
	alert.Set("name", "Reboot")
	require.NoError(t, hub.Save(alert))

	reportUptime := func(uptime uint64) (*system.Info, bool) {
		before, err := hub.CountRecords("alerts_history", dbx.HashExp{"name": "Reboot"})
		require.NoError(t, err)
		record, err := monitoredSystem.CreateRecords(&system.CombinedData{
			Info: system.Info{Uptime: uptime},
		})
		require.NoError(t, err)
		after, err := hub.CountRecords("alerts_history", dbx.HashExp{"name": "Reboot"})
		require.NoError(t, err)
		var stored system.Info
		require.NoError(t, record.UnmarshalJSONField("info", &stored))
		return &stored, after > before
	}

	stored, rebooted := reportUptime(86400)
	assert.False(t, rebooted, "the first observation establishes the baseline")
	assert.EqualValues(t, 86400, stored.Uptime)

	_, rebooted = reportUptime(86460)
	assert.False(t, rebooted, "a climbing uptime is not a restart")

	stored, rebooted = reportUptime(20)
	assert.True(t, rebooted, "uptime collapsing past the tolerance is a restart")
	assert.EqualValues(t, 20, stored.Uptime, "the agent's reading is persisted as sent")

	_, rebooted = reportUptime(80)
	assert.False(t, rebooted, "climbing again after the restart must not re-alert")

	// A failed read reports zero. It must neither alert nor be mistaken for a
	// fresh boot, and the hub must not invent a carried-forward value.
	stored, rebooted = reportUptime(0)
	assert.False(t, rebooted, "an unreadable uptime must not look like a restart")
	assert.EqualValues(t, 0, stored.Uptime, "the hub must not write back a substitute reading")
}

func TestSystemRecordsPersistConfirmedMissingFilesystems(t *testing.T) {
	hub, user := beszelTests.GetHubWithUser(t)
	defer hub.Cleanup()

	systemRecords, err := beszelTests.CreateSystems(hub, 1, user.Id, "up")
	require.NoError(t, err)
	monitoredSystem, err := hub.GetSystemManager().GetSystemFromStore(systemRecords[0].Id)
	require.NoError(t, err)
	monitoredSystem.StopUpdater()

	writeAvailability := func(availability map[string]bool) system.Info {
		record, err := monitoredSystem.CreateRecords(&system.CombinedData{Info: system.Info{ExtraFsAvailable: availability}})
		require.NoError(t, err)
		var stored system.Info
		require.NoError(t, record.UnmarshalJSONField("info", &stored))
		return stored
	}

	stored := writeAvailability(map[string]bool{"backup": true, "media": true})
	assert.Empty(t, stored.ExtraFsMissing)
	stored = writeAvailability(map[string]bool{"backup": false, "media": true})
	assert.Empty(t, stored.ExtraFsMissing, "the first missing sample should remain pending")
	stored = writeAvailability(map[string]bool{"backup": false, "media": true})
	assert.Equal(t, []string{"backup"}, stored.ExtraFsMissing)
	stored = writeAvailability(map[string]bool{"backup": false, "media": false})
	assert.Equal(t, []string{"backup"}, stored.ExtraFsMissing, "media should remain pending for one sample")
	stored = writeAvailability(map[string]bool{"backup": true, "media": false})
	assert.Equal(t, []string{"media"}, stored.ExtraFsMissing)
}
