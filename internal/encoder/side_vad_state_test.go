package encoder

import (
	"math"
	"testing"
)

func TestSideVADReturnsOpusAdjustedState(t *testing.T) {
	pcm := make([]float32, 320)
	for i := range pcm {
		pcm[i] = float32(.6 * math.Sin(2*math.Pi*440*float64(i)/16000))
	}
	for _, active := range []bool{false, true} {
		name := "inactive"
		if active {
			name = "active"
		}
		t.Run(name, func(t *testing.T) {
			e := NewEncoder(48000, 2)
			e.lastOpusVADValid = true
			e.lastOpusVADActive = active
			state, flag := e.computeSilkVADSide(pcm, len(pcm), 16)
			raw := e.silkVADSide
			if !state.Valid || raw.SpeechActivityQ8 < speechActivityThresholdQ8 {
				t.Fatal("fixture must produce active SILK VAD")
			}
			want := raw.SpeechActivityQ8
			if !active {
				want = speechActivityThresholdQ8 - 1
			}
			if state.SpeechActivityQ8 != want || flag != active {
				t.Fatalf("activity=%d flag=%v, want %d/%v", state.SpeechActivityQ8, flag, want, active)
			}
			if state.InputTiltQ15 != raw.InputTiltQ15 || state.InputQualityBandsQ15 != raw.InputQualityBandsQ15 {
				t.Fatal("adjusted state lost VAD shaping controls")
			}
		})
	}
}
