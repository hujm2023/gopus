package celt

import "testing"

func TestVBRPayloadCapExcludesTOC(t *testing.T) {
	e := NewEncoder(1)
	for _, tc := range []struct{ frameSize, want int }{{120, 159}, {240, 318}, {480, 637}, {960, 1275}} {
		if got := e.vbrMaxPayloadBytes(tc.frameSize); got != tc.want {
			t.Errorf("frameSize=%d payload cap=%d want=%d", tc.frameSize, got, tc.want)
		}
	}
}
