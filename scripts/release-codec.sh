#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# Tags a codec release for an existing core release (issue #280).
#
# codec is versioned in lockstep with core: codec vX.Y.Z requires core vX.Y.Z
# exactly. On main, codec/go.mod builds against the in-repo core through a
# replace directive. This script makes the release commit on a detached HEAD
# (never on main): it pins the core require to the release tag, proves the module
# builds and tests the way a consumer resolves it (replace dropped, core fetched
# from the tag), commits, and tags codec/vX.Y.Z. Pushing the tag is left to the
# caller.
#
# Usage:
#   ./scripts/release-codec.sh v9.1.0

set -euo pipefail

version="${1:-}"
if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "usage: $0 vX.Y.Z (got '${version}')"
    exit 1
fi

root=$(git rev-parse --show-toplevel)
cd "$root"

core=$(go mod edit -json | jq -r '.Module.Path')
codec=$(cd codec && go mod edit -json | jq -r '.Module.Path')
tag="codec/$version"

major="v$(echo "${version#v}" | cut -d. -f1)"
for path in "$core" "$codec"; do
    if [[ "$path" != */"$major" ]]; then
        echo "module $path does not match release major $major"
        exit 1
    fi
done

if ! git rev-parse -q --verify "refs/tags/$version" >/dev/null; then
    echo "core tag $version does not exist; release core first"
    exit 1
fi
if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
    echo "$tag already exists"
    exit 1
fi

git checkout -q --detach "$version"

(cd codec && go mod edit -require="$core@$version")

# Build and test a copy with the replace dropped, so the core comes from the
# tag exactly as a consumer would resolve it. GOPROXY=direct because the proxy
# may not have seen a tag pushed moments ago.
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
cp -R codec/. "$tmp"
(
    cd "$tmp"
    go mod edit -dropreplace="$core"
    GOPROXY=direct GOFLAGS=-mod=mod go mod tidy
    go build ./...
    go test ./...
)
cp "$tmp/go.sum" codec/go.sum

git add codec/go.mod codec/go.sum
git commit -q -m "chore(codec): release $version"
git tag -a "$tag" -m "codec $version"

echo "OK: tagged $tag (codec requires $core $version)"
