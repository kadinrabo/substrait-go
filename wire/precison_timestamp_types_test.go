// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"testing"

	"github.com/substrait-io/substrait-go/v9/types"
)

func TestPrecisionTimestampTypesToProto(t *testing.T) {
	req := types.NullabilityRequired
	roundTripType(t, "precision_time", &types.PrecisionTimeType{
		Precision: types.PrecisionMicroSeconds, Nullability: req, TypeVariationRef: tvr})
	roundTripType(t, "precision_timestamp", &types.PrecisionTimestampType{
		Precision: types.PrecisionMilliSeconds, Nullability: req, TypeVariationRef: tvr})
	roundTripType(t, "precision_timestamp_tz", &types.PrecisionTimestampTzType{PrecisionTimestampType: types.PrecisionTimestampType{
		Precision: types.PrecisionSeconds, Nullability: req, TypeVariationRef: tvr}})
}
