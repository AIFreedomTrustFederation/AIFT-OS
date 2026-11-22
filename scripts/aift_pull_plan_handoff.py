#!/usr/bin/env python3
"""Decode one operator-reviewed v2 plan on a GitHub-hosted publisher, no token."""
from __future__ import annotations

import argparse
import base64
import binascii
import hashlib
import hmac
import os
from pathlib import Path
import re
import sys


class InvalidHandoff(ValueError):
    pass


def decode_reviewed_plan(encoded: str, expected_digest: str, destination: Path) -> str:
    """Fail closed on bounds/digest; never log or execute untrusted plan text."""
    if not isinstance(encoded, str) or not 0 < len(encoded) <= 11000:
        raise InvalidHandoff("Encoded plan missing or oversized")
    if not isinstance(expected_digest, str) or re.fullmatch(r"[0-9a-f]{64}", expected_digest) is None:
        raise InvalidHandoff("Reviewed plan digest missing or malformed")
    try:
        data = base64.b64decode(encoded, validate=True)
    except (binascii.Error, ValueError) as error:
        raise InvalidHandoff("Invalid encoded plan") from error
    if not 0 < len(data) <= 8192:
        raise InvalidHandoff("Decoded plan missing or oversized")
    digest = hashlib.sha256(data).hexdigest()
    if not hmac.compare_digest(digest, expected_digest):
        raise InvalidHandoff("Reviewed plan digest mismatch")
    # Use O_EXCL to avoid following or overwriting an existing path. The Actions
    # temp directory is fresh; a preexisting plan is suspicious and blocks.
    fd = os.open(destination, os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_NOFOLLOW, 0o600)
    try:
        with os.fdopen(fd, "wb") as file:
            file.write(data)
            file.flush()
            os.fsync(file.fileno())
    except Exception:
        destination.unlink(missing_ok=True)
        raise
    return digest


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    try:
        digest = decode_reviewed_plan(
            os.environ.get("PLAN_B64", ""),
            os.environ.get("PLAN_SHA256", ""),
            args.output,
        )
    except (InvalidHandoff, OSError) as error:
        print(f"BLOCKED: {error}", file=sys.stderr)
        return 1
    print(f"Reviewed plan digest verified: {digest}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
