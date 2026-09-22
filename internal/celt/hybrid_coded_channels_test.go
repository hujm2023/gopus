package celt

import (
	"fmt"
	"testing"
)

func TestHybridCodedChannelSpectrum(t *testing.T) {
	const frameSize = 960
	for _, channels := range []int{1, 2} {
		for _, upsample := range []int{1, 2} {
			t.Run(fmt.Sprintf("coded%d/upsample%d", channels, upsample), func(t *testing.T) {
				e := NewEncoder(2)
				e.SetStreamChannels(channels)
				e.SetUpsample(upsample)
				e.EnsureScratch(frameSize)
				coeffs := make([]float32, frameSize*2)
				want := make([]float32, frameSize*channels)
				for i := range frameSize {
					coeffs[i] = float32(i+1) * .125
					coeffs[frameSize+i] = float32(i-7) * .25
				}
				for c := range channels {
					for i := range frameSize {
						if i < frameSize/upsample {
							v := coeffs[c*frameSize+i]
							if channels == 1 {
								v = .5*coeffs[i] + .5*coeffs[frameSize+i]
							}
							want[c*frameSize+i] = v * float32(upsample)
						}
					}
				}
				got := e.PrepareHybridMDCT(coeffs, frameSize)
				if len(got) != len(want) {
					t.Fatalf("spectrum length=%d want%d", len(got), len(want))
				}
				for i := range want {
					if got[i] != want[i] {
						t.Fatalf("bin%d=%g want%g", i, got[i], want[i])
					}
				}
				for _, bands := range []int{19, 21} {
					energy := e.ComputeBandEnergiesF32(got, bands, frameSize)
					if len(energy) != bands*channels {
						t.Fatalf("energy length=%d want%d", len(energy), bands*channels)
					}
					wantEnergy := make([]celtGLog, len(energy))
					computeBandEnergiesGLogF32Into(want, bands, frameSize, channels, 8, wantEnergy)
					for i := range energy {
						if energy[i] != wantEnergy[i] {
							t.Fatalf("energy%d=%g want%g", i, energy[i], wantEnergy[i])
						}
					}
				}
			})
		}
	}
	t.Run("transient_long_analysis", func(t *testing.T) {
		e := NewEncoder(2)
		e.SetStreamChannels(1)
		e.SetComplexity(10)
		e.EnsureScratch(frameSize)
		pcm := make([]float32, frameSize*2)
		for i := 480; i < 540; i++ {
			pcm[2*i] = 10000
			pcm[2*i+1] = -10000
		}
		transient, _, _, _, _, _, energy := e.TransientAnalysisHybrid(pcm, frameSize, 21, 3, false)
		if !transient || len(energy) != 21 {
			t.Fatalf("transient=%v energies=%d", transient, len(energy))
		}
		want := make([]celtGLog, 21)
		computeBandEnergiesGLogF32Into(make([]float32, frameSize), 21, frameSize, 1, 8, want)
		for i := range want {
			want[i] += 1.5
			if energy[i] != want[i] {
				t.Fatalf("antiphase mono energy%d=%g want%g", i, energy[i], want[i])
			}
		}
	})
}
