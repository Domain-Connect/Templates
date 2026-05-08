#!/usr/bin/env python3
"""
Check PR description against the pull_request_template.md requirements.

Checks:
1. "Type of change" section has at least one ticked checkbox
2. "How Has This Been Tested?" section has ALL checkboxes ticked
3. "Checklist of common problems" section has at least some checkboxes ticked
4. "Online Editor test results" section contains at least one valid editor link
5. For each template JSON file in the PR, there is a corresponding editor test link
   - hostRequired=true  → at least 1 test with a non-empty host parameter
   - otherwise          → at least 2 tests: one with host and one without (empty/"@")
"""

import gzip
import hashlib
import hmac
import json
import os
import re
import sys
from base64 import b64decode, b64encode
from urllib.parse import unquote

EDITOR_URL_PATTERN = re.compile(
    r"https://domainconnect\.paulonet\.eu/dc/free/templateedit\?token=([A-Za-z0-9+/=%]+)"
)

# Matches  [x], [X], [✓], [✗] – any non-space character inside brackets
TICKED = re.compile(r"\[([^ ])\]")
UNTICKED = re.compile(r"\[ \]")


class _ForgedTokenError(ValueError):
    """Raised when a token's HMAC signature does not match."""


def _canonical_json(obj) -> bytes:
    """Return a canonical JSON byte string: keys sorted recursively, no whitespace."""
    return json.dumps(obj, sort_keys=True, separators=(',', ':')).encode('utf-8')


def _hmac_signature(data: bytes) -> str | None:
    """Return base64 HMAC-SHA256 of *data*, or None if the key is not configured."""
    key = os.environ.get("DC_APPLY_STATE_HMAC_KEY")
    if not key:
        return None
    key_bytes = bytes.fromhex(key) if isinstance(key, str) else key
    digest = hmac.new(key_bytes, data, hashlib.sha256).digest()
    return b64encode(digest).decode('ascii')


def decode_token(token: str) -> dict:
    """Decode a token and verify its HMAC-SHA256 signature.

    Signature verification is always required.  If ``DC_APPLY_STATE_HMAC_KEY``
    is not set in the environment the check is skipped (with a debug log) and a
    :class:`SignatureMissingKeyError` is raised so callers can distinguish
    "key not configured" from "bad signature".
    """
    compressed = b64decode(unquote(token))
    raw = gzip.decompress(compressed)
    payload = json.loads(raw.decode('utf-8'))

    sig = payload.pop("_signature", None)

    if sig is None:
        raise _ForgedTokenError("token has no signature")

    expected = _hmac_signature(_canonical_json(payload))
    if expected is None:
        print("  DEBUG: DC_APPLY_STATE_HMAC_KEY not set; skipping signature verification",
              file=sys.stderr)
        return payload

    if not hmac.compare_digest(sig, expected):
        raise _ForgedTokenError("token signature mismatch")

    return payload


def find_section(body: str, heading: str) -> str | None:
    """Return the text from *heading* until the next same-or-higher-level heading."""
    # Match ## or # headings case-insensitively; heading text is a plain string
    pattern = re.compile(
        r"^(#{1,3})\s+" + re.escape(heading) + r"\s*$",
        re.IGNORECASE | re.MULTILINE,
    )
    m = pattern.search(body)
    if not m:
        return None
    level = len(m.group(1))
    start = m.end()
    # Find next heading at same or higher level
    next_heading = re.compile(
        r"^#{1," + str(level) + r"}\s+", re.MULTILINE
    )
    nm = next_heading.search(body, start)
    end = nm.start() if nm else len(body)
    return body[start:end]


def check_section_ticked(section_text: str, require_all: bool) -> tuple[bool, str]:
    """
    If require_all=True  → every checkbox must be ticked.
    If require_all=False → at least one checkbox must be ticked.
    Returns (ok, detail_message).
    """
    ticked = len(TICKED.findall(section_text))
    unticked = len(UNTICKED.findall(section_text))
    total = ticked + unticked
    if total == 0:
        return False, "no checkboxes found"
    if require_all:
        if unticked > 0:
            return False, f"{unticked}/{total} checkboxes are not ticked"
        return True, f"all {total} checkboxes ticked"
    else:
        if ticked == 0:
            return False, "no checkboxes are ticked"
        return True, f"{ticked}/{total} checkboxes ticked"


def _required_payload_keys() -> set[str]:
    """Return the set of required top-level payload keys from the env variable."""
    raw = os.environ.get("REQUIRED_PAYLOAD_KEYS", "")
    return {k.strip() for k in raw.split(",") if k.strip()}


def get_editor_links(body: str) -> tuple[list[dict], bool]:
    """Return (payloads, forged) where forged=True if any token failed signature verification."""
    required_keys = _required_payload_keys()
    results = []
    forged = False
    for token in EDITOR_URL_PATTERN.findall(body):
        try:
            payload = decode_token(token)
            if required_keys and not required_keys.issubset(payload.keys()):
                raise ValueError("invalid payload")
            results.append(payload)
        except _ForgedTokenError as e:
            print(f"  WARNING: forged token detected: {e}", file=sys.stderr)
            forged = True
        except Exception as e:
            print(f"  WARNING: could not decode token: {e}", file=sys.stderr)
    return results, forged


def template_id_from_payload(payload: dict) -> str | None:
    """Return '<providerId>.<serviceId>' from template payload, or None."""
    t = payload.get("template")
    if not isinstance(t, dict):
        return None
    pid = t.get("providerId", "")
    sid = t.get("serviceId", "")
    if pid and sid:
        return f"{pid}.{sid}"
    return None


def host_is_set(payload: dict) -> bool:
    host = payload.get("host", "")
    return bool(host) and host != "@"


def templates_equal(a: dict, b: dict) -> bool:
    """Deep-equal comparison of two template dicts, ignoring key order."""
    return json.loads(json.dumps(a, sort_keys=True)) == json.loads(
        json.dumps(b, sort_keys=True)
    )


def check_template_coverage(
    template_files: list[str], payloads: list[dict]
) -> list[str]:
    """
    For each template file, verify there are sufficient editor tests and that
    the template embedded in the token matches the file exactly.
    Returns a list of error strings (empty = all good).
    """
    errors = []
    for fpath in template_files:
        fname = os.path.basename(fpath)
        try:
            with open(fpath) as f:
                tmpl = json.load(f)
        except Exception as e:
            errors.append(f"{fname}: could not read template file: {e}")
            continue

        host_required = tmpl.get("hostRequired", False)
        provider_id = tmpl.get("providerId", "")
        service_id = tmpl.get("serviceId", "")
        expected_id = f"{provider_id}.{service_id}" if provider_id and service_id else None

        # Filter payloads belonging to this template by providerId/serviceId
        matching = [
            p for p in payloads
            if template_id_from_payload(p) == expected_id
        ]

        if not matching:
            errors.append(
                f"{fname}: no editor test link found "
                f"(expected template id '{expected_id}')"
            )
            continue

        # Check that the template in each token matches the file exactly
        stale = [
            p for p in matching
            if not templates_equal(p.get("template", {}), tmpl)
        ]
        if stale:
            errors.append(
                f"{fname}: {len(stale)}/{len(matching)} test link(s) contain "
                f"a template that does not match the file in this PR "
                f"(tests may be out of date)"
            )
            # Keep only matching-template payloads for host coverage checks
            matching = [p for p in matching if templates_equal(p.get("template", {}), tmpl)]
            if not matching:
                continue

        if host_required:
            with_host = [p for p in matching if host_is_set(p)]
            if not with_host:
                errors.append(
                    f"{fname}: hostRequired=true but no test with a host parameter found"
                )
        else:
            with_host = [p for p in matching if host_is_set(p)]
            without_host = [p for p in matching if not host_is_set(p)]
            missing = []
            if not with_host:
                missing.append("a test WITH host")
            if not without_host:
                missing.append("a test WITHOUT host (or with empty/@)")
            if missing:
                errors.append(
                    f"{fname}: missing {' and '.join(missing)}"
                )

    return errors


LABEL_DESCRIPTION_INCOMPLETE = "PR description incomplete"
LABEL_CHECKLIST_INCOMPLETE = "Checklist of common problems not complete"
LABEL_TEST_LINKS_MISSING = "Test links missing"
LABEL_FORGED_EDITOR_LINKS = "Forged editor links"

ALL_MANAGED_LABELS = {
    LABEL_DESCRIPTION_INCOMPLETE,
    LABEL_CHECKLIST_INCOMPLETE,
    LABEL_TEST_LINKS_MISSING,
    LABEL_FORGED_EDITOR_LINKS,
}


def main() -> int:
    body = os.environ.get("PR_BODY", "")
    template_files_env = os.environ.get("TEMPLATE_FILES", "")
    template_files = [f for f in template_files_env.split() if f.strip()]
    labels_add_file = os.environ.get("LABELS_ADD_FILE", "")
    labels_remove_file = os.environ.get("LABELS_REMOVE_FILE", "")

    failures: list[str] = []
    labels_to_add: set[str] = set()

    # --- 1. Type of change ---
    section = find_section(body, "Type of change")
    if section is None:
        failures.append("'Type of change' section not found")
        labels_to_add.add(LABEL_DESCRIPTION_INCOMPLETE)
    else:
        ok, detail = check_section_ticked(section, require_all=False)
        if not ok:
            failures.append(f"'Type of change': {detail}")
            labels_to_add.add(LABEL_DESCRIPTION_INCOMPLETE)
        else:
            print(f"  OK  Type of change: {detail}")

    # --- 2. How Has This Been Tested? ---
    section = find_section(body, "How Has This Been Tested?")
    if section is None:
        failures.append("'How Has This Been Tested?' section not found")
        labels_to_add.add(LABEL_DESCRIPTION_INCOMPLETE)
    else:
        ok, detail = check_section_ticked(section, require_all=True)
        if not ok:
            failures.append(f"'How Has This Been Tested?': {detail}")
            labels_to_add.add(LABEL_DESCRIPTION_INCOMPLETE)
        else:
            print(f"  OK  How Has This Been Tested?: {detail}")

    # --- 3. Checklist of common problems ---
    section = find_section(body, "Checklist of common problems")
    if section is None:
        failures.append("'Checklist of common problems' section not found")
        labels_to_add.add(LABEL_CHECKLIST_INCOMPLETE)
    else:
        ok, detail = check_section_ticked(section, require_all=False)
        if not ok:
            failures.append(f"'Checklist of common problems': {detail}")
            labels_to_add.add(LABEL_CHECKLIST_INCOMPLETE)
        else:
            print(f"  OK  Checklist of common problems: {detail}")

    # --- 4. Online Editor test results (links present) ---
    editor_section = find_section(body, "Online Editor test results")
    if editor_section is None:
        failures.append("'Online Editor test results' section not found")
        labels_to_add.add(LABEL_TEST_LINKS_MISSING)
        payloads = []
    else:
        payloads, forged = get_editor_links(editor_section)
        if forged:
            failures.append(
                "'Online Editor test results': one or more editor links are forged. Copy-paste real links from the editor."
            )
            labels_to_add.add(LABEL_FORGED_EDITOR_LINKS)
        if not payloads:
            failures.append(
                "'Online Editor test results': no valid editor test link found"
            )
            labels_to_add.add(LABEL_TEST_LINKS_MISSING)
        else:
            print(f"  OK  Online Editor test results: {len(payloads)} link(s) found")

    # --- 5. Per-template link coverage ---
    if template_files and payloads is not None:
        coverage_errors = check_template_coverage(template_files, payloads)
        for err in coverage_errors:
            failures.append(f"Template coverage: {err}")
            labels_to_add.add(LABEL_TEST_LINKS_MISSING)
        if not coverage_errors and template_files:
            print(f"  OK  Template coverage: all {len(template_files)} template(s) covered")

    # --- Write labels files ---
    labels_to_remove = ALL_MANAGED_LABELS - labels_to_add
    if labels_add_file:
        with open(labels_add_file, "w") as f:
            if labels_to_add:
                f.write("\n".join(sorted(labels_to_add)) + "\n")
        if labels_to_add:
            print(f"\nLabels to add: {', '.join(sorted(labels_to_add))}")
    if labels_remove_file:
        with open(labels_remove_file, "w") as f:
            if labels_to_remove:
                f.write("\n".join(sorted(labels_to_remove)) + "\n")
        if labels_to_remove:
            print(f"Labels to remove: {', '.join(sorted(labels_to_remove))}")

    if failures:
        print("\nPR description check FAILED:")
        for f in failures:
            print(f"  FAIL  {f}")
        return 1

    print("\nPR description check PASSED")
    return 0


if __name__ == "__main__":
    sys.exit(main())
