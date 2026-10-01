# substrait-go/codec

Conversion between substrait-go's domain types and the generated Substrait
protobuf messages: encode with the `*ToProto` helpers, decode with `*FromProto`.

It lives in its own module so `substrait-protobuf` stays out of core's dependency
graph. A consumer that only builds and inspects plans imports the core packages
and never compiles the protobuf codegen; a consumer that needs the wire format
imports this module on purpose. See issue #280.

## Install

    go get github.com/substrait-io/substrait-go/codec/v9

```go
import (
    "github.com/substrait-io/substrait-go/v9/plan"
    "github.com/substrait-io/substrait-go/codec/v9" // package codec
)
```

codec is versioned in lockstep with core: codec `v9.x.y` requires core `v9.x.y`
exactly, so matching versions always work together. A core major bump moves
codec to the same major.

## Local development

`codec/go.mod` carries `replace github.com/substrait-io/substrait-go/v9 => ../`,
so a checkout builds codec against the core in this repo:

    cd codec
    go build ./...
    go test ./...

Consumers ignore replace directives in their dependencies, so it never reaches
them. On main the core `require` may lag behind; only the release commit has to
be exact.

## Releasing

Automatic. When the release workflow tags core `vX.Y.Z`, it runs
`scripts/release-codec.sh vX.Y.Z`, which on a detached HEAD at that tag:

1. pins codec's core `require` to `vX.Y.Z`,
2. builds and tests a copy with the replace dropped, resolving core from the tag
   the way a consumer does,
3. commits and tags `codec/vX.Y.Z`.

The workflow then pushes the tag. The release commit lives only under the tag,
so main is never written to. To cut one by hand, run the script and push the tag.
