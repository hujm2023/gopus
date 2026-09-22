package encoder

import (
	"math"
	"testing"

	"github.com/hujm2023/gopus/types"
)

func TestHybridToSILKUsesFreshPacketCoder(t *testing.T) {
	e := NewEncoder(48000, 1)
	e.SetBitrate(24000)
	e.SetMode(ModeHybrid)
	pcm := make([]float32, 480)
	for i := range pcm {
		pcm[i] = float32(.4 * math.Sin(2*math.Pi*440*float64(i)/48000))
	}
	if _, err := e.EncodeFloat32(pcm, 480); err != nil {
		t.Fatal(err)
	}
	if e.silkEncoder.GetRangeEncoderPtr() == nil {
		t.Fatal("Hybrid did not leave a shared coder for the transition fixture")
	}
	e.SetMode(ModeSILK)
	e.SetBandwidth(types.BandwidthWideband)
	e.SetMaxBandwidth(types.BandwidthWideband)
	packet, err := e.EncodeFloat32(pcm, 480)
	if err != nil {
		t.Fatal(err)
	}
	if len(packet) < 2 || e.silkEncoder.GetRangeEncoderPtr() != nil {
		t.Fatal("standalone SILK retained the shared Hybrid packet coder")
	}
}
