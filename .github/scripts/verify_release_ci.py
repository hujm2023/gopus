"""Require the tagged commit's successful CI push run and its release gates."""

import json
import sys

REQUIRED_CHECKS = (
    "lint-static-analysis", "test-linux", "perf-linux", "test-macos", "test-windows",
)


def verify(run, pages, sha, repository):
    expected = {
        "head_sha": sha,
        "head_branch": "master",
        "path": ".github/workflows/ci.yml",
        "event": "push",
        "status": "completed",
        "conclusion": "success",
    }
    for key, value in expected.items():
        if run.get(key) != value:
            raise ValueError(f"CI run {key}={run.get(key)!r}, expected {value!r}")
    if run.get("head_repository", {}).get("full_name") != repository:
        raise ValueError("CI run does not belong to the release repository")
    jobs = [job for page in pages for job in page["jobs"]]
    for name in REQUIRED_CHECKS:
        matches = [job for job in jobs if job["name"] == name]
        if len(matches) != 1:
            raise ValueError(f"required CI job {name!r} is missing or duplicated")
        job = matches[0]
        if (job.get("run_id") != run["id"] or job.get("head_sha") != sha
                or job.get("status") != "completed" or job.get("conclusion") != "success"):
            raise ValueError(f"required CI job {name!r} is not successful on this run and SHA")


if __name__ == "__main__":
    with open(sys.argv[1], encoding="utf-8") as source:
        run = json.load(source)
    with open(sys.argv[2], encoding="utf-8") as source:
        pages = json.load(source)
    verify(run, pages, sys.argv[3], sys.argv[4])
    print(f"Verified CI run {run['id']}: {run['html_url']}")
