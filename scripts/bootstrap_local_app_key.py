#!/usr/bin/env python3
"""Discover exactly one local GitHub App PEM and hand it to the existing secure setup.

This script never prints key contents, never sends a key to chat, and does not
retrieve a previously generated key from GitHub. Preview is the default.
"""
import argparse
import os
from pathlib import Path
import stat
import subprocess
import sys
import tempfile

HERE = Path(__file__).resolve().parent
SETUP = HERE / "setup_federation_evidence.py"


def candidates(directory):
    if not directory.is_dir():
        raise ValueError("Downloads directory is unavailable; pass --downloads-dir")
    matches = []
    for path in directory.iterdir():
        if path.suffix.lower() != ".pem" or path.is_symlink():
            continue
        try:
            with path.open("rb") as handle:
                head = handle.read(32)
            if head.startswith((b"-----BEGIN RSA PRIVATE KEY-----", b"-----BEGIN PRIVATE KEY-----")):
                matches.append(path)
        except (OSError, PermissionError):
            continue
    return matches


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--downloads-dir", type=Path, required=True,
                        help="Explicit local directory containing the downloaded App PEM")
    parser.add_argument("--client-id", required=True, help="Public GitHub App Client ID")
    parser.add_argument("--apply", action="store_true",
                        help="Explicitly authorize writing the key as a GitHub Actions secret")
    args = parser.parse_args()
    try:
        found = candidates(args.downloads_dir)
        if len(found) != 1:
            raise ValueError("Expected exactly one recognizable PEM in the selected directory; "
                             "move other PEM files elsewhere and retry")
        # The shared Android Downloads filesystem may not support chmod 600.
        # Copy only to an owner-private temporary file within Ubuntu.
        if not args.apply:
            print("Found exactly one candidate PEM. No key contents displayed or copied.")
            return subprocess.call([sys.executable, str(SETUP)])
        with tempfile.TemporaryDirectory(prefix="aift-app-key-") as private_dir:
            os.chmod(private_dir, 0o700)
            private_path = Path(private_dir) / "app.pem"
            flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL
            fd = os.open(private_path, flags, 0o600)
            try:
                with found[0].open("rb") as source, os.fdopen(fd, "wb") as target:
                    data = source.read(65537)
                    if len(data) > 65536 or not data.startswith(
                        (b"-----BEGIN RSA PRIVATE KEY-----", b"-----BEGIN PRIVATE KEY-----")
                    ):
                        raise ValueError("Invalid PEM size or header")
                    target.write(data)
            except BaseException:
                private_path.unlink(missing_ok=True)
                raise
            # Existing setup verifies admin access and passes the key through gh stdin.
            return subprocess.call([sys.executable, str(SETUP), "--apply",
                                    "--client-id", args.client_id,
                                    "--private-key", str(private_path)])
    except (ValueError, OSError) as exc:
        print(f"Setup not performed: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
