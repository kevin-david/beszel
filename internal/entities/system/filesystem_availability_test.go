package system

import (
	"encoding/json"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesystemAvailabilityJSONCompatibility(t *testing.T) {
	var oldAgentInfo Info
	require.NoError(t, json.Unmarshal([]byte(`{"u":1}`), &oldAgentInfo))
	assert.Nil(t, oldAgentInfo.ExtraFsAvailable)

	newAgentInfo, err := json.Marshal(Info{ExtraFsAvailable: map[string]bool{}})
	require.NoError(t, err)
	assert.Contains(t, string(newAgentInfo), `"efa":{}`)

	encoded, err := cbor.Marshal(Info{ExtraFsAvailable: map[string]bool{}})
	require.NoError(t, err)
	var decoded Info
	require.NoError(t, cbor.Unmarshal(encoded, &decoded))
	assert.NotNil(t, decoded.ExtraFsAvailable)
}

func TestLegacyForkAndUpstreamInfoCBOR(t *testing.T) {
	for _, tc := range []struct {
		name         string
		fields       map[uint64]any
		availability map[string]bool
		updates      []uint16
	}{
		{"legacy fork", map[uint64]any{25: map[string]bool{"backup": false}}, map[string]bool{"backup": false}, nil},
		{"legacy fork empty inventory", map[uint64]any{25: map[string]bool{}}, map[string]bool{}, nil},
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
