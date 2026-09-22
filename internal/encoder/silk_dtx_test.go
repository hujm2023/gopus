package encoder

import "testing"

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
