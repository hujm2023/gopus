package silk

// ControlBandwidth implements libopus silk_control_audio_bandwidth. Rates are
// in Hz; current may be zero after an encoder reset. The caller owns the Opus
// redundancy handshake and applies the selected rate to both SILK channels.
func (lp *LPState) ControlBandwidth(current, desired, minimum, maximum, api int, allow, canSwitch bool) (rate int, ready bool) {
	if current == 0 {
		current = int(lp.SavedFsKHz) * 1000
	}
	rate = current
	if rate == 0 {
		return min(desired, api), false
	}
	if rate > api || rate > maximum || rate < minimum {
		return max(min(api, maximum), minimum), false
	}
	if lp.TransitionFrameNo >= transitionFrames {
		lp.Mode = 0
	}
	if !allow && !canSwitch {
		return rate, false
	}
	switch {
	case current > desired:
		if lp.Mode == 0 {
			lp.TransitionFrameNo = transitionFrames
			lp.InLPState = [2]int32{}
		}
		if canSwitch {
			lp.Mode = 0
			rate = 8000
			if current == 16000 {
				rate = 12000
			}
		} else if lp.TransitionFrameNo <= 0 {
			ready = true
		} else {
			lp.Mode = -2
		}
	case current < desired:
		if canSwitch {
			rate = 16000
			if current == 8000 {
				rate = 12000
			}
			lp.TransitionFrameNo = 0
			lp.InLPState = [2]int32{}
			lp.Mode = 1
		} else if lp.Mode == 0 {
			ready = true
		} else {
			lp.Mode = 1
		}
	default:
		if lp.Mode < 0 {
			lp.Mode = 1
		}
	}
	return rate, ready
}
