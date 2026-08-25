//go:build testing

package systems

import (
	"errors"
	"testing"
	"time"

	"github.com/henrygd/beszel/internal/entities/zfs"
	"github.com/henrygd/beszel/internal/hub/expirymap"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveZfsPoolsEmptyPayloadPrunesRemovedPools(t *testing.T) {
	sys, app := newTestSystemWithHub(t)
	collection, err := app.FindCachedCollectionByNameOrId("zfs_pools")
	require.NoError(t, err)
	record := core.NewRecord(collection)
	record.Set("system", sys.Id)
	record.Set("name", "removed-pool")
	require.NoError(t, app.SaveNoValidate(record))

	require.NoError(t, sys.saveZfsPools(&zfs.ZfsData{}))
	remaining, err := app.FindRecordsByFilter("zfs_pools", "system={:system}", "", 0, 0, dbx.Params{"system": sys.Id})
	require.NoError(t, err)
	assert.Empty(t, remaining)
}

func TestRecordZfsFetchResult(t *testing.T) {
	sm := &SystemManager{zfsFetchMap: expirymap.New[zfsFetchState](time.Hour)}
	t.Cleanup(sm.zfsFetchMap.StopCleaner)

	sys := &System{
		Id:          "system-1",
		manager:     sm,
		zfsInterval: time.Hour,
	}

	// Successful fetch with pools
	sys.recordZfsFetchResult(nil, 2)
	state, ok := sm.zfsFetchMap.GetOk(sys.Id)
	assert.True(t, ok, "expected zfs fetch result to be stored")
	assert.True(t, state.Successful, "expected successful fetch state to be recorded")

	// Failed fetch
	sys.recordZfsFetchResult(errors.New("failed"), 0)
	state, ok = sm.zfsFetchMap.GetOk(sys.Id)
	assert.True(t, ok, "expected failed zfs fetch state to be stored")
	assert.False(t, state.Successful, "expected failed zfs fetch state to be marked unsuccessful")

	// Successful fetch but no pools
	sys.recordZfsFetchResult(nil, 0)
	state, ok = sm.zfsFetchMap.GetOk(sys.Id)
	assert.True(t, ok, "expected fetch with zero pools to be stored")
	assert.False(t, state.Successful, "expected fetch with zero pools to be marked unsuccessful")
}

func TestShouldFetchZfs(t *testing.T) {
	sm := &SystemManager{zfsFetchMap: expirymap.New[zfsFetchState](time.Hour)}
	t.Cleanup(sm.zfsFetchMap.StopCleaner)

	sys := &System{
		Id:          "system-1",
		manager:     sm,
		zfsInterval: time.Hour,
	}

	assert.True(t, sys.shouldFetchZfs(), "expected initial zfs fetch to be allowed")

	sys.recordZfsFetchResult(errors.New("failed"), 0)
	assert.False(t, sys.shouldFetchZfs(), "expected zfs fetch to be blocked while interval entry exists")

	sm.zfsFetchMap.Remove(sys.Id)
	assert.True(t, sys.shouldFetchZfs(), "expected zfs fetch to be allowed after interval entry is cleared")
}

func TestZfsFetchIntervalDefault(t *testing.T) {
	sys := &System{}
	assert.Equal(t, time.Hour, sys.zfsFetchInterval())

	sys.zfsInterval = 5 * time.Minute
	assert.Equal(t, 5*time.Minute, sys.zfsFetchInterval())
}

func TestResetFailedZfsFetchState(t *testing.T) {
	sm := &SystemManager{zfsFetchMap: expirymap.New[zfsFetchState](time.Hour)}
	t.Cleanup(sm.zfsFetchMap.StopCleaner)

	sm.zfsFetchMap.Set("system-1", zfsFetchState{LastAttempt: time.Now().UnixMilli(), Successful: false}, time.Hour)
	sm.resetFailedZfsFetchState("system-1")
	_, ok := sm.zfsFetchMap.GetOk("system-1")
	assert.False(t, ok, "expected failed zfs fetch state to be cleared on reconnect")

	sm.zfsFetchMap.Set("system-1", zfsFetchState{LastAttempt: time.Now().UnixMilli(), Successful: true}, time.Hour)
	sm.resetFailedZfsFetchState("system-1")
	_, ok = sm.zfsFetchMap.GetOk("system-1")
	assert.True(t, ok, "expected successful zfs fetch state to be preserved")
}
