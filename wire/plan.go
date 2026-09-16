// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"github.com/substrait-io/substrait-go/v9/plan"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

// PlanToProto encodes a plan as its protobuf message.
func PlanToProto(p *plan.Plan) (*proto.Plan, error) {
	reg := p.ExtensionRegistry()
	urns, decls := reg.ExtensionsToProto()

	rels := p.Relations()
	relations := make([]*proto.PlanRel, len(rels))
	for i := range rels {
		relations[i] = relationToProto(&rels[i])
	}

	var bindings []*proto.DynamicParameterBinding
	if bs := p.ParameterBindings(); len(bs) > 0 {
		bindings = make([]*proto.DynamicParameterBinding, len(bs))
		for i, b := range bs {
			bindings[i] = &proto.DynamicParameterBinding{
				ParameterAnchor: b.ParameterAnchor,
				Value:           LiteralToProto(b.Value),
			}
		}
	}

	return &proto.Plan{
		Version:            VersionToProto(p.Version()),
		ExpectedTypeUrls:   p.ExpectedTypeURLs(),
		AdvancedExtensions: p.GetAdvancedExtension(),
		Relations:          relations,
		Extensions:         decls,
		ExtensionUrns:      urns,
		ParameterBindings:  bindings,
	}, nil
}
