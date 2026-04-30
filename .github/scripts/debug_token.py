#!/usr/bin/env python3
"""
Reproduce token verification against the real decode_token() pipeline.

Usage:
  python debug_token.py <url_with_token_param>
  python debug_token.py <raw_token_string>

Set DC_APPLY_STATE_HMAC_KEY in the environment to test signature verification.
"""

import sys
import traceback
from urllib.parse import urlparse, parse_qs

# Import the production function (and its private helpers for diagnostics)
from check_pr_description import (
    decode_token,
    _ForgedTokenError,
    _hmac_signature,
    _canonical_json,
)
import json
import gzip
import hmac as _hmac
from base64 import b64decode
from urllib.parse import unquote


def extract_token(raw: str) -> str:
    parsed = urlparse(raw)
    if parsed.scheme in ("http", "https"):
        params = parse_qs(parsed.query)
        if "token" not in params:
            print("Error: URL provided but no 'token' query parameter found", file=sys.stderr)
            sys.exit(1)
        return params["token"][0]
    return raw.strip()


def diagnose(token: str) -> None:
    """Decode the token without verifying the signature and print diagnostics."""
    try:
        compressed = b64decode(unquote(token))
        raw = gzip.decompress(compressed)
        payload = json.loads(raw.decode("utf-8"))
    except Exception as e:
        print(f"  [diag] Failed to decode token bytes: {e}", file=sys.stderr)
        return

    sig = payload.get("_signature")
    print(f"  [diag] Signature in token : {sig}", file=sys.stderr)

    payload_copy = dict(payload)
    payload_copy.pop("_signature", None)
    expected = _hmac_signature(_canonical_json(payload_copy))
    if expected is None:
        print("  [diag] DC_APPLY_STATE_HMAC_KEY not set – cannot compute expected signature",
              file=sys.stderr)
    else:
        print(f"  [diag] Expected signature  : {expected}", file=sys.stderr)
        match = _hmac.compare_digest(sig or "", expected)
        print(f"  [diag] Signatures match    : {match}", file=sys.stderr)

    print("\n  [diag] Payload (without _signature):", file=sys.stderr)
    print(json.dumps(payload_copy, indent=2), file=sys.stderr)


def main() -> int:
    if len(sys.argv) < 2:
        print(__doc__)
        return 1

    raw_input = sys.argv[1]
    token = extract_token(raw_input)

    print("--- Calling decode_token() ---")
    try:
        result = decode_token(token)
        print("SUCCESS – decoded payload:")
        print(json.dumps(result, indent=2, ensure_ascii=False))
        return 0
    except _ForgedTokenError as e:
        print(f"FAIL – ForgedTokenError: {e}", file=sys.stderr)
    except Exception as e:
        print(f"FAIL – {type(e).__name__}: {e}", file=sys.stderr)
        traceback.print_exc()

    print("\n--- Diagnostics ---", file=sys.stderr)
    diagnose(token)
    return 1


if __name__ == "__main__":
    sys.exit(main())
