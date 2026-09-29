package gopus_test

import (
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestTrustDocsContract(t *testing.T) {
	readme := mustReadDocForTest(t, "README.md")
	for _, needle := range []string{
		"## Trust And Verification",
		"## Fork origin and maintenance",
		"independently maintained fork of [`thesyncim/gopus`]",
		"bd8db897c2681f6b5e3e5b9d16cf7f9e50f32781",
		"Fork releases and their evidence:",
		"https://github.com/hujm2023/gopus/releases",
		"https://github.com/hujm2023/gopus/actions/workflows/ci.yml",
		"it does not certify this fork",
		"Required branch checks:",
		"[SECURITY.md](SECURITY.md)",
		"[examples/external-consumer-smoke/smoke_test.go](examples/external-consumer-smoke/smoke_test.go)",
	} {
		if !strings.Contains(readme, needle) {
			t.Fatalf("README.md trust section missing %q", needle)
		}
	}

	for _, command := range []string{
		"go test ./...",
		"make test-doc-contract",
		"make lint",
		"make test-consumer-smoke",
		"make test-examples-smoke",
		"make verify-production",
		"make verify-production-exhaustive",
		"make release-evidence",
	} {
		if !strings.Contains(readme, command) {
			t.Fatalf("README.md release checklist missing required command %s", command)
		}
	}

	security := mustReadDocForTest(t, "SECURITY.md")
	for _, needle := range []string{
		"Do not open a public issue",
		"Prefer GitHub private vulnerability reporting",
		"email `thesyncim@gmail.com`",
	} {
		if !strings.Contains(security, needle) {
			t.Fatalf("SECURITY.md missing %q", needle)
		}
	}

	requiredChecks := extractRequiredChecks(t, readme)
	wantChecks := []string{"lint-static-analysis", "test-linux", "perf-linux", "test-macos", "test-windows"}
	if !reflect.DeepEqual(requiredChecks, wantChecks) {
		t.Fatalf("required checks = %v, want %v", requiredChecks, wantChecks)
	}
	ciJobs := workflowJobNames(t, ".github/workflows/ci.yml")
	for _, check := range requiredChecks {
		if !ciJobs[check] {
			t.Fatalf("README.md lists stale required check %q; actual CI job names are %v", check, sortedKeys(ciJobs))
		}
	}
}

func TestTrustSensitiveFilesHaveCodeOwners(t *testing.T) {
	codeowners := mustReadDocForTest(t, ".github/CODEOWNERS")
	for _, pattern := range []string{
		".github/workflows/*",
		".github/CODEOWNERS",
		".github/scripts/*",
		"tools/gen_release_evidence.sh",
		"SECURITY.md",
		"README.md",
		"tools/ensure_libopus.sh",
		"Makefile",
	} {
		if !strings.Contains(codeowners, pattern+" @hujm2023") {
			t.Fatalf(".github/CODEOWNERS missing owner for %s", pattern)
		}
	}
}

func TestReleaseNotesSourceIsReadme(t *testing.T) {
	readme := mustReadDocForTest(t, "README.md")
	for _, needle := range []string{
		"Fork releases and their evidence:",
		"Verify the exact commit SHA",
		"is not proof of a verified GitHub Release or passing required checks",
		"make release-evidence",
	} {
		if !strings.Contains(readme, needle) {
			t.Fatalf("README.md release notes source missing %q", needle)
		}
	}
}

func extractRequiredChecks(t *testing.T, doc string) []string {
	t.Helper()
	const start = "<!-- required-checks:start -->"
	const end = "<!-- required-checks:end -->"
	startAt := strings.Index(doc, start)
	endAt := strings.Index(doc, end)
	if startAt < 0 || endAt < 0 || endAt <= startAt {
		t.Fatalf("README.md missing required-checks markers")
	}

	block := doc[startAt+len(start) : endAt]
	var checks []string
	for line := range strings.SplitSeq(block, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- `") && strings.HasSuffix(line, "`") {
			checks = append(checks, strings.TrimSuffix(strings.TrimPrefix(line, "- `"), "`"))
		}
	}
	return checks
}

func workflowJobNames(t *testing.T, path string) map[string]bool {
	t.Helper()
	data := mustReadDocForTest(t, path)
	names := workflowJobNamesFromText(t, path, data)
	crlfData := strings.ReplaceAll(data, "\n", "\r\n")
	crlfNames := workflowJobNamesFromText(t, path+" with CRLF", crlfData)
	if !reflect.DeepEqual(names, crlfNames) {
		t.Fatalf("workflow job names differ under CRLF line endings: lf=%v crlf=%v", sortedKeys(names), sortedKeys(crlfNames))
	}
	return names
}

func workflowJobNamesFromText(t *testing.T, path, data string) map[string]bool {
	t.Helper()
	names := make(map[string]bool)
	inJobs := false
	inJob := false

	for line := range strings.SplitSeq(data, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "jobs:" {
			inJobs = true
			continue
		}
		if !inJobs {
			continue
		}
		if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") && strings.HasSuffix(strings.TrimSpace(line), ":") {
			inJob = true
			continue
		}
		if inJob && strings.HasPrefix(line, "    name:") {
			name := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "name:"))
			name = strings.Trim(name, `"'`)
			names[name] = true
		}
	}
	if len(names) == 0 {
		t.Fatalf("no job names parsed from %s", path)
	}
	return names
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func TestCIWorkflowContract(t *testing.T) {
	ci := mustReadDocForTest(t, ".github/workflows/ci.yml")
	_, aggregate, ok := strings.Cut(ci, "\n  test-linux:\n")
	if !ok {
		t.Fatal("missing test-linux aggregate")
	}
	aggregate, _, _ = strings.Cut(aggregate, "\n  perf-linux:")
	for _, job := range []string{"lint-tag-matrix", "test-linux-conformance", "test-linux-arm64-fixtures", "vulnerability-scan"} {
		if !strings.Contains(aggregate, "      - "+job+"\n") || !strings.Contains(aggregate, "needs."+job+".result") {
			t.Fatalf("test-linux must require and check %s", job)
		}
	}
	paths, err := filepath.Glob(".github/workflows/*.yml")
	if err != nil || len(paths) == 0 {
		t.Fatalf("workflow files: %v", err)
	}
	pinned := regexp.MustCompile(`@[0-9a-f]{40}(\s|$)`)
	for _, path := range paths {
		for line := range strings.SplitSeq(mustReadDocForTest(t, path), "\n") {
			if strings.Contains(line, "uses:") && !pinned.MatchString(line) {
				t.Errorf("%s: action must use a commit SHA: %s", path, line)
			}
		}
	}
	release := mustReadDocForTest(t, ".github/workflows/release.yml")
	_, publish, ok := strings.Cut(release, "\n  publish-release:\n")
	if !ok || !strings.Contains(publish, "needs: verify-release") || strings.Contains(publish, "actions/checkout@") {
		t.Fatal("publish must depend on verification and must not check out repository code")
	}
	if strings.Count(release, "contents: write") != 1 || !strings.Contains(publish, "contents: write") {
		t.Fatal("only publication may receive contents: write")
	}
	if strings.Contains(release, `tag="${{`) || strings.Contains(release, `TAG="${{`) {
		t.Fatal("release tag inputs must pass through environment variables")
	}
	exhaustive := mustReadDocForTest(t, ".github/workflows/verify-production-exhaustive.yml")
	if !strings.Contains(exhaustive, "run: make verify-production-exhaustive") {
		t.Fatal("exhaustive workflow must execute the exhaustive gate")
	}
}
