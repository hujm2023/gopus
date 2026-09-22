package celt

import "testing"

func TestTemporalVBRUsesEquivalentRateAndPacketCap(t *testing.T) {
	enc := NewEncoder(2)
	enc.SetVBR(true)
	enc.SetConstrainedVBR(true)
	enc.SetBitrate(128000)
	enc.lastCodedBands = 3
	enc.intensity = 19
	enc.lastStereoSaving = -0.000721026212
	enc.lastTemporalVBR = 3
	enc.lastDynalloc.MaxDepth = 24.6021194
	// C compute_vbr's 2.5 ms equivalent rate is 93000 after 35000 bps
	// overhead. The traced target is 1619 before its +45 temporal boost.
	if got := enc.computeVBRTarget(1680, 120, 0, false); got != 1664 {
		t.Fatalf("short-frame temporal target=%d, want 1664", got)
	}
	// A caller cap of 30 bytes lowers equivalent rate further to 61000;
	// the temporal boost reaches its 32000-bps calibration ceiling.
	enc.SetMaxPayloadBytes(30)
	if got := enc.computeVBRTarget(1680, 120, 0, false); got != 2100 {
		t.Fatalf("buffer-limited temporal target=%d, want 2100", got)
	}
}
