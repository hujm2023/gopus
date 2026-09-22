package encoder_test

import (
	"fmt"
	"testing"

	"github.com/hujm2023/gopus"
	"github.com/hujm2023/gopus/internal/encoder"
	"github.com/hujm2023/gopus/types"
)

func TestSILKCBRInternalRateLimit(t *testing.T) {
	for _, tc := range []struct {
		frameSize int
		bitrate   int
		requested types.Bandwidth
		want      types.Bandwidth
	}{
		{480, 8000, types.BandwidthMediumband, types.BandwidthNarrowband},
		{480, 11200, types.BandwidthWideband, types.BandwidthMediumband},
		{480, 12000, types.BandwidthWideband, types.BandwidthWideband},
		{960, 6400, types.BandwidthWideband, types.BandwidthNarrowband},
		{960, 7600, types.BandwidthWideband, types.BandwidthMediumband},
		{960, 8000, types.BandwidthWideband, types.BandwidthWideband},
		{960, 7600, types.BandwidthNarrowband, types.BandwidthNarrowband},
	} {
		t.Run(fmt.Sprintf("samples%d/rate%d/bw%d", tc.frameSize, tc.bitrate, tc.requested), func(t *testing.T) {
			enc := encoder.NewEncoder(48000, 1)
			enc.SetMode(encoder.ModeSILK)
			enc.SetBandwidth(tc.requested)
			enc.SetBitrate(tc.bitrate)
			enc.SetBitrateMode(encoder.ModeCBR)
			packet, err := encodeTest(enc, make([]float32, tc.frameSize), tc.frameSize)
			if err != nil {
				t.Fatal(err)
			}
			if len(packet) == 0 {
				t.Fatal("missing SILK packet")
			}
			if toc := gopus.ParseTOC(packet[0]); toc.Mode != gopus.ModeSILK || types.Bandwidth(toc.Bandwidth) != tc.want {
				t.Fatalf("packet TOC = %+v, want SILK bandwidth %v", toc, tc.want)
			}
			if enc.Bandwidth() != tc.requested {
				t.Fatalf("internal rate limit changed selected bandwidth: %v", enc.Bandwidth())
			}
		})
	}
}
