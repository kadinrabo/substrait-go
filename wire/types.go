// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"github.com/substrait-io/substrait-go/v9/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

// VersionToProto encodes a version as its protobuf message.
func VersionToProto(v types.Version) *proto.Version {
	return &proto.Version{
		MajorNumber: v.MajorNumber,
		MinorNumber: v.MinorNumber,
		PatchNumber: v.PatchNumber,
		GitHash:     v.GitHash,
		Producer:    v.Producer,
	}
}

// NamedStructToProto encodes a named struct as its protobuf message.
func NamedStructToProto(n types.NamedStruct) *proto.NamedStruct {
	return &proto.NamedStruct{
		Names:  n.Names,
		Struct: structTypeToProto(&n.Struct).GetStruct(),
	}
}

// TypeToProto constructs the protobuf message for the given type.
func TypeToProto(t types.Type) *proto.Type {
	switch t := t.(type) {
	case *types.BooleanType:
		return &proto.Type{Kind: &proto.Type_Bool{
			Bool: &proto.Type_Boolean{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.Int8Type:
		return &proto.Type{Kind: &proto.Type_I8_{
			I8: &proto.Type_I8{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.Int16Type:
		return &proto.Type{Kind: &proto.Type_I16_{
			I16: &proto.Type_I16{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.Int32Type:
		return &proto.Type{Kind: &proto.Type_I32_{
			I32: &proto.Type_I32{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.Int64Type:
		return &proto.Type{Kind: &proto.Type_I64_{
			I64: &proto.Type_I64{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.Float32Type:
		return &proto.Type{Kind: &proto.Type_Fp32{
			Fp32: &proto.Type_FP32{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.Float64Type:
		return &proto.Type{Kind: &proto.Type_Fp64{
			Fp64: &proto.Type_FP64{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.StringType:
		return &proto.Type{Kind: &proto.Type_String_{
			String_: &proto.Type_String{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.BinaryType:
		return &proto.Type{Kind: &proto.Type_Binary_{
			Binary: &proto.Type_Binary{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.DateType:
		return &proto.Type{Kind: &proto.Type_Date_{
			Date: &proto.Type_Date{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.TimeType:
		return &proto.Type{Kind: &proto.Type_Time_{
			Time: &proto.Type_Time{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.TimestampTzType:
		return &proto.Type{Kind: &proto.Type_TimestampTz{
			TimestampTz: &proto.Type_TimestampTZ{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.TimestampType:
		return &proto.Type{Kind: &proto.Type_Timestamp_{
			Timestamp: &proto.Type_Timestamp{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.IntervalYearType:
		return &proto.Type{Kind: &proto.Type_IntervalYear_{
			IntervalYear: &proto.Type_IntervalYear{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.UUIDType:
		return &proto.Type{Kind: &proto.Type_Uuid{
			Uuid: &proto.Type_UUID{
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.FixedCharType:
		return &proto.Type{Kind: &proto.Type_FixedChar_{
			FixedChar: &proto.Type_FixedChar{
				Length:                 t.Length,
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.VarCharType:
		return &proto.Type{Kind: &proto.Type_Varchar{
			Varchar: &proto.Type_VarChar{
				Length:                 t.Length,
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.FixedBinaryType:
		return &proto.Type{Kind: &proto.Type_FixedBinary_{
			FixedBinary: &proto.Type_FixedBinary{
				Length:                 t.Length,
				Nullability:            proto.Type_Nullability(t.Nullability),
				TypeVariationReference: t.TypeVariationRef}}}
	case *types.DecimalType:
		return decimalTypeToProto(t)
	case *types.StructType:
		return structTypeToProto(t)
	case *types.FuncType:
		return funcTypeToProto(t)
	case *types.ListType:
		return listTypeToProto(t)
	case *types.MapType:
		return mapTypeToProto(t)
	}
	panic("unimplemented type")
}

func decimalTypeToProto(s *types.DecimalType) *proto.Type {
	return &proto.Type{Kind: &proto.Type_Decimal_{
		Decimal: &proto.Type_Decimal{
			Scale: s.Scale, Precision: s.Precision,
			Nullability:            proto.Type_Nullability(s.Nullability),
			TypeVariationReference: s.TypeVariationRef}}}
}

func structTypeToProto(t *types.StructType) *proto.Type {
	children := make([]*proto.Type, len(t.Types))
	for i, c := range t.Types {
		children[i] = TypeToProto(c)
	}

	return &proto.Type{Kind: &proto.Type_Struct_{
		Struct: &proto.Type_Struct{Types: children,
			TypeVariationReference: t.TypeVariationRef,
			Nullability:            proto.Type_Nullability(t.Nullability)}}}
}

func funcTypeToProto(f *types.FuncType) *proto.Type {
	params := make([]*proto.Type, len(f.ParameterTypes))
	for i, p := range f.ParameterTypes {
		params[i] = TypeToProto(p)
	}

	return &proto.Type{Kind: &proto.Type_Func_{
		Func: &proto.Type_Func{
			ParameterTypes: params,
			ReturnType:     TypeToProto(f.ReturnType),
			Nullability:    proto.Type_Nullability(f.Nullability),
		}}}
}

func listTypeToProto(t *types.ListType) *proto.Type {
	return &proto.Type{Kind: &proto.Type_List_{
		List: &proto.Type_List{Nullability: proto.Type_Nullability(t.Nullability),
			Type:                   TypeToProto(t.Type),
			TypeVariationReference: t.TypeVariationRef}}}
}

func mapTypeToProto(t *types.MapType) *proto.Type {
	return &proto.Type{Kind: &proto.Type_Map_{
		Map: &proto.Type_Map{Nullability: proto.Type_Nullability(t.Nullability),
			TypeVariationReference: t.TypeVariationRef,
			Key:                    TypeToProto(t.Key),
			Value:                  TypeToProto(t.Value)}}}
}
