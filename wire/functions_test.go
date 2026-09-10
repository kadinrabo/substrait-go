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
// SortFieldToProto encodes both the direction and comparison-function kinds.
func TestSortFieldToProtoKinds(t *testing.T) {
	lit := expr.NewPrimitiveLiteral[int32](1, false)

	direction := SortFieldToProto(&expr.SortField{Expr: lit, Kind: types.SortAscNullsFirst})
	wantDirection := &proto.SortField{
		Expr:     ExprToProto(lit),
		SortKind: &proto.SortField_Direction{Direction: proto.SortField_SortDirection(types.SortAscNullsFirst)},
	}
	if diff := cmp.Diff(wantDirection, direction, protocmp.Transform()); diff != "" {
		t.Errorf("SortFieldToProto(direction) mismatch (-want +got):\n%s", diff)
	}

	comparison := SortFieldToProto(&expr.SortField{Expr: lit, Kind: types.FunctionRef(3)})
	wantComparison := &proto.SortField{
		Expr:     ExprToProto(lit),
		SortKind: &proto.SortField_ComparisonFunctionReference{ComparisonFunctionReference: 3},
	}
	if diff := cmp.Diff(wantComparison, comparison, protocmp.Transform()); diff != "" {
		t.Errorf("SortFieldToProto(comparison) mismatch (-want +got):\n%s", diff)
	}
}

// BoundToProto round-trips each bound kind through the decoder unchanged.
func TestBoundToProtoRoundTrips(t *testing.T) {
	cases := []struct {
		name  string
		bound expr.Bound
	}{
		{"preceding", expr.PrecedingBound(3)},
		{"following", expr.FollowingBound(4)},
		{"current_row", expr.CurrentRow{}},
		{"unbounded", expr.Unbounded{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := BoundToProto(tc.bound)
			if diff := cmp.Diff(BoundToProto(expr.BoundFromProto(got)), got, protocmp.Transform()); diff != "" {
				t.Errorf("BoundToProto(%s) did not round-trip (-want +got):\n%s", tc.name, diff)
			}
		})
	}
}
