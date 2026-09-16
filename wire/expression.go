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
	case *expr.Cast:
		return castToProto(e)
	case *expr.DynamicParameter:
		return dynamicParameterToProto(e)
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

func castToProto(ex *expr.Cast) *proto.Expression {
	return &proto.Expression{
		RexType: &proto.Expression_Cast_{
			Cast: &proto.Expression_Cast{
				Type:            TypeToProto(ex.Type),
				Input:           ExprToProto(ex.Input),
				FailureBehavior: proto.Expression_Cast_FailureBehavior(ex.FailureBehavior),
			},
		},
	}
}

func dynamicParameterToProto(dp *expr.DynamicParameter) *proto.Expression {
	return &proto.Expression{
		RexType: &proto.Expression_DynamicParameter{
			DynamicParameter: &proto.DynamicParameter{
				Type:               TypeToProto(dp.OutputType),
				ParameterReference: dp.ParameterReference,
			},
		},
	}
}
