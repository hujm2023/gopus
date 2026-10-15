# CI and releases

`CI` runs on every pull request and every push to `master`, including documentation
changes. The release checks retain these stable names:

- `lint-static-analysis`: documentation/type contracts, workflow validation, release-gate fixtures, PGO validation, and lint.
- `test-linux`: fast no-cgo tests on Go 1.27 and stable; reachable-vulnerability scanning; tag lint; RFC conformance; native ARM64 fixtures; existing codec, optional-feature, consumer and fuzz gates. Pushes also require quality, race, provenance, and repeated parity checks.
- `perf-linux`, `test-macos`, and `test-windows`: performance and native-platform gates.

The scalar parity lane uses `purego` with `GOPUS_LIBOPUS_REF_SCALAR=1`; native
lanes use their native libopus build. `CGO_ENABLED=0` proves the Go library does
not require cgo; separate C executables remain the codec test oracle.

`Verify Production Exhaustive` runs the existing exhaustive production target
before capturing supplementary release evidence. `Verify Safety` retains its
scheduled Go/platform checks. `PGO refresh` is a manual, reviewable pull request.

`Release` accepts a `v*` tag push or an existing version tag through manual
dispatch. A read-only job resolves the commit and checks the latest CI push run
on `master` for that exact SHA, repository, workflow, and run attempt. Every
required job must succeed. Its evidence bundle records CI run/job metadata and
checksums. A separate job receives `contents: write`, downloads that bundle,
checks the remote tag still resolves to the verified SHA, and publishes the
release without checking out or executing repository code.

Actions use full commit SHAs; Dependabot checks their version updates weekly.
The CI container uses Go 1.27.2 with a pinned multi-architecture image digest;
Dependabot also checks Docker image updates.
Checkout credentials are not persisted except in the PGO workflow that pushes
its refresh branch. Pull request jobs have read-only repository contents
permissions; CodeQL jobs receive permission to upload security results.

`vulnerability-scan` scans each module separately: the root library, both WebRTC
examples, and the external consumer smoke module. The WebRTC examples require
Go 1.26 or newer; the root library requires Go 1.25 or newer. The desktop
example is scanned with its GUI entry point and native build dependencies.
Each WebRTC module's `osv-scanner.toml` excludes only `GO-2026-5932`: Pion DTLS
uses `x/crypto/cryptobyte` and ChaCha20-Poly1305, while `openpgp` and its
subpackages are absent from the import graph. `govulncheck` confirms no affected
imported package or reachable symbol. The CI dependency guard includes tests
and optional codec builds and fails if `golang.org/x/crypto/openpgp` or any
subpackage enters the graph. Other vulnerabilities remain subject to the
existing vulnerability gate.
`CodeQL` runs extended Go security queries on the library and both WebRTC
examples for pull requests, pushes to `master`, and the weekly schedule.

Repository settings are separate from these files. Import
[`rulesets/master.json`](rulesets/master.json) under Settings → Rules → Rulesets
after the CI and CodeQL checks succeed. The rule requires pull requests, the
five release checks above, and all three CodeQL jobs; prohibits force pushes and
branch deletion; and has no bypass actors. It requires zero approval votes for
the single-maintainer workflow, because authors cannot approve their own pull
requests. This does not satisfy Scorecard's independent human-review check.
Restrict release-tag updates to maintainers. Workflow files and the importable
rule do not demonstrate that repository settings are enabled.

Local workflow checks:

```sh
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s .github/scripts -p 'test_*.py'
bash -n tools/gen_release_evidence.sh
go test . -run '^(TestTrustDocsContract|TestTrustSensitiveFilesHaveCodeOwners|TestReleaseNotesSourceIsReadme|TestCIWorkflowContract)$' -count=1
```
