package silk

import (
	"testing"

	"github.com/hujm2023/gopus/internal/rangecoding"
)

func TestFinalizePacketRangePreservesTellLength(t *testing.T) {
	newRange := func(bits, storage int) *rangecoding.Encoder {
		buf := make([]byte, 8)
		for i := range buf {
			buf[i] = 0xa5
		}
		re := &rangecoding.Encoder{}
		re.Init(buf)
		re.Limit(uint32(storage))
		for i := 0; i < bits; i++ {
			re.EncodeBit(i&1, 1)
		}
		return re
	}
	t.Run("zero_before_redundancy", func(t *testing.T) {
		reference := newRange(7, 8)
		reference.EncodeBit(1, 1)
		wantLen := (reference.Tell() + 7) / 8
		packed := reference.Done()
		if wantLen != 2 || len(packed) != 1 {
			t.Fatalf("fixture tell length=%d packed=%d", wantLen, len(packed))
		}

		re := newRange(7, 8)
		e := &Encoder{}
		calls := 0
		e.SetPacketTermination(func(re *rangecoding.Encoder) { calls++; re.EncodeBit(1, 1) })
		got := e.finalizePacketRange(re)
		if calls != 1 || len(got) != wantLen {
			t.Fatalf("termination calls=%d length=%d want%d", calls, len(got), wantLen)
		}
		if got[0] != packed[0] || got[1] != 0 {
			t.Fatalf("payload=%x, want %02x00", got, packed[0])
		}
		if e.lastRng != reference.Range() {
			t.Fatalf("final range=%08x want%08x", e.lastRng, reference.Range())
		}
	})
	t.Run("bounded_by_active_storage", func(t *testing.T) {
		re := newRange(16, 2)
		if n := (re.Tell() + 7) / 8; n <= re.Storage() {
			t.Fatalf("fixture tell length=%d storage=%d", n, re.Storage())
		}
		e := &Encoder{}
		got := e.finalizePacketRange(re)
		if len(got) != 2 {
			t.Fatalf("length=%d want active storage2 (backing length8)", len(got))
		}
	})
}
