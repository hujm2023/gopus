package multistream

import "testing"

func TestHybridToSilkFadeOutRequiresPreviousFrame(t *testing.T) {
	for _, reset := range []bool{false, true} {
		name := "fresh"
		if reset {
			name = "reset"
		}
		t.Run(name, func(t *testing.T) {
			d := newStreamDecoder(48000, 1)
			if reset {
				d.haveDecoded = true
				d.Reset()
			}
			out := make([]float32, 480)
			if err := d.addHybridToSilkFadeOut(out); err != nil {
				t.Fatal(err)
			}
			for i, sample := range out {
				if sample != 0 {
					t.Fatalf("sample %d = %g: no previous Hybrid frame exists to fade out", i, sample)
				}
			}
		})
	}
}
