package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substrait-io/substrait-go/v9/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

func TestLiteralFromProtoLiteral(t *testing.T) {
	intDayToSecVal := &proto.Expression_Literal_IntervalDayToSecond{Days: 1, Seconds: 2, PrecisionMode: &proto.Expression_Literal_IntervalDayToSecond_Precision{Precision: 5}}
	for _, tc := range []struct {
		name             string
		constructedProto *proto.Expression_Literal
		expectedLiteral  interface{}
	}{
		{"TimeType",
			&proto.Expression_Literal{LiteralType: &proto.Expression_Literal_PrecisionTime_{PrecisionTime: &proto.Expression_Literal_PrecisionTime{Precision: 4, Value: 12345678}}, Nullable: true},
			&ProtoLiteral{Value: int64(12345678), Type: types.NewPrecisionTimeType(types.PrecisionEMinus4Seconds).WithNullability(types.NullabilityNullable)},
		},
		{"TimeStampType",
			&proto.Expression_Literal{LiteralType: &proto.Expression_Literal_PrecisionTimestamp_{PrecisionTimestamp: &proto.Expression_Literal_PrecisionTimestamp{Precision: 4, Value: 12345678}}, Nullable: true},
			&ProtoLiteral{Value: int64(12345678), Type: types.NewPrecisionTimestampType(types.PrecisionEMinus4Seconds).WithNullability(types.NullabilityNullable)},
		},
		{"TimeStampTzType",
			&proto.Expression_Literal{LiteralType: &proto.Expression_Literal_PrecisionTimestampTz{PrecisionTimestampTz: &proto.Expression_Literal_PrecisionTimestamp{Precision: 9, Value: 12345678}}, Nullable: true},
			&ProtoLiteral{Value: int64(12345678), Type: types.NewPrecisionTimestampTzType(types.PrecisionNanoSeconds).WithNullability(types.NullabilityNullable)},
		},
		{"IntervalDayType",
			&proto.Expression_Literal{LiteralType: &proto.Expression_Literal_IntervalDayToSecond_{IntervalDayToSecond: intDayToSecVal}, Nullable: true},
			&ProtoLiteral{Value: &types.IntervalDayToSecond{Days: 1, Seconds: 2, Precision: types.PrecisionEMinus5Seconds}, Type: &types.IntervalDayType{Precision: types.PrecisionEMinus5Seconds, Nullability: types.NullabilityNullable}},
		},
		{"IntervalYearToMonthType",
			&proto.Expression_Literal{LiteralType: &proto.Expression_Literal_IntervalYearToMonth_{IntervalYearToMonth: &proto.Expression_Literal_IntervalYearToMonth{Years: 1234, Months: 5}}, Nullable: true},
			IntervalYearToMonthLiteral{Years: 1234, Months: 5, Nullability: types.NullabilityNullable},
		},
		{"IntervalCompoundType",
			&proto.Expression_Literal{LiteralType: &proto.Expression_Literal_IntervalCompound_{
				IntervalCompound: &proto.Expression_Literal_IntervalCompound{
					IntervalYearToMonth: &proto.Expression_Literal_IntervalYearToMonth{Years: 1234, Months: -5},
					IntervalDayToSecond: &proto.Expression_Literal_IntervalDayToSecond{Days: 6, Seconds: -7,
						PrecisionMode: &proto.Expression_Literal_IntervalDayToSecond_Precision{Precision: 8},
						Subseconds:    -9,
					},
				}}, Nullable: true},
			IntervalCompoundLiteral{Years: 1234, Months: -5, Days: 6, Seconds: -7, SubSecondPrecision: 8, SubSeconds: -9, Nullability: types.NullabilityNullable},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			literal := LiteralFromProto(tc.constructedProto)
			assert.Equal(t, tc.expectedLiteral, literal)
		})

	}
}
