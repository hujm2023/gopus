package celt

import "testing"

func TestResetClearsTemporalVBRHistory(t *testing.T) {
	enc := NewEncoder(1)
	enc.specAvg = 0.17344456
	enc.lastTemporalVBR = 2.7487864

	// Opus mode transitions call Reset before prefilling CELT again.
	enc.Reset()
	if enc.specAvg != 0 || enc.lastTemporalVBR != 0 {
		t.Fatalf("Reset retained temporal history: spectral average=%v temporal VBR=%v", enc.specAvg, enc.lastTemporalVBR)
	}
}
