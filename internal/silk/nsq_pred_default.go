//go:build !amd64 || purego

package silk

// shortTermPrediction16State is the order-16 short-term (LPC) prediction used by
// the noise-shaping quantizer. Targets without an amd64 SSE4.1 kernel use the
// scalar reference directly.
func shortTermPrediction16State(sLPCQ14 *[maxSubFrameLength + nsqLpcBufLength]int32, idx int, aQ12 *[16]int16) int32 {
	return shortTermPrediction16StateGo(sLPCQ14, idx, aQ12)
}
