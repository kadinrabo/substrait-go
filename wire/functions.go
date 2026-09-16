// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"fmt"

	"github.com/substrait-io/substrait-go/v9/expr"
	"github.com/substrait-io/substrait-go/v9/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

// FunctionOptionsToProto encodes domain FunctionOptions as their protobuf messages.
func FunctionOptionsToProto(opts []*types.FunctionOption) []*proto.FunctionOption {
	if opts == nil {
		return nil
	}
	out := make([]*proto.FunctionOption, len(opts))
	for i, o := range opts {
		if o == nil {
			continue
		}
		out[i] = &proto.FunctionOption{Name: o.Name, Preference: o.Preference}
	}
	return out
}

func scalarFunctionToProto(s *expr.ScalarFunction) *proto.Expression {
	args := make([]*proto.FunctionArgument, s.NArgs())
	for i := range args {
		args[i] = FuncArgToProto(s.Arg(i))
	}

	return &proto.Expression{
		RexType: &proto.Expression_ScalarFunction_{
			ScalarFunction: &proto.Expression_ScalarFunction{
				FunctionReference: s.FuncRef(),
				Options:           FunctionOptionsToProto(s.GetOptions()),
				OutputType:        TypeToProto(s.GetType()),
				Arguments:         args,
			},
		},
	}
}
