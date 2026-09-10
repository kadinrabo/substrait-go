// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"testing"

	"github.com/substrait-io/substrait-go/v9/expr"
	"github.com/substrait-io/substrait-go/v9/types"
)

func TestIntervalCompoundLiteralToProto(t *testing.T) {
	roundTripLiteral(t, "interval_compound", expr.IntervalCompoundLiteral{
		Years: 1, Months: 2, Days: 3, Seconds: 4, SubSeconds: 5,
		SubSecondPrecision: types.PrecisionMicroSeconds, Nullability: types.NullabilityRequired})
}
