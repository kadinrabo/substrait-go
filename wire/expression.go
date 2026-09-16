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
	case *expr.IfThen:
		return ifThenToProto(e)
	case *expr.SwitchExpr:
		return switchExprToProto(e)
	case *expr.SingularOrList:
		return singularOrListToProto(e)
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

func ifThenToProto(ex *expr.IfThen) *proto.Expression {
	clauses := make([]*proto.Expression_IfThen_IfClause, ex.NIfs())
	for i := range clauses {
		pair := ex.IfPair(i)
		clauses[i] = &proto.Expression_IfThen_IfClause{
			If:   ExprToProto(pair.If),
			Then: ExprToProto(pair.Then),
		}
	}

	var elseClause *proto.Expression
	if e := ex.Else(); e != nil {
		elseClause = ExprToProto(e)
	}
	return &proto.Expression{
		RexType: &proto.Expression_IfThen_{
			IfThen: &proto.Expression_IfThen{
				Ifs:  clauses,
				Else: elseClause,
			},
		},
	}
}

func switchExprToProto(ex *expr.SwitchExpr) *proto.Expression {
	var elseExpr *proto.Expression
	if e := ex.Else(); e != nil {
		elseExpr = ExprToProto(e)
	}

	cases := make([]*proto.Expression_SwitchExpression_IfValue, ex.NCases())
	for i := range cases {
		c := ex.Case(i)
		cases[i] = &proto.Expression_SwitchExpression_IfValue{
			If:   LiteralToProto(c.If),
			Then: ExprToProto(c.Then),
		}
	}

	return &proto.Expression{
		RexType: &proto.Expression_SwitchExpression_{
			SwitchExpression: &proto.Expression_SwitchExpression{
				Match: ExprToProto(ex.MatchExpr()),
				Ifs:   cases,
				Else:  elseExpr,
			},
		},
	}
}

func singularOrListToProto(ex *expr.SingularOrList) *proto.Expression {
	opts := make([]*proto.Expression, len(ex.Options))
	for i, o := range ex.Options {
		opts[i] = ExprToProto(o)
	}
	return &proto.Expression{
		RexType: &proto.Expression_SingularOrList_{
			SingularOrList: &proto.Expression_SingularOrList{
				Value:   ExprToProto(ex.Value),
				Options: opts,
			},
		},
	}
}
