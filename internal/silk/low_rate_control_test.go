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

func TestStereoPrefillRateIgnoresPriorPacketLayout(t *testing.T) {
	e := NewEncoder(BandwidthMediumband)
	e.nFramesPerPacket = 3
	e.nFramesEncoded = 2
	e.nBitsUsedLBRR = 400
	e.nBitsExceeded = 100
	if got := e.StereoPrefillTargetRate(23867); got != 23600 {
		t.Fatalf("10 ms prefill target=%d want 23600", got)
	}
	if e.nBitsUsedLBRR != 400 || e.nFramesPerPacket != 3 || e.nBitsExceeded != 100 {
		t.Fatal("prefill calculation changed the pending packet state")
	}
}

func TestTransitionPrefillAdvancesLBRRControl(t *testing.T) {
	e := NewEncoder(BandwidthWideband)
	e.SetFEC(true)
	e.SetPacketLoss(20)
	e.ResetTransitionPrefillState()
	// No PrefillFrame call: even a mid-only side channel runs the controls.
	e.ResetPacketState()
	if e.lbrrGainIncreases != 4 {
		t.Fatalf("post-prefill LBRR gain=%d want 4", e.lbrrGainIncreases)
	}
	e.SetFEC(false)
	e.ResetTransitionPrefillState()
	e.SetFEC(true)
	e.ResetPacketState()
	if e.lbrrGainIncreases != 7 {
		t.Fatalf("disabled prefill retained LBRR history: gain=%d want 7", e.lbrrGainIncreases)
	}
}

// TestLTPScaleUsesCurrentPacketLBRRFlag pins the libopus silk_LTP_scale_ctrl_FLP
// condition: the round-loss reduction depends on psEnc->sCmn.LBRR_flag, i.e.
// whether THIS packet actually carries LBRR data, not on whether LBRR was enabled
// for an earlier packet. silk/fixed/LTP_scale_ctrl_FIX.c reads the same field.
func TestLTPScaleUsesCurrentPacketLBRRFlag(t *testing.T) {
	e := NewEncoder(BandwidthWideband)
	e.SetFEC(true)
	e.SetPacketLoss(20)
	e.nFramesPerPacket = 1
	e.snrDBQ7 = 3150
	const gainQ7 = 399 // 3.1171875 dB, the value libopus silk_SMULBB truncates to 3.

	// No LBRR in the packet: round_loss stays 20, so both thresholds are crossed.
	e.lbrrFlag = 0
	if got := e.computeLTPScaleIndex(gainQ7, codeIndependently); got != 2 {
		t.Fatalf("packet without LBRR: idx=%d want 2", got)
	}
	// LBRR present: round_loss drops to 6, so only the first threshold is crossed.
	e.lbrrFlag = 1
	if got := e.computeLTPScaleIndex(gainQ7, codeIndependently); got != 1 {
		t.Fatalf("packet with LBRR: idx=%d want 1", got)
	}
}
