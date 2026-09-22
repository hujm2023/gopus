package silk

import (
	"testing"

	"github.com/hujm2023/gopus/internal/rangecoding"
)

func TestSILKDTXPacketLatch(t *testing.T) {
	e := NewEncoder(BandwidthNarrowband)
	e.SetDTX(true)
	for i := 0; i < 8; i++ {
		e.advanceDTXVAD(2)
	}
	e.ResetPacketState()
	for _, ready := range []bool{false, true, true} {
		e.advanceDTXVAD(2)
		if e.DTXReady() != ready || e.InDTX() {
			t.Fatal("threshold-crossing packet must remain coded")
		}
	}
	e.ResetPacketState()
	for i := 0; i < 3; i++ {
		e.advanceDTXVAD(2)
	}
	if !e.InDTX() {
		t.Fatal("next fully inactive packet should suppress")
	}
	// The 31st inactive frame refreshes and cannot re-enable the same packet.
	for i := 14; i < 30; i++ {
		e.advanceDTXVAD(2)
	}
	e.ResetPacketState()
	e.advanceDTXVAD(2)
	e.advanceDTXVAD(2)
	if e.InDTX() || !e.DTXReady() {
		t.Fatal("refresh latch or threshold mismatch")
	}
	e.ResetPacketState()
	e.advanceDTXVAD(200)
	e.advanceDTXVAD(2)
	if e.InDTX() || e.DTXReady() {
		t.Fatal("active subframe must reset both state machines")
	}
	e.Reset()
	if e.InDTX() || e.DTXReady() {
		t.Fatal("reset left DTX history")
	}
}

func TestSILKDTXPrefillAndDisabledHistory(t *testing.T) {
	e := NewEncoder(BandwidthNarrowband)
	for i := 0; i < 9; i++ {
		e.advanceDTXVAD(2)
	}
	e.ResetPacketState()
	e.SetVADState(2, 0, [4]int32{})
	e.PrefillFrame(make([]float32, 80))
	if !e.DTXReady() || e.InDTX() {
		t.Fatal("prefill must advance one VAD frame while DTX disabled")
	}
	e.SetDTX(true)
	e.ResetPacketState()
	e.advanceDTXVAD(2)
	if !e.InDTX() {
		t.Fatal("disabled DTX lost no-speech history")
	}
	e.ResetForBandwidthPrefill()
	if e.DTXReady() || e.InDTX() {
		t.Fatal("bandwidth prefill reset retained DTX history")
	}
}

func TestSILKDTXFrameCountsOnceAndSkipsTermination(t *testing.T) {
	e := NewEncoder(BandwidthNarrowband)
	e.SetDTX(true)
	e.SetComplexity(0)
	for i := 0; i < 9; i++ {
		e.advanceDTXVAD(2)
	}
	pcm := make([]float32, 160)
	e.SetVADState(2, 0, [4]int32{})
	if got := e.EncodeFrame(pcm, nil, false); len(got) == 0 || e.InDTX() || !e.DTXReady() {
		t.Fatal("tenth actual frame must be coded exactly once")
	}
	called := false
	e.SetPacketTermination(func(_ *rangecoding.Encoder) { called = true })
	e.SetVADState(2, 0, [4]int32{})
	if got := e.EncodeFrame(pcm, nil, false); len(got) != 0 || !e.InDTX() {
		t.Fatal("eleventh actual frame must suppress")
	}
	if called {
		t.Fatal("internal DTX executed termination callback")
	}
	if e.nFramesEncoded != 1 {
		t.Fatal("suppression skipped frame state")
	}
}

func TestSILKDTXMultiFrameReservoir(t *testing.T) {
	e := NewEncoder(BandwidthNarrowband)
	e.SetDTX(true)
	e.SetComplexity(0)
	e.SetBitrate(8000)
	for i := 0; i < 10; i++ {
		e.advanceDTXVAD(2)
	}
	e.nBitsExceeded = 1000
	states := []VADFrameState{{SpeechActivityQ8: 2, Valid: true}, {SpeechActivityQ8: 2, Valid: true}, {SpeechActivityQ8: 2, Valid: true}}
	if got := e.EncodePacketWithFECWithVADStates(make([]float32, 480), nil, []bool{false, false, false}, states); len(got) != 0 || !e.InDTX() {
		t.Fatal("inactive 60 ms packet was not suppressed")
	}
	if e.nBitsExceeded != 520 {
		t.Fatalf("DTX charged encoded bytes to reservoir: got %d want 520", e.nBitsExceeded)
	}
	if e.nFramesEncoded != 3 {
		t.Fatalf("encoded frames = %d want 3", e.nFramesEncoded)
	}
}

func TestSILKDTXStereoRequiresBothChannels(t *testing.T) {
	for _, sideEnabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "side-disabled", true: "both-enabled"}[sideEnabled], func(t *testing.T) {
			mid, side := NewEncoder(BandwidthNarrowband), NewEncoder(BandwidthNarrowband)
			mid.SetDTX(true)
			side.SetDTX(sideEnabled)
			for _, e := range []*Encoder{mid, side} {
				e.SetComplexity(0)
				e.SetBitrate(16000)
				for i := 0; i < 10; i++ {
					e.advanceDTXVAD(2)
				}
			}
			called := false
			mid.SetPacketTermination(func(_ *rangecoding.Encoder) { called = true })
			states := []VADFrameState{{SpeechActivityQ8: 2, Valid: true}}
			pcm := make([]float32, 160)
			got, err := EncodeStereoWithEncoderVADFlagsAndStatesWithSide(mid, side, pcm, pcm, BandwidthNarrowband, []bool{false}, states, []bool{false}, states)
			if err != nil {
				t.Fatal(err)
			}
			if (len(got) == 0) != sideEnabled || called == sideEnabled {
				t.Fatalf("side DTX=%v payload=%d termination=%v", sideEnabled, len(got), called)
			}
		})
	}
}
