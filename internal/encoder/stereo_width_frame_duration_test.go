package encoder

import (
	"fmt"
	"math"
	"testing"
)

func TestStereoWidthSmoothingFollowsFrameDuration(t *testing.T) {
	for _, frameSize := range []int{480, 960, 1920, 2880} {
		t.Run(fmt.Sprint(frameSize), func(t *testing.T) {
			enc := NewEncoder(48000, 2)
			// Equal channel energies and zero input make instantaneous width zero.
			// Existing width and peak history must decay by the actual frame duration.
			enc.widthMem = StereoWidthMem{XX: 1, YY: 1, SmoothedWidth: .25, MaxFollower: .5}
			enc.computeStereoWidthForMode(make([]opusRes, frameSize*2), frameSize)
			frameRate := float64(48000 / frameSize)
			wantSmooth := .25 * (1 - 1/frameRate)
			wantPeak := .5 - .02/frameRate
			if math.Abs(float64(enc.widthMem.SmoothedWidth)-wantSmooth) > 1e-7 {
				t.Errorf("smoothed width = %.9g, want %.9g", enc.widthMem.SmoothedWidth, wantSmooth)
			}
			if math.Abs(float64(enc.widthMem.MaxFollower)-wantPeak) > 1e-7 {
				t.Errorf("peak follower = %.9g, want %.9g", enc.widthMem.MaxFollower, wantPeak)
			}
		})
	}
}
