package silk

import (
	"math"
	"testing"
)

// Reference semantics for the saturating 32-bit primitives, pinned before any
// vector implementation of them exists. The vector port must reproduce these
// exactly, including the saturation value's unsigned conversion.
//
//	silk_ADD_SAT32(a,b): r = a+b; overflow iff ((a^r) & (b^r)) < 0
//	silk_SUB_SAT32(a,b): r = a-b; overflow iff ((a^b) & (a^r)) < 0
//	saturation value:    int32((uint32(a) >> 31) + 0x7fffffff)
//
// The saturation value derives the target bound from the SIGN OF a, not from
// the operation: (uint32(a)>>31) is 1 for negative a and 0 otherwise, so a
// negative a saturates to MinInt32 and a non-negative a to MaxInt32.
func TestSaturating32BitReferenceCorners(t *testing.T) {
	const (
		max32 = int32(math.MaxInt32)
		min32 = int32(math.MinInt32)
	)

	addCases := []struct {
		a, b, want int32
		why        string
	}{
		{max32, 1, max32, "positive overflow at the top"},
		{max32, max32, max32, "positive overflow, both maximal"},
		{min32, -1, min32, "negative overflow; signed >>31 would yield 2147483646"},
		{min32, min32, min32, "negative overflow, both minimal"},
		{0, min32, min32, "exact, no saturation"},
		{max32, min32, -1, "exact, no saturation"},
		{123, -23, 100, "passthrough"},
		{-1, 0, -1, "passthrough negative"},
	}
	for _, c := range addCases {
		if got := silk_ADD_SAT32(c.a, c.b); got != c.want {
			t.Errorf("silk_ADD_SAT32(%d, %d) = %d, want %d (%s)", c.a, c.b, got, c.want, c.why)
		}
	}

	subCases := []struct {
		a, b, want int32
		why        string
	}{
		{max32, -1, max32, "positive overflow; the add condition would miss it and return MinInt32"},
		{max32, -2, max32, "positive overflow by two"},
		{min32, 1, min32, "negative overflow; signed >>31 would yield 2147483646"},
		{min32, max32, min32, "negative overflow across the range"},
		{0, 0, 0, "exact"},
		{min32, min32, 0, "exact, no saturation"},
		{-1, 1, -2, "passthrough"},
		{max32, max32, 0, "exact"},
	}
	for _, c := range subCases {
		if got := silk_SUB_SAT32(c.a, c.b); got != c.want {
			t.Errorf("silk_SUB_SAT32(%d, %d) = %d, want %d (%s)", c.a, c.b, got, c.want, c.why)
		}
	}

	// The two operations must NOT share an overflow predicate: this pair is the
	// shortest discriminator between them.
	if got := silk_SUB_SAT32(max32, -1); got != max32 {
		t.Errorf("silk_SUB_SAT32(MaxInt32, -1) = %d, want MaxInt32; the add predicate was used", got)
	}
	if got := silk_ADD_SAT32(min32, -1); got != min32 {
		t.Errorf("silk_ADD_SAT32(MinInt32, -1) = %d, want MinInt32; the saturation value lost its unsigned conversion", got)
	}
}

// TestSaturating32BitReferenceExhaustiveBoundary walks every a in a window
// around both 32-bit boundaries against a set of b values that can and cannot
// overflow, so the predicate is checked on both sides of each edge.
func TestSaturating32BitReferenceExhaustiveBoundary(t *testing.T) {
	const max32 = int32(math.MaxInt32)
	const min32 = int32(math.MinInt32)

	bs := []int32{0, 1, -1, 2, -2, max32, min32, max32 - 1, min32 + 1}
	as := []int32{}
	for d := int32(-4); d <= 4; d++ {
		as = append(as, max32+d, min32+d)
	}
	as = append(as, 0, 1, -1)

	for _, a := range as {
		for _, b := range bs {
			// Reference computed in int64, with the documented target bound
			// chosen by the sign of a.
			sum := int64(a) + int64(b)
			bound := int64(max32)
			if a < 0 {
				bound = int64(min32)
			}
			wantAdd := int32(sum)
			if sum > int64(max32) || sum < int64(min32) {
				wantAdd = int32(bound)
			}
			if got := silk_ADD_SAT32(a, b); got != wantAdd {
				t.Errorf("silk_ADD_SAT32(%d, %d) = %d, want %d", a, b, got, wantAdd)
			}

			diff := int64(a) - int64(b)
			wantSub := int32(diff)
			if diff > int64(max32) || diff < int64(min32) {
				wantSub = int32(bound)
			}
			if got := silk_SUB_SAT32(a, b); got != wantSub {
				t.Errorf("silk_SUB_SAT32(%d, %d) = %d, want %d", a, b, got, wantSub)
			}
		}
	}
}
