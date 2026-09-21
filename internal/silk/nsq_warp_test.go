package silk

import (
	"math/rand"
	"testing"

	"github.com/hujm2023/gopus/internal/cpufeat"
)

var benchWarpSink [maxDelDecStates]int32

// refWarpedARFeedback computes warped AR feedback in pure Go (reference).
func refWarpedARFeedback(sAR []int32, diffQ14 int32, arShpQ13 []int16, warpQ16 int32, order int) int32 {
	w := int64(warpQ16)
	tmp2 := diffQ14 + int32((int64(sAR[0])*w)>>16)
	tmp1 := sAR[0] + int32((int64(sAR[1]-tmp2)*w)>>16)
	sAR[0] = tmp2
	acc := int32(order>>1) + int32((int64(tmp2)*int64(arShpQ13[0]))>>16)

	for j := 2; j < order; j += 2 {
		tmp2 = sAR[j-1] + int32((int64(sAR[j]-tmp1)*w)>>16)
		sAR[j-1] = tmp1
		acc += int32((int64(tmp1) * int64(arShpQ13[j-1])) >> 16)
		tmp1 = sAR[j] + int32((int64(sAR[j+1]-tmp2)*w)>>16)
		sAR[j] = tmp2
		acc += int32((int64(tmp2) * int64(arShpQ13[j])) >> 16)
	}
	sAR[order-1] = tmp1
	acc += int32((int64(tmp1) * int64(arShpQ13[order-1])) >> 16)
	return acc
}

func TestWarpedARFeedback24(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for trial := range 1000 {
		var sAR [maxShapeLpcOrder]int32
		var sARRef [maxShapeLpcOrder]int32
		var arShpQ13 [24]int16
		for i := range 24 {
			v := rng.Int31n(1<<20) - (1 << 19)
			sAR[i] = v
			sARRef[i] = v
		}
		for i := range arShpQ13 {
			arShpQ13[i] = int16(rng.Int31n(1<<13) - (1 << 12))
		}
		diffQ14 := rng.Int31n(1<<20) - (1 << 19)
		warpQ16 := rng.Int31n(1<<15) - (1 << 14) // typical range

		got := warpedARFeedback24(&sAR, diffQ14, &arShpQ13, warpQ16)
		want := refWarpedARFeedback(sARRef[:], diffQ14, arShpQ13[:], warpQ16, 24)

		if got != want {
			t.Fatalf("trial %d: warpedARFeedback24 mismatch: got %d, want %d", trial, got, want)
		}
		for i := range 24 {
			if sAR[i] != sARRef[i] {
				t.Fatalf("trial %d: sAR[%d] mismatch: got %d, want %d", trial, i, sAR[i], sARRef[i])
			}
		}
	}
}

func TestWarpedARFeedback16(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	for trial := range 1000 {
		var sAR [maxShapeLpcOrder]int32
		var sARRef [maxShapeLpcOrder]int32
		var arShpQ13 [16]int16
		for i := range 24 {
			v := rng.Int31n(1<<20) - (1 << 19)
			sAR[i] = v
			sARRef[i] = v
		}
		for i := range arShpQ13 {
			arShpQ13[i] = int16(rng.Int31n(1<<13) - (1 << 12))
		}
		diffQ14 := rng.Int31n(1<<20) - (1 << 19)
		warpQ16 := rng.Int31n(1<<15) - (1 << 14)

		got := warpedARFeedback16(&sAR, diffQ14, &arShpQ13, warpQ16)
		want := refWarpedARFeedback(sARRef[:], diffQ14, arShpQ13[:], warpQ16, 16)

		if got != want {
			t.Fatalf("trial %d: warpedARFeedback16 mismatch: got %d, want %d", trial, got, want)
		}
		for i := range 16 {
			if sAR[i] != sARRef[i] {
				t.Fatalf("trial %d: sAR[%d] mismatch: got %d, want %d", trial, i, sAR[i], sARRef[i])
			}
		}
	}
}

func TestWarpedARFeedback24EdgeCases(t *testing.T) {
	// Zero warping
	var sAR [maxShapeLpcOrder]int32
	var sARRef [maxShapeLpcOrder]int32
	var arShpQ13 [24]int16
	for i := range 24 {
		sAR[i] = int32(i * 1000)
		sARRef[i] = sAR[i]
		arShpQ13[i] = int16(100 + i)
	}
	got := warpedARFeedback24(&sAR, 500, &arShpQ13, 0)
	want := refWarpedARFeedback(sARRef[:], 500, arShpQ13[:], 0, 24)
	if got != want {
		t.Fatalf("zero warp: got %d, want %d", got, want)
	}

	// Max values
	for i := range 24 {
		sAR[i] = 0x7FFFF
		sARRef[i] = sAR[i]
		arShpQ13[i] = 0x7FFF
	}
	got = warpedARFeedback24(&sAR, 0x7FFFF, &arShpQ13, 0x7FFF)
	want = refWarpedARFeedback(sARRef[:], 0x7FFFF, arShpQ13[:], 0x7FFF, 24)
	if got != want {
		t.Fatalf("max values: got %d, want %d", got, want)
	}

	// Negative values
	for i := range 24 {
		sAR[i] = -0x7FFFF
		sARRef[i] = sAR[i]
		arShpQ13[i] = -0x7FFF
	}
	got = warpedARFeedback24(&sAR, -0x7FFFF, &arShpQ13, -0x7FFF)
	want = refWarpedARFeedback(sARRef[:], -0x7FFFF, arShpQ13[:], -0x7FFF, 24)
	if got != want {
		t.Fatalf("negative values: got %d, want %d", got, want)
	}
}

func buildWarpAR(states []nsqDelDecState) *nsqWarpAR {
	ar := new(nsqWarpAR)
	for k := 0; k < maxDelDecStates; k++ {
		ar.diff[k] = int64(states[k].diffQ14)
		for j := 0; j < maxShapeLpcOrder; j++ {
			ar.set(k, j, states[k].sAR2Q14[j])
		}
	}
	return ar
}

func TestWarpedARFeedback24States4MatchesScalar(t *testing.T) {
	rng := rand.New(rand.NewSource(123))
	for trial := range 1000 {
		var states [maxDelDecStates]nsqDelDecState
		var strided [maxDelDecStates]nsqDelDecState
		var arShpQ13 [24]int16
		for k := range maxDelDecStates {
			states[k].diffQ14 = rng.Int31n(1<<20) - (1 << 19)
			strided[k].diffQ14 = states[k].diffQ14
			for i := range 24 {
				v := rng.Int31n(1<<20) - (1 << 19)
				states[k].sAR2Q14[i] = v
				strided[k].sAR2Q14[i] = v
			}
		}
		for i := range arShpQ13 {
			arShpQ13[i] = int16(rng.Int31n(1<<13) - (1 << 12))
		}
		warpQ16 := rng.Int31n(1<<15) - (1 << 14)
		ar := buildWarpAR(states[:])
		warpedARFeedback24States4(ar, &arShpQ13, warpQ16)
		var want [maxDelDecStates]int32
		warpedARFeedback24States4Strided(strided[:], &arShpQ13, warpQ16, &want)
		if ar.out != want {
			t.Fatalf("trial %d: got %v want %v", trial, ar.out, want)
		}
		for k := range maxDelDecStates {
			for i := range 24 {
				if ar.get(k, i) != strided[k].sAR2Q14[i] {
					t.Fatalf("trial %d state %d sAR[%d] mismatch", trial, k, i)
				}
			}
		}
	}
}

func TestWarpedARFeedback24States4Corners(t *testing.T) {
	vals := []int32{0, 1, -1, 1 << 15, -(1 << 15), 0x7FFFFFFF, -0x80000000, 0x7FFFFF, -0x800000}
	warps := []int32{0, 1, -1, 0x7FFF, -0x8000, 1 << 14, -(1 << 14)}
	coefs := []int16{0, 1, -1, 0x7FFF, -0x8000, 1 << 12, -(1 << 12)}
	for _, w := range warps {
		for _, cv := range coefs {
			for _, v := range vals {
				var states [maxDelDecStates]nsqDelDecState
				var strided [maxDelDecStates]nsqDelDecState
				var arShpQ13 [24]int16
				for k := range maxDelDecStates {
					states[k].diffQ14 = v
					strided[k].diffQ14 = v
					for i := range 24 {
						states[k].sAR2Q14[i] = v
						strided[k].sAR2Q14[i] = v
					}
				}
				for i := range arShpQ13 {
					arShpQ13[i] = cv
				}
				ar := buildWarpAR(states[:])
				warpedARFeedback24States4(ar, &arShpQ13, w)
				var want [maxDelDecStates]int32
				warpedARFeedback24States4Strided(strided[:], &arShpQ13, w, &want)
				if ar.out != want {
					t.Fatalf("w=%d c=%d v=%d: got %v want %v", w, cv, v, ar.out, want)
				}
				for k := range maxDelDecStates {
					for i := range 24 {
						if ar.get(k, i) != strided[k].sAR2Q14[i] {
							t.Fatalf("w=%d c=%d v=%d state %d sAR[%d] mismatch", w, cv, v, k, i)
						}
					}
				}
			}
		}
	}
}

// TestWarpedARFeedback24States4Repeat calls the kernel many times on the same
// state. nsqWarpAR is only 8-byte aligned, so an SSE instruction that requires an
// aligned memory operand on it would fault here rather than on the first call.
func TestWarpedARFeedback24States4Repeat(t *testing.T) {
	var arShpQ13 [24]int16
	for i := range arShpQ13 {
		arShpQ13[i] = int16(1000 + i)
	}
	ar := new(nsqWarpAR)
	for k := 0; k < maxDelDecStates; k++ {
		ar.diff[k] = 4096
		for j := 0; j < maxShapeLpcOrder; j++ {
			ar.set(k, j, 12345)
		}
	}
	for i := 0; i < 100000; i++ {
		warpedARFeedback24States4(ar, &arShpQ13, 16384)
	}
	if ar.out[0] == 0 {
		t.Fatalf("kernel produced no output")
	}
}

func TestWarpedARFeedback24States4StackState(t *testing.T) {
	if !cpufeat.AMD64.HasSSE41 {
		t.Skip("scalar path has no such constraint")
	}
	var stackAR nsqWarpAR
	var arShpQ13 [24]int16
	heapAR := new(nsqWarpAR)
	warpedARFeedback24States4(heapAR, &arShpQ13, 0)
	t.Log("heap-resident transposed state: ok")
	warpedARFeedback24States4(&stackAR, &arShpQ13, 0)
	t.Log("stack-resident transposed state: ok")
}

func TestWarpedARFeedback20States3MatchesScalar(t *testing.T) {
	rng := rand.New(rand.NewSource(124))
	for trial := range 1000 {
		var states [maxDelDecStates]nsqDelDecState
		var scalar [maxDelDecStates]nsqDelDecState
		var arShpQ13 [20]int16
		for k := range 3 {
			states[k].diffQ14 = rng.Int31n(1<<20) - (1 << 19)
			scalar[k].diffQ14 = states[k].diffQ14
			for i := range 20 {
				v := rng.Int31n(1<<20) - (1 << 19)
				states[k].sAR2Q14[i] = v
				scalar[k].sAR2Q14[i] = v
			}
		}
		for i := range arShpQ13 {
			arShpQ13[i] = int16(rng.Int31n(1<<13) - (1 << 12))
		}
		warpQ16 := rng.Int31n(1<<15) - (1 << 14)

		var got [maxDelDecStates]int32
		warpedARFeedback20States3(states[:], &arShpQ13, warpQ16, &got)

		for k := range 3 {
			want := refWarpedARFeedback(scalar[k].sAR2Q14[:], scalar[k].diffQ14, arShpQ13[:], warpQ16, 20)
			if got[k] != want {
				t.Fatalf("trial %d state %d: got %d want %d", trial, k, got[k], want)
			}
			for i := range 20 {
				if states[k].sAR2Q14[i] != scalar[k].sAR2Q14[i] {
					t.Fatalf("trial %d state %d sAR[%d]: got %d want %d", trial, k, i, states[k].sAR2Q14[i], scalar[k].sAR2Q14[i])
				}
			}
		}
	}
}

func BenchmarkWarpedARFeedback24(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	var sAR [maxShapeLpcOrder]int32
	var arShpQ13 [24]int16
	for i := range sAR {
		sAR[i] = rng.Int31()
	}
	for i := range arShpQ13 {
		arShpQ13[i] = int16(rng.Int31())
	}
	diffQ14 := rng.Int31()
	warpQ16 := int32(rng.Int31n(1 << 15))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		warpedARFeedback24(&sAR, diffQ14, &arShpQ13, warpQ16)
	}
}

func BenchmarkWarpedARFeedback24States4(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	var states [maxDelDecStates]nsqDelDecState
	var arShpQ13 [24]int16
	for k := range states {
		states[k].diffQ14 = rng.Int31()
		for i := range states[k].sAR2Q14 {
			states[k].sAR2Q14[i] = rng.Int31()
		}
	}
	for i := range arShpQ13 {
		arShpQ13[i] = int16(rng.Int31())
	}
	warpQ16 := int32(rng.Int31n(1 << 15))
	ar := buildWarpAR(states[:])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		warpedARFeedback24States4(ar, &arShpQ13, warpQ16)
	}
	benchWarpSink = ar.out
}

func BenchmarkWarpedARFeedback24States4Strided(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	var states [maxDelDecStates]nsqDelDecState
	var arShpQ13 [24]int16
	for k := range states {
		states[k].diffQ14 = rng.Int31()
		for i := range states[k].sAR2Q14 {
			states[k].sAR2Q14[i] = rng.Int31()
		}
	}
	for i := range arShpQ13 {
		arShpQ13[i] = int16(rng.Int31())
	}
	warpQ16 := int32(rng.Int31n(1 << 15))
	var sink [maxDelDecStates]int32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		warpedARFeedback24States4Strided(states[:], &arShpQ13, warpQ16, &sink)
	}
	benchWarpSink = sink
}

func BenchmarkWarpedARFeedback20States3(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	var states [maxDelDecStates]nsqDelDecState
	var arShpQ13 [20]int16
	for k := range 3 {
		states[k].diffQ14 = rng.Int31()
		for i := range 20 {
			states[k].sAR2Q14[i] = rng.Int31()
		}
	}
	for i := range arShpQ13 {
		arShpQ13[i] = int16(rng.Int31())
	}
	warpQ16 := int32(rng.Int31n(1 << 15))
	var out [maxDelDecStates]int32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		warpedARFeedback20States3(states[:], &arShpQ13, warpQ16, &out)
	}
}

func BenchmarkWarpedARFeedback20States3Scalar(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	var states [maxDelDecStates]nsqDelDecState
	var arShpQ13 [20]int16
	for k := range 3 {
		states[k].diffQ14 = rng.Int31()
		for i := range 20 {
			states[k].sAR2Q14[i] = rng.Int31()
		}
	}
	for i := range arShpQ13 {
		arShpQ13[i] = int16(rng.Int31())
	}
	warpQ16 := int32(rng.Int31n(1 << 15))
	var out [maxDelDecStates]int32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k := range 3 {
			out[k] = warpedARFeedbackGeneric(states[k].sAR2Q14[:], states[k].diffQ14, arShpQ13[:], warpQ16, 20)
		}
	}
}

func BenchmarkWarpedARFeedback16(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	var sAR [maxShapeLpcOrder]int32
	var arShpQ13 [16]int16
	for i := range sAR {
		sAR[i] = rng.Int31()
	}
	for i := range arShpQ13 {
		arShpQ13[i] = int16(rng.Int31())
	}
	diffQ14 := rng.Int31()
	warpQ16 := int32(rng.Int31n(1 << 15))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		warpedARFeedback16(&sAR, diffQ14, &arShpQ13, warpQ16)
	}
}

func TestWarpedARFeedback24States4Allocs(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(*nsqWarpAR, *[24]int16, int32)
	}{
		{"dispatch", warpedARFeedback24States4},
		{"scalar", warpedARFeedback24States4Go},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ar := new(nsqWarpAR)
			var coefficients [24]int16
			for k := range maxDelDecStates {
				ar.diff[k] = int64(k+1) * 12345
			}
			for j := range coefficients {
				coefficients[j] = int16(j*17 - 200)
			}
			tc.call(ar, &coefficients, 16384)
			if got := testing.AllocsPerRun(1000, func() {
				tc.call(ar, &coefficients, 16384)
			}); got != 0 {
				t.Fatalf("steady-state allocations = %g, want 0", got)
			}
		})
	}
}
