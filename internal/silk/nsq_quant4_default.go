//go:build !amd64 || purego

package silk

func quantize4(batch *nsqQuant4, offset, lambda int32) { quantize4Go(batch, offset, lambda) }

func nsqDelDecWideReconstruction() bool { return false }
