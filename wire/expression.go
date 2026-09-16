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
	case *expr.MultiOrList:
		return multiOrListToProto(e)
	case *expr.MapExpr:
		return mapExprToProto(e)
	case *expr.StructExpr:
		return structExprToProto(e)
	case *expr.ListExpr:
		return listExprToProto(e)
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

func multiOrListToProto(ex *expr.MultiOrList) *proto.Expression {
	toSlice := func(exprs []expr.Expression) []*proto.Expression {
		out := make([]*proto.Expression, len(exprs))
		for i, e := range exprs {
			out[i] = ExprToProto(e)
		}
		return out
	}

	opts := make([]*proto.Expression_MultiOrList_Record, len(ex.Options))
	for i, o := range ex.Options {
		opts[i] = &proto.Expression_MultiOrList_Record{Fields: toSlice(o)}
	}

	return &proto.Expression{
		RexType: &proto.Expression_MultiOrList_{
			MultiOrList: &proto.Expression_MultiOrList{
				Value:   toSlice(ex.Value),
				Options: opts,
			},
		},
	}
}

func mapExprToProto(ex *expr.MapExpr) *proto.Expression {
	kvs := make([]*proto.Expression_Nested_Map_KeyValue, len(ex.KeyValues))
	for i, kv := range ex.KeyValues {
		kvs[i] = &proto.Expression_Nested_Map_KeyValue{
			Key:   ExprToProto(kv.Key),
			Value: ExprToProto(kv.Value),
		}
	}
	return &proto.Expression{
		RexType: &proto.Expression_Nested_{
			Nested: &proto.Expression_Nested{
				Nullable:               ex.Nullable,
				TypeVariationReference: ex.TypeVariationRef,
				NestedType: &proto.Expression_Nested_Map_{
					Map: &proto.Expression_Nested_Map{KeyValues: kvs},
				},
			},
		},
	}
}

func structExprToProto(ex *expr.StructExpr) *proto.Expression {
	fields := make([]*proto.Expression, len(ex.Fields))
	for i, f := range ex.Fields {
		fields[i] = ExprToProto(f)
	}
	return &proto.Expression{
		RexType: &proto.Expression_Nested_{
			Nested: &proto.Expression_Nested{
				Nullable:               ex.Nullable,
				TypeVariationReference: ex.TypeVariationRef,
				NestedType: &proto.Expression_Nested_Struct_{
					Struct: &proto.Expression_Nested_Struct{Fields: fields},
				},
			},
		},
	}
}

func listExprToProto(ex *expr.ListExpr) *proto.Expression {
	vals := make([]*proto.Expression, len(ex.Values))
	for i, v := range ex.Values {
		vals[i] = ExprToProto(v)
	}
	return &proto.Expression{
		RexType: &proto.Expression_Nested_{
			Nested: &proto.Expression_Nested{
				Nullable:               ex.Nullable,
				TypeVariationReference: ex.TypeVariationRef,
				NestedType: &proto.Expression_Nested_List_{
					List: &proto.Expression_Nested_List{Values: vals},
				},
			},
		},
	}
}

// ExpressionReferenceToProto encodes an expression reference as its protobuf
// message.
func ExpressionReferenceToProto(er *expr.ExpressionReference) *proto.ExpressionReference {
	out := &proto.ExpressionReference{OutputNames: er.OutputNames}
	switch {
	case er.GetExpr() != nil:
		out.ExprType = &proto.ExpressionReference_Expression{
			Expression: ExprToProto(er.GetExpr()),
		}
	case er.GetMeasure() != nil:
		out.ExprType = &proto.ExpressionReference_Measure{
			Measure: AggregateFunctionToProto(er.GetMeasure()),
		}
	}
	return out
}
