package encoder

import "testing"

func TestSILKInputBitrateFrameBudget(t *testing.T) {
	// opus_encoder.c computes bits_to_bitrate(bitrate_to_bits(rate)-8).
	for _, tc := range []struct {
		rate, frameSize, want int
	}{
		{6000, 960, 5600},
		{6001, 960, 5600},
		{6000, 1920, 5800},
		{6000, 2880, 5866},
		{0, 2880, 0},
		{6000, 0, 0},
		{100, 960, 0},
	} {
		for _, fs := range []int{8000, 12000, 16000, 24000, 48000} {
			e := Encoder{sampleRate: int32(fs), bitrate: int32(tc.rate)}
			frameSize := tc.frameSize * fs / 48000
			if got := e.silkInputBitrate(frameSize); got != tc.want {
				t.Errorf("rate=%d fs=%d frame=%d: got %d want %d", tc.rate, fs, frameSize, got, tc.want)
			}
		}
	}
}
