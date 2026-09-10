// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"testing"

	"github.com/substrait-io/substrait-go/v9/expr"
	"github.com/substrait-io/substrait-go/v9/types"
)

func TestIntervalYearToMonthLiteralToProto(t *testing.T) {
	roundTripLiteral(t, "interval_year_to_month", expr.IntervalYearToMonthLiteral{
		Years: 7, Months: 8, Nullability: types.NullabilityRequired})
}
