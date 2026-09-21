//go:build amd64 && !purego

package cpufeat

func init() {
	_, _, ecx1, _ := cpuid(1, 0)

	// SSE4.1 is CPUID.01H:ECX bit 19. It needs neither AVX nor the OS XSAVE
	// state, so it is recorded before the AVX early-outs below.
	AMD64.HasSSE41 = isSet(19, ecx1)

	hasOSXSAVE := isSet(27, ecx1)
	hasAVX := isSet(28, ecx1)
	hasFMA := isSet(12, ecx1)

	if !hasOSXSAVE || !hasAVX {
		return
	}

	eax, _ := xgetbv()
	osSupportsAVX := isSet(1, eax) && isSet(2, eax)
	if !osSupportsAVX {
		return
	}

	_, ebx7, _, _ := cpuid(7, 0)
	AMD64.HasAVX2 = isSet(5, ebx7)
	AMD64.HasFMA = hasFMA
}

func isSet(bitpos uint, value uint32) bool {
	return value&(1<<bitpos) != 0
}

func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)
func xgetbv() (eax, edx uint32)
