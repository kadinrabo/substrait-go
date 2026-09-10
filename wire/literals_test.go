// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/substrait-io/substrait-go/v9/expr"
	"github.com/substrait-io/substrait-go/v9/types"
	"google.golang.org/protobuf/testing/protocmp"
)

// roundTripLiteral asserts LiteralToProto followed by LiteralFromProto and a
// re-encode reproduces the same protobuf message.
func roundTripLiteral(t *testing.T, name string, lit expr.Literal) {
	t.Helper()
	got := LiteralToProto(lit)
	if diff := cmp.Diff(LiteralToProto(expr.LiteralFromProto(got)), got, protocmp.Transform()); diff != "" {
		t.Errorf("LiteralToProto(%s) did not round-trip (-want +got):\n%s", name, diff)
	}
}

// LiteralToProto must produce byte-for-byte what each literal's own encoding
// does. The corpus covers every literal the dispatcher routes.
func TestLiteralToProtoMatchesCore(t *testing.T) {
	req := types.NullabilityRequired
	prim := expr.NewPrimitiveLiteral[int32](1, false)

	cases := []struct {
		name string
		lit  expr.Literal
	}{
		{"bool", &expr.PrimitiveLiteral[bool]{Value: true, Type: &types.BooleanType{Nullability: req, TypeVariationRef: 4}}},
		{"i8", &expr.PrimitiveLiteral[int8]{Value: -8, Type: &types.Int8Type{Nullability: req}}},
		{"i16", &expr.PrimitiveLiteral[int16]{Value: -16, Type: &types.Int16Type{Nullability: req}}},
		{"i32", &expr.PrimitiveLiteral[int32]{Value: -32, Type: &types.Int32Type{Nullability: req}}},
		{"i64", &expr.PrimitiveLiteral[int64]{Value: -64, Type: &types.Int64Type{Nullability: req}}},
		{"fp32", &expr.PrimitiveLiteral[float32]{Value: 1.5, Type: &types.Float32Type{Nullability: req}}},
		{"fp64", &expr.PrimitiveLiteral[float64]{Value: 2.5, Type: &types.Float64Type{Nullability: req}}},
		{"string", &expr.PrimitiveLiteral[string]{Value: "hi", Type: &types.StringType{Nullability: req}}},
		{"timestamp", &expr.PrimitiveLiteral[types.Timestamp]{Value: 111, Type: &types.TimestampType{Nullability: req}}},
		{"date", &expr.PrimitiveLiteral[types.Date]{Value: 222, Type: &types.DateType{Nullability: req}}},
		{"time", &expr.PrimitiveLiteral[types.Time]{Value: 333, Type: &types.TimeType{Nullability: req}}},
		{"fixed_char", &expr.PrimitiveLiteral[types.FixedChar]{Value: "abc", Type: &types.FixedCharType{Length: 3, Nullability: req}}},
		{"timestamp_tz", &expr.PrimitiveLiteral[types.TimestampTz]{Value: 444, Type: &types.TimestampTzType{Nullability: req}}},
		{"null", expr.NewNullLiteral(&types.Int32Type{Nullability: types.NullabilityNullable, TypeVariationRef: 2})},
		{"nested_struct", &expr.NestedLiteral[expr.StructLiteralValue]{
			Value: expr.StructLiteralValue{prim, expr.NewPrimitiveLiteral("x", false)},
			Type:  &types.StructType{Nullability: req, Types: []types.Type{&types.Int32Type{Nullability: req}, &types.StringType{Nullability: req}}}}},
		{"nested_list", &expr.NestedLiteral[expr.ListLiteralValue]{
			Value: expr.ListLiteralValue{prim},
			Type:  &types.ListType{Nullability: req, Type: &types.Int32Type{Nullability: req}}}},
		{"nested_list_empty", &expr.NestedLiteral[expr.ListLiteralValue]{
			Value: expr.ListLiteralValue{},
			Type:  &types.ListType{Nullability: req, Type: &types.Int32Type{Nullability: req}}}},
		{"map", &expr.MapLiteral{
			Value: expr.MapLiteralValue{{Key: expr.NewPrimitiveLiteral("k", false), Value: prim}},
			Type:  &types.MapType{Nullability: req, Key: &types.StringType{Nullability: req}, Value: &types.Int32Type{Nullability: req}}}},
		{"map_empty", &expr.MapLiteral{
			Value: expr.MapLiteralValue{},
			Type:  &types.MapType{Nullability: req, Key: &types.StringType{Nullability: req}, Value: &types.Int32Type{Nullability: req}}}},
		{"binary", &expr.ByteSliceLiteral[[]byte]{Value: []byte{1, 2, 3}, Type: &types.BinaryType{Nullability: req}}},
		{"fixed_binary", &expr.ByteSliceLiteral[types.FixedBinary]{Value: types.FixedBinary{4, 5}, Type: &types.FixedBinaryType{Length: 2, Nullability: req}}},
		{"uuid", &expr.ByteSliceLiteral[types.UUID]{Value: types.UUID(make([]byte, 16)), Type: &types.UUIDType{Nullability: req}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { roundTripLiteral(t, tc.name, tc.lit) })
	}
}
