# CI and releases

`CI` runs on every pull request and every push to `master`, including documentation
changes. The release checks retain these stable names:

- `lint-static-analysis`: documentation/type contracts, workflow validation, release-gate fixtures, PGO validation, and lint.
- `test-linux`: fast no-cgo tests on Go 1.25 and stable; reachable-vulnerability scanning; tag lint; RFC conformance; native ARM64 fixtures; existing codec, optional-feature, consumer and fuzz gates. Pushes also require quality, race, provenance, and repeated parity checks.
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
Checkout credentials are not persisted except in the PGO workflow that pushes
its refresh branch. Pull request jobs have read-only repository permissions.

Repository settings are separate from these files. Configure the five checks
above as required checks on `master`, require code-owner review, and restrict
release-tag updates to maintainers. Workflow changes do not configure those
settings or demonstrate that they are enabled.

Local workflow checks:

```sh
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s .github/scripts -p 'test_*.py'
bash -n tools/gen_release_evidence.sh
go test . -run '^(TestTrustDocsContract|TestTrustSensitiveFilesHaveCodeOwners|TestReleaseNotesSourceIsReadme|TestCIWorkflowContract)$' -count=1
```
