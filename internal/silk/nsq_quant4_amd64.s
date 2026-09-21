//go:build amd64 && !purego

#include "textflag.h"

// Four independent states, two candidates each; silk/NSQ_del_dec.c arithmetic.
// All adds/multiplies wrap at 32 bits. SMULBB operands are explicitly narrowed
// to signed 16 bits, including lambda. Equal costs select candidate 2 first.
TEXT ·quantize4SSE41(SB), NOSPLIT, $0-16
	MOVQ batch+0(FP), AX
	MOVL offset+8(FP), CX
	MOVD CX, X14
	PSHUFD $0, X14, X14
	MOVL lambda+12(FP), DX
	MOVOU 0(AX), X1 // residual
	MOVOU X1, X2
	PSUBL X14, X2 // residual-offset
	PXOR X15, X15
	CMPL DX, $2048
	JLE no_rdo
	MOVL DX, CX
	SARL $1, CX
	SUBL $512, CX
	MOVD CX, X3
	PSHUFD $0, X3, X3 // rdo offset
	MOVOU X2, X4
	PSRAL $31, X4 // sign
	MOVOU X3, X5
	PXOR X4, X5
	PSUBL X4, X5 // signed rdo offset
	MOVOU X2, X6
	PSUBL X5, X6
	PSRAL $10, X6
	MOVOU X2, X7
	PABSD X7, X7
	PCMPGTL X3, X7 // abs(delta)>rdo
	PXOR X4, X6
	PAND X7, X6
	PXOR X4, X6
	MOVOU X6, X2
	JMP levels
no_rdo:
	PSRAL $10, X2
levels:
	MOVL $80, CX
	MOVD CX, X13
	PSHUFD $0, X13, X13
	// q1 = q*1024+offset + (q<0 ? 80 : q>0 ? -80 : 0).
	MOVOU X2, X3
	PSRAL $31, X3
	PAND X13, X3
	MOVOU X2, X4
	PCMPGTL X15, X4
	PAND X13, X4
	MOVOU X2, X5
	PSLLL $10, X5
	PADDD X14, X5
	PADDD X3, X5
	PSUBL X4, X5 // q1
	PCMPEQL X12, X12
	PSUBL X12, X2 // q+1
	MOVOU X2, X3
	PSRAL $31, X3
	PAND X13, X3
	MOVOU X2, X4
	PCMPGTL X15, X4
	PAND X13, X4
	PSLLL $10, X2
	PADDD X14, X2
	PADDD X3, X2
	PSUBL X4, X2 // q2
	MOVD DX, X14
	PSLLL $16, X14
	PSRAL $16, X14
	PSHUFD $0, X14, X14
	MOVOU X5, X6
	PABSD X6, X6
	PSLLL $16, X6
	PSRAL $16, X6
	PMULLD X14, X6
	MOVOU X1, X7
	PSUBL X5, X7
	PSLLL $16, X7
	PSRAL $16, X7
	PMULLD X7, X7
	PADDD X7, X6
	PSRAL $10, X6 // rd1
	MOVOU X2, X7
	PABSD X7, X7
	PSLLL $16, X7
	PSRAL $16, X7
	PMULLD X14, X7
	PSUBL X2, X1
	PSLLL $16, X1
	PSRAL $16, X1
	PMULLD X1, X1
	PADDD X1, X7
	PSRAL $10, X7 // rd2
	MOVOU X7, X8
	PCMPGTL X6, X8 // rd1 < rd2 (strict)
	MOVOU X5, X9
	PXOR X2, X9
	PAND X8, X9
	MOVOU X2, X10
	PXOR X9, X10
	PXOR X9, X5
	MOVOU X10, 32(AX) // best q
	MOVOU X5, 48(AX) // second q
	MOVOU X6, X9
	PXOR X7, X9
	PAND X8, X9
	MOVOU X7, X10
	PXOR X9, X10
	PXOR X9, X6
	MOVOU 16(AX), X11
	PADDD X11, X10
	PADDD X11, X6
	MOVOU X10, 64(AX)
	MOVOU X6, 80(AX)
	RET
