//go:build amd64 && !purego

package silk

import "github.com/hujm2023/gopus/internal/cpufeat"

func quantize4(batch *nsqQuant4, offset, lambda int32) {
	if cpufeat.AMD64.HasSSE41 {
		quantize4SSE41(batch, offset, lambda)
		return
	}
	quantize4Go(batch, offset, lambda)
}

//go:noescape
func quantize4SSE41(batch *nsqQuant4, offset, lambda int32)
