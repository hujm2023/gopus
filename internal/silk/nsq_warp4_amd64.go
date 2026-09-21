//go:build amd64 && !purego

package silk

import "github.com/hujm2023/gopus/internal/cpufeat"

// warpedARFeedback24States4 dispatches to the SSE4.1 kernel when the running CPU
// has SSE4.1 and to the scalar routine otherwise. The Go amd64 baseline is SSE2,
// so the kernel's PMULDQ and PMOVSXDQ must be gated at runtime.
func warpedARFeedback24States4(ar *nsqWarpAR, c *[24]int16, w int32) {
	if cpufeat.AMD64.HasSSE41 {
		warpedARFeedback24States4SSE41(ar, c, w)
		return
	}
	warpedARFeedback24States4Go(ar, c, w)
}

//go:noescape
func warpedARFeedback24States4SSE41(ar *nsqWarpAR, c *[24]int16, w int32)
