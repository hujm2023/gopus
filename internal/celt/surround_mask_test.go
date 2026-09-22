package celt

import "testing"

func TestEncoderSetEnergyMask(t *testing.T) {
	enc := NewEncoder(2)

	mask := make([]float32, 2*MaxBands)
	for i := range mask {
		mask[i] = float32(i) * 0.01
	}
	enc.SetEnergyMask(mask)
	got := enc.EnergyMask()
	if len(got) != len(mask) {
		t.Fatalf("EnergyMask len=%d want=%d", len(got), len(mask))
	}
	for i := range mask {
		want := celtGLog(mask[i])
		if got[i] != want {
			t.Fatalf("EnergyMask[%d]=%f want=%f", i, got[i], want)
		}
	}

	enc.SetEnergyMask(mask[:MaxBands-1])
	if len(enc.EnergyMask()) != 0 {
		t.Fatalf("invalid-size mask should clear state, got len=%d", len(enc.EnergyMask()))
	}
}

func TestComputeSurroundDynallocFromMask(t *testing.T) {
	enc := NewEncoder(2)
	enc.lastCodedBands = 17

	mask := make([]float32, 2*MaxBands)
	for i := range MaxBands {
		mask[i] = -2.0
		mask[MaxBands+i] = -2.0
	}
	mask[5] = 0
	mask[6] = 0
	mask[MaxBands+5] = 0
	mask[MaxBands+6] = 0
	enc.SetEnergyMask(mask)

	out := make([]celtGLog, MaxBands)
	trim, ok := enc.computeSurroundDynallocFromMask(MaxBands, out)
	if !ok {
		t.Fatalf("computeSurroundDynallocFromMask returned ok=false")
	}
	if trim == 0 {
		t.Fatalf("expected non-zero surround trim from mask")
	}
	nonZero := 0
	for i := range MaxBands {
		if out[i] > 0 {
			nonZero++
		}
	}
	if nonZero == 0 {
		t.Fatalf("expected non-zero surround dynalloc bands")
	}
}

func TestSurroundMaskFeedsVBRTarget(t *testing.T) {
	enc := NewEncoder(1)
	enc.lastCodedBands = 17
	enc.lastDynalloc.MaxDepth = 100
	mask := make([]float32, MaxBands)
	for i := range mask {
		mask[i] = -2
	}
	baseline := enc.computeVBRTarget(20000, 960, .044, false)
	enc.SetEnergyMask(mask)
	enc.computeSurroundDynallocFromMask(MaxBands, make([]celtGLog, MaxBands))
	if enc.surroundMasking < -1.600001 || enc.surroundMasking > -1.599999 {
		t.Fatalf("mask=%g, want -1.6 after libopus's two 0.2 offsets", enc.surroundMasking)
	}
	if got := enc.computeVBRTarget(20000, 960, .044, false); got >= baseline {
		t.Fatalf("surround masking did not reduce target: got=%d baseline=%d", got, baseline)
	}
	enc.SetEnergyMask(nil)
	enc.computeSurroundDynallocFromMask(MaxBands, make([]celtGLog, MaxBands))
	if enc.surroundMasking != 0 || enc.computeVBRTarget(20000, 960, .044, false) != baseline {
		t.Fatal("removed mask left a VBR adjustment")
	}
}

func TestLFEVBRUsesAnalysisBoost(t *testing.T) {
	enc := NewEncoder(1)
	enc.SetVBR(true)
	enc.SetConstrainedVBR(true)
	enc.SetLFE(true)
	enc.SetBitrate(8150)
	enc.lastDynalloc.MaxDepth = 20.167902
	// libopus's LFE analysis boost is zero even when coding band zero
	// consumes 288 Q3 boost bits. Those bits affect only the minimum size.
	enc.computeFinalVBRTargetBytes(960, 0, false, 395, 288, 40)
	if enc.vbrOffset != 9 {
		t.Fatalf("LFE VBR offset=%d, want libopus trace value 9", enc.vbrOffset)
	}
}
