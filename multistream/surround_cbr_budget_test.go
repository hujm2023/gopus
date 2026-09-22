package multistream

import "testing"

func TestSurroundCBRBudgetDropsInnerPadding(t *testing.T) {
	enc, err := NewEncoderDefault(48000, 6)
	if err != nil {
		t.Fatal(err)
	}
	enc.SetBitrate(64000)
	enc.SetVBR(false)
	pcm := generateSurroundSweep(6, 480, 4)
	for f := range 4 {
		frame := pcm[f*480*6 : (f+1)*480*6]
		packet, err := enc.EncodeFloat32WithAnalysisMaxBytes(frame, 480, frame, 4000)
		if err != nil {
			t.Fatal(err)
		}
		// The third stream needs ordinary padding on frame 2. Its removed
		// bytes must be available to the final LFE stream, preserving CBR.
		if len(packet) != 80 {
			t.Fatalf("frame %d: %d bytes, want 80", f, len(packet))
		}
	}
}
