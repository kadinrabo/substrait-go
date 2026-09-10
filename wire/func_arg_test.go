// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/substrait-io/substrait-go/v9/expr"
	"github.com/substrait-io/substrait-go/v9/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
	"google.golang.org/protobuf/testing/protocmp"
)

// FuncArgToProto routes each argument kind to the matching protobuf arm.
func TestFuncArgToProtoArms(t *testing.T) {
	lit := expr.NewPrimitiveLiteral[int32](5, false)
	typeArg := &types.Int32Type{Nullability: types.NullabilityRequired}

	cases := []struct {
		name string
		arg  types.FuncArg
		want *proto.FunctionArgument
	}{
		{"type", typeArg, &proto.FunctionArgument{ArgType: &proto.FunctionArgument_Type{Type: TypeToProto(typeArg)}}},
		{"enum", types.Enum("day"), &proto.FunctionArgument{ArgType: &proto.FunctionArgument_Enum{Enum: "day"}}},
		{"value", lit, &proto.FunctionArgument{ArgType: &proto.FunctionArgument_Value{Value: ExprToProto(lit)}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if diff := cmp.Diff(tc.want, FuncArgToProto(tc.arg), protocmp.Transform()); diff != "" {
				t.Errorf("FuncArgToProto(%s) mismatch (-want +got):\n%s", tc.name, diff)
			}
		})
	}
}
