package system

import (
	"encoding/json"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesystemAvailabilitySerialization(t *testing.T) {
	var unavailableInfo Info
	require.NoError(t, json.Unmarshal([]byte(`{"u":1}`), &unavailableInfo))
	assert.Nil(t, unavailableInfo.ExtraFsAvailable)

	newAgentInfo, err := json.Marshal(Info{ExtraFsAvailable: map[string]bool{}})
	require.NoError(t, err)
	assert.Contains(t, string(newAgentInfo), `"efa":{}`)

	encoded, err := cbor.Marshal(Info{ExtraFsAvailable: map[string]bool{}})
	require.NoError(t, err)
	var decoded Info
	require.NoError(t, cbor.Unmarshal(encoded, &decoded))
	assert.NotNil(t, decoded.ExtraFsAvailable)
}

func TestCurrentForkAndUpstreamInfoCBOR(t *testing.T) {
	for _, tc := range []struct {
		name         string
		fields       map[uint64]any
		availability map[string]bool
		updates      []uint16
	}{
		{"upstream", map[uint64]any{25: []uint16{12, 3}}, nil, []uint16{12, 3}},
		{"updated fork", map[uint64]any{25: []uint16{12, 3}, 1000: map[string]bool{"backup": true}}, map[string]bool{"backup": true}, []uint16{12, 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := cbor.Marshal(tc.fields)
			require.NoError(t, err)
			var info Info
			require.NoError(t, cbor.Unmarshal(payload, &info))
			assert.Equal(t, tc.availability, info.ExtraFsAvailable)
			assert.Equal(t, tc.updates, info.PackageUpdates)
		})
	}
}

func TestCurrentCombinedDataCBOR(t *testing.T) {
	payload, err := cbor.Marshal(map[uint64]any{
		0: map[uint64]any{38: [2]uint64{100, 200}, 39: map[string]any{"tank": map[uint64]any{0: 100.0, 1: 60.0, 4: "ONLINE"}}},
		1: map[uint64]any{25: []uint16{12, 3}, 1000: map[string]bool{"backup": true}},
	})
	require.NoError(t, err)
	var decoded CombinedData
	require.NoError(t, cbor.Unmarshal(payload, &decoded))
	assert.Equal(t, [2]uint64{100, 200}, decoded.Stats.DiskIOTotal)
	assert.Equal(t, []uint16{12, 3}, decoded.Info.PackageUpdates)
	assert.Equal(t, map[string]bool{"backup": true}, decoded.Info.ExtraFsAvailable)
	require.Contains(t, decoded.Stats.ZfsPools, "tank")
	assert.Equal(t, "ONLINE", decoded.Stats.ZfsPools["tank"].Health)
	assert.Equal(t, 60.0, decoded.Stats.ZfsPools["tank"].Used)
}
