package encoder

import (
	"testing"

	"github.com/hujm2023/gopus/internal/silk"
	"github.com/hujm2023/gopus/types"
)

func TestSILKBandwidthResetAndConstraints(t *testing.T) {
	e := NewEncoder(48000, 1)
	e.bandwidth = types.BandwidthMediumband
	e.silkMaxInternalRate = 16000
	e.controlSILKBandwidth(ModeSILK)
	e.ensureSILKEncoder()
	e.first = false
	old := e.silkEncoder
	e.bandwidth = types.BandwidthWideband
	e.controlSILKBandwidth(ModeSILK)
	e.ensureSILKEncoder()
	if e.silkEncoder != old || e.silkBandwidth() != silk.BandwidthMediumband {
		t.Fatal("active speech replaced the encoder or changed its actual bandwidth")
	}
	e.silkMaxInternalRate = 8000
	e.controlSILKBandwidth(ModeSILK)
	e.ensureSILKEncoder()
	if e.silkEncoder.SampleRate() != 8000 || e.silkResamplerRate != 8000 {
		t.Fatal("hard ceiling did not change both encoder and resampler")
	}
	e.Reset()
	e.bandwidth = types.BandwidthWideband
	e.silkMaxInternalRate = 16000
	e.controlSILKBandwidth(ModeSILK)
	e.ensureSILKEncoder()
	if e.silkEncoder.SampleRate() != 16000 || e.silkOpusCanSwitch || e.silkBWSwitch {
		t.Fatal("reset retained old bandwidth or pending handshake")
	}
}
