package silk

import "testing"

func TestLowRateControlMatchesUnorderedLimits(t *testing.T) {
	for _, tc := range []struct{ rate, excess, want int32 }{
		{4930, 0, 4930}, {4930, -100, 5000}, {6400, 0, 6400}, {6400, 1000, 5000},
	} {
		mono := NewEncoder(BandwidthWideband)
		mono.SetBitrate(int(tc.rate))
		mono.nBitsExceeded = tc.excess
		mono.EncodeFrame(make([]float32, 320), nil, false)
		if mono.lastControlTargetRateBps != tc.want {
			t.Fatalf("mono rate=%d excess=%d: got %d want %d", tc.rate, tc.excess, mono.lastControlTargetRateBps, tc.want)
		}
		stereo := NewEncoder(BandwidthWideband)
		stereo.nBitsExceeded = tc.excess
		if got := stereoAllocationTargetRate(stereo, int(tc.rate), 320, 0); got != int(tc.want) {
			t.Fatalf("stereo rate=%d excess=%d: got %d want %d", tc.rate, tc.excess, got, tc.want)
		}
	}
}
