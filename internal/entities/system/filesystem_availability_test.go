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
