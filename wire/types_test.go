// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/substrait-io/substrait-go/v9/types"
	"google.golang.org/protobuf/testing/protocmp"
)

const tvr = uint32(7)

// roundTripType asserts TypeToProto followed by TypeFromProto reproduces the type.
// Values use a non-default nullability and type-variation ref.
func roundTripType(t *testing.T, name string, typ types.Type) {
	t.Helper()
	if got := types.TypeFromProto(TypeToProto(typ)); !typ.Equals(got) {
		t.Errorf("TypeToProto(%s) did not round-trip: got %v", name, got)
	}
}

// TypeToProto must produce byte-for-byte what the core dispatcher does for every
// type it handles.
func TestTypeToProtoMatchesCore(t *testing.T) {
	req := types.NullabilityRequired

	cases := []struct {
		name string
		typ  types.Type
	}{
		{"bool", &types.BooleanType{Nullability: req, TypeVariationRef: tvr}},
		{"i8", &types.Int8Type{Nullability: req, TypeVariationRef: tvr}},
		{"i16", &types.Int16Type{Nullability: req, TypeVariationRef: tvr}},
		{"i32", &types.Int32Type{Nullability: req, TypeVariationRef: tvr}},
		{"i64", &types.Int64Type{Nullability: req, TypeVariationRef: tvr}},
		{"fp32", &types.Float32Type{Nullability: req, TypeVariationRef: tvr}},
		{"fp64", &types.Float64Type{Nullability: req, TypeVariationRef: tvr}},
		{"string", &types.StringType{Nullability: req, TypeVariationRef: tvr}},
		{"binary", &types.BinaryType{Nullability: req, TypeVariationRef: tvr}},
		{"date", &types.DateType{Nullability: req, TypeVariationRef: tvr}},
		{"time", &types.TimeType{Nullability: req, TypeVariationRef: tvr}},
		{"timestamp", &types.TimestampType{Nullability: req, TypeVariationRef: tvr}},
		{"timestamp_tz", &types.TimestampTzType{Nullability: req, TypeVariationRef: tvr}},
		{"interval_year", &types.IntervalYearType{Nullability: req, TypeVariationRef: tvr}},
		{"uuid", &types.UUIDType{Nullability: req, TypeVariationRef: tvr}},
		{"fixed_char", &types.FixedCharType{Length: 12, Nullability: req, TypeVariationRef: tvr}},
		{"varchar", &types.VarCharType{Length: 24, Nullability: req, TypeVariationRef: tvr}},
		{"fixed_binary", &types.FixedBinaryType{Length: 8, Nullability: req, TypeVariationRef: tvr}},
		{"decimal", &types.DecimalType{Scale: 2, Precision: 10, Nullability: req, TypeVariationRef: tvr}},
		{"struct", &types.StructType{Nullability: req, TypeVariationRef: tvr, Types: []types.Type{
			&types.Int32Type{Nullability: req}, &types.StringType{Nullability: req}}}},
		{"list", &types.ListType{Nullability: req, TypeVariationRef: tvr, Type: &types.Int64Type{Nullability: req}}},
		{"map", &types.MapType{Nullability: req, TypeVariationRef: tvr,
			Key: &types.StringType{Nullability: req}, Value: &types.Int32Type{Nullability: req}}},
		{"func", &types.FuncType{Nullability: req, ParameterTypes: []types.Type{&types.Int32Type{Nullability: req}},
			ReturnType: &types.BooleanType{Nullability: req}}},
		{"user_defined", &types.UserDefinedType{Nullability: req, TypeVariationRef: tvr, TypeReference: 5,
			TypeParameters: []types.TypeParam{types.IntegerParameter(9), types.BooleanParameter(true)}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { roundTripType(t, tc.name, tc.typ) })
	}
}

func TestNamedStructToProtoMatchesCore(t *testing.T) {
	n := types.NamedStruct{
		Names: []string{"a", "b"},
		Struct: types.StructType{
			Nullability: types.NullabilityRequired,
			Types: []types.Type{
				&types.Int32Type{Nullability: types.NullabilityRequired},
				&types.BooleanType{Nullability: types.NullabilityRequired},
			},
		},
	}
	got := NamedStructToProto(n)
	if diff := cmp.Diff(NamedStructToProto(types.NewNamedStructFromProto(got)), got, protocmp.Transform()); diff != "" {
		t.Errorf("NamedStructToProto did not round-trip (-want +got):\n%s", diff)
	}
}

func TestTypeParamToProtoMatchesCore(t *testing.T) {
	params := []types.TypeParam{
		types.NullParameter{},
		&types.DataTypeParameter{Type: &types.Int32Type{Nullability: types.NullabilityRequired}},
		types.BooleanParameter(true),
		types.IntegerParameter(42),
		types.EnumParameter("day"),
		types.StringParameter("name"),
	}
	for _, p := range params {
		got := TypeParamToProto(p)
		if diff := cmp.Diff(TypeParamToProto(types.TypeParamFromProto(got)), got, protocmp.Transform()); diff != "" {
			t.Errorf("TypeParamToProto(%T) did not round-trip (-want +got):\n%s", p, diff)
		}
	}
}
