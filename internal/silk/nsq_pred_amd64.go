//go:build amd64 && !purego

package silk

import "github.com/hujm2023/gopus/internal/cpufeat"

// silkUseShortTermPredictionSSE41 gates the SSE4.1 kernel on runtime CPU
// support. The kernel uses PMULDQ and PMOVSXWD, which are SSE4.1 instructions,
// while the amd64 baseline the Go toolchain targets is SSE2 only (GOAMD64=v1),
// so the kernel must never be called unconditionally.
var silkUseShortTermPredictionSSE41 = cpufeat.AMD64.HasSSE41

// shortTermPrediction16State is the order-16 short-term (LPC) prediction used by
// the noise-shaping quantizer. It dispatches to the SSE4.1 kernel when the CPU
// supports SSE4.1 and to the scalar reference otherwise; both produce identical
// results, which TestShortTermPrediction16StateDispatch asserts on every host.
func shortTermPrediction16State(sLPCQ14 *[maxSubFrameLength + nsqLpcBufLength]int32, idx int, aQ12 *[16]int16) int32 {
	if silkUseShortTermPredictionSSE41 {
		return shortTermPrediction16StateSSE41(sLPCQ14, idx, aQ12)
	}
	return shortTermPrediction16StateGo(sLPCQ14, idx, aQ12)
}

//go:noescape
func shortTermPrediction16StateSSE41(sLPCQ14 *[maxSubFrameLength + nsqLpcBufLength]int32, idx int, aQ12 *[16]int16) int32
