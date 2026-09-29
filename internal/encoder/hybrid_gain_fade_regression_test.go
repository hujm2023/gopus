package encoder

import (
	"bytes"
	"fmt"
	"math"
	"os/exec"
	"testing"

	"github.com/hujm2023/gopus/internal/libopustest"
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

func TestHybridGainFadeMatchesLibopus(t *testing.T) {
	libopustest.RequireOracle(t)
	bin, err := libopustest.BuildCHelper(libopustest.CHelperConfig{
		Label:      "gain fade",
		OutputBase: "gopus_libopus_gain_fade",
		SourceFile: "libopus_gain_fade_info.c",
		CFlags:     []string{"-DHAVE_CONFIG_H", "-O3", "-DNDEBUG"},
		Libs:       []string{libopustest.RefPath(".libs", "libopus.a"), "-lm"},
	})
	if err != nil {
		libopustest.HelperUnavailable(t, "gain fade", err)
	}
	for _, rate := range []int{48000, 24000} {
		for _, channels := range []int{1, 2} {
			for _, gains := range [][2]opusVal16{{.9993, .9993}, {1, .5}, {.7, .3}, {.25, .9993}} {
				t.Run(fmt.Sprintf("%dHz/%dch/%g-%g", rate, channels, gains[0], gains[1]), func(t *testing.T) {
					out, err := exec.Command(bin, fmt.Sprint(rate), fmt.Sprint(channels), fmt.Sprint(gains[0]), fmt.Sprint(gains[1])).Output()
					if err != nil {
						t.Fatal(err)
					}
					e := NewEncoder(rate, channels)
					e.hybridState = &HybridState{prevHBGain: gains[0]}
					samples := make([]opusRes, rate/50*channels)
					for i := range samples {
						samples[i] = .25
					}
					got := e.applyHBGainFade(samples, gains[1])
					reader := bytes.NewReader(out)
					for i, sample := range got {
						var want uint32
						if _, err := fmt.Fscanf(reader, "%x\n", &want); err != nil {
							t.Fatalf("oracle sample %d: %v", i, err)
						}
						if bits := math.Float32bits(sample); bits != want {
							t.Fatalf("sample %d bits=%08x want libopus=%08x", i, bits, want)
						}
					}
					if reader.Len() != 0 {
						t.Fatalf("oracle has %d trailing bytes", reader.Len())
					}
				})
			}
		}
	}
}
