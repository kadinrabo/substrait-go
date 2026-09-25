#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# Pre-release gate for the codec module (issue #280).
#
# codec depends on the core module in the same repo. Local dev resolves that
# through the committed go.work, and codec/go.mod pins the latest core tag as a
# placeholder. Before tagging a codec release the pin has to point at a real,
# non-prerelease core version that contains the proto removal, with no leftover
# replace directive. Otherwise the published module fails to compile for every
# consumer (it would resolve a core that lacks codec's API).
#
# This is not part of the per-push CI; run it right before `git tag
# codec/vX.Y.Z` (or wire it into codec's release). See codec/README.md.

set -euo pipefail

core='github.com/substrait-io/substrait-go/v9'
mod=$(cd codec && go mod edit -json)
fail=0

# A replace directive must never ship in a released go.mod.
if echo "$mod" | jq -e --arg m "$core" '(.Replace // []) | any(.Old.Path == $m)' >/dev/null; then
    echo "codec/go.mod still has a replace for $core; a released module must not."
    fail=1
fi

version=$(echo "$mod" | jq -r --arg m "$core" '(.Require // [])[] | select(.Path == $m) | .Version')
if [ -z "$version" ]; then
    echo "codec/go.mod has no require for $core."
    fail=1
elif [[ "$version" == *-* ]]; then
    # a hyphen means a pre-release (v9.0.0-alpha.1) or pseudo-version, not a real tag
    echo "codec/go.mod pins a pre-release core version ($version)."
    echo "Point it at a released core tag that contains the #280 proto removal:"
    echo "  cd codec && go get $core@vX.Y.Z && go mod tidy"
    fail=1
fi

if [ "$fail" -ne 0 ]; then
    echo
    echo "codec is not release-ready (issue #280)."
    exit 1
fi

echo "OK: codec pins a released core version ($version), no replace"
