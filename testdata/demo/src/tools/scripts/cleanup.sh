#!/usr/bin/env bash
# Remove build output and caches
set -euo pipefail
rm -rf bin dist
echo "removed bin/ and dist/"
