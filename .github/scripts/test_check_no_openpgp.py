import os
from pathlib import Path
import subprocess
import tempfile
import unittest


SCRIPT = Path(__file__).resolve().parents[2] / "tools" / "check_no_openpgp.sh"


class OpenPGPDependencyCheckTest(unittest.TestCase):
    def run_check(self, dependencies, exit_code=0, flags=()):
        with tempfile.TemporaryDirectory() as directory:
            go = Path(directory) / "go"
            go.write_text(
                "#!/usr/bin/env python3\n"
                "import os, sys\n"
                "assert sys.argv[1:5] == ['list', '-deps', '-test', '-f']\n"
                "assert sys.argv[6:] == ['-tags', 'purego', './...'] if "
                "os.environ.get('TEST_TAGS') else sys.argv[6:] == ['./...']\n"
                "print(os.environ['TEST_DEPENDENCIES'])\n"
                "sys.exit(int(os.environ['TEST_EXIT_CODE']))\n",
                encoding="utf-8",
            )
            go.chmod(0o700)
            env = dict(os.environ, GO=str(go), TEST_DEPENDENCIES=dependencies,
                       TEST_EXIT_CODE=str(exit_code), TEST_TAGS="1" if flags else "")
            return subprocess.run(["bash", str(SCRIPT), *flags], cwd=directory,
                                  env=env, capture_output=True, text=True)

    def test_allows_crypto_packages_and_similar_names(self):
        result = self.run_check("\n".join((
            "golang.org/x/crypto/cryptobyte", "golang.org/x/crypto/chacha20poly1305",
            "golang.org/x/crypto/openpgphelper", "example.org/openpgp",
        )), flags=("-tags", "purego"))
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_rejects_openpgp_and_each_subpackage(self):
        for package in ("openpgp", "openpgp/packet", "openpgp/armor",
                        "openpgp/clearsign", "openpgp/errors", "openpgp/elgamal",
                        "openpgp/s2k"):
            with self.subTest(package=package):
                dependency = "golang.org/x/crypto/" + package
                result = self.run_check("fmt\n" + dependency + "\nio")
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(dependency, result.stderr)

    def test_rejects_incomplete_dependency_graph(self):
        result = self.run_check("fmt\nio", exit_code=3)
        self.assertEqual(result.returncode, 3)
        self.assertNotIn("No golang.org/x/crypto/openpgp", result.stdout)


if __name__ == "__main__":
    unittest.main()
