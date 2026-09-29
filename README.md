# gopus

Pure-Go Opus codec — RFC 6716 / RFC 8251, targeting bit-exact and quality parity
with pinned libopus 1.6.1, with no cgo.

Encoder, decoder, multistream, projection/ambisonics, Ogg, and RTP RED — all in
plain Go, with caller-owned, zero-allocation encode and decode hot paths. Codec
math and bitstream decisions are checked against the pinned reference by a
live C oracle (see [Parity & testing](#parity--testing)).

## Fork origin and maintenance

[`github.com/hujm2023/gopus`](https://github.com/hujm2023/gopus) is an
independently maintained fork of [`thesyncim/gopus`](https://github.com/thesyncim/gopus).
The upstream base is
[`bd8db897c2681f6b5e3e5b9d16cf7f9e50f32781`](https://github.com/thesyncim/gopus/commit/bd8db897c2681f6b5e3e5b9d16cf7f9e50f32781);
its immediate child, [`55ebdc97`](https://github.com/hujm2023/gopus/commit/55ebdc971105717b15b8ac76f8d5dc04024485ca),
sets this fork's module path to `github.com/hujm2023/gopus`. Use that path to
consume this fork. The [BSD 3-Clause license and upstream copyright](LICENSE)
apply here.

Beyond that base, this fork maintains these additional changes:

| Area | Fork changes and source evidence |
| --- | --- |
| SIMD / NSQ performance | CPU-gated SSE4.1 short-term prediction, batched delayed-decision candidate costs, and four-state warped AR feedback. See [dispatch](internal/silk/nsq_pred_amd64.go), [candidate kernel](internal/silk/nsq_quant4.go), and [warped AR checks](internal/silk/nsq_warp_test.go). |
| Stateful SILK / CELT / Hybrid transitions | Mode-switch prefill, stereo/mono history, shared channel-0 VAD state, redundancy, DTX and VBR state continuity. See [encoder](internal/encoder/encoder.go), [VAD continuity checks](internal/encoder/hybrid_stereo_vad_test.go), and [stateful transition oracle](encode_stateful_transition_fuzz_test.go). |
| Native rates, multistream and controls | Native-rate Hybrid preprocessing, packet/surround/projection budgets, and bandwidth/frame/channel control alignment. See [native-rate checks](internal/celt/hybrid_native_rate_test.go), [multistream budget checks](multistream/surround_cbr_budget_test.go), and [control alignment](https://github.com/hujm2023/gopus/commit/f6a2335a). |
| Test and release harness | Cross-platform oracle and fixture checks, PGO profile validation, documentation contracts, and release verification tied to the exact commit SHA. See [CI and releases](.github/CI.md) and [release evidence tooling](tools/gen_release_evidence.sh). |

These links describe implementation and test coverage; passing results belong to
the tested commit, build tags, architecture and oracle configuration. See
[Trust And Verification](#trust-and-verification) for this fork's release evidence.

## Install

```sh
go get github.com/hujm2023/gopus
```

Requires Go 1.25 or newer.

## Quick start

The hot-path API takes caller-owned buffers and returns the number of bytes /
samples written, so the encode and decode loops allocate nothing:

```go
func (e *Encoder) Encode(pcm []float32, data []byte) (int, error)
func (d *Decoder) Decode(data []byte, pcm []float32) (int, error)
```

Encode one 20 ms stereo frame at 48 kHz, then decode it back:

```go
package main

import (
	"log"

	"github.com/hujm2023/gopus"
)

func main() {
	const (
		sampleRate = 48000
		channels   = 2
		frameSize  = 960 // 20 ms at 48 kHz
	)

	enc, err := gopus.NewEncoder(gopus.EncoderConfig{
		SampleRate:  sampleRate,
		Channels:    channels,
		Application: gopus.ApplicationAudio,
	})
	if err != nil {
		log.Fatal(err)
	}

	dec, err := gopus.NewDecoder(gopus.DefaultDecoderConfig(sampleRate, channels))
	if err != nil {
		log.Fatal(err)
	}

	pcm := make([]float32, frameSize*channels) // your interleaved input
	packet := make([]byte, 4000)               // reusable encode buffer
	out := make([]float32, frameSize*channels) // reusable decode buffer

	n, err := enc.Encode(pcm, packet) // n = bytes written to packet
	if err != nil {
		log.Fatal(err)
	}

	samples, err := dec.Decode(packet[:n], out) // samples = per-channel samples
	if err != nil {
		log.Fatal(err)
	}
	_ = out[:samples*channels]
}
```

`int16` and `int24` PCM use the same caller-buffer shape — only the slice
element type changes:

```go
pcm16 := make([]int16, frameSize*channels) // interleaved 16-bit input
packet := make([]byte, 4000)
out16 := make([]int16, frameSize*channels)

n, err := enc.EncodeInt16(pcm16, packet)   // also EncodeInt24([]int32, …)
// …
samples, err := dec.DecodeInt16(packet[:n], out16) // also DecodeInt24(…, []int32)
```

Tune the encoder through libopus-style CTL methods (`SetBitrate`, `SetVBR`,
`SetComplexity`, `SetInBandFEC`, `SetDTX`, …). Pass a nil packet to `Decode` to
run packet-loss concealment for a dropped frame.

See [examples/](examples/) for Ogg files, ffmpeg interop, RED loss recovery,
WebRTC control, and benchmarks.

## Features

| Area | gopus |
| --- | --- |
| Coding modes | SILK, CELT, Hybrid, with automatic mode selection |
| Sample rates | 8, 12, 16, 24, 48 kHz (native sub-48 kHz encode, no upsampling) |
| Channels | Mono, stereo, multistream, projection / ambisonics |
| Frame sizes | 2.5–120 ms |
| Bitrate control | CBR, VBR, CVBR, low-delay, DTX |
| Resilience | Packet loss concealment, in-band FEC / LBRR |
| PCM formats | `float32`, `int16`, `int24` (single-stream and multistream) |
| Containers | `container/ogg` (Ogg read/write), `container/red` (RFC 2198 RTP RED parse/build/recover) |
| libopus surface | Full public API: the libopus CTL surface, packet parsing, soft clipping, and matching error codes |

## Public API

The importable surface is four packages. Everything else lives under `internal/`
and is not importable.

| Package | Use it for |
| --- | --- |
| `github.com/hujm2023/gopus` | `Encoder` / `Decoder` (float32 / int16 / int24), streaming `Reader` / `Writer`, multistream and DRED constructors, packet parsing, repacketizer, soft clip, CTLs, error codes |
| `github.com/hujm2023/gopus/multistream` | Lower-level multistream `Encoder` / `Decoder` and projection / ambisonics (`NewProjectionEncoder` / `NewProjectionDecoder`) |
| `github.com/hujm2023/gopus/container/ogg` | Read and write Ogg Opus files (RFC 7845) |
| `github.com/hujm2023/gopus/container/red` | `Encoder` / `Decoder` structs (plus `Build` / `Parse` / `FindRecovery`) to build, parse, and recover RFC 2198 RTP RED payloads |
| `github.com/hujm2023/gopus/types` | Shared `Mode` / `Bandwidth` / `Signal` enums |

Multistream is reachable two ways: `gopus.NewMultistreamEncoder` /
`gopus.NewMultistreamDecoder` (and the `…Default` constructors for 1–8 channels
in Vorbis order) wrap the lower-level `multistream` package, which also carries
the projection / ambisonics constructors (mapping families 0/1/3/255).

Write an Ogg Opus file with the `container/ogg` writer:

```go
w, err := ogg.NewWriter(file, sampleRate, channels)
if err != nil {
	log.Fatal(err)
}
defer w.Close()

n, _ := enc.Encode(pcm, packet)
if err := w.WritePacket(packet[:n], frameSize); err != nil {
	log.Fatal(err)
}
```

## Optional features behind build tags

The default build is core encode/decode/multistream/Ogg/RED — matching a default
libopus `./configure`. Optional features are exposed exactly the way libopus
exposes them: behind a compile flag in libopus, behind the matching Go build tag
here. The default build links ZERO of their code (enforced by
`TestDefaultBuildIsZeroCostForGatedFeatures`).

| gopus build tag | libopus flag |
| --- | --- |
| `gopus_dred` | `--enable-dred` |
| `gopus_osce` | `--enable-osce` (+ `ENABLE_DEEP_PLC`) |
| `gopus_qext` | `--enable-qext` |
| `gopus_custom_modes` | `--enable-custom-modes` |
| `gopus_fixed_point` | `--enable-fixed-point` |

Under their tag these are parity-complete — none are experimental:

- **`gopus_dred`** — DRED (RDOVAE), control + standalone surfaces.
- **`gopus_osce`** — OSCE BWE / LACE / NoLACE plus the deep-PLC family
  (PitchDNN / FARGAN), exactly as `--enable-osce`.
- **`gopus_qext`** — QEXT framing and native 96 kHz (Opus HD): decode is
  sample-exact, and the public `Encode` at `Fs=96000` is byte-exact (TOC, padding,
  main CELT payload, reserved QEXT extension) vs libopus `--enable-qext`. 96 kHz is
  CELT-only fullband (mirroring libopus) and accepted only under this tag;
  default-build API rates stay 8/12/16/24/48 kHz.
- **`gopus_custom_modes`** — Opus Custom standard modes.
- **`gopus_fixed_point`** — integer CELT/SILK pipeline (libopus `FIXED_POINT`);
  public decode and encode are bit-exact vs the `--enable-fixed-point` oracle.

One more tag is orthogonal to the feature flags above and has no libopus
equivalent:

- **`purego`** — forces the scalar Go code path, disabling the architecture
  assembly kernels (arm64 NEON, amd64 AVX2). Output is identical to the default
  build except that this is the bit-exact tier on every architecture; use it when
  you want the reference numeric path or to build for a target without an asm
  kernel. The default build (no tag) already selects asm only where libopus does.

Default builds expose no optional extensions; `SetDNNBlob(...)` is a no-op
returning `ErrOptionalExtensionUnavailable`. This matches a default libopus build,
where the DNN / PitchDNN / FARGAN / RDOVAE neural code is empty and none of it is
compiled; gopus keeps those packages out of the default import graph. DNN blob
loading (USE_WEIGHTS_FILE model loading) requires `-tags gopus_dred` or
`-tags gopus_osce`; QEXT requires `-tags gopus_qext`; DRED
control/standalone surfaces require `-tags gopus_dred`; OSCE BWE/LACE/NoLACE
require `-tags gopus_osce`. Under their build tag these are
parity-complete and supported, exactly as libopus exposes them behind the
corresponding compile flag.

| Extension | Status | Probe |
| --- | --- | --- |
| DNN blob loading | Supported under `gopus_dred` / `gopus_osce` | `OptionalExtensionDNNBlob` |
| QEXT | Supported under `gopus_qext` | `OptionalExtensionQEXT` |
| DRED | Supported under `gopus_dred` (control + standalone) | `OptionalExtensionDRED` |
| OSCE BWE | Supported under `gopus_osce` | `OptionalExtensionOSCEBWE` |

The `gopus_osce` tag enables the OSCE and deep-PLC family exactly as
libopus's `--enable-osce` does. These features are supported under the tag and
link zero code into the default build.

```sh
go test -tags gopus_qext ./...
go test -tags gopus_dred ./...
go test -tags gopus_osce ./...
```

```sh
make test-dnn-blob-parity
make test-qext-parity
make test-dred-tag
make test-extra-controls-parity
make test-custom-parity
```

## Performance

gopus is built for real-time use, where steady allocation is the enemy:

- **Zero-allocation hot paths.** `Encode` / `Decode` (and their `int16` / `int24`
  variants) reuse caller-owned buffers; all scratch is pre-allocated at
  construction, so a steady-state encode or decode loop performs no heap
  allocations.
- **Allocation-free containers too.** `container/ogg` (`Reader.ReadPacketInto` /
  `Writer.WritePacket`) and `container/red` (`Decoder.Parse` / `Encoder.Encode`)
  own their buffers and the redundant-frame history, so steady-state demux/mux and
  RED packetization allocate nothing once warm — each locked by an
  `AllocsPerRun == 0` test.
- **SIMD where libopus has it.** On amd64 the float pitch cross-correlation uses
  an AVX2 kernel that mirrors libopus's `celt_pitch_xcorr_avx2`, computing several
  correlation lags per FMA instead of one scalar FMA per element — bit-identical
  output, materially faster stereo CELT and Hybrid encode. Encode-side SILK
  kernels also dispatch to the same architecture-specific lanes as libopus where
  the pinned reference provides them.

The required `perf-linux` CI lane publishes both steady-state Go benchmark
guardrails and libopus-relative ratios. This table comes from `perf-linux` run
`28648782556` on a GitHub linux/amd64 runner (`go1.25.0`, AMD EPYC 7763):

| Benchmark | CI result |
| --- | ---: |
| `BenchmarkEncoderEncode_CallerBuffer` | 94,064 ns/op, 0 B/op, 0 allocs/op |
| `BenchmarkEncoderEncodeInt16` | 94,773 ns/op, 0 B/op, 0 allocs/op |
| `BenchmarkDecoderDecode_CELT` | 20,832 ns/op, 0 B/op, 0 allocs/op |
| `BenchmarkDecoderDecodeInt16` | 22,504 ns/op, 0 B/op, 0 allocs/op |

The same lane compares gopus against libopus 1.6.1 on the same runner; lower
ratios are closer to libopus, and every gopus path below reports 0 allocs/op:

| Path | gopus/libopus |
| --- | ---: |
| Decode vectors, Float32 | 1.334x |
| Decode vectors, Int16 | 1.326x |
| Encode, all Float32 cases | 1.489x |
| Encode, CELT-FB-20ms-stereo-128k | 1.556x |
| Encode, CELT-FB-5ms-mono-64k | 1.471x |
| Encode, Hybrid-FB-20ms-mono-64k | 1.504x |
| Encode, Hybrid-FB-20ms-stereo-96k | 1.685x |
| Encode, SILK-WB-20ms-mono-32k | 1.395x |

Run the benchmarks for numbers on your machine:

```sh
go run ./examples/bench-encode
go run ./examples/bench-decode
```

`make bench-guard` runs the benchmark guardrails used in CI.

## Parity & testing

gopus is codec-complete against libopus 1.6.1: the full public API and CTL
surface, plus the optional surface mirrored tag-for-flag (above). The pinned
`tmp_check/opus-1.6.1/` is the reference — when behavior is uncertain, gopus
matches libopus unless fixture evidence says otherwise.

Parity is proven on two tiers, against a live libopus C oracle:

- **Bit-exact kernel oracles.** Isolated kernels (range coder, NLSF/LPC/gain,
  PVQ/bands, MDCT/KISS-FFT, resamplers, DNN matmuls) are compared bit-for-bit
  against C. Every public decode entry point is two-sided differential-fuzzed
  against the same oracle.
- **`opus_compare` quality on real audio.** End-to-end audio is judged by
  libopus's own `opus_compare` (RFC 8251's conformance metric), tier-matched so
  gopus tracks the reference at least as closely as libopus tracks itself across
  builds. SILK decode is bit-exact; CELT/Hybrid sit inside the near-exact
  envelope. The encoder precision guard runs on representative real recordings,
  where `opus_compare` Q is a genuine quality measure.

One residual is documented: a few CELT float kernels drift by ≤1 ULP on
darwin/arm64 (a per-arch float budget). amd64/CI is bit-exact; the default arm64
build is quality-gated for that tail, exactly as libopus's NEON path is relative
to its own scalar build.

Pre-v1: use this fork's release evidence in [Trust And Verification](#trust-and-verification).

## Verification

Run focused tests while iterating. Before merge-ready codec changes, run:

```sh
go test ./...
make test-doc-contract
make lint
make test-consumer-smoke
make test-examples-smoke
make verify-production
```

```sh
make verify-production-exhaustive
make release-evidence
```

Before a tag is published the tagged commit must be green on the required branch
checks (below), and `make release-evidence` must produce a PASS summary.

## Trust And Verification

Fork releases and their evidence: [hujm2023/gopus Releases](https://github.com/hujm2023/gopus/releases).
Verify the exact commit SHA against this fork's [CI runs](https://github.com/hujm2023/gopus/actions/workflows/ci.yml)
and the release's attached evidence. A downloadable Git tag, including `v0.2.0`,
is not proof of a verified GitHub Release or passing required checks.

The upstream [`v0.1.1` release](https://github.com/thesyncim/gopus/releases/tag/v0.1.1)
is evidence for upstream code; it does not certify this fork. Release status comes
from the fork's Releases page and per-commit checks, not an upstream version number.

Required branch checks:

<!-- required-checks:start -->
- `lint-static-analysis`
- `test-linux`
- `perf-linux`
- `test-macos`
- `test-windows`
<!-- required-checks:end -->

These aggregate gates make the libopus C-oracle parity suites mandatory across
Linux, macOS, and Windows; each lane builds the pinned libopus C reference first
under `GOWORK=off GOPUS_TEST_TIER=parity GOPUS_STRICT_LIBOPUS_REF=1` and compares
against committed arch-matched fixtures. They are the authoritative codec gate:
`release.yml` publishes a tag only after verifying these checks are green on the
tagged commit. `make release-evidence` then captures the supplementary safety and
performance gates that are not in the required CI set, plus build provenance; it
does not re-run the codec suites, since doing so against a live native libopus
reference compares gopus's single portable float order against another toolchain's
rounding rather than measuring a defect.

Workflow triggers, trust boundaries, and local validation: [CI and releases](.github/CI.md).

Security policy: [SECURITY.md](SECURITY.md). Consumer smoke test:
[examples/external-consumer-smoke/smoke_test.go](examples/external-consumer-smoke/smoke_test.go).

## Docs

- [CONTRIBUTING.md](CONTRIBUTING.md)
- [SECURITY.md](SECURITY.md)
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- [examples/README.md](examples/README.md)

## License

See [LICENSE](LICENSE).
