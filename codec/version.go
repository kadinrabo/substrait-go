// SPDX-License-Identifier: Apache-2.0

package codec

import (
	"github.com/substrait-io/substrait-go/v8/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

// VersionToProto encodes a domain version as its protobuf message. A nil version stays absent
// rather than becoming an empty message, since the field is optional on the wire and an empty
// message reads as a plan claiming version 0.0.0.
func VersionToProto(v *types.Version) *proto.Version {
	if v == nil {
		return nil
	}
	return &proto.Version{
		MajorNumber: v.MajorNumber,
		MinorNumber: v.MinorNumber,
		PatchNumber: v.PatchNumber,
		GitHash:     v.GitHash,
		Producer:    v.Producer,
	}
}

// VersionFromProto decodes a version message. Fields are carried across as they stand: the spec
// asks producers for a 40 character lowercase hex git hash, but enforcing that here would reject or
// rewrite plans substrait-go can otherwise read. Unknown fields are the one thing lost, the domain
// struct has nowhere to keep them.
func VersionFromProto(v *proto.Version) *types.Version {
	if v == nil {
		return nil
	}
	return &types.Version{
		MajorNumber: v.MajorNumber,
		MinorNumber: v.MinorNumber,
		PatchNumber: v.PatchNumber,
		GitHash:     v.GitHash,
		Producer:    v.Producer,
	}
}
