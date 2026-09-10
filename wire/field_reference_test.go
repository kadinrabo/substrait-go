// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/substrait-io/substrait-go/v9/expr"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
	"google.golang.org/protobuf/testing/protocmp"
)

// RefSegmentToProto must match each segment's own encoding, including nesting
// through a child of every segment kind and a map-key literal.
func TestRefSegmentToProtoMatchesCore(t *testing.T) {
	seg := &expr.StructFieldRef{
		Field: 2,
		Child: &expr.ListElementRef{
			Offset: 5,
			Child:  &expr.MapKeyRef{MapKey: expr.NewPrimitiveLiteral("k", false)},
		},
	}
	got := RefSegmentToProto(seg)
	if diff := cmp.Diff(RefSegmentToProto(expr.RefSegmentFromProto(got)), got, protocmp.Transform()); diff != "" {
		t.Errorf("RefSegmentToProto did not round-trip (-want +got):\n%s", diff)
	}
}

// leafSelect is a minimal struct select used as a required child.
func leafSelect() *proto.Expression_MaskExpression_Select {
	return &proto.Expression_MaskExpression_Select{
		Type: &proto.Expression_MaskExpression_Select_Struct{
			Struct: &proto.Expression_MaskExpression_StructSelect{
				StructItems: []*proto.Expression_MaskExpression_StructItem{{Field: 0}},
			},
		},
	}
}

// MaskExpressionToProto must match the mask's own encoding across struct, list,
// and both map select kinds, plus list element and slice items.
func TestMaskExpressionToProtoMatchesCore(t *testing.T) {
	p := &proto.Expression_MaskExpression{
		MaintainSingularStruct: true,
		Select: &proto.Expression_MaskExpression_StructSelect{
			StructItems: []*proto.Expression_MaskExpression_StructItem{
				{Field: 0},
				{Field: 1, Child: leafSelect()},
				{Field: 2, Child: &proto.Expression_MaskExpression_Select{
					Type: &proto.Expression_MaskExpression_Select_List{
						List: &proto.Expression_MaskExpression_ListSelect{
							Selection: []*proto.Expression_MaskExpression_ListSelect_ListSelectItem{
								{Type: &proto.Expression_MaskExpression_ListSelect_ListSelectItem_Item{
									Item: &proto.Expression_MaskExpression_ListSelect_ListSelectItem_ListElement{Field: 7}}},
								{Type: &proto.Expression_MaskExpression_ListSelect_ListSelectItem_Slice{
									Slice: &proto.Expression_MaskExpression_ListSelect_ListSelectItem_ListSlice{Start: 1, End: 5}}},
							},
							Child: leafSelect(),
						},
					},
				}},
				{Field: 3, Child: &proto.Expression_MaskExpression_Select{
					Type: &proto.Expression_MaskExpression_Select_Map{
						Map: &proto.Expression_MaskExpression_MapSelect{
							Child:  leafSelect(),
							Select: &proto.Expression_MaskExpression_MapSelect_Key{Key: &proto.Expression_MaskExpression_MapSelect_MapKey{MapKey: "k"}},
						},
					},
				}},
				{Field: 4, Child: &proto.Expression_MaskExpression_Select{
					Type: &proto.Expression_MaskExpression_Select_Map{
						Map: &proto.Expression_MaskExpression_MapSelect{
							Child:  leafSelect(),
							Select: &proto.Expression_MaskExpression_MapSelect_Expression{Expression: &proto.Expression_MaskExpression_MapSelect_MapKeyExpression{MapKeyExpression: "e"}},
						},
					},
				}},
			},
		},
	}

	m := expr.MaskExpressionFromProto(p)
	got := MaskExpressionToProto(m)
	if diff := cmp.Diff(MaskExpressionToProto(expr.MaskExpressionFromProto(got)), got, protocmp.Transform()); diff != "" {
		t.Errorf("MaskExpressionToProto did not round-trip (-want +got):\n%s", diff)
	}
}
