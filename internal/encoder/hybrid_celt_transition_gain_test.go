package encoder

import "testing"

func TestCELTTransitionRestoresHybridHighBandGain(t *testing.T) {
	e := NewEncoder(48000, 1)
	e.hybridState = &HybridState{prevHBGain: .5}
	pcm := make([]opusRes, 960)
	for i := range pcm {
		pcm[i] = 1
	}
	got := e.applyNonHybridStereoWidthFade(pcm, 960, ModeCELT)
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
	got = e.applyNonHybridStereoWidthFade(pcm, 960, ModeCELT)
	for i, v := range got {
		if v != 1 {
			t.Fatalf("steady CELT sample[%d]=%g want 1", i, v)
		}
	}
}

func TestSILKFrameRetainsWidthForHybrid(t *testing.T) {
	e := NewEncoder(48000, 2)
	e.SetBitrate(24000)
	e.SetPacketLoss(20)
	e.SetMode(ModeSILK)
	pcm := make([]opusRes, 1920)
	for i := range pcm {
		pcm[i] = .1
	}
	data, err := e.encodeSILKFrameWithDREDAndMax(pcm, nil, 960, 24000, 0, 1276)
	if err != nil || len(data) == 0 {
		t.Fatalf("encode: bytes=%d err=%v", len(data), err)
	}
	if e.hybridState == nil || e.hybridState.stereoWidthQ14 >= 16384 {
		t.Fatal("coded SILK frame did not retain its reduced width for Hybrid")
	}
	silkWidth := e.hybridState.stereoWidthQ14
	e.applyNonHybridStereoWidthFade(nil, 960, ModeCELT)
	if e.hybridState.stereoWidthQ14 == silkWidth {
		t.Fatal("SILK and CELT effective rates incorrectly use the same mode")
	}
}
