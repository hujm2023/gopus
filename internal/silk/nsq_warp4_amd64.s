//go:build amd64 && !purego

#include "textflag.h"

// func warpedARFeedback24States4SSE41(ar *nsqWarpAR, c *[24]int16, w int32)
//
// Four-state warped AR feedback, one tap index across the four states held in a
// 128-bit register. ar.state is the transposed state: row j is 32 bytes holding
// the four states, one 64-bit slot each, with the int32 value in the low 32 bits.
// ar.diff holds the per-state diff_Q14 inputs and ar.out receives the results, so
// the whole kernel needs a single pointer and no frame-relative vector loads.
//
// Every multiply is ((int32) * (int16) >> 16): the caller truncates warpingQ16 to
// int16 before the call, so the plain 64-bit product shifted right by 16 is
// exactly silk_SMULWB. PMULDQ multiplies the low 32 bits of each 64-bit lane, so
// one PMULDQ covers two states exactly, and PSRLQ $16 leaves bits 16..47 in the
// low half of each lane -- the low 32 bits of the arithmetic shift. No lane
// crossing or repacking is needed, so the serial recurrence chain is
// PSUBL -> PMULDQ -> PSRLQ -> PADDD.
//
// Alignment: nsqWarpAR is 8-byte aligned (its widest field is int64), so state
// rows are only 16-byte aligned by luck. Every vector load and store uses MOVOU,
// which is unaligned-safe, and every arithmetic instruction takes register
// operands only: legacy SSE integer arithmetic with a memory operand raises #GP
// on a misaligned address.
//
// nsqWarpAR layout: state at 0 (768 bytes), diff at 768 (32 bytes), out at 800
// (16 bytes).
TEXT ·warpedARFeedback24States4SSE41(SB), NOSPLIT, $0-20
	MOVQ ar+0(FP), DI
	MOVQ c+8(FP), SI

	// X0 = w broadcast to all four 32-bit lanes.
	MOVL w+16(FP), AX
	MOVQ AX, X0
	PSHUFD $0, X0, X0

	// --- prologue -------------------------------------------------------
	// t2 = diff + smulwb(state[0], w)
	MOVOU 768(DI), X1 // diff, states 0..1
	MOVOU 784(DI), X2 // diff, states 2..3
	MOVOU 0(DI), X3   // state[0], states 0..1  (kept for t1)
	MOVOU 16(DI), X4  // state[0], states 2..3
	MOVOU X3, X7
	PMULDQ X0, X7
	PSRLQ $16, X7
	PADDD X7, X1 // X1 = t2, states 0..1
	MOVOU X4, X8
	PMULDQ X0, X8
	PSRLQ $16, X8
	PADDD X8, X2 // X2 = t2, states 2..3

	// t1 = state[0] + smulwb(state[1] - t2, w)   (uses the original state[0])
	MOVOU 32(DI), X7
	MOVOU 48(DI), X8
	PSUBL X1, X7
	PSUBL X2, X8
	PMULDQ X0, X7
	PSRLQ $16, X7
	PADDD X3, X7 // X7 = t1, states 0..1
	PMULDQ X0, X8
	PSRLQ $16, X8
	PADDD X4, X8 // X8 = t1, states 2..3
	MOVOU X7, X3
	MOVOU X8, X4

	// state[0] = t2
	MOVOU X1, 0(DI)
	MOVOU X2, 16(DI)

	// acc = 12 + smulwb(t2, c[0])
	MOVWQSX 0(SI), AX
	MOVQ AX, X10
	PSHUFD $0, X10, X10
	MOVOU X1, X7
	PMULDQ X10, X7
	PSRLQ $16, X7
	MOVOU X2, X8
	PMULDQ X10, X8
	PSRLQ $16, X8
	MOVL $12, AX
	MOVQ AX, X5
	PSHUFD $0, X5, X5
	MOVOU X5, X6
	PADDD X7, X5
	PADDD X8, X6

	// --- j = 2, 4, ..., 22 ----------------------------------------------
	MOVQ $64, R9   // 32*j, j starts at 2
	MOVQ $4, R10   // 2*j, j starts at 2
	MOVQ $768, R11 // 32*(22+2): the value R9 takes after the last iteration
loop:
	// state[j] and state[j-1] are loaded with MOVOU so that no arithmetic
	// instruction needs an aligned memory operand.
	MOVOU (DI)(R9*1), X9
	MOVOU 16(DI)(R9*1), X13
	MOVOU -32(DI)(R9*1), X14
	MOVOU -16(DI)(R9*1), X15

	// t2 = state[j-1] + smulwb(state[j] - t1, w)
	MOVOU X9, X7
	PSUBL X3, X7
	PMULDQ X0, X7
	PSRLQ $16, X7
	PADDD X14, X7
	MOVOU X13, X8
	PSUBL X4, X8
	PMULDQ X0, X8
	PSRLQ $16, X8
	PADDD X15, X8

	// state[j-1] = t1
	MOVOU X3, -32(DI)(R9*1)
	MOVOU X4, -16(DI)(R9*1)

	// acc += smulwb(t1, c[j-1])
	MOVWQSX -2(SI)(R10*1), AX
	MOVQ AX, X10
	PSHUFD $0, X10, X10
	MOVOU X3, X11
	PMULDQ X10, X11
	PSRLQ $16, X11
	MOVOU X4, X12
	PMULDQ X10, X12
	PSRLQ $16, X12
	PADDD X11, X5
	PADDD X12, X6

	// t1 = state[j] + smulwb(state[j+1] - t2, w)
	MOVOU 32(DI)(R9*1), X3
	MOVOU 48(DI)(R9*1), X4
	PSUBL X7, X3
	PMULDQ X0, X3
	PSRLQ $16, X3
	PADDD X9, X3
	PSUBL X8, X4
	PMULDQ X0, X4
	PSRLQ $16, X4
	PADDD X13, X4

	// state[j] = t2
	MOVOU X7, (DI)(R9*1)
	MOVOU X8, 16(DI)(R9*1)

	// acc += smulwb(t2, c[j])
	MOVWQSX (SI)(R10*1), AX
	MOVQ AX, X10
	PSHUFD $0, X10, X10
	MOVOU X7, X11
	PMULDQ X10, X11
	PSRLQ $16, X11
	MOVOU X8, X12
	PMULDQ X10, X12
	PSRLQ $16, X12
	PADDD X11, X5
	PADDD X12, X6

	ADDQ $64, R9 // j += 2
	ADDQ $4, R10 // coefficient index += 2
	CMPQ R9, R11
	JNE  loop

	// --- epilogue -------------------------------------------------------
	// state[23] = t1
	MOVOU X3, 736(DI)
	MOVOU X4, 752(DI)

	// acc += smulwb(t1, c[23])
	MOVWQSX 46(SI), AX
	MOVQ AX, X10
	PSHUFD $0, X10, X10
	MOVOU X3, X11
	PMULDQ X10, X11
	PSRLQ $16, X11
	MOVOU X4, X12
	PMULDQ X10, X12
	PSRLQ $16, X12
	PADDD X11, X5
	PADDD X12, X6

	// X5 = [a0, ?, a1, ?], X6 = [a2, ?, a3, ?] -> [a0, a1, a2, a3]
	PSHUFD    $0x88, X5, X5
	PSHUFD    $0x88, X6, X6
	PUNPCKLLQ X6, X5
	PSHUFD    $0xD8, X5, X5
	MOVOU     X5, 800(DI)
	RET
