package silk

import "testing"

// TestStereoCondCodingUsesSilkPacketFrameIndex pins the libopus enc_API.c
// conditional-coding rule that the Opus-level Hybrid stereo leg must feed:
// condCoding is chosen from state_Fxx[0].nFramesEncoded (the internal frame index
// within the current SILK packet) minus the channel index, so the single internal
// frame of a 10/20 ms SILK packet codes both channels independently, while the
// second internal frame of a 40/60 ms packet codes conditionally.
func TestStereoCondCodingUsesSilkPacketFrameIndex(t *testing.T) {
	for _, tc := range []struct {
		name             string
		midFramesEncoded int
		channelIdx       int
		prevDecodeOnly   int32
		want             int
	}{
		{"20ms packet mid", 0, 0, 0, codeIndependently},
		{"20ms packet side", 1, 1, 0, codeIndependently},
		{"20ms packet side after mid-only", 1, 1, 1, codeIndependently},
		{"40ms packet second frame mid", 1, 0, 0, codeConditionally},
		{"40ms packet second frame side", 2, 1, 0, codeConditionally},
		{"40ms packet side after mid-only", 2, 1, 1, codeIndependentlyNoLtpScaling},
	} {
		if got := stereoSelectCondCoding(tc.midFramesEncoded, tc.channelIdx, tc.prevDecodeOnly); got != tc.want {
			t.Errorf("%s: stereoSelectCondCoding(%d, %d, %d) = %d, want %d",
				tc.name, tc.midFramesEncoded, tc.channelIdx, tc.prevDecodeOnly, got, tc.want)
		}
	}
}

// TestSetStereoCondContextPinsChannelContext checks the exported setter the Opus
// wrapper uses, and that a new packet starts from a clean context.
func TestSetStereoCondContextPinsChannelContext(t *testing.T) {
	mid := NewEncoder(BandwidthWideband)
	side := NewEncoder(BandwidthWideband)

	mid.SetStereoCondContext(mid, int(mid.NFramesEncoded()), 0, 0)
	if mid.stereoCondMid != mid || mid.stereoChannelIdx != 0 {
		t.Fatalf("mid context: condMid=%v channelIdx=%d", mid.stereoCondMid != nil, mid.stereoChannelIdx)
	}
	side.SetStereoCondContext(mid, int(mid.NFramesEncoded()), 1, 1)
	if side.stereoCondMid != mid || side.stereoChannelIdx != 1 || side.stereoPrevDecodeOnlyMiddle != 1 {
		t.Fatalf("side context: mid=%v channelIdx=%d prevDom=%d",
			side.stereoCondMid == mid, side.stereoChannelIdx, side.stereoPrevDecodeOnlyMiddle)
	}

	mid.ResetPacketState()
	if mid.stereoCondMid != nil || mid.stereoChannelIdx != 0 || mid.stereoCondMidFramesEncoded != 0 {
		t.Fatal("ResetPacketState must clear the stereo condCoding context")
	}
}
