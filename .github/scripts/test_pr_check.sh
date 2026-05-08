#!/usr/bin/env bash
# Usage: ./test_pr_check.sh <PR_NUMBER> [repo]
# Example: ./test_pr_check.sh 763
#          ./test_pr_check.sh 763 Domain-Connect/Templates

set -euo pipefail

PR_NUMBER="${1:?Usage: $0 <PR_NUMBER> [owner/repo]}"
REPO="${2:-Domain-Connect/Templates}"

echo "=== Fetching PR #${PR_NUMBER} from ${REPO} ==="

# Fetch PR metadata (body + head sha + head repo)
PR_META=$(curl -sf "https://api.github.com/repos/${REPO}/pulls/${PR_NUMBER}")
PR_BODY=$(echo "$PR_META" | python3 -c "import sys,json; print(json.load(sys.stdin)['body'] or '')")
HEAD_SHA=$(echo "$PR_META" | python3 -c "import sys,json; print(json.load(sys.stdin)['head']['sha'])")
HEAD_REPO=$(echo "$PR_META" | python3 -c "import sys,json; print(json.load(sys.stdin)['head']['repo']['full_name'])")

echo "Head: ${HEAD_REPO}@${HEAD_SHA}"

# Fetch changed JSON template files (filenames only, space-separated)
TEMPLATE_FILES=$(curl -sf "https://api.github.com/repos/${REPO}/pulls/${PR_NUMBER}/files" \
  | python3 -c "
import sys, json
files = [f['filename'] for f in json.load(sys.stdin) if f['filename'].endswith('.json')]
print(' '.join(files))
")

echo "Template files in PR: ${TEMPLATE_FILES:-<none>}"

# Download template files from the PR's head commit so they match what was tested
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

DOWNLOADED_FILES=""
for f in $TEMPLATE_FILES; do
  dest="$TMPDIR/$f"
  mkdir -p "$(dirname "$dest")"
  url="https://raw.githubusercontent.com/${HEAD_REPO}/${HEAD_SHA}/${f}"
  if curl -sf "$url" -o "$dest"; then
    echo "  Downloaded: $f (from PR head)"
    DOWNLOADED_FILES="$DOWNLOADED_FILES $dest"
  else
    echo "  WARNING: could not download $f from PR head, skipping"
  fi
done

LABELS_FILE=$(mktemp)

echo ""
echo "=== Running check ==="
set +e
PR_BODY="$PR_BODY" \
TEMPLATE_FILES="$DOWNLOADED_FILES" \
LABELS_FILE="$LABELS_FILE" \
  python3 "$(dirname "$0")/check_pr_description.py"
EXIT_CODE=$?
set -e

echo ""
echo "=== Labels that would be applied ==="
if [ -s "$LABELS_FILE" ]; then
  cat "$LABELS_FILE"
else
  echo "(none)"
fi

rm -f "$LABELS_FILE"
exit $EXIT_CODE
