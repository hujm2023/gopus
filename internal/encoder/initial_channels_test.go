package encoder

import "testing"

func TestFirstFrameHasNoStereoTransition(t *testing.T) {
	enc := NewEncoder(48000, 2)
	enc.SetBitrate(14000)
	enc.SetForceChannels(1)
	for _, reset := range []bool{false, true} {
		if reset {
			enc.Reset()
		}
		packet, err := enc.EncodeFloat32(make([]float32, 480*2), 480)
		if err != nil {
			t.Fatal(err)
		}
		if len(packet) == 0 || packet[0]&4 != 0 || enc.toMono != 0 {
			t.Fatalf("reset=%v: spurious stereo transition: packet=%x toMono=%d", reset, packet, enc.toMono)
		}
	}
}
