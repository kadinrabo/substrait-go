# substrait-go/codec

Converts substrait-go's domain types to and from the generated Substrait protobuf
messages. It is a separate module so core needs no `substrait-protobuf` dependency;
import it only when you need the wire format.

    go get github.com/substrait-io/substrait-go/codec/v9

codec is released in lockstep with core: `codec/vX.Y.Z` requires core `vX.Y.Z`
exactly. Local dev builds against the in-repo core through the `replace` in
`codec/go.mod`; `scripts/release-codec.sh` cuts a release (pinning core, dropping
the replace) and the release workflow tags it.
