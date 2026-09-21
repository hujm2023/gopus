package silk

import "testing"

var quant4Sink [4]nsqSamplePair

func BenchmarkQuantize4(b *testing.B) {
	for _, tc := range []struct {
		name string
		fn   func(*nsqQuant4, int32, int32)
	}{
		{"scalar", quantize4Go}, {"dispatch", quantize4},
	} {
		b.Run(tc.name, func(b *testing.B) {
			// Include AoS input gathering and candidate scattering, as at the call site.
			input := [4]struct{ r, rd int32 }{{-17321, 31}, {-250, 99}, {4701, 324}, {23103, 571}}
			var batch nsqQuant4
			b.ReportAllocs()
			for b.Loop() {
				for k := range 4 {
					batch.r[k], batch.rd[k] = input[k].r, input[k].rd
				}
				tc.fn(&batch, offsetUVHQ10, 2049)
				for k := range 4 {
					quant4Sink[k][0].qQ10, quant4Sink[k][1].qQ10 = batch.q0[k], batch.q1[k]
					quant4Sink[k][0].rdQ10, quant4Sink[k][1].rdQ10 = batch.cost0[k], batch.cost1[k]
				}
			}
		})
	}
}
