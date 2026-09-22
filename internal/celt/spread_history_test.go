package celt

import (
	"math"
	"testing"
)

func TestEncodeFrameSpreadHistory(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		complexity, bitrate     int
		silence, transient, lfe bool
		want                    int32
	}{
		{name: "silence", complexity: 10, bitrate: 24000, silence: true, want: spreadNormal},
		{name: "complexity_zero", complexity: 0, bitrate: 24000, want: spreadNone},
		{name: "complexity_two", complexity: 2, bitrate: 24000, want: spreadNormal},
		{name: "low_rate", complexity: 10, bitrate: 3200, want: spreadNormal},
		{name: "transient", complexity: 10, bitrate: 24000, transient: true, want: spreadNormal},
		{name: "lfe", complexity: 10, bitrate: 24000, lfe: true, want: spreadNormal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEncoder(1)
			e.SetBitrate(tc.bitrate)
			e.SetComplexity(tc.complexity)
			e.SetLFE(tc.lfe)
			e.spreadDecision = spreadAggressive
			pcm := make([]float32, 960)
			if !tc.silence {
				for i := range pcm {
					if !tc.transient || i >= 480 {
						pcm[i] = float32(.2 * math.Sin(float64(i)*.17))
					}
				}
			}
			if _, err := e.EncodeFrame(pcm, 960); err != nil {
				t.Fatal(err)
			}
			if tc.transient && e.consecTransient == 0 {
				t.Fatal("fixture did not select transient branch")
			}
			if e.spreadDecision != tc.want {
				t.Fatalf("spread history=%d, want %d", e.spreadDecision, tc.want)
			}
		})
	}
}
