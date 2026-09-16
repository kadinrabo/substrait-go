// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"fmt"

	"github.com/substrait-io/substrait-go/v9/expr"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

// RefSegmentToProto encodes a reference segment as its protobuf message.
func RefSegmentToProto(r expr.ReferenceSegment) *proto.Expression_ReferenceSegment {
	switch r := r.(type) {
	case *expr.MapKeyRef:
		return mapKeyRefToProto(r)
	case *expr.StructFieldRef:
		return structFieldRefToProto(r)
	case *expr.ListElementRef:
		return listElementRefToProto(r)
	default:
		panic(fmt.Sprintf("wire: unhandled reference segment %T", r))
	}
}

func maskSelectToProto(s expr.MaskSelect) *proto.Expression_MaskExpression_Select {
	switch s := s.(type) {
	case expr.MaskStructSelect:
		return &proto.Expression_MaskExpression_Select{
			Type: &proto.Expression_MaskExpression_Select_Struct{Struct: maskStructSelectToProto(s)},
		}
	case *expr.MaskListSelect:
		return maskListSelectToProto(s)
	case *expr.MaskMapSelect:
		return maskMapSelectToProto(s)
	default:
		panic(fmt.Sprintf("wire: unhandled mask selection %T", s))
	}
}
