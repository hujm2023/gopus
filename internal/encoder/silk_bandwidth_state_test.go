package encoder

import (
	"github.com/hujm2023/gopus/internal/silk"
	"github.com/hujm2023/gopus/types"
	"testing"
)

func TestBandwidthControlRetainsSideFilterHistory(t *testing.T) {
	e := NewEncoder(48000, 2)
	e.SetBandwidth(types.BandwidthWideband)
	e.silkInternalRate = 16000
	e.ensureSILKEncoder()
	e.ensureSILKSideEncoder()
	e.first = false
	e.prevMode = ModeSILK
	main := silk.LPState{Mode: 1, TransitionFrameNo: 17, InLPState: [2]int32{10, 11}}
	side := silk.LPState{Mode: 1, TransitionFrameNo: 8, InLPState: [2]int32{20, 21}}
	e.silkEncoder.SetLPState(main)
	e.silkSideEncoder.SetLPState(side)
	e.controlSILKBandwidth(ModeSILK)
	if got := e.silkSideEncoder.GetLPState(); got.InLPState != side.InLPState || got.TransitionFrameNo != side.TransitionFrameNo {
		t.Fatalf("side history overwritten: got %+v want %+v", got, side)
	}
}
