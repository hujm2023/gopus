//go:build amd64 && !purego

package silk

import "github.com/hujm2023/gopus/internal/cpufeat"

// quantize4 selects the four-int32-lane kernel only with runtime SSE4.1 support.
// GOAMD64=v1 alone does not permit PMULLD. The purego/non-amd64 build uses the
// same scalar reference, so callers need no architecture-specific state layout.
func quantize4(batch *nsqQuant4, offset, lambda int32) {
	if cpufeat.AMD64.HasSSE41 {
		quantize4SSE41(batch, offset, lambda)
		return
	}
	quantize4Go(batch, offset, lambda)
}

//go:noescape
func quantize4SSE41(batch *nsqQuant4, offset, lambda int32)
