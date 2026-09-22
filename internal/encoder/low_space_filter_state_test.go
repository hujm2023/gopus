package encoder

import "testing"

func TestLowSpacePacketPreservesInputFilterState(t *testing.T) {
	for _, voip := range []bool{false, true} {
		enc := NewEncoder(48000, 1)
		enc.SetVoIPApplication(voip)
		enc.SetBitrate(800)
		enc.SetBitrateMode(ModeCBR)
		pcm := make([]float32, 480)
		for i := range pcm {
			pcm[i] = .25
		}
		packet, err := enc.EncodeFloat32WithAnalysisMaxBytes(pcm, 480, pcm, 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(packet) != 1 {
			t.Fatalf("packet length=%d, want 1", len(packet))
		}
		if enc.hpMem != [4]float32{} || enc.variableHPSmth2Inited {
			t.Fatalf("voip=%v: low-space packet advanced input filtering: hp=%v smoother=%v", voip, enc.hpMem, enc.variableHPSmth2Inited)
		}
	}
}
