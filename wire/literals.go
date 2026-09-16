// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"fmt"

	"github.com/substrait-io/substrait-go/v9/expr"
	"github.com/substrait-io/substrait-go/v9/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

// LiteralToProto encodes a literal as its protobuf message.
func LiteralToProto(l expr.Literal) *proto.Expression_Literal {
	switch l := l.(type) {
	case *expr.NullLiteral:
		return nullLiteralToProto(l)
	default:
		panic(fmt.Sprintf("wire: unhandled literal %T", l))
	}
}

func nullLiteralToProto(n *expr.NullLiteral) *proto.Expression_Literal {
	return &proto.Expression_Literal{
		Nullable:               true,
		TypeVariationReference: n.Type.GetTypeVariationReference(),
		LiteralType:            &proto.Expression_Literal_Null{Null: TypeToProto(n.Type)},
	}
}
