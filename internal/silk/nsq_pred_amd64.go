//go:build amd64 && !purego

package silk

// shortTermPrediction16State is the order-16 short-term (LPC) prediction used by
// the noise-shaping quantizer. On amd64 it runs the SSE4.1 kernel; the scalar
// reference is shortTermPrediction16StateGo, which the parity test compares
// against bit for bit.
func shortTermPrediction16State(sLPCQ14 *[maxSubFrameLength + nsqLpcBufLength]int32, idx int, aQ12 *[16]int16) int32 {
	return shortTermPrediction16StateSSE41(sLPCQ14, idx, aQ12)
}

//go:noescape
func shortTermPrediction16StateSSE41(sLPCQ14 *[maxSubFrameLength + nsqLpcBufLength]int32, idx int, aQ12 *[16]int16) int32
