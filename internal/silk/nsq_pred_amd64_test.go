//go:build amd64 && !purego

package silk

import (
	"math/rand"
	"testing"

	"github.com/hujm2023/gopus/internal/cpufeat"
)

const nsqPredTestBufLen = maxSubFrameLength + nsqLpcBufLength

// nsqPredCase builds one reproducible input triple.
func nsqPredCase(rng *rand.Rand, scale int32, saturated bool) (st [nsqPredTestBufLen]int32, a [16]int16) {
	for i := range st {
		switch {
		case saturated && scale > 0:
			st[i] = 1<<31 - 1
		case saturated:
			st[i] = -1 << 31
		default:
			span := int64(2)*int64(scale) + 1
			st[i] = int32(int64(rng.Int63n(span)) - int64(scale))
		}
	}
	for i := range a {
		a[i] = int16(rng.Intn(65536) - 32768)
	}
	return st, a
}

// TestShortTermPrediction16StateKernelParity compares the SSE4.1 kernel against
// the scalar reference directly. It is skipped when the host lacks SSE4.1, which
// is exactly the case the runtime dispatch exists for.
func TestShortTermPrediction16StateKernelParity(t *testing.T) {
	if !cpufeat.AMD64.HasSSE41 {
		t.Skip("host CPU does not report SSE4.1; the kernel is not reachable here")
	}
	rng := rand.New(rand.NewSource(0x5eed16))
	scales := []int32{1, 2, 16, 256, 4096, 1 << 15, 1 << 20, 1 << 26, 1 << 30}
	for _, scale := range scales {
		for trial := 0; trial < 64; trial++ {
			st, a := nsqPredCase(rng, scale, scale == 1<<30 && trial%2 == 0)
			for _, idx := range []int{15, 16, nsqPredTestBufLen - 1} {
				want := shortTermPrediction16StateGo(&st, idx, &a)
				got := shortTermPrediction16StateSSE41(&st, idx, &a)
				if got != want {
					t.Fatalf("scale=%d trial=%d idx=%d: kernel %d want %d", scale, trial, idx, got, want)
				}
			}
		}
	}

	// Saturation corners: coefficients and states at both extremes.
	for _, sv := range []int32{0, 1, -1, 1 << 30, -1 << 30, 1<<31 - 1, -1 << 31} {
		for _, cv := range []int16{0, 1, -1, 32767, -32768} {
			var st [nsqPredTestBufLen]int32
			for i := range st {
				st[i] = sv
			}
			var a [16]int16
			for i := range a {
				a[i] = cv
			}
			want := shortTermPrediction16StateGo(&st, 15, &a)
			got := shortTermPrediction16StateSSE41(&st, 15, &a)
			if got != want {
				t.Fatalf("corner state=%d coef=%d: kernel %d want %d", sv, cv, got, want)
			}
		}
	}
}

// TestShortTermPrediction16StateDispatch exercises both branches of the runtime
// dispatch on any host: with the gate forced off the scalar reference must be
// selected, and with it forced on the result must equal the kernel's.
func TestShortTermPrediction16StateDispatch(t *testing.T) {
	saved := silkUseShortTermPredictionSSE41
	t.Cleanup(func() { silkUseShortTermPredictionSSE41 = saved })

	rng := rand.New(rand.NewSource(0xd15a7c4))
	type tc struct {
		st  [nsqPredTestBufLen]int32
		a   [16]int16
		idx int
	}
	cases := make([]tc, 0, 128)
	for i := 0; i < 128; i++ {
		st, a := nsqPredCase(rng, int32(1)<<uint(i%31), i%3 == 0)
		cases = append(cases, tc{st: st, a: a, idx: []int{15, 16, nsqPredTestBufLen - 1}[i%3]})
	}

	// Path 1: gate off must select the scalar reference.
	silkUseShortTermPredictionSSE41 = false
	for i, c := range cases {
		got := shortTermPrediction16State(&c.st, c.idx, &c.a)
		want := shortTermPrediction16StateGo(&c.st, c.idx, &c.a)
		if got != want {
			t.Fatalf("gate-off case %d: dispatch %d want scalar %d", i, got, want)
		}
	}

	// Path 2: gate on must agree with the scalar reference. On a host without
	// SSE4.1 the production gate stays false, so forcing it on would fault and
	// only path 1 is meaningful.
	if !cpufeat.AMD64.HasSSE41 {
		t.Log("host CPU does not report SSE4.1; only the scalar dispatch path is exercised")
		return
	}
	silkUseShortTermPredictionSSE41 = true
	for i, c := range cases {
		got := shortTermPrediction16State(&c.st, c.idx, &c.a)
		want := shortTermPrediction16StateGo(&c.st, c.idx, &c.a)
		if got != want {
			t.Fatalf("gate-on case %d: dispatch %d want scalar %d", i, got, want)
		}
	}
}

func BenchmarkShortTermPrediction16StateKernel(b *testing.B) {
	var st [nsqPredTestBufLen]int32
	for i := range st {
		st[i] = int32(i) * 1000
	}
	var a [16]int16
	for i := range a {
		a[i] = int16(1000 - i*7)
	}
	if !cpufeat.AMD64.HasSSE41 {
		b.Skip("host CPU does not report SSE4.1")
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sink int32
	for i := 0; i < b.N; i++ {
		sink += shortTermPrediction16StateSSE41(&st, 15, &a)
	}
	_ = sink
}

func BenchmarkShortTermPrediction16StateScalar(b *testing.B) {
	var st [nsqPredTestBufLen]int32
	for i := range st {
		st[i] = int32(i) * 1000
	}
	var a [16]int16
	for i := range a {
		a[i] = int16(1000 - i*7)
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sink int32
	for i := 0; i < b.N; i++ {
		sink += shortTermPrediction16StateGo(&st, 15, &a)
	}
	_ = sink
}
