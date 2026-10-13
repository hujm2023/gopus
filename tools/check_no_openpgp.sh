#!/usr/bin/env bash
# Run from a module directory. Additional go list flags select build tags.
set -euo pipefail

go_bin="${GO:-go}"
dependencies="$(mktemp)"
trap 'rm -f "$dependencies"' EXIT

# Include transitive and test dependencies; an incomplete graph fails the check.
GOWORK=off "$go_bin" list -deps -test -f '{{.ImportPath}}' "$@" ./... > "$dependencies"
matches="$(awk '$0 == "golang.org/x/crypto/openpgp" || index($0, "golang.org/x/crypto/openpgp/") == 1' "$dependencies")"
if [[ -n "$matches" ]]; then
  printf 'Unsupported OpenPGP dependencies invalidate the GO-2026-5932 exception:\n%s\n' "$matches" >&2
  exit 1
fi

echo "No golang.org/x/crypto/openpgp dependencies (including tests)."
