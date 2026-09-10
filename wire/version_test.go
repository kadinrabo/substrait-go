// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"testing"

	"github.com/substrait-io/substrait-go/v9/types"
)

func TestVersionToProtoMatchesCore(t *testing.T) {
	v := types.Version{MajorNumber: 1, MinorNumber: 2, PatchNumber: 3, GitHash: "abc", Producer: "prod"}
	if got := types.VersionFromProto(VersionToProto(v)); got != v {
		t.Errorf("VersionToProto did not round-trip: got %v want %v", got, v)
	}
}
