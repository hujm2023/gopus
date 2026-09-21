//go:build amd64 && !purego

#include "textflag.h"

// func shortTermPrediction16StateSSE41(sLPCQ14 *int32, idx int, aQ12 *[16]int16) int32
//
// Order-16 short-term (LPC) prediction, mirroring libopus
// silk/x86/NSQ_del_dec_sse4_1.c. Every term is ((state * coef) >> 16)
// truncated to int32, and the result is the int32-wrapping sum of all sixteen
// terms plus the order>>1 bias, so regrouping the sum is bit-identical.
//
// PMULDQ multiplies the low 32 bits of each 64-bit lane, giving two signed
// 32x32 products per instruction; PSRLQ $16 then leaves bits 16..47 in the low
// half of each lane, which are exactly the low 32 bits of the arithmetic shift.
// Only int32 lanes 0 and 2 carry terms, so the reduction folds lane 2 into
// lane 0 and discards the high halves.
//
// States are read in ascending memory order from &sLPCQ14[idx-15], so the
// coefficients are reversed to pair ascending states with descending indices.
TEXT ·shortTermPrediction16StateSSE41(SB), NOSPLIT, $0-28
	MOVQ sLPCQ14+0(FP), SI
	MOVQ idx+8(FP), R8
	MOVQ aQ12+16(FP), R9

	// States: X0 = s[idx-15..idx-12], X1 = s[idx-11..idx-8],
	// X2 = s[idx-7..idx-4], X3 = s[idx-3..idx].
	LEAQ (SI)(R8*4), AX
	SUBQ $60, AX
	MOVOU 0(AX), X0
	MOVOU 16(AX), X1
	MOVOU 32(AX), X2
	MOVOU 48(AX), X3

	// Coefficients: sixteen int16 sign-extended to int32, then reversed within
	// each four-lane group.
	MOVOU 0(R9), X4
	MOVOU 16(R9), X5
	PMOVSXWD X4, X6
	PSRLDQ   $8, X4
	PMOVSXWD X4, X7
	PMOVSXWD X5, X8
	PSRLDQ   $8, X5
	PMOVSXWD X5, X9
	PSHUFD   $0x1B, X6, X6 // c3, c2, c1, c0
	PSHUFD   $0x1B, X7, X7 // c7, c6, c5, c4
	PSHUFD   $0x1B, X8, X8 // c11, c10, c9, c8
	PSHUFD   $0x1B, X9, X9 // c15, c14, c13, c12

	PXOR X10, X10 // accumulator

	// DOT4 pairs four states with four coefficients: one PMULDQ covers lanes 0
	// and 2, and rotating both operands by one lane covers lanes 1 and 3.
#define DOT4(sreg, creg, t1, t2) \
	MOVOU sreg, t1;              \
	PMULDQ creg, t1;             \
	PSRLQ $16, t1;               \
	PADDD t1, X10;               \
	MOVOU sreg, t2;              \
	PSHUFD $0x39, t2, t2;        \
	MOVOU creg, t1;              \
	PSHUFD $0x39, t1, t1;        \
	PMULDQ t1, t2;               \
	PSRLQ $16, t2;               \
	PADDD t2, X10

	DOT4(X0, X9, X11, X12)
	DOT4(X1, X8, X11, X12)
	DOT4(X2, X7, X11, X12)
	DOT4(X3, X6, X11, X12)

	// Fold lane 2 into lane 0 and take the low lane: the odd lanes hold the high
	// halves of the 64-bit products and are not part of the sum.
	PSHUFD $0x0E, X10, X11
	PADDD  X11, X10
	MOVD   X10, AX
	ADDL   $8, AX // predictLPCOrder >> 1
	MOVL   AX, ret+24(FP)
	RET
