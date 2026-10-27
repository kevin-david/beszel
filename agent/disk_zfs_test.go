//go:build testing

package agent

import (
	"errors"
	"testing"
	"time"

	"github.com/henrygd/beszel/agent/zfs"
	"github.com/henrygd/beszel/internal/entities/system"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFailedDatasetUsageDoesNotPublishStaleOrStatfsCapacity(t *testing.T) {
	b := &poolBackend{name: "zfs"}
	b.storeDatasetUsage([]zfs.Dataset{{Mountpoint: "/", Used: 10, Avail: 30}}, nil)
	m := &StoragePoolManager{backends: []*poolBackend{b}}
	a := &Agent{fsStats: map[string]*system.FsStats{"root": {Root: true, Mountpoint: "/"}}, storagePoolManager: m}
	var stats system.Stats
	a.updateDiskUsage(&stats)
	assert.Equal(t, 25.0, stats.DiskPct)

	previous := b.datasetUsage
	b.storeDatasetUsage(nil, errors.New("zfs collector failed"))
	a.updateDiskUsage(&stats)
	assert.True(t, b.datasetUsage["/"].unavailable)
	assert.Equal(t, uint64(10), previous["/"].used, "published cache references must not be mutated")
	assert.Zero(t, stats.DiskTotal)
	assert.Zero(t, stats.DiskUsed)
	assert.Zero(t, stats.DiskPct)

	b.storeDatasetUsage([]zfs.Dataset{{Mountpoint: "/", Used: 20, Avail: 20}}, nil)
	a.updateDiskUsage(&stats)
	assert.Equal(t, 50.0, stats.DiskPct)
}

func TestObservedZfsMountSurvivesMissingUtility(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	b := newZfsBackend()
	b.storeDatasetUsage([]zfs.Dataset{{Mountpoint: "/tank", Used: 10, Avail: 30}}, nil)
	datasets, err := b.datasets()
	require.Error(t, err, "the real backend must preserve missing utility or interface errors")
	b.storeDatasetUsage(datasets, err)
	require.Contains(t, b.datasetUsage, "/tank")
	assert.True(t, b.datasetUsage["/tank"].unavailable)
}

// TestUpdateDiskUsageZfsMountpoint verifies that a filesystem whose mountpoint
// is a ZFS dataset reports `zfs list` usage (which includes child datasets)
// instead of the dataset-scoped statfs values (#1541).
func TestUpdateDiskUsageZfsMountpoint(t *testing.T) {
	zm := &StoragePoolManager{detailInterval: time.Hour, backends: []*poolBackend{{name: "zfs"}}}
	zm.backends[0].datasetsFn = func() ([]zfs.Dataset, error) {
		return []zfs.Dataset{
			{Name: "tank", Used: 12000000000000, Avail: 11999000000000, Mountpoint: "/tank"},
		}, nil
	}
	agent := &Agent{
		fsStats: map[string]*system.FsStats{
			"tank": {Root: false, Mountpoint: "/tank"},
		},
		storagePoolManager: zm,
	}

	var stats system.Stats
	agent.updateDiskUsage(&stats)

	fs := agent.fsStats["tank"]
	require.NotNil(t, fs)
	assert.Equal(t, 22350.81, fs.DiskTotal) // (used + avail) in GiB
	assert.Equal(t, 11175.87, fs.DiskUsed)
	// Non-root filesystems do not populate system-level stats.
	assert.Equal(t, float64(0), stats.DiskTotal)
}

// TestUpdateDiskUsageZfsRootPopulatesSystemStats verifies the root disk values
// are derived from ZFS usage when the root mountpoint is a ZFS dataset.
func TestUpdateDiskUsageZfsRootPopulatesSystemStats(t *testing.T) {
	zm := &StoragePoolManager{detailInterval: time.Hour, backends: []*poolBackend{{name: "zfs"}}}
	zm.backends[0].datasetsFn = func() ([]zfs.Dataset, error) {
		return []zfs.Dataset{
			{Name: "rpool/ROOT/pve-1", Used: 900000000000, Avail: 300000000000, Mountpoint: "/"},
		}, nil
	}
	agent := &Agent{
		fsStats: map[string]*system.FsStats{
			"rpool/ROOT/pve-1": {Root: true, Mountpoint: "/"},
		},
		storagePoolManager: zm,
	}

	var stats system.Stats
	agent.updateDiskUsage(&stats)

	assert.Equal(t, 1117.59, agent.fsStats["rpool/ROOT/pve-1"].DiskTotal)
	assert.Equal(t, 838.19, agent.fsStats["rpool/ROOT/pve-1"].DiskUsed)
	assert.Equal(t, 75.0, stats.DiskPct)
	assert.Equal(t, 1117.59, stats.DiskTotal)
	assert.Equal(t, 838.19, stats.DiskUsed)
}

// TestUpdateDiskUsageWithoutZfsManager falls back to statfs when no manager is
// present (e.g. tests constructing bare Agent values).
func TestUpdateDiskUsageWithoutZfsManager(t *testing.T) {
	agent := &Agent{
		fsStats: map[string]*system.FsStats{
			"root": {Root: true, Mountpoint: "/"},
		},
	}

	var stats system.Stats
	agent.updateDiskUsage(&stats)

	assert.True(t, agent.fsStats["root"].DiskTotal > 0, "root usage should come from statfs")
	assert.True(t, stats.DiskTotal > 0)
}

// TestInitializeDiskIoStatsSkipsZfsMountpoints verifies ZFS filesystems are
// excluded from diskstats I/O tracking instead of warning about a missing device.
func TestInitializeDiskIoStatsSkipsZfsMountpoints(t *testing.T) {
	zm := &StoragePoolManager{detailInterval: time.Hour, backends: []*poolBackend{{name: "zfs"}}}
	zm.backends[0].datasetsFn = func() ([]zfs.Dataset, error) {
		return []zfs.Dataset{{Name: "tank", Mountpoint: "/tank"}}, nil
	}
	agent := &Agent{
		fsStats: map[string]*system.FsStats{
			"tank": {Root: false, Mountpoint: "/tank"},
			"sda1": {Root: false, Mountpoint: "/mnt/data"},
		},
		storagePoolManager: zm,
		diskPrev:           make(map[uint16]map[string]prevDisk),
	}

	agent.initializeDiskIoStats(map[string]disk.IOCountersStat{
		"sda1": {Name: "sda1", ReadBytes: 100, WriteBytes: 100},
	})

	assert.Equal(t, []string{"sda1"}, agent.fsNames)
	assert.Equal(t, uint64(100), agent.fsStats["sda1"].TotalRead)
	// ZFS entry is present but untouched by diskstats initialization.
	assert.Equal(t, uint64(0), agent.fsStats["tank"].TotalRead)
}
