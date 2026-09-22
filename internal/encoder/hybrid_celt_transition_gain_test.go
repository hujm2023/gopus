package encoder

import "testing"

func TestCELTTransitionRestoresHybridHighBandGain(t *testing.T) {
	e := NewEncoder(48000, 1)
	e.hybridState = &HybridState{prevHBGain: .5}
	pcm := make([]opusRes, 960)
	for i := range pcm {
		pcm[i] = 1
	}
	got := e.applyCELTStereoWidthFade(pcm, 960)
	// The shared CELT preprocessing runs for mono too: the first sample retains
	// the Hybrid gain, then the 2.5 ms fade restores the full CELT input level.
	if got[0] != .5 || got[120] != 1 {
		t.Fatalf("transition gain: first=%g after overlap=%g", got[0], got[120])
	}
	if e.hybridState.prevHBGain != 1 {
		t.Fatalf("previous HB gain=%g, want 1", e.hybridState.prevHBGain)
	}
	for i := range pcm {
		pcm[i] = 1
	}
	got = e.applyCELTStereoWidthFade(pcm, 960)
	for i, v := range got {
		if v != 1 {
			t.Fatalf("steady CELT sample[%d]=%g want 1", i, v)
		}
	}
}
