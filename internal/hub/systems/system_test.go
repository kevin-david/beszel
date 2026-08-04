//go:build testing

package systems

import (
	"context"
	"testing"

	"github.com/henrygd/beszel/internal/entities/system"
)

func TestRebootedFromUptime(t *testing.T) {
	const day = 86400
	tests := []struct {
		name     string
		previous uint64
		current  uint64
		rebooted bool
	}{
		// A restart drops uptime far below the previous observation.
		{name: "reboot after a day", previous: day, current: 30, rebooted: true},
		{name: "reboot just past tolerance", previous: 300, current: 300 - rebootUptimeTolerance - 1, rebooted: true},

		// The agent cannot report before it is running, so a zero reading means
		// the platform is unsupported or the read failed - never a fresh boot.
		{name: "current read failed after long uptime", previous: day, current: 0},
		{name: "unsupported platform reports zero forever", previous: 0, current: 0},

		// Falls out of the arithmetic with no special case.
		{name: "first observation", previous: 0, current: day},

		// Steady state.
		{name: "normal increase", previous: day, current: day + 60},
		{name: "unchanged", previous: day, current: day},

		// Wall-clock derived uptime (darwin/BSD/solaris) wobbles without a reboot.
		{name: "small backwards clock step", previous: day, current: day - 5},
		{name: "backwards step exactly at tolerance", previous: day, current: day - rebootUptimeTolerance},

		// Accepted limitation: a restart on a host that had only just booted
		// drops uptime by less than the tolerance, and is not reported.
		{name: "restart too soon after booting is missed", previous: 130, current: 50},

		// A clock step past boot time underflows uint64 there; neither the
		// absurd value nor recovering from it may be reported as a reboot.
		{name: "underflowed current", previous: day, current: ^uint64(0) - 100},
		{name: "underflowed previous recovering", previous: ^uint64(0) - 100, current: day},

		// The implausibility ceiling must not swallow genuinely long uptimes.
		{name: "reboot after ten years up", previous: 10 * 365 * day, current: 45, rebooted: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rebootedFromUptime(tt.previous, tt.current); got != tt.rebooted {
				t.Fatalf("rebootedFromUptime(%d, %d) = %v, want %v", tt.previous, tt.current, got, tt.rebooted)
			}
		})
	}
}

func TestCombinedData_MigrateDeprecatedFields(t *testing.T) {
	t.Run("Migrate NetworkSent and NetworkRecv to Bandwidth", func(t *testing.T) {
		cd := &system.CombinedData{
			Stats: system.Stats{
				NetworkSent: 1.5, // 1.5 MB
				NetworkRecv: 2.5, // 2.5 MB
			},
		}
		migrateDeprecatedFields(cd, true)

		expectedSent := uint64(1.5 * 1024 * 1024)
		expectedRecv := uint64(2.5 * 1024 * 1024)

		if cd.Stats.Bandwidth[0] != expectedSent {
			t.Errorf("expected Bandwidth[0] %d, got %d", expectedSent, cd.Stats.Bandwidth[0])
		}
		if cd.Stats.Bandwidth[1] != expectedRecv {
			t.Errorf("expected Bandwidth[1] %d, got %d", expectedRecv, cd.Stats.Bandwidth[1])
		}
		if cd.Stats.NetworkSent != 0 || cd.Stats.NetworkRecv != 0 {
			t.Errorf("expected NetworkSent and NetworkRecv to be reset, got %f, %f", cd.Stats.NetworkSent, cd.Stats.NetworkRecv)
		}
	})

	t.Run("Migrate Info.Bandwidth to Info.BandwidthBytes", func(t *testing.T) {
		cd := &system.CombinedData{
			Info: system.Info{
				Bandwidth: 10.0, // 10 MB
			},
		}
		migrateDeprecatedFields(cd, true)

		expected := uint64(10 * 1024 * 1024)
		if cd.Info.BandwidthBytes != expected {
			t.Errorf("expected BandwidthBytes %d, got %d", expected, cd.Info.BandwidthBytes)
		}
		if cd.Info.Bandwidth != 0 {
			t.Errorf("expected Info.Bandwidth to be reset, got %f", cd.Info.Bandwidth)
		}
	})

	t.Run("Migrate DiskReadPs and DiskWritePs to DiskIO", func(t *testing.T) {
		cd := &system.CombinedData{
			Stats: system.Stats{
				DiskReadPs:  3.0, // 3 MB
				DiskWritePs: 4.0, // 4 MB
			},
		}
		migrateDeprecatedFields(cd, true)

		expectedRead := uint64(3 * 1024 * 1024)
		expectedWrite := uint64(4 * 1024 * 1024)

		if cd.Stats.DiskIO[0] != expectedRead {
			t.Errorf("expected DiskIO[0] %d, got %d", expectedRead, cd.Stats.DiskIO[0])
		}
		if cd.Stats.DiskIO[1] != expectedWrite {
			t.Errorf("expected DiskIO[1] %d, got %d", expectedWrite, cd.Stats.DiskIO[1])
		}
		if cd.Stats.DiskReadPs != 0 || cd.Stats.DiskWritePs != 0 {
			t.Errorf("expected DiskReadPs and DiskWritePs to be reset, got %f, %f", cd.Stats.DiskReadPs, cd.Stats.DiskWritePs)
		}
	})

	t.Run("Migrate Info fields to Details struct", func(t *testing.T) {
		cd := &system.CombinedData{
			Stats: system.Stats{
				Mem: 16.0, // 16 GB
			},
			Info: system.Info{
				Hostname:      "test-host",
				KernelVersion: "6.8.0",
				Cores:         8,
				Threads:       16,
				CpuModel:      "Intel i7",
				Podman:        true,
				Os:            system.Linux,
			},
		}
		migrateDeprecatedFields(cd, true)

		if cd.Details == nil {
			t.Fatal("expected Details struct to be created")
		}
		if cd.Details.Hostname != "test-host" {
			t.Errorf("expected Hostname 'test-host', got '%s'", cd.Details.Hostname)
		}
		if cd.Details.Kernel != "6.8.0" {
			t.Errorf("expected Kernel '6.8.0', got '%s'", cd.Details.Kernel)
		}
		if cd.Details.Cores != 8 {
			t.Errorf("expected Cores 8, got %d", cd.Details.Cores)
		}
		if cd.Details.Threads != 16 {
			t.Errorf("expected Threads 16, got %d", cd.Details.Threads)
		}
		if cd.Details.CpuModel != "Intel i7" {
			t.Errorf("expected CpuModel 'Intel i7', got '%s'", cd.Details.CpuModel)
		}
		if cd.Details.Podman != true {
			t.Errorf("expected Podman true, got %v", cd.Details.Podman)
		}
		if cd.Details.Os != system.Linux {
			t.Errorf("expected Os Linux, got %d", cd.Details.Os)
		}
		expectedMem := uint64(16 * 1024 * 1024 * 1024)
		if cd.Details.MemoryTotal != expectedMem {
			t.Errorf("expected MemoryTotal %d, got %d", expectedMem, cd.Details.MemoryTotal)
		}

		if cd.Info.Hostname != "" || cd.Info.KernelVersion != "" || cd.Info.Cores != 0 || cd.Info.CpuModel != "" || cd.Info.Podman != false || cd.Info.Os != 0 {
			t.Errorf("expected Info fields to be reset, got %+v", cd.Info)
		}
	})

	t.Run("Do not migrate if Details already exists", func(t *testing.T) {
		cd := &system.CombinedData{
			Details: &system.Details{Hostname: "existing-host"},
			Info: system.Info{
				Hostname: "deprecated-host",
			},
		}
		migrateDeprecatedFields(cd, true)

		if cd.Details.Hostname != "existing-host" {
			t.Errorf("expected Hostname 'existing-host', got '%s'", cd.Details.Hostname)
		}
		if cd.Info.Hostname != "deprecated-host" {
			t.Errorf("expected Info.Hostname to remain 'deprecated-host', got '%s'", cd.Info.Hostname)
		}
	})

	t.Run("Do not create details if migrateDetails is false", func(t *testing.T) {
		cd := &system.CombinedData{
			Info: system.Info{
				Hostname: "deprecated-host",
			},
		}
		migrateDeprecatedFields(cd, false)

		if cd.Details != nil {
			t.Fatal("expected Details struct to not be created")
		}

		if cd.Info.Hostname != "" {
			t.Errorf("expected Info.Hostname to be reset, got '%s'", cd.Info.Hostname)
		}
	})
}

func TestSetDownAfterContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// manager is nil on purpose: setDown must bail out before touching the app
	sys := &System{Status: up, ctx: ctx}

	if err := sys.setDown(nil); err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if sys.Status != up {
		t.Fatalf("status should be untouched, got %q", sys.Status)
	}
}
