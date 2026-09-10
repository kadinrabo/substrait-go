// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"github.com/substrait-io/substrait-go/v9/plan"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

func relCommonToProto(rc *plan.RelCommon) *proto.RelCommon {
	ret := &proto.RelCommon{
		Hint:              rc.Hint(),
		AdvancedExtension: rc.GetAdvancedExtension(),
	}
	if mapping := rc.OutputMapping(); mapping == nil {
		ret.EmitKind = &proto.RelCommon_Direct_{Direct: &proto.RelCommon_Direct{}}
	} else {
		ret.EmitKind = &proto.RelCommon_Emit_{Emit: &proto.RelCommon_Emit{OutputMapping: mapping}}
	}
	return ret
}
