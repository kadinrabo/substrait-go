# substrait-go/codec

Conversion between substrait-go's domain types and the generated Substrait
protobuf messages: encode with the `*ToProto` helpers, decode with `*FromProto`.

It lives in its own module so `substrait-protobuf` stays out of core's dependency
graph. A consumer that only builds and inspects plans imports the core packages
and never compiles the protobuf codegen; a consumer that needs the wire format
imports this module on purpose. See issue #280.

## Install

    go get github.com/substrait-io/substrait-go/codec

codec depends on the core module (`github.com/substrait-io/substrait-go/v9`), so
`go get` pulls that in too.

## Local development

The repo commits a `go.work` tying the core and codec modules together, so a
checkout builds codec against the local core instead of a published release. From
the repo root or from `codec/`:

    go build ./...
    go test ./...

No `replace` directive is needed, and the committed `codec/go.mod` stays
release-shaped.

## Releasing

codec depends on the core module in the same repo, so they release in order:

1. Release core (the root module) the usual way. That tags `v9.x.y`.
2. Point codec at that release and commit the result:

       cd codec
       go get github.com/substrait-io/substrait-go/v9@v9.x.y
       go mod tidy

3. Confirm the module is releasable, then tag it with its path prefix and push:

       ./scripts/check-codec-release.sh
       git tag codec/v0.1.0
       git push origin codec/v0.1.0

   `check-codec-release.sh` fails if codec/go.mod still pins a pre-release core
   version or carries a replace directive, so a broken module never gets tagged.
   It fails today on purpose: codec isn't releasable until step 2 points it at a
   real core release.

The `require` in `codec/go.mod` pins `v9.0.0-alpha.1` for now, the latest core
tag, so the workspace resolves and CI builds. Before codec's first real release,
step 2 has to point it at a core release that actually contains the proto removal
from #280. `go.work` only affects this repo's local builds; consumers resolve the
real `require`.
