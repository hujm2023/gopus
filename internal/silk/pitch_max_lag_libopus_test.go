package silk

import (
	"math"
	"testing"

	"github.com/hujm2023/gopus/internal/libopustest"
)

func TestPitchAnalysisRetainsMaximumCoarseLag(t *testing.T) {
	libopustest.RequireOracle(t)
	// A sharp 18 ms periodic signal puts the strongest coarse candidate at
	// lag 72. A narrow candidate threshold isolates its expansion to lag 143.
	frame := make([]float32, 480)
	for i := range frame {
		for harmonic := 1; harmonic <= 12; harmonic++ {
			frame[i] += float32(700 * math.Sin(2*math.Pi*float64(harmonic*i)/288))
		}
	}
	tc := libopusSILKPitchAnalysisCase{
		bandwidth:    BandwidthWideband,
		nbSubfr:      2,
		complexity:   2,
		prevLag:      96,
		ltpCorr:      0.852043986,
		searchThres1: 0.99,
		searchThres2: 0.262,
		frame:        frame,
	}
	want, _, err := probeLibopusSILKPitchAnalysis([]libopusSILKPitchAnalysisCase{tc})
	if err != nil {
		t.Fatal(err)
	}
	enc := NewEncoder(tc.bandwidth)
	enc.pitchEstimationComplexity = int32(tc.complexity)
	enc.pitchState.prevLag = int32(tc.prevLag)
	enc.pitchState.ltpCorr = tc.ltpCorr
	lags, lagIndex, contourIndex := enc.detectPitch(tc.frame, tc.nbSubfr, tc.searchThres1, tc.searchThres2)
	if want[0].ret != 0 {
		t.Fatalf("oracle did not classify boundary pitch as voiced: %+v", want[0])
	}
	for i, got := range lags {
		if got != int32(want[0].pitchOut[i]) {
			t.Fatalf("pitch[%d]=%d want %d", i, got, want[0].pitchOut[i])
		}
	}
	if lagIndex != want[0].lagIndex || contourIndex != want[0].contourIndex || enc.pitchState.ltpCorr != want[0].ltpCorr {
		t.Fatalf("got index=%d contour=%d corr=%g want %+v", lagIndex, contourIndex, enc.pitchState.ltpCorr, want[0])
	}
}
