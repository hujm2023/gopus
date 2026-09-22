package encoder

import (
	"testing"

	"github.com/hujm2023/gopus/internal/testsignal"
	"github.com/hujm2023/gopus/types"
)

func TestHybridStereoContinuesPrefillMidVAD(t *testing.T) {
	pcm, err := testsignal.GenerateCorpusSignal(testsignal.CorpusMixedV1, 48000, 3*960, 2)
	if err != nil {
		t.Fatal(err)
	}
	enc := NewEncoder(48000, 2)
	enc.SetMode(ModeHybrid)
	enc.SetFrameSize(480)
	enc.SetBandwidth(types.BandwidthSuperwideband)
	enc.SetMaxBandwidth(types.BandwidthSuperwideband)
	enc.SetBitrate(24000)
	enc.SetBitrateMode(ModeVBR)
	enc.SetComplexity(10)
	enc.SetForceChannels(2)
	enc.SetSignalType(types.SignalVoice)
	enc.prevMode = ModeCELT
	enc.prevPacketMode = ModeCELT
	enc.delayBuffer = make([]opusRes, 960)
	for i := range enc.delayBuffer {
		enc.delayBuffer[i] = opusRes(pcm[i])
	}
	enc.maybePrefillSILKOnModeTransitionWithOptions(ModeHybrid, false, false, enc.silkInputBitrate(480))
	if enc.silkVADMidFeedback == nil {
		t.Fatal("stereo prefill did not initialize mid VAD")
	}
	previous := enc.silkVADMidFeedback.Counter
	// The prefill is complete; encode consecutive Hybrid packets using that history.
	enc.prevMode = ModeHybrid
	enc.prevPacketMode = ModeHybrid
	for frame := 1; frame < 3; frame++ {
		if _, err := enc.Encode(pcm[frame*960:(frame+1)*960], 480); err != nil {
			t.Fatal(err)
		}
		if got := enc.silkVADMidFeedback.Counter; got != previous+1 {
			t.Fatalf("frame %d: mid VAD counter=%d, want %d after prefill", frame, got, previous+1)
		}
		previous++
	}
}
