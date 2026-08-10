// SPDX-License-Identifier: Apache-2.0

package codec_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substrait-io/substrait-go/codec"
	"github.com/substrait-io/substrait-go/v8/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const versionGolden = "testdata/version.golden"

// versionGoldenHeader names the columns and doubles as the parse error message.
const versionGoldenHeader = "name\tmajor\tminor\tpatch\tgitHash\tproducer\twireHex"

// versionCase is a named domain version, shared with the differential test against the core.
type versionCase struct {
	name    string
	version types.Version
}

// versionValues is the golden's row order. The rows are picked so that a conversion that reads the
// wrong field, normalizes a value, or drops one shows up as a diff: every number field is maxed on
// its own, and the two strings are set on their own.
var versionValues = []versionCase{
	{"zero", types.Version{}},
	{"populated", types.Version{
		MajorNumber: 1,
		MinorNumber: 29,
		PatchNumber: 3,
		GitHash:     "0123456789abcdef0123456789abcdef01234567",
		// a real producer string, spaces included, which is why the golden is tab separated
		Producer: "substrait-go v8.0.0 darwin/arm64",
	}},
	{"major_max", types.Version{MajorNumber: math.MaxUint32}},
	{"minor_max", types.Version{MinorNumber: math.MaxUint32}},
	{"patch_max", types.Version{PatchNumber: math.MaxUint32}},
	{"git_hash_only", types.Version{GitHash: "abc123"}},
	{"producer_only", types.Version{Producer: "substrait-go"}},
	// a producer is free text, so its bytes have to survive as bytes rather than as characters
	{"multibyte_producer", types.Version{Producer: "substrait-go (café build)"}},
	// the spec asks for 40 lowercase hex characters. This is neither, and has to survive anyway.
	{"loose_git_hash", types.Version{GitHash: "HEAD"}},
	// the populated row's hash is exactly the length the spec asks for, so without a longer one a
	// conversion that trimmed to 40 would pass every row here
	{"long_git_hash", types.Version{GitHash: "0123456789ABCDEF0123456789abcdef0123456789abcdef"}},
}

// versionRecord is one golden row: a domain version and the bytes a plan carrying it serializes to.
type versionRecord struct {
	name    string
	version types.Version
	wire    []byte
}

// Tab separated rather than space separated like the enum goldens, because a producer is free text
// and carries spaces in practice.
func readVersionGolden(t *testing.T, path string) []versionRecord {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)

	var records []versionRecord
	for i, line := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		columns := strings.Split(line, "\t")
		require.Lenf(t, columns, 7, "%s:%d is not %q", path, i+1, versionGoldenHeader)

		numbers := make([]uint32, 3)
		for j, column := range columns[1:4] {
			n, err := strconv.ParseUint(column, 10, 32)
			require.NoErrorf(t, err, "%s:%d column %d", path, i+1, j+2)
			numbers[j] = uint32(n)
		}
		gitHash, err := strconv.Unquote(columns[4])
		require.NoErrorf(t, err, "%s:%d gitHash is not a quoted string", path, i+1)
		producer, err := strconv.Unquote(columns[5])
		require.NoErrorf(t, err, "%s:%d producer is not a quoted string", path, i+1)
		wire, err := hex.DecodeString(columns[6])
		require.NoErrorf(t, err, "%s:%d wireHex is not hex", path, i+1)

		records = append(records, versionRecord{
			name: columns[0],
			version: types.Version{
				MajorNumber: numbers[0],
				MinorNumber: numbers[1],
				PatchNumber: numbers[2],
				GitHash:     gitHash,
				Producer:    producer,
			},
			wire: wire,
		})
	}
	require.NotEmpty(t, records)
	return records
}

// codecVersionWire encodes one version through codec. substrait.Plan is where the spec puts the
// version field, and going through it keeps an absent version distinguishable from an empty one,
// which marshaling the bare message would not.
func codecVersionWire(t *testing.T, v *types.Version) []byte {
	t.Helper()
	b, err := protobuf.MarshalOptions{Deterministic: true}.Marshal(&proto.Plan{Version: codec.VersionToProto(v)})
	require.NoError(t, err)
	return b
}

func renderVersionGolden(t *testing.T) string {
	t.Helper()
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# %s\n", versionGoldenHeader)
	buf.WriteString("# substrait.Plan{version} encoded through codec, tab separated.\n")
	buf.WriteString("# Strings are Go quoted, so an empty one is visible as \"\".\n")
	for _, v := range versionValues {
		fmt.Fprintf(&buf, "%s\t%d\t%d\t%d\t%q\t%q\t%s\n",
			v.name, v.version.MajorNumber, v.version.MinorNumber, v.version.PatchNumber,
			v.version.GitHash, v.version.Producer, hex.EncodeToString(codecVersionWire(t, &v.version)))
	}
	return buf.String()
}

// TestVersionGoldenMatchesCodec compares the committed bytes against what codec encodes now. The
// golden is a snapshot, so a later change to the conversion shows up as a diff.
func TestVersionGoldenMatchesCodec(t *testing.T) {
	if *update {
		require.NoError(t, os.WriteFile(versionGolden, []byte(renderVersionGolden(t)), 0o644))
		return
	}

	records := readVersionGolden(t, versionGolden)
	require.Len(t, records, len(versionValues), "the golden does not have a row per value")
	for i, v := range versionValues {
		t.Run(v.name, func(t *testing.T) {
			assert.Equal(t, v.name, records[i].name)
			assert.Equal(t, v.version, records[i].version, "the golden's field columns")
			assert.Equal(t, hex.EncodeToString(codecVersionWire(t, &v.version)), hex.EncodeToString(records[i].wire))
		})
	}
}

// TestVersionGoldenWireCarriesTheFields reads each row's bytes and checks them against that row's
// field columns without going through the conversion. Decoding through VersionFromProto would only
// prove the pair agrees with itself, so a drift in both directions would regenerate the golden with
// wrong bytes and still pass.
func TestVersionGoldenWireCarriesTheFields(t *testing.T) {
	for _, record := range readVersionGolden(t, versionGolden) {
		t.Run(record.name, func(t *testing.T) {
			var decoded proto.Plan
			require.NoError(t, protobuf.Unmarshal(record.wire, &decoded))
			wire := decoded.GetVersion()
			require.NotNil(t, wire, "the row's bytes carry no version at all")

			assert.Equal(t, record.version.MajorNumber, wire.GetMajorNumber(), "major_number")
			assert.Equal(t, record.version.MinorNumber, wire.GetMinorNumber(), "minor_number")
			assert.Equal(t, record.version.PatchNumber, wire.GetPatchNumber(), "patch_number")
			assert.Equal(t, record.version.GitHash, wire.GetGitHash(), "git_hash")
			assert.Equal(t, record.version.Producer, wire.GetProducer(), "producer")

			// and the decode half lands on the same value those bytes carry
			assert.Equal(t, &record.version, codec.VersionFromProto(wire))
		})
	}
}

// TestVersionGoldenCoversDescriptor fails if the spec gains a version field no row writes, so the
// golden cannot quietly stop covering the message.
func TestVersionGoldenCoversDescriptor(t *testing.T) {
	written := map[protoreflect.Name]bool{}
	for _, record := range readVersionGolden(t, versionGolden) {
		var decoded proto.Plan
		require.NoError(t, protobuf.Unmarshal(record.wire, &decoded))
		require.NotNil(t, decoded.GetVersion())
		decoded.GetVersion().ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
			written[fd.Name()] = true
			return true
		})
	}

	fields := (&proto.Version{}).ProtoReflect().Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		name := fields.Get(i).Name()
		assert.Truef(t, written[name], "no golden row puts %s on the wire", name)
	}
}

// The version field is optional on the wire, so an absent version has to stay absent in both
// directions. Mapping nil onto an empty message would make a plan that carries no version claim
// 0.0.0 instead.
func TestVersionAbsentStaysAbsent(t *testing.T) {
	assert.Nil(t, codec.VersionToProto(nil))
	assert.Nil(t, codec.VersionFromProto(nil))

	assert.Empty(t, codecVersionWire(t, nil), "an absent version must not write any bytes")
	// an empty version is a different thing: field 6, length zero
	assert.Equal(t, "3200", hex.EncodeToString(codecVersionWire(t, &types.Version{})))
}

// The domain value and the message must not share anything, so a caller editing one cannot reach
// into the other. types.Version used to be the generated message, where handing out the same
// pointer let a caller rewrite plan.CurrentVersion.
func TestVersionConversionsCopy(t *testing.T) {
	domain := &types.Version{MinorNumber: 29, Producer: "substrait-go"}

	wire := codec.VersionToProto(domain)
	wire.Producer = "someone else"
	assert.Equal(t, "substrait-go", domain.Producer)

	back := codec.VersionFromProto(wire)
	back.Producer = "a third party"
	assert.Equal(t, "someone else", wire.Producer)
}

// Unknown fields are the one thing a round trip loses, since the domain struct has nowhere to keep
// them. Pinned here so it stays a decision rather than a surprise.
func TestVersionDropsUnknownFields(t *testing.T) {
	fromANewerSpec := &proto.Version{MinorNumber: 29}
	// field 6 as a varint, which this version of the spec does not define
	fromANewerSpec.ProtoReflect().SetUnknown(protoreflect.RawFields([]byte{0x30, 0x01}))

	domain := codec.VersionFromProto(fromANewerSpec)
	assert.Equal(t, &types.Version{MinorNumber: 29}, domain)
	assert.Empty(t, codec.VersionToProto(domain).ProtoReflect().GetUnknown())
}
