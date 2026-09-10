// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"testing"

	"github.com/substrait-io/substrait-go/v9/types"
)

func TestIntervalDayTypeToProto(t *testing.T) {
	roundTripType(t, "interval_day", &types.IntervalDayType{
		Precision: types.PrecisionNanoSeconds, Nullability: types.NullabilityRequired, TypeVariationRef: tvr})
}
