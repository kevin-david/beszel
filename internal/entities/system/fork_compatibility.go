package system

import "github.com/fxamacker/cbor/v2"

// UnmarshalCBOR accepts the deployed fork's pool metrics before upstream's
// disk-I/O counters reused their field number. Info disambiguates reports in
// which pool collection failed but dataset measurements are still present.
func (data *CombinedData) UnmarshalCBOR(payload []byte) error {
	const statsKey = 0
	const infoKey = 1
	const legacyAvailabilityKey = 25
	const legacyPoolsKey = 38
	const poolsKey = 39
	const cborMapMajorType = 5
	var fields map[uint64]cbor.RawMessage
	if err := cbor.Unmarshal(payload, &fields); err != nil {
		return err
	}
	var info, stats map[uint64]cbor.RawMessage
	if raw := fields[infoKey]; len(raw) > 0 {
		if err := cbor.Unmarshal(raw, &info); err != nil {
			return err
		}
	}
	if raw := fields[statsKey]; len(raw) > 0 {
		if err := cbor.Unmarshal(raw, &stats); err != nil {
			return err
		}
	}
	availability, pools := info[legacyAvailabilityKey], stats[legacyPoolsKey]
	legacyFork := len(availability) > 0 && availability[0]>>5 == cborMapMajorType
	legacyPools := len(pools) > 0 && pools[0]>>5 == cborMapMajorType
	if legacyFork || legacyPools {
		// Legacy key 39 contains datasets, which must never become pool metrics.
		delete(stats, poolsKey)
		if legacyPools {
			stats[poolsKey] = pools
		}
		delete(stats, legacyPoolsKey)
		var err error
		fields[statsKey], err = cbor.Marshal(stats)
		if err != nil {
			return err
		}
		payload, err = cbor.Marshal(fields)
		if err != nil {
			return err
		}
	}
	type plainCombinedData CombinedData
	return cbor.Unmarshal(payload, (*plainCombinedData)(data))
}
