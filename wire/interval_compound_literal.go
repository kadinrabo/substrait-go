// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"github.com/substrait-io/substrait-go/v9/expr"
	"github.com/substrait-io/substrait-go/v9/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

func intervalCompoundLiteralToProto(m expr.IntervalCompoundLiteral) *proto.Expression_Literal {
	t := types.NewIntervalCompoundType().WithPrecision(m.SubSecondPrecision).WithNullability(m.Nullability)
	intrCompPB := &proto.Expression_Literal_IntervalCompound{}

	if m.Years != 0 || m.Months != 0 {
		intrCompPB.IntervalYearToMonth = &proto.Expression_Literal_IntervalYearToMonth{
			Years:  m.Years,
			Months: m.Months,
		}
	}

	if m.Days != 0 || m.Seconds != 0 || m.SubSeconds != 0 {
		intrCompPB.IntervalDayToSecond = &proto.Expression_Literal_IntervalDayToSecond{
			Days:          m.Days,
			Seconds:       m.Seconds,
			PrecisionMode: &proto.Expression_Literal_IntervalDayToSecond_Precision{Precision: m.SubSecondPrecision.ToProtoVal()},
			Subseconds:    m.SubSeconds,
		}
	}

	return &proto.Expression_Literal{
		LiteralType:            &proto.Expression_Literal_IntervalCompound_{IntervalCompound: intrCompPB},
		Nullable:               t.GetNullability() == types.NullabilityNullable,
		TypeVariationReference: t.GetTypeVariationReference(),
	}
}
