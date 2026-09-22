package gopus

import (
	"bytes"
	"testing"

	"github.com/hujm2023/gopus/internal/libopustest"
	"github.com/hujm2023/gopus/internal/testsignal"
)

// Active speech holds MB, then silence permits the two-packet redundancy
// handshake to WB. Speech resumes to verify the preserved prediction state.
func TestEncoderSILKBandwidthSwitchLibopusParity(t *testing.T) {
	libopustest.RequireOracle(t)
	var spec encDiffSpec
	for _, candidate := range buildEncDiffSweep() {
		if candidate.name == "silk_wb_ch1_10ms_24000bps_vbr0_fectrue_dtxfalse" {
			spec = candidate
			break
		}
	}
	if spec.name == "" {
		t.Fatal("bandwidth-switch fixture missing from differential matrix")
	}
	const frames = 80
	pcm, err := testsignal.GenerateCorpusSignal(testsignal.CorpusSpeechInNoiseV1, 48000, 480*frames, 1)
	if err != nil {
		t.Fatal(err)
	}
	clear(pcm[8*480 : 60*480])
	recs, err := libopustest.ProbeEncodeDiff(libopustest.EncodeDiffParams{
		SampleRate: 48000, Channels: 1,
		Application:  libopustest.EncodeDiffApplicationAudio,
		ForceMode:    libopustest.EncodeDiffForceModeSILKOnly,
		Bandwidth:    libopustest.EncodeDiffBandwidthWideband,
		MaxBandwidth: libopustest.EncodeDiffBandwidthWideband,
		Bitrate:      24000, Complexity: 10, Signal: libopustest.EncodeDiffSignalVoice,
		VBR: true, ForceChannels: 1, InbandFEC: 1, PacketLoss: 20,
		FrameSize: 480, FrameCount: frames, PCM: pcm,
	})
	if err != nil {
		t.Fatal(err)
	}
	enc, ok := configureEncDiff(t, spec)
	if !ok {
		t.Fatal("config")
	}
	if len(recs) != frames {
		t.Fatalf("oracle returned %d frames, want %d", len(recs), frames)
	}
	seenMB, seenWB := false, false
	for i, rec := range recs {
		buf, err := enc.EncodeFloat32(pcm[i*480 : (i+1)*480])
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, rec.Packet) || enc.FinalRange() != rec.FinalRange {
			t.Fatalf("frame %d: packet %x, want %x; range %d, want %d", i, buf, rec.Packet, enc.FinalRange(), rec.FinalRange)
		}
		if len(buf) > 0 {
			seenMB = seenMB || buf[0]>>5 == 1
			seenWB = seenWB || buf[0]>>5 == 2
		}
	}
	if !seenMB || !seenWB {
		t.Fatal("sequence did not exercise both MB and WB packets")
	}
}
