// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"fmt"
	"github.com/substrait-io/substrait-go/v9/expr"
	"github.com/substrait-io/substrait-go/v9/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

func scalarFunctionToProto(s *expr.ScalarFunction) *proto.Expression {
	args := make([]*proto.FunctionArgument, s.NArgs())
	for i := range args {
		args[i] = FuncArgToProto(s.Arg(i))
	}

	return &proto.Expression{
		RexType: &proto.Expression_ScalarFunction_{
			ScalarFunction: &proto.Expression_ScalarFunction{
				FunctionReference: s.FuncRef(),
				Options:           s.GetOptions(),
				OutputType:        TypeToProto(s.GetType()),
				Arguments:         args,
			},
		},
	}
}

func windowFunctionToProto(w *expr.WindowFunction) *proto.Expression {
	var (
		args       []*proto.FunctionArgument
		sorts      []*proto.SortField
		parts      []*proto.Expression
		upperBound *proto.Expression_WindowFunction_Bound
		lowerBound *proto.Expression_WindowFunction_Bound
	)

	if w.NArgs() > 0 {
		args = make([]*proto.FunctionArgument, w.NArgs())
		for i := range args {
			args[i] = FuncArgToProto(w.Arg(i))
		}
	}

	if len(w.Sorts) > 0 {
		sorts = make([]*proto.SortField, len(w.Sorts))
		for i, s := range w.Sorts {
			sorts[i] = SortFieldToProto(&s)
		}
	}

	if len(w.Partitions) > 0 {
		parts = make([]*proto.Expression, len(w.Partitions))
		for i, p := range w.Partitions {
			parts[i] = ExprToProto(p)
		}
	}

	if w.UpperBound != nil {
		upperBound = BoundToProto(w.UpperBound)
	}

	if w.LowerBound != nil {
		lowerBound = BoundToProto(w.LowerBound)
	}

	return &proto.Expression{
		RexType: &proto.Expression_WindowFunction_{
			WindowFunction: &proto.Expression_WindowFunction{
				FunctionReference: w.FuncRef(),
				Arguments:         args,
				Options:           w.GetOptions(),
				OutputType:        TypeToProto(w.GetType()),
				Phase:             w.Phase(),
				Sorts:             sorts,
				Invocation:        w.Invocation(),
				Partitions:        parts,
				BoundsType:        proto.Expression_WindowFunction_BoundsType(w.BoundsType),
				LowerBound:        lowerBound,
				UpperBound:        upperBound,
			},
		},
	}
}

// AggregateFunctionToProto encodes an aggregate function as its protobuf message.
func AggregateFunctionToProto(a *expr.AggregateFunction) *proto.AggregateFunction {
	var (
		args  []*proto.FunctionArgument
		sorts []*proto.SortField
	)
	if a.NArgs() > 0 {
		args = make([]*proto.FunctionArgument, a.NArgs())
		for i := range args {
			args[i] = FuncArgToProto(a.Arg(i))
		}
	}

	if len(a.Sorts) > 0 {
		sorts = make([]*proto.SortField, len(a.Sorts))
		for i, s := range a.Sorts {
			sorts[i] = SortFieldToProto(&s)
		}
	}

	return &proto.AggregateFunction{
		FunctionReference: a.FuncRef(),
		Arguments:         args,
		Options:           a.GetOptions(),
		OutputType:        TypeToProto(a.GetType()),
		Phase:             a.Phase(),
		Sorts:             sorts,
		Invocation:        a.Invocation(),
	}
}

// SortFieldToProto encodes a sort field as its protobuf message.
func SortFieldToProto(s *expr.SortField) *proto.SortField {
	ret := &proto.SortField{Expr: ExprToProto(s.Expr)}
	switch k := s.Kind.(type) {
	case types.SortDirection:
		ret.SortKind = &proto.SortField_Direction{
			Direction: proto.SortField_SortDirection(k)}
	case types.FunctionRef:
		ret.SortKind = &proto.SortField_ComparisonFunctionReference{
			ComparisonFunctionReference: uint32(k)}
	}
	return ret
}

// BoundToProto encodes a window-function bound as its protobuf message.
func BoundToProto(b expr.Bound) *proto.Expression_WindowFunction_Bound {
	switch b := b.(type) {
	case expr.PrecedingBound:
		return &proto.Expression_WindowFunction_Bound{
			Kind: &proto.Expression_WindowFunction_Bound_Preceding_{
				Preceding: &proto.Expression_WindowFunction_Bound_Preceding{Offset: int64(b)},
			},
		}
	case expr.FollowingBound:
		return &proto.Expression_WindowFunction_Bound{
			Kind: &proto.Expression_WindowFunction_Bound_Following_{
				Following: &proto.Expression_WindowFunction_Bound_Following{Offset: int64(b)},
			},
		}
	case expr.CurrentRow:
		return &proto.Expression_WindowFunction_Bound{
			Kind: &proto.Expression_WindowFunction_Bound_CurrentRow_{
				CurrentRow: &proto.Expression_WindowFunction_Bound_CurrentRow{},
			},
		}
	case expr.Unbounded:
		return &proto.Expression_WindowFunction_Bound{
			Kind: &proto.Expression_WindowFunction_Bound_Unbounded_{
				Unbounded: &proto.Expression_WindowFunction_Bound_Unbounded{},
			},
		}
	default:
		panic(fmt.Sprintf("wire: unhandled bound %T", b))
	}
}
