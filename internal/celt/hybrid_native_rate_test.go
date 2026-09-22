package celt

import "testing"

func TestHybridNativePCMUsesCoreFrame(t *testing.T) {
	e := NewEncoder(2)
	e.SetUpsample(2)
	got, frameSize := e.PrepareHybridPCM([]float32{1, 2, 3, 4}, 2)
	want := []float32{1, 2, 0, 0, 3, 4, 0, 0}
	if frameSize != 4 || len(got) != len(want) {
		t.Fatalf("core frame=%d samples=%d", frameSize, len(got))
	}
	for i, v := range want {
		if got[i] != v {
			t.Fatalf("sample[%d]=%g want=%g", i, got[i], v)
		}
	}
	coeffs := []float32{1, 2, 3, 4, 5, 6, 7, 8}
	e.ScaleHybridMDCT(coeffs, 4)
	for i, v := range []float32{2, 4, 0, 0, 10, 12, 0, 0} {
		if coeffs[i] != v {
			t.Fatalf("bin[%d]=%g want=%g", i, coeffs[i], v)
		}
	}
}

func TestHybridComplexityZeroDisablesTransientAnalysis(t *testing.T) {
	for _, channels := range []int{1, 2} {
		e := NewEncoder(channels)
		e.SetComplexity(0)
		e.EnsureScratch(480)
		pcm := make([]float32, 480*channels)
		for i := 200 * channels; i < 240*channels; i++ {
			pcm[i] = 10000
		}
		transient, weak, estimate, _, _, shortBlocks, energies := e.TransientAnalysisHybrid(pcm, 480, 19, 2, false)
		if transient || weak || estimate != 0 || shortBlocks != 1 || energies != nil {
			t.Fatalf("channels%d transient=%v weak=%v estimate=%g blocks=%d", channels, transient, weak, estimate, shortBlocks)
		}
	}
}
