package system

import (
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLegacyForkCombinedDataCBOR(t *testing.T) {
	// Literal field numbers match c4387a84, independent of the current tags.
	legacyPool := map[uint64]any{0: 100.0, 1: 60.0, 4: "ONLINE"}
	legacyDataset := map[uint64]any{0: 20.0, 1: 10.0}
	for _, poolsPresent := range []bool{false, true} {
		stats := map[uint64]any{39: map[string]any{"tank/apps": legacyDataset}}
		if poolsPresent {
			stats[38] = map[string]any{"tank": legacyPool}
		}
		payload, err := cbor.Marshal(map[uint64]any{
			0: stats,
			1: map[uint64]any{10: "0.18.8", 25: map[string]bool{"backup": false}},
		})
		require.NoError(t, err)
		var decoded CombinedData
		require.NoError(t, cbor.Unmarshal(payload, &decoded))
		assert.Equal(t, map[string]bool{"backup": false}, decoded.Info.ExtraFsAvailable)
		assert.Zero(t, decoded.Stats.DiskIOTotal)
		assert.NotContains(t, decoded.Stats.ZfsPools, "tank/apps")
		if poolsPresent {
			require.Contains(t, decoded.Stats.ZfsPools, "tank")
			assert.Equal(t, "ONLINE", decoded.Stats.ZfsPools["tank"].Health)
			assert.Equal(t, 60.0, decoded.Stats.ZfsPools["tank"].Used)
		} else {
			assert.Empty(t, decoded.Stats.ZfsPools)
		}
	}
}

func TestUpstreamCombinedDataCBORKeepsCountersAndPools(t *testing.T) {
	payload, err := cbor.Marshal(map[uint64]any{
		0: map[uint64]any{38: [2]uint64{100, 200}, 39: map[string]any{"tank": map[uint64]any{0: 100.0, 1: 60.0, 4: "ONLINE"}}},
		1: map[uint64]any{25: []uint16{12, 3}},
	})
	require.NoError(t, err)
	var decoded CombinedData
	require.NoError(t, cbor.Unmarshal(payload, &decoded))
	assert.Equal(t, [2]uint64{100, 200}, decoded.Stats.DiskIOTotal)
	assert.Equal(t, []uint16{12, 3}, decoded.Info.PackageUpdates)
	require.Contains(t, decoded.Stats.ZfsPools, "tank")
	assert.Equal(t, "ONLINE", decoded.Stats.ZfsPools["tank"].Health)
}
