package celt

import "testing"

func TestPreemphasisSilenceUsesCodedChannels(t *testing.T) {
	for _, twoTap := range []bool{false, true} {
		name := "single_tap"
		if twoTap {
			name = "two_tap"
		}
		t.Run(name, func(t *testing.T) {
			enc := NewEncoder(2)
			enc.SetStreamChannels(1)
			if twoTap {
				enc.hd96kPreemph = [4]float32{0.8, 0.1, 0.9, 1}
			}
			pcm := make([]float32, 240)
			pcm[144] = 0.25
			pcm[145] = -0.5
			out := make([]float32, len(pcm))
			if !enc.applyPreemphasisWithScalingAndSilenceCore(pcm, out, 120, 120) {
				t.Fatal("coded mono prefix is silent even though the input tail is not")
			}
			if out[144] == 0 || out[145] == 0 {
				t.Fatal("silence scan must not truncate either input channel's pre-emphasis")
			}
			enc.overlapMax = 0.25
			clear(pcm)
			if enc.applyPreemphasisWithScalingAndSilenceCore(pcm, out, 120, 120) {
				t.Fatal("previous overlap energy must prevent silence")
			}
			if !enc.applyPreemphasisWithScalingAndSilenceCore(pcm, out, 120, 120) {
				t.Fatal("silent overlap must replace previous overlap energy")
			}
			enc.SetStreamChannels(2)
			pcm[144] = 0.25
			if enc.applyPreemphasisWithScalingAndSilenceCore(pcm, out, 120, 120) {
				t.Fatal("coded stereo must scan the complete input")
			}
		})
	}
}
