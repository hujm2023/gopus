package encoder

import (
	"github.com/hujm2023/gopus/types"
	"slices"
	"testing"
)

func TestSILKDTXSelection(t *testing.T) {
	e := NewEncoder(48000, 1)
	e.SetDTX(true)
	for _, tc := range []struct{ analysis, silence, internal bool }{
		{false, false, true}, {false, true, false}, {true, false, false},
	} {
		e.lastAnalysisValid = tc.analysis
		e.updateSILKDTXMode(tc.silence)
		e.ensureSILKEncoder()
		if e.silkUseDTX != tc.internal {
			t.Fatalf("selection %+v: got %v", tc, e.silkUseDTX)
		}
		e.silkEncoder.ResetPacketState()
		if e.silkEncoder.InDTX() != tc.internal {
			t.Fatal("SILK did not receive the packet DTX selection")
		}
		if tc.internal && e.decideDTXSuppress(false, 2880) {
			t.Fatal("Opus DTX ran alongside SILK DTX")
		}
	}
}

func TestSILKInternalDTXPreservesDelayHistory(t *testing.T) {
	for _, tc := range []struct {
		mode Mode
		size int
		bw   types.Bandwidth
	}{
		{ModeSILK, 960, types.BandwidthWideband},
		{ModeHybrid, 960, types.BandwidthSuperwideband},
		{ModeHybrid, 2880, types.BandwidthSuperwideband},
	} {
		e := NewEncoder(48000, 2)
		e.SetMode(tc.mode)
		e.SetBandwidth(tc.bw)
		e.SetMaxBandwidth(tc.bw)
		e.SetForceChannels(2)
		e.SetComplexity(0)
		e.SetBitrate(32000)
		e.SetDTX(true)
		pcm := make([]opusRes, tc.size*2)
		suppressed := false
		for f := 0; f < 35; f++ {
			for i := range pcm {
				pcm[i] = .00001 * float32(1+f%3)
			}
			before := slices.Clone(e.delayBuffer)
			packet, err := e.Encode(pcm, tc.size)
			if err != nil {
				t.Fatal(err)
			}
			if len(packet) <= 2 && e.silkUseDTX {
				if !slices.Equal(before, e.delayBuffer) {
					t.Fatalf("mode=%v size=%d: suppressed packet advanced delay history", tc.mode, tc.size)
				}
				suppressed = true
				break
			}
		}
		if !suppressed {
			t.Fatalf("mode=%v size=%d: fixture did not enter native DTX", tc.mode, tc.size)
		}
	}
}
