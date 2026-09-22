package celt

import "testing"

func TestUpdateStereoSavingPreservesRawEstimate(t *testing.T) {
	const frameSize = 960
	left := make([]celtNorm, frameSize)
	right := make([]celtNorm, frameSize)
	for band := 0; band < MaxBands; band++ {
		start := ScaledBandStart(band, frameSize)
		left[start] = 1
		right[start+1] = 1
	}
	// Orthogonal normalized bands make log2(1.001-correlation^2) positive.
	if got := UpdateStereoSaving(0, left, right, MaxBands, 3, 15); got >= 0 {
		t.Fatalf("uncorrelated stereo saving=%g, want a negative estimate", got)
	}
	// Sustained identical channels can accumulate a state above one; only
	// compute_vbr clamps its local copy, leaving this history intact.
	if got := UpdateStereoSaving(2, left, left, MaxBands, 3, 15); got != 2.25 {
		t.Fatalf("correlated stereo saving=%g, want 2.25", got)
	}
}
