//go:build testing

package hub_test

import (
	"testing"

	"github.com/henrygd/beszel/internal/entities/system"
	beszelTests "github.com/henrygd/beszel/internal/tests"
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

	reportUptime := func(uptime uint64) (*system.Info, bool) {
		record, rebooted, err := monitoredSystem.CreateRecordsWithReboot(&system.CombinedData{
			Info: system.Info{Uptime: uptime},
		})
		require.NoError(t, err)
		var stored system.Info
		require.NoError(t, record.UnmarshalJSONField("info", &stored))
		return &stored, rebooted
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
