//go:build !amd64 || purego

package silk

// warpedARFeedback24States4 computes the 24-tap warped AR feedback of all
// delayed-decision states on the scalar path.
func warpedARFeedback24States4(ar *nsqWarpAR, c *[24]int16, w int32) {
	warpedARFeedback24States4Go(ar, c, w)
}
