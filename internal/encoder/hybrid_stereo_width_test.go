package encoder

import (
	"testing"

	"github.com/hujm2023/gopus/internal/testsignal"
	"github.com/hujm2023/gopus/types"
)

func TestHybridHighBandRetainsSmoothedStereoWidth(t *testing.T) {
	pcm, err := testsignal.GenerateCorpusSignal(testsignal.CorpusMixedV1, 48000, 8*960, 2)
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
	if _, err := enc.Encode(pcm[:960], 480); err != nil {
		t.Fatal(err)
	}
	// At this rate SILK collapses its quantized width to zero. The high band
	// retains the initial smoothed width, matching silk enc_API's control output.
	width := enc.hybridState.stereoWidthQ14
	if width != 16384 {
		t.Fatalf("high-band width=%d, want initial smoothed width 16384", width)
	}
	if want := enc.silkEncoder.SmoothedStereoWidthQ14(); width != want {
		t.Fatalf("high-band width=%d, smoothed SILK width=%d", width, want)
	}
	enc.SetMode(ModeAuto)
	enc.SetForceChannels(1)
	if _, err := enc.Encode(pcm[:960], 480); err != nil {
		t.Fatal(err)
	}
	if enc.toMono != 1 || enc.streamChannels != 2 {
		t.Fatalf("transition state toMono=%d channels=%d", enc.toMono, enc.streamChannels)
	}
	if width := enc.hybridState.stereoWidthQ14; width != 0 {
		t.Fatalf("stereo-to-mono high-band width=%d, want 0", width)
	}

}
