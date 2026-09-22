package encoder

import (
	"testing"

	"github.com/hujm2023/gopus/types"
)

func TestSurroundSILKRateAndGain(t *testing.T) {
	for _, tc := range []struct {
		name     string
		channels int
		bw       types.Bandwidth
		mask     float32
		hybrid   bool
		want     int
	}{
		{"NB floor", 1, types.BandwidthNarrowband, -2, false, 3334},
		{"WB stereo", 2, types.BandwidthWideband, .5, false, 29200},
		{"Hybrid split", 2, types.BandwidthFullband, .5, true, 21520},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEncoder(48000, tc.channels)
			e.SetBandwidth(tc.bw)
			e.SetBitrateMode(ModeVBR)
			mask := make([]float32, 21*tc.channels)
			for i := range mask {
				mask[i] = tc.mask
			}
			e.SetCELTEnergyMask(mask)
			if got := e.surroundSILKBitrate(10000, tc.hybrid); got != tc.want {
				t.Fatalf("rate=%d want=%d", got, tc.want)
			}
			if e.computeHBGain(2000) != 1 {
				t.Fatal("surround must retain full high-band gain")
			}
			e.SetBitrateMode(ModeCBR)
			if e.surroundSILKBitrate(10000, tc.hybrid) != 10000 {
				t.Fatal("CBR rate changed")
			}
		})
	}
}
