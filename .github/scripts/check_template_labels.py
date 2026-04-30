#!/usr/bin/env python3
"""
Check changed template JSON files and emit labels based on their properties.

Labels managed:
  - "syncBlock"            → set if any template has syncBlock set (truthy)
  - "no syncPubKeyDomain"  → set if any template lacks syncPubKeyDomain AND syncBlock is not set
  - "warnPhishing"         → set if any template has warnPhishing set (truthy) AND syncBlock is not set
  - "hostRequired"         → set if any template has hostRequired set (truthy)

Each label is added when the condition is met in at least one template,
and removed when no template in the PR satisfies the condition.
"no syncPubKeyDomain" and "warnPhishing" are suppressed for templates where syncBlock is set.

Environment variables:
  TEMPLATE_FILES      space-separated list of changed *.json template paths
  LABELS_ADD_FILE     path to write labels that should be added (one per line)
  LABELS_REMOVE_FILE  path to write labels that should be removed (one per line)
"""

import json
import os
import sys

LABEL_SYNC_BLOCK = "syncBlock"
LABEL_NO_SYNC_PUB_KEY_DOMAIN = "no syncPubKeyDomain"
LABEL_WARN_PHISHING = "warnPhishing"
LABEL_HOST_REQUIRED = "hostRequired"

ALL_MANAGED_LABELS = {
    LABEL_SYNC_BLOCK,
    LABEL_NO_SYNC_PUB_KEY_DOMAIN,
    LABEL_WARN_PHISHING,
    LABEL_HOST_REQUIRED,
}


def main() -> int:
    template_files_env = os.environ.get("TEMPLATE_FILES", "")
    template_files = [f for f in template_files_env.split() if f.strip()]
    labels_add_file = os.environ.get("LABELS_ADD_FILE", "")
    labels_remove_file = os.environ.get("LABELS_REMOVE_FILE", "")

    if not template_files:
        print("No template files to check.")
        labels_to_add: set[str] = set()
    else:
        sync_block = False
        no_sync_pub_key_domain = False
        warn_phishing = False
        host_required = False

        for fpath in template_files:
            fname = os.path.basename(fpath)
            try:
                with open(fpath) as f:
                    tmpl = json.load(f)
            except Exception as e:
                print(f"  WARNING: could not read {fname}: {e}", file=sys.stderr)
                continue

            tmpl_sync_block = bool(tmpl.get("syncBlock"))

            if tmpl_sync_block:
                sync_block = True
                print(f"  {fname}: syncBlock={tmpl['syncBlock']!r}")

            if not tmpl.get("syncPubKeyDomain") and not tmpl_sync_block:
                no_sync_pub_key_domain = True
                print(f"  {fname}: syncPubKeyDomain not set")

            if tmpl.get("warnPhishing") and not tmpl_sync_block:
                warn_phishing = True
                print(f"  {fname}: warnPhishing={tmpl['warnPhishing']!r}")

            if tmpl.get("hostRequired"):
                host_required = True
                print(f"  {fname}: hostRequired={tmpl['hostRequired']!r}")

        labels_to_add = set()
        if sync_block:
            labels_to_add.add(LABEL_SYNC_BLOCK)
        if no_sync_pub_key_domain:
            labels_to_add.add(LABEL_NO_SYNC_PUB_KEY_DOMAIN)
        if warn_phishing:
            labels_to_add.add(LABEL_WARN_PHISHING)
        if host_required:
            labels_to_add.add(LABEL_HOST_REQUIRED)

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

    return 0


if __name__ == "__main__":
    sys.exit(main())
