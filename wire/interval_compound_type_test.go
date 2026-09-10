// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"testing"

	"github.com/substrait-io/substrait-go/v9/types"
)

func TestIntervalCompoundTypeToProto(t *testing.T) {
	compound := types.NewIntervalCompoundType().
		WithPrecision(types.PrecisionMicroSeconds).
		WithTypeVariationRef(tvr).
		WithNullability(types.NullabilityRequired)
	roundTripType(t, "interval_compound", compound)
}
