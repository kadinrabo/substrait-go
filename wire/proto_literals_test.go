// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"testing"

	"github.com/substrait-io/substrait-go/v9/expr"
	"github.com/substrait-io/substrait-go/v9/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

// ProtoLiteral carries an already-encoded value; LiteralToProto must round-trip
// each kind the dispatcher routes through it.
func TestProtoLiteralToProto(t *testing.T) {
	req := types.NullabilityRequired
	prim := expr.NewPrimitiveLiteral[int32](1, false)

	cases := []struct {
		name string
		lit  expr.Literal
	}{
		{"proto_varchar", &expr.ProtoLiteral{Value: "vc", Type: &types.VarCharType{Length: 10, Nullability: req}}},
		{"proto_decimal", &expr.ProtoLiteral{Value: make([]byte, 16), Type: &types.DecimalType{Precision: 10, Scale: 2, Nullability: req}}},
		{"proto_precision_time", &expr.ProtoLiteral{Value: int64(123), Type: types.NewPrecisionTimeType(types.PrecisionEMinus4Seconds).WithNullability(req)}},
		{"proto_precision_timestamp", &expr.ProtoLiteral{Value: int64(456), Type: types.NewPrecisionTimestampType(types.PrecisionMilliSeconds).WithNullability(req)}},
		{"proto_precision_timestamp_tz", &expr.ProtoLiteral{Value: int64(789), Type: types.NewPrecisionTimestampTzType(types.PrecisionNanoSeconds).WithNullability(req)}},
		{"proto_interval_year", &expr.ProtoLiteral{Value: &types.IntervalYearToMonth{Years: 3, Months: 4}, Type: &types.IntervalYearType{Nullability: req}}},
		{"proto_interval_day", &expr.ProtoLiteral{Value: &types.IntervalDayToSecond{Days: 1, Seconds: 2, PrecisionMode: &proto.Expression_Literal_IntervalDayToSecond_Precision{Precision: 5}}, Type: &types.IntervalDayType{Precision: types.PrecisionEMinus5Seconds, Nullability: req}}},
		{"proto_user_defined", &expr.ProtoLiteral{
			Value: &proto.Expression_Literal_UserDefined_Struct{Struct: &proto.Expression_Literal_Struct{Fields: []*proto.Expression_Literal{LiteralToProto(prim)}}},
			Type:  &types.UserDefinedType{Nullability: req, TypeReference: 3, TypeParameters: []types.TypeParam{types.IntegerParameter(9)}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { roundTripLiteral(t, tc.name, tc.lit) })
	}
}
