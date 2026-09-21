package silk

import (
	"math/rand"
	"testing"
)

func TestQuantize4Parity(t *testing.T) {
	// Exhaust clamped residuals at all codec offsets and lambda/RDO boundaries.
	for _, offset := range []int32{offsetVLQ10, offsetVHQ10, offsetUVLQ10, offsetUVHQ10} {
		for _, lambda := range []int32{0, 1, 1024, 2048, 2049, 3072, 32767, 32768, 60000} {
			for r := int32(-31 << 10); r <= 30<<10; r++ {
				batch := nsqQuant4{r: [4]int32{r, -31 << 10, 0, 30 << 10}, rd: [4]int32{0, 2147483647, -2147483648, -1}}
				want := batch
				quantize4Go(&want, offset, lambda)
				quantize4(&batch, offset, lambda)
				if batch != want {
					t.Fatalf("offset=%d lambda=%d r=%d: got %+v want %+v", offset, lambda, r, batch, want)
				}
			}
		}
	}
	rng := rand.New(rand.NewSource(42))
	for range 10000 {
		var batch nsqQuant4
		for k := range 4 {
			batch.r[k] = rng.Int31n(61<<10+1) - (31 << 10)
			batch.rd[k] = int32(rng.Uint32())
		}
		offset := []int32{offsetVLQ10, offsetVHQ10, offsetUVLQ10, offsetUVHQ10}[rng.Intn(4)]
		lambda := rng.Int31n(65536)
		want := batch
		quantize4Go(&want, offset, lambda)
		quantize4(&batch, offset, lambda)
		if batch != want {
			t.Fatalf("random offset=%d lambda=%d got %+v want %+v", offset, lambda, batch, want)
		}
	}
	var batch nsqQuant4
	if n := testing.AllocsPerRun(1000, func() { quantize4(&batch, offsetVLQ10, 2049) }); n != 0 {
		t.Fatalf("allocs=%g", n)
	}
}
