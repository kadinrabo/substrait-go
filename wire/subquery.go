// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"github.com/substrait-io/substrait-go/v9/plan"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

func scalarSubqueryToProto(s *plan.ScalarSubquery) *proto.Expression {
	return &proto.Expression{
		RexType: &proto.Expression_Subquery_{
			Subquery: &proto.Expression_Subquery{
				SubqueryType: &proto.Expression_Subquery_Scalar_{
					Scalar: &proto.Expression_Subquery_Scalar{
						Input: RelToProto(s.Input),
					},
				},
			},
		},
	}
}
