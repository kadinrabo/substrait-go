// SPDX-License-Identifier: Apache-2.0

// The only file here that calls a core conversion, so it goes away with them.

package codec_test

import (
	"encoding/hex"
	"math"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substrait-io/substrait-go/codec"
	"github.com/substrait-io/substrait-go/v8/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
	protobuf "google.golang.org/protobuf/proto"
)

// coreVersionWire is the encode path codec is replacing.
func coreVersionWire(t *testing.T, v *types.Version) []byte {
	t.Helper()
	b, err := protobuf.MarshalOptions{Deterministic: true}.Marshal(&proto.Plan{Version: types.VersionToProto(v)})
	require.NoError(t, err)
	return b
}

// TestVersionMatchesCore checks codec against the core in both directions, over the golden rows plus
// values the golden does not carry. Whatever the core writes today codec has to write byte for byte,
// so which path built a plan cannot be told from the plan.
func TestVersionMatchesCore(t *testing.T) {
	values := append(slices.Clone(versionValues),
		versionCase{"all_fields_set", types.Version{
			MajorNumber: math.MaxUint32,
			MinorNumber: math.MaxUint32,
			PatchNumber: math.MaxUint32,
			GitHash:     "0123456789abcdef0123456789abcdef01234567",
			Producer:    "a producer with spaces, punctuation and version numbers 1.2.3",
		}},
		// the shape of plan.CurrentVersion, spelled out rather than imported so the codec module
		// does not take on the plan package's dependencies for one test row
		versionCase{"current_version", types.Version{MinorNumber: 29, Producer: "substrait-go v8.0.0 darwin/arm64"}},
	)

	for _, v := range values {
		t.Run(v.name, func(t *testing.T) {
			wire := coreVersionWire(t, &v.version)
			assert.Equal(t, hex.EncodeToString(wire), hex.EncodeToString(codecVersionWire(t, &v.version)))

			var decoded proto.Plan
			require.NoError(t, protobuf.Unmarshal(wire, &decoded))
			assert.Equal(t, types.VersionFromProto(decoded.GetVersion()), codec.VersionFromProto(decoded.GetVersion()))
		})
	}

	// the absent case matters most, since mapping nil onto an empty message would change what a
	// versionless plan reports
	assert.Nil(t, codec.VersionToProto(nil))
	assert.Equal(t, types.VersionToProto(nil), codec.VersionToProto(nil))
	assert.Nil(t, codec.VersionFromProto(nil))
	assert.Equal(t, types.VersionFromProto(nil), codec.VersionFromProto(nil))
}
