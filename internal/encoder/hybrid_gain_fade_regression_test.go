package encoder

import (
	"math"
	"testing"
)

func TestHybridGainFadeReferenceRounding(t *testing.T) {
	// Values from libopus 1.6.1 gain_fade and window120, float build.
	for _, tc := range []struct {
		name              string
		rate              int
		previous, current opusVal16
		index             int
		want              uint32
	}{
		{"equal gains retain window rounding", 48000, .9993, .9993, 5, 0x3e7fd221},
		{"native 24k window stride", 24000, 1, .5, 30, 0x3e3eaf26},
		{"native 24k final fade sample", 24000, 1, .5, 59, 0x3e000003},
		{"native 24k fade ends at 60", 24000, 1, .5, 60, 0x3e000000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEncoder(tc.rate, 1)
			e.hybridState = &HybridState{prevHBGain: tc.previous}
			samples := make([]opusRes, tc.rate/50)
			for i := range samples {
				samples[i] = .25
			}
			got := e.applyHBGainFade(samples, tc.current)
			if bits := math.Float32bits(float32(got[tc.index])); bits != tc.want {
				t.Fatalf("sample %d bits=%08x want=%08x", tc.index, bits, tc.want)
			}
		})
	}
}
