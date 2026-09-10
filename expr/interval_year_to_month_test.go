package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substrait-io/substrait-go/v9/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

func TestIntervalYearToMonthFromProto(t *testing.T) {
	nullable := true
	nullability := types.NullabilityNullable
	var oneYear int32 = 1
	var oneMonth int32 = 1
	for _, tc := range []struct {
		name            string
		inputProto      *proto.Expression_Literal
		expectedLiteral IntervalYearToMonthLiteral
	}{
		{"OnlyYearToMonth",
			&proto.Expression_Literal{
				LiteralType: &proto.Expression_Literal_IntervalYearToMonth_{
					IntervalYearToMonth: &proto.Expression_Literal_IntervalYearToMonth{Years: oneYear, Months: oneMonth}},
				Nullable: nullable},
			IntervalYearToMonthLiteral{Years: oneYear, Months: oneMonth, Nullability: nullability},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gotLiteral := intervalYearToMonthLiteralFromProto(tc.inputProto)
			assert.Equal(t, tc.expectedLiteral, gotLiteral)
			// verify equal method too returns true
			assert.True(t, tc.expectedLiteral.Equals(gotLiteral))
			assert.True(t, gotLiteral.IsScalar())
			// got literal after serialization is different from empty literal
			assert.False(t, IntervalYearToMonthLiteral{}.Equals(gotLiteral))
		})

	}
}
