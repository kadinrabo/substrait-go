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

func inPredicateSubqueryToProto(s *plan.InPredicateSubquery) *proto.Expression {
	needles := make([]*proto.Expression, len(s.Needles))
	for i, needle := range s.Needles {
		needles[i] = ExprToProto(needle)
	}

	return &proto.Expression{
		RexType: &proto.Expression_Subquery_{
			Subquery: &proto.Expression_Subquery{
				SubqueryType: &proto.Expression_Subquery_InPredicate_{
					InPredicate: &proto.Expression_Subquery_InPredicate{
						Needles:  needles,
						Haystack: RelToProto(s.Haystack),
					},
				},
			},
		},
	}
}

func setPredicateSubqueryToProto(s *plan.SetPredicateSubquery) *proto.Expression {
	return &proto.Expression{
		RexType: &proto.Expression_Subquery_{
			Subquery: &proto.Expression_Subquery{
				SubqueryType: &proto.Expression_Subquery_SetPredicate_{
					SetPredicate: &proto.Expression_Subquery_SetPredicate{
						PredicateOp: s.Operation,
						Tuples:      RelToProto(s.Tuples),
					},
				},
			},
		},
	}
}

func setComparisonSubqueryToProto(s *plan.SetComparisonSubquery) *proto.Expression {
	return &proto.Expression{
		RexType: &proto.Expression_Subquery_{
			Subquery: &proto.Expression_Subquery{
				SubqueryType: &proto.Expression_Subquery_SetComparison_{
					SetComparison: &proto.Expression_Subquery_SetComparison{
						ReductionOp:  proto.Expression_Subquery_SetComparison_ReductionOp(s.ReductionOp),
						ComparisonOp: proto.Expression_Subquery_SetComparison_ComparisonOp(s.ComparisonOp),
						Left:         ExprToProto(s.Left),
						Right:        RelToProto(s.Right),
					},
				},
			},
		},
	}
}
