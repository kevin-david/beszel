//go:build testing

package hub_test

import (
	"testing"

	"github.com/henrygd/beszel/internal/entities/system"
	beszelTests "github.com/henrygd/beszel/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemRecordsDetectAndRetainBootIdentity(t *testing.T) {
	hub, user := beszelTests.GetHubWithUser(t)
	defer hub.Cleanup()

	systemRecords, err := beszelTests.CreateSystems(hub, 1, user.Id, "up")
	require.NoError(t, err)
	monitoredSystem, err := hub.GetSystemManager().GetSystemFromStore(systemRecords[0].Id)
	require.NoError(t, err)
	monitoredSystem.StopUpdater()

	bootA := "boot-a"
	_, rebooted, err := monitoredSystem.CreateRecordsWithReboot(&system.CombinedData{Info: system.Info{BootID: &bootA}})
	require.NoError(t, err)
	assert.False(t, rebooted, "the first boot identity establishes the baseline")

	bootB := "boot-b"
	_, rebooted, err = monitoredSystem.CreateRecordsWithReboot(&system.CombinedData{Info: system.Info{BootID: &bootB}})
	require.NoError(t, err)
	assert.True(t, rebooted)

	record, rebooted, err := monitoredSystem.CreateRecordsWithReboot(&system.CombinedData{})
	require.NoError(t, err)
	assert.False(t, rebooted, "a missing sample must not look like another reboot")
	var storedInfo system.Info
	require.NoError(t, record.UnmarshalJSONField("info", &storedInfo))
	require.NotNil(t, storedInfo.BootID)
	assert.Equal(t, bootB, *storedInfo.BootID, "a missing sample must not erase the last valid identity")
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
