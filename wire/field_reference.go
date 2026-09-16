// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"fmt"

	"github.com/substrait-io/substrait-go/v9/expr"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

// RefSegmentToProto encodes a reference segment as its protobuf message.
func RefSegmentToProto(r expr.ReferenceSegment) *proto.Expression_ReferenceSegment {
	switch r := r.(type) {
	case *expr.MapKeyRef:
		return mapKeyRefToProto(r)
	case *expr.StructFieldRef:
		return structFieldRefToProto(r)
	case *expr.ListElementRef:
		return listElementRefToProto(r)
	default:
		panic(fmt.Sprintf("wire: unhandled reference segment %T", r))
	}
}

func mapKeyRefToProto(r *expr.MapKeyRef) *proto.Expression_ReferenceSegment {
	var child *proto.Expression_ReferenceSegment
	if r.Child != nil {
		child = RefSegmentToProto(r.Child)
	}
	return &proto.Expression_ReferenceSegment{
		ReferenceType: &proto.Expression_ReferenceSegment_MapKey_{
			MapKey: &proto.Expression_ReferenceSegment_MapKey{
				MapKey: LiteralToProto(r.MapKey),
				Child:  child,
			},
		},
	}
}

func structFieldRefToProto(r *expr.StructFieldRef) *proto.Expression_ReferenceSegment {
	var child *proto.Expression_ReferenceSegment
	if r.Child != nil {
		child = RefSegmentToProto(r.Child)
	}
	return &proto.Expression_ReferenceSegment{
		ReferenceType: &proto.Expression_ReferenceSegment_StructField_{
			StructField: &proto.Expression_ReferenceSegment_StructField{
				Field: r.Field,
				Child: child,
			},
		},
	}
}

func listElementRefToProto(r *expr.ListElementRef) *proto.Expression_ReferenceSegment {
	var child *proto.Expression_ReferenceSegment
	if r.Child != nil {
		child = RefSegmentToProto(r.Child)
	}
	return &proto.Expression_ReferenceSegment{
		ReferenceType: &proto.Expression_ReferenceSegment_ListElement_{
			ListElement: &proto.Expression_ReferenceSegment_ListElement{
				Offset: r.Offset,
				Child:  child,
			},
		},
	}
}

// MaskExpressionToProto encodes a mask expression as its protobuf message.
func MaskExpressionToProto(e *expr.MaskExpression) *proto.Expression_MaskExpression {
	return &proto.Expression_MaskExpression{
		Select:                 maskStructSelectToProto(e.Select()),
		MaintainSingularStruct: e.MaintainSingularStruct(),
	}
}

func maskStructSelectToProto(m expr.MaskStructSelect) *proto.Expression_MaskExpression_StructSelect {
	items := make([]*proto.Expression_MaskExpression_StructItem, len(m))
	for i := range m {
		items[i] = maskStructItemToProto(&m[i])
	}
	return &proto.Expression_MaskExpression_StructSelect{StructItems: items}
}

func maskSelectToProto(s expr.MaskSelect) *proto.Expression_MaskExpression_Select {
	switch s := s.(type) {
	case expr.MaskStructSelect:
		return &proto.Expression_MaskExpression_Select{
			Type: &proto.Expression_MaskExpression_Select_Struct{Struct: maskStructSelectToProto(s)},
		}
	case *expr.MaskListSelect:
		return maskListSelectToProto(s)
	case *expr.MaskMapSelect:
		return maskMapSelectToProto(s)
	default:
		panic(fmt.Sprintf("wire: unhandled mask selection %T", s))
	}
}

func maskStructItemToProto(m *expr.MaskStructItem) *proto.Expression_MaskExpression_StructItem {
	var child *proto.Expression_MaskExpression_Select
	if c := m.Child(); c != nil {
		child = maskSelectToProto(c)
	}
	return &proto.Expression_MaskExpression_StructItem{
		Field: m.Field(),
		Child: child,
	}
}

func maskListSelectToProto(m *expr.MaskListSelect) *proto.Expression_MaskExpression_Select {
	sel := m.Selection()
	items := make([]*proto.Expression_MaskExpression_ListSelect_ListSelectItem, len(sel))
	for i, s := range sel {
		items[i] = maskListSelectItemToProto(s)
	}
	return &proto.Expression_MaskExpression_Select{
		Type: &proto.Expression_MaskExpression_Select_List{
			List: &proto.Expression_MaskExpression_ListSelect{
				Selection: items,
				Child:     maskSelectToProto(m.Child()),
			},
		},
	}
}

func maskListSelectItemToProto(s expr.MaskListSelectItem) *proto.Expression_MaskExpression_ListSelect_ListSelectItem {
	switch s := s.(type) {
	case *expr.MaskListElement:
		return &proto.Expression_MaskExpression_ListSelect_ListSelectItem{
			Type: &proto.Expression_MaskExpression_ListSelect_ListSelectItem_Item{
				Item: &proto.Expression_MaskExpression_ListSelect_ListSelectItem_ListElement{
					Field: s.GetField(),
				},
			},
		}
	case *expr.MaskListSlice:
		start, end := s.GetBounds()
		return &proto.Expression_MaskExpression_ListSelect_ListSelectItem{
			Type: &proto.Expression_MaskExpression_ListSelect_ListSelectItem_Slice{
				Slice: &proto.Expression_MaskExpression_ListSelect_ListSelectItem_ListSlice{
					Start: start,
					End:   end,
				},
			},
		}
	default:
		panic(fmt.Sprintf("wire: unhandled mask selection %T", s))
	}
}

func maskMapSelectToProto(m *expr.MaskMapSelect) *proto.Expression_MaskExpression_Select {
	mapSelect := &proto.Expression_MaskExpression_Select_Map{
		Map: &proto.Expression_MaskExpression_MapSelect{
			Child: maskSelectToProto(m.Child()),
		},
	}

	if m.KeyKind() == expr.MapSelectKey {
		mapSelect.Map.Select = &proto.Expression_MaskExpression_MapSelect_Key{
			Key: &proto.Expression_MaskExpression_MapSelect_MapKey{MapKey: m.Key()},
		}
	} else {
		mapSelect.Map.Select = &proto.Expression_MaskExpression_MapSelect_Expression{
			Expression: &proto.Expression_MaskExpression_MapSelect_MapKeyExpression{MapKeyExpression: m.Key()},
		}
	}

	return &proto.Expression_MaskExpression_Select{Type: mapSelect}
}

// FieldReferenceToProto encodes a field reference as its protobuf message.
func FieldReferenceToProto(f *expr.FieldReference) *proto.Expression {
	return &proto.Expression{
		RexType: &proto.Expression_Selection{Selection: fieldReferenceRefToProto(f)},
	}
}

func fieldReferenceRefToProto(f *expr.FieldReference) *proto.Expression_FieldReference {
	ret := &proto.Expression_FieldReference{}
	switch r := f.Reference.(type) {
	case expr.ReferenceSegment:
		ret.ReferenceType = &proto.Expression_FieldReference_DirectReference{
			DirectReference: RefSegmentToProto(r),
		}
	case *expr.MaskExpression:
		ret.ReferenceType = &proto.Expression_FieldReference_MaskedReference{
			MaskedReference: MaskExpressionToProto(r),
		}
	}

	if f.Root != expr.RootReference {
		switch r := f.Root.(type) {
		case expr.Expression:
			ret.RootType = &proto.Expression_FieldReference_Expression{
				Expression: ExprToProto(r),
			}
		case expr.OuterReference:
			ret.RootType = &proto.Expression_FieldReference_OuterReference_{
				OuterReference: &proto.Expression_FieldReference_OuterReference{
					StepsOut: uint32(r),
				},
			}
		case expr.LambdaParameterReference:
			ret.RootType = &proto.Expression_FieldReference_LambdaParameterReference_{
				LambdaParameterReference: &proto.Expression_FieldReference_LambdaParameterReference{
					StepsOut: r.StepsOut,
				},
			}
		}
	} else {
		ret.RootType = &proto.Expression_FieldReference_RootReference_{
			RootReference: &proto.Expression_FieldReference_RootReference{},
		}
	}

	return ret
}
