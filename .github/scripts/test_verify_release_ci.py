import copy
import unittest

from verify_release_ci import REQUIRED_CHECKS, verify


class ReleaseCITest(unittest.TestCase):
    def setUp(self):
        self.sha = "a" * 40
        self.repository = "hujm2023/gopus"
        self.run = {
            "id": 42, "head_sha": self.sha, "head_branch": "master",
            "path": ".github/workflows/ci.yml", "event": "push",
            "status": "completed", "conclusion": "success",
            "head_repository": {"full_name": self.repository},
        }
        jobs = [{"name": name, "run_id": 42, "head_sha": self.sha,
                 "status": "completed", "conclusion": "success"}
                for name in REQUIRED_CHECKS]
        self.pages = [{"jobs": jobs[:2]}, {"jobs": jobs[2:]}]

    def test_success_across_pages(self):
        verify(self.run, self.pages, self.sha, self.repository)

    def test_rejects_wrong_run_identity_or_result(self):
        for key, value in (
            ("head_sha", "b" * 40), ("head_branch", "feature"),
            ("path", ".github/workflows/other.yml"), ("event", "pull_request"),
            ("status", "in_progress"), ("conclusion", "failure"),
            ("head_repository", {"full_name": "another/fork"}),
        ):
            with self.subTest(key=key), self.assertRaises(ValueError):
                verify(dict(self.run, **{key: value}), self.pages, self.sha, self.repository)

    def test_rejects_missing_or_duplicate_jobs(self):
        for duplicate in (False, True):
            pages = copy.deepcopy(self.pages)
            if duplicate:
                pages[0]["jobs"].append(pages[0]["jobs"][0])
            else:
                pages[1]["jobs"].pop()
            with self.subTest(duplicate=duplicate), self.assertRaises(ValueError):
                verify(self.run, pages, self.sha, self.repository)

    def test_rejects_wrong_job_identity_or_result(self):
        for key, value in (("run_id", 43), ("head_sha", "b" * 40),
                           ("status", "in_progress"), ("conclusion", "failure"),
                           ("conclusion", "skipped")):
            pages = copy.deepcopy(self.pages)
            pages[0]["jobs"][0][key] = value
            with self.subTest(key=key, value=value), self.assertRaises(ValueError):
                verify(self.run, pages, self.sha, self.repository)


if __name__ == "__main__":
    unittest.main()
