package agent

import (
	"testing"

	"github.com/henrygd/beszel/internal/entities/system"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/stretchr/testify/assert"
)

func TestExpectedFilesystemDisappearanceAndRecovery(t *testing.T) {
	stats := &system.FsStats{
		Mountpoint: "/usb-hdd", Setting: "/usb-hdd", Expected: true,
		Available: true, DiskTotal: 100, DiskUsed: 90,
	}
	filesystems := map[string]*system.FsStats{"usb": stats}
	refreshExpectedFilesystemAvailability(filesystems, nil)
	assert.False(t, stats.Available)
	assert.Zero(t, stats.DiskTotal)
	assert.Zero(t, stats.DiskUsed)

	refreshExpectedFilesystemAvailability(filesystems, []disk.PartitionStat{{Device: "/dev/sdd", Mountpoint: "/usb-hdd"}})
	assert.True(t, stats.Available)
}
