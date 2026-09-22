package encoder

import (
	"slices"
	"testing"
)

func TestRedundancyWidthFadePreservesMainInput(t *testing.T) {
	e := NewEncoder(48000, 2)
	e.hybridState = &HybridState{prevHBGain: 1, stereoWidthQ14: 16384, silkStereoWidthQ14: 0}
	pcm := make([]opusRes, 960)
	for i := 0; i < len(pcm); i += 2 {
		pcm[i] = .25
		pcm[i+1] = -.25
	}
	original := slices.Clone(pcm)
	redundancy := e.prepareCELTTransitionRedundancyInput(pcm, 1)
	redundancy = e.applyStereoWidthFade(redundancy, 16384, 0)
	if !slices.Equal(pcm, original) {
		t.Fatal("redundancy processing changed main input")
	}
	if len(redundancy) != 480 || redundancy[478] != 0 || redundancy[479] != 0 {
		t.Fatal("redundancy did not collapse to mono after the fade")
	}
	if e.hybridState.stereoWidthQ14 != 16384 {
		t.Fatal("redundancy advanced the width needed by main processing")
	}
}
