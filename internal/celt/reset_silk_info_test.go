package celt

import "testing"

func TestResetClearsHybridSideInformation(t *testing.T) {
	enc := NewEncoder(1)
	enc.SetSilkInfo(2, 100)
	enc.Reset()
	signal, offset := enc.SilkInfo()
	if signal != 0 || offset != 0 {
		t.Fatalf("reset retained SILK info: signal=%d offset=%d", signal, offset)
	}
}
