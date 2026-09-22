package silk

import "testing"

func TestNSQDelDecReconstructOverflow(t *testing.T) {
	for _, states := range []int{1, 2, 3, 4} {
		for _, sign := range []int32{-1, 1} {
			// SMULWW wraps this product to zero before the scalar shift.
			// AVX2 retains it and saturates the final reconstruction instead.
			want := int16(0)
			if nsqDelDecWideReconstruction() && states >= 3 {
				want = 32767
				if sign < 0 {
					want = -32768
				}
			}
			for _, shift := range []int{8, 14} {
				if got := nsqDelDecReconstruct(sign*(1<<30), 1<<18, shift, states); got != want {
					t.Errorf("states=%d sign=%d shift=%d: got %d, want %d", states, sign, shift, got, want)
				}
			}
		}
	}
}
