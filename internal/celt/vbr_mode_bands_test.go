package celt

import "testing"

func TestSurroundVBRUsesFullModeBands(t *testing.T) {
	enc := NewEncoder(2)
	enc.SetBandwidth(CELTWideband)
	enc.SetEnergyMask(make([]float32, 2*MaxBands))
	enc.intensity = 9
	enc.lastStereoSaving = .028437763
	enc.analysisValid = true
	enc.analysisActivity = .552490354
	enc.analysisTonality = .279120386
	enc.lastDynalloc.MaxDepth = 25.6757622
	enc.surroundMasking = -.537500024
	// libopus 1.6.1 surround 6.1, 20ms/64kbps CVBR, frame 0 stream 2:
	// compute_vbr returns 3308 Q3 before the shared entropy coder's 1232 Q3 tell.
	if got := enc.computeVBRTargetWithBoost(2368, 960, .992824256, false, 1344, 19800); got != 3308 {
		t.Fatalf("first-frame VBR target = %d, want oracle 3308", got)
	}
	// A remembered coded-band count still controls shaping, while the depth
	// ceiling remains a property of the complete mode at every bandwidth.
	enc.lastCodedBands = 17
	enc.lastDynalloc.MaxDepth = .1
	wide := enc.computeVBRTargetWithBoost(2368, 960, .992824256, false, 1344, 19800)
	enc.SetBandwidth(CELTFullband)
	full := enc.computeVBRTargetWithBoost(2368, 960, .992824256, false, 1344, 19800)
	if wide != full {
		t.Fatalf("same coded history and depth changed target with bandwidth: wide=%d full=%d", wide, full)
	}
}
