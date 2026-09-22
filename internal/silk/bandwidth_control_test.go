package silk

import "testing"

func TestControlBandwidth(t *testing.T) {
	// Decisions from libopus silk/control_audio_bandwidth.c. In particular,
	// desired bandwidth alone never changes a valid rate during active speech.
	cases := []struct {
		name                                    string
		lp                                      LPState
		current, desired, minimum, maximum, api int
		allow, can                              bool
		wantRate                                int
		wantReady                               bool
		wantMode                                int
		wantFrame                               int32
	}{
		{"initial", LPState{}, 0, 12000, 8000, 16000, 48000, false, false, 12000, false, 0, 0},
		{"active speech holds", LPState{}, 12000, 16000, 8000, 16000, 48000, false, false, 12000, false, 0, 0},
		{"up ready", LPState{}, 12000, 16000, 8000, 16000, 48000, true, false, 12000, true, 0, 0},
		{"up permitted one step", LPState{}, 8000, 16000, 8000, 16000, 48000, false, true, 12000, false, 1, 0},
		{"down start", LPState{}, 16000, 8000, 8000, 16000, 48000, true, false, 16000, false, -2, 256},
		{"down ready", LPState{Mode: -2}, 16000, 8000, 8000, 16000, 48000, true, false, 16000, true, -2, 0},
		{"down permitted one step", LPState{Mode: -2}, 16000, 8000, 8000, 16000, 48000, false, true, 12000, false, 0, 0},
		{"reverse down", LPState{Mode: -2, TransitionFrameNo: 80}, 16000, 16000, 8000, 16000, 48000, true, false, 16000, false, 1, 80},
		{"finish up", LPState{Mode: 1, TransitionFrameNo: 256}, 16000, 16000, 8000, 16000, 48000, false, false, 16000, false, 0, 256},
		{"forced maximum", LPState{}, 16000, 8000, 8000, 12000, 48000, false, false, 12000, false, 0, 0},
		{"forced hybrid minimum", LPState{}, 12000, 16000, 16000, 16000, 48000, false, false, 16000, false, 0, 0},
		{"API ceiling", LPState{}, 16000, 16000, 8000, 16000, 12000, false, false, 12000, false, 0, 0},
		{"prefill saved rate", LPState{SavedFsKHz: 12}, 0, 16000, 8000, 16000, 48000, false, true, 16000, false, 1, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lp := tc.lp
			rate, ready := lp.ControlBandwidth(tc.current, tc.desired, tc.minimum, tc.maximum, tc.api, tc.allow, tc.can)
			if rate != tc.wantRate || ready != tc.wantReady || lp.Mode != tc.wantMode || lp.TransitionFrameNo != tc.wantFrame {
				t.Fatalf("rate=%d ready=%t LP=%+v; want rate=%d ready=%t mode=%d frame=%d", rate, ready, lp, tc.wantRate, tc.wantReady, tc.wantMode, tc.wantFrame)
			}
		})
	}
}
