//go:build amd64 && !purego

package silk

import (
	"math/rand"
	"testing"
)

// TestShortTermPrediction16StateParity compares the amd64 SSE4.1 kernel against
// the scalar reference over the state and coefficient ranges the quantizer
// actually produces, including the extremes that force the widest products.
func TestShortTermPrediction16StateParity(t *testing.T) {
	rng := rand.New(rand.NewSource(0x5eed16))
	const bufLen = maxSubFrameLength + nsqLpcBufLength

	scales := []int32{1, 2, 16, 256, 4096, 1 << 15, 1 << 20, 1 << 26, 1 << 30, -1 << 30}
	for _, scale := range scales {
		for trial := 0; trial < 64; trial++ {
			var st [bufLen]int32
			for i := range st {
				switch {
				case scale == -1<<30:
					st[i] = -1 << 30
				case scale >= 1<<30:
					// Keep the draw inside int32 while still reaching the top
					// of the range on some samples.
					if trial%2 == 0 {
						st[i] = 1<<31 - 1
					} else {
						st[i] = int32(rng.Intn(1<<20)) << 11
					}
				default:
					span := int64(2)*int64(scale) + 1
					st[i] = int32(int64(rng.Int63n(span)) - int64(scale))
				}
			}
			var a [16]int16
			for i := range a {
				a[i] = int16(rng.Intn(65536) - 32768)
			}
			for _, idx := range []int{15, 16, bufLen - 1} {
				want := shortTermPrediction16StateGo(&st, idx, &a)
				got := shortTermPrediction16State(&st, idx, &a)
				if got != want {
					t.Fatalf("scale=%d trial=%d idx=%d: got %d want %d", scale, trial, idx, got, want)
				}
			}
		}
	}

	// Saturation-shaped corners: coefficients at both extremes with states at
	// both extremes, which is where a lane or shift mistake shows up first.
	for _, sv := range []int32{0, 1, -1, 1 << 30, -1 << 30, 1<<31 - 1, -1 << 31} {
		for _, cv := range []int16{0, 1, -1, 32767, -32768} {
			var st [bufLen]int32
			for i := range st {
				st[i] = sv
			}
			var a [16]int16
			for i := range a {
				a[i] = cv
			}
			want := shortTermPrediction16StateGo(&st, 15, &a)
			got := shortTermPrediction16State(&st, 15, &a)
			if got != want {
				t.Fatalf("corner state=%d coef=%d: got %d want %d", sv, cv, got, want)
			}
		}
	}
}

func BenchmarkShortTermPrediction16State(b *testing.B) {
	const bufLen = maxSubFrameLength + nsqLpcBufLength
	var st [bufLen]int32
	for i := range st {
		st[i] = int32(i) * 1000
	}
	var a [16]int16
	for i := range a {
		a[i] = int16(1000 - i*7)
	}
	idx := 15
	b.ReportAllocs()
	b.ResetTimer()
	var sink int32
	for i := 0; i < b.N; i++ {
		sink += shortTermPrediction16State(&st, idx, &a)
	}
	_ = sink
}

func BenchmarkShortTermPrediction16StateGo(b *testing.B) {
	const bufLen = maxSubFrameLength + nsqLpcBufLength
	var st [bufLen]int32
	for i := range st {
		st[i] = int32(i) * 1000
	}
	var a [16]int16
	for i := range a {
		a[i] = int16(1000 - i*7)
	}
	idx := 15
	b.ReportAllocs()
	b.ResetTimer()
	var sink int32
	for i := 0; i < b.N; i++ {
		sink += shortTermPrediction16StateGo(&st, idx, &a)
	}
	_ = sink
}
