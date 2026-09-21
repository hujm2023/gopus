# SILK NSQ performance implementation

The NSQ fast paths reduce repeated scalar work while preserving the quantizer's
integer results and state transitions. They operate inside one encoder instance;
none of the four SIMD lanes represents another stream or a future audio sample.
The generic path and the scalar kernels remain executable references.

## Where the kernels run

`noiseShapeQuantizerDelDec` selects the two specialized sample loops only for
24 shaping taps, 16 LPC prediction taps, four delayed-decision states, and valid
history windows. One loop handles voiced input, the other unvoiced input.
Other configurations use `noiseShapeQuantizerDelDecGeneric`.

Inside each specialized loop, one sample passes through:

1. Warped AR feedback for the four competing histories.
2. Per-state LPC prediction, feedback, dithering and residual clamping.
3. Batched two-candidate quantization for those four residuals.
4. Candidate reconstruction, history replacement and winner/state updates.

Step 4 completes before the next sample starts. Vectorizing across consecutive
samples would violate this dependency. The generic loop can also use the order-16
LPC helper, but its four-state AR helper retains the strided state representation.

Algorithm selection and CPU selection are separate. On amd64 without `purego`,
`nsq_pred_amd64.go`, `nsq_quant4_amd64.go` and `nsq_warp4_amd64.go` check runtime
SSE4.1 support. `GOAMD64=v1` only guarantees SSE2. The `purego` and non-amd64
variants call scalar Go implementations of these helpers.

## Order-16 LPC prediction: independent products and wrapping sums

Source: [nsq_pred.go](nsq_pred.go), [nsq_pred_amd64.s](nsq_pred_amd64.s).

The result is `8 + sum(int32((int64(state) * int64(coef)) >> 16))`, with 16
products, signed int16 coefficients and int32-wrapping additions. The scalar
implementation uses four accumulator chains, allowing independent multiplications
to proceed without a single long add dependency. The SSE4.1 implementation
computes signed products with `PMULDQ`, shifts each product, then folds the
useful int32 lanes.

The order of the *wrapping integer additions* may change; each product's shift
and truncation must stay before the sum. Accumulating full products and shifting
once would introduce carries from low bits that the reference discards.
This integer argument does not apply to floating-point reductions.

The assembly loads history in ascending memory order and reverses coefficients
to preserve the original state/coefficient pairing. `PMULDQ` uses the low int32
of each 64-bit lane. After `PSRLQ $16`, bits 16..47 occupy the low int32 and match
the low 32 bits of the arithmetic-shift result, including negative products.
The high halves are not terms in the sum.

## Four-state candidate quantization: batch only the independent block

Source: [nsq_quant4.go](nsq_quant4.go), [nsq_quant4_amd64.s](nsq_quant4_amd64.s),
[nsq_del_dec.go](nsq_del_dec.go).

`nsqQuant4` stores one four-element array per field: residual, accumulated cost,
two quantized outputs and two costs. This contiguous layout lets the SIMD kernel
process four int32 lanes without gathering fields from separate history structs.
The caller prepares the residuals and restores the results to per-state candidate
records; that packing and reconstruction still has a cost.

Each residual is dithered and clamped to `[-31<<10, 30<<10]` before the call.
Offsets are the codec's 32, 100 or 240 values. The kernel preserves:

- The RDO branch at `lambda > 2048` and the signed quantization-level cases.
- Signed int16 operand narrowing in `SMULBB`/`SMLABB`, including lambda, even
  when a positive int32 operand does not fit in int16.
- Int32 wraparound and arithmetic right shifts at the same calculation stages.
- Strict `rd1 < rd2`: equal costs put candidate 2 first. A different tie rule
  changes the subsequent history selection and can change the bitstream.

The kernel does not batch candidate reconstruction, delayed-decision pruning or
state replacement. Measuring `quantize4` alone therefore cannot establish the
speedup of the complete NSQ loop or encoder.

## Warped AR: preserve the recurrence, transpose the independent histories

Source: [nsq_warp.go](nsq_warp.go), [nsq_warp4_amd64.s](nsq_warp4_amd64.s),
[NSQState.delDecAR](nsq.go), and the two specialized loops in
[nsq_del_dec.go](nsq_del_dec.go).

A state's next tap depends on the previous tap. The four histories are independent
at that stage, so the vectorization dimension is the state index rather than tap
or sample time. `state[j][k]` stores tap `j` of state `k` in a 64-bit slot; only
the low 32 bits represent the signed state value.

| One tap row, 32 bytes | First 128-bit register | Second 128-bit register |
| --- | --- | --- |
| State lanes | states 0 and 1, one 64-bit slot each | states 2 and 3, one 64-bit slot each |
| Meaningful value | low int32 of each slot | low int32 of each slot |

Two `PMULDQ` operations cover all four states. The padded layout avoids shuffling
between every multiply and recurrence update, at the cost of 768 bytes for the
24-by-4 tap matrix instead of 384 bytes for tightly packed int32 values.
`diff` and `out` add 32 and 16 bytes, respectively, to the workspace.

The caller truncates warping to signed int16 before dispatch. Each multiply then
matches `SMULWB`; int32 subtraction/addition and per-product truncation preserve
the scalar recurrence. Read state through `get` and write through `set`: high
slot bits left by SIMD arithmetic are not a valid int64 state value.

`NSQState` owns this reusable workspace. A specialized subframe copies its
histories in once, makes the transposed values authoritative throughout the sample
loop, and writes them back once. Copying a winning history requires copying its
AR column as well as `nsqDelDecStateTail`; the latter contains stale AR values
until writeback. Entry/exit transposition and column copies belong in full-path
performance measurements. A kernel-only timing excludes them.

The struct's alignment is eight bytes. Vector memory accesses use `MOVOU`, and
legacy SSE integer arithmetic uses register operands. An arithmetic instruction
with an unaligned memory operand can fault; heap allocation does not guarantee
that every such address is aligned to 16 bytes. Layout changes must keep the
assembly offsets (`state=0`, `diff=768`, `out=800`) synchronized.

## Correctness and allocation checks

| Constraint | Existing runnable check |
| --- | --- |
| Order-16 scalar/assembly equality and CPU fallback | `TestShortTermPrediction16StateKernelParity`, `TestShortTermPrediction16StateDispatch` |
| Quantization residuals, signed narrowing, ties and allocations | `TestQuantize4Parity` |
| Warped AR result **and mutated history** | `TestWarpedARFeedback24States4MatchesScalar`, `TestWarpedARFeedback24States4Corners` |
| Repeated calls and stack/heap state forms | `TestWarpedARFeedback24States4Repeat`, `TestWarpedARFeedback24States4StackState` |
| Warm dispatch/scalar kernel allocations | `TestWarpedARFeedback24States4Allocs` |

For these local contracts, run the focused set in default and scalar builds:

```sh
go test ./internal/silk -run '^(TestShortTermPrediction16.*|TestQuantize4Parity|TestWarpedARFeedback24States4.*)$' -count=1
go test -tags purego ./internal/silk -run '^(TestShortTermPrediction16.*|TestQuantize4Parity|TestWarpedARFeedback24States4.*)$' -count=1
```

Kernel equality is one layer of evidence. Caller changes also require complete
state/bitstream and encoder-path validation. Keep the project's live C oracle,
quality and release gates; a Go-to-Go comparison does not replace them.

Zero-allocation kernel checks warm the state first. The complete encoder may
initialize SILK lazily on its first encode, so benchmark initialization, operation
count and allocation units must be explicit. `testing.B.ResetTimer` resets
allocation counters as well as time; allocations after it are still measured.
A zero-allocation inner kernel does not imply that an adapter or RTP pipeline
allocates nothing.
