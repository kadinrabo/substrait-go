#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# Tag codec/vX.Y.Z for an existing core release vX.Y.Z (lockstep, #280). The
# commit is made on a detached HEAD, so main keeps its dev setup; the caller
# pushes the tag. Re-running an existing codec tag is a no-op.
# Usage: ./scripts/release-codec.sh v9.1.0

set -euo pipefail

version="${1:-}"
[[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "usage: $0 vX.Y.Z"; exit 1; }
major="${version%%.*}"

cd "$(git rev-parse --show-toplevel)"
core=$(go mod edit -json | jq -r .Module.Path)
codec=$(cd codec && go mod edit -json | jq -r .Module.Path)
for m in "$core" "$codec"; do
    [[ "$m" == */"$major" ]] || { echo "$m is not major $major; bump the module paths first"; exit 1; }
done

git rev-parse -q --verify "refs/tags/$version" >/dev/null || { echo "core tag $version missing; release core first"; exit 1; }
git rev-parse -q --verify "refs/tags/codec/$version" >/dev/null && { echo "codec/$version exists; nothing to do"; exit 0; }

git checkout -q --detach "$version"

# Pin core to the tag, then prove codec builds the way a consumer resolves it.
# GOWORK=off ignores the dev go.work so core resolves from the tag, not this tree;
# GOPROXY=direct because the core tag may be seconds old.
(cd codec
    export GOWORK=off
    go mod edit -require="$core@$version"
    GOPROXY=direct GOFLAGS=-mod=mod go mod tidy
    go build ./... && go test ./...)

git commit -q -m "chore(codec): release $version" codec/go.mod codec/go.sum
git tag -a "codec/$version" -m "codec $version"
echo "OK: tagged codec/$version"
