// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"fmt"

	"github.com/substrait-io/substrait-go/v9/expr"
	"github.com/substrait-io/substrait-go/v9/plan"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

// ExprToProto encodes an expression as its protobuf message.
func ExprToProto(e expr.Expression) *proto.Expression {
	switch e := e.(type) {
	case *expr.ScalarFunction:
		return scalarFunctionToProto(e)
	case *expr.WindowFunction:
		return windowFunctionToProto(e)
	case *expr.FieldReference:
		return FieldReferenceToProto(e)
	case expr.Literal:
		return &proto.Expression{
			RexType: &proto.Expression_Literal_{Literal: LiteralToProto(e)},
		}
	default:
		panic(fmt.Sprintf("wire: unhandled expression %T", e))
	}
}
