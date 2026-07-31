// SPDX-License-Identifier: Apache-2.0

package extensions_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substrait-io/substrait-go/v8/extensions"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
	extensionspb "github.com/substrait-io/substrait-protobuf/go/substraitpb/extensions"
)

const sortSampleURN = "extension:test:sample"

func typeDecl(anchor uint32, name string) *extensionspb.SimpleExtensionDeclaration {
	return &extensionspb.SimpleExtensionDeclaration{
		MappingType: &extensionspb.SimpleExtensionDeclaration_ExtensionType_{
			ExtensionType: &extensionspb.SimpleExtensionDeclaration_ExtensionType{
				ExtensionUrnReference: 1,
				TypeAnchor:            anchor,
				Name:                  name,
			},
		},
	}
}

func typeVariationDecl(anchor uint32, name string) *extensionspb.SimpleExtensionDeclaration {
	return &extensionspb.SimpleExtensionDeclaration{
		MappingType: &extensionspb.SimpleExtensionDeclaration_ExtensionTypeVariation_{
			ExtensionTypeVariation: &extensionspb.SimpleExtensionDeclaration_ExtensionTypeVariation{
				ExtensionUrnReference: 1,
				TypeVariationAnchor:   anchor,
				Name:                  name,
			},
		},
	}
}

func functionDecl(anchor uint32, name string) *extensionspb.SimpleExtensionDeclaration {
	return &extensionspb.SimpleExtensionDeclaration{
		MappingType: &extensionspb.SimpleExtensionDeclaration_ExtensionFunction_{
			ExtensionFunction: &extensionspb.SimpleExtensionDeclaration_ExtensionFunction{
				ExtensionUrnReference: 1,
				FunctionAnchor:        anchor,
				Name:                  name,
			},
		},
	}
}

// newSortTestSet round-trips the declarations through GetExtensionSet so that
// the anchors are the ones given here rather than sequentially assigned.
func newSortTestSet(t *testing.T, decls ...*extensionspb.SimpleExtensionDeclaration) (extensions.Set, *extensions.Collection) {
	t.Helper()

	c := &extensions.Collection{}
	require.NoError(t, c.Load(strings.NewReader(sampleYAML)))

	plan := &proto.Plan{
		ExtensionUrns: []*extensionspb.SimpleExtensionURN{
			{ExtensionUrnAnchor: 1, Urn: sortSampleURN},
		},
		Extensions: decls,
	}

	extSet, err := extensions.GetExtensionSet(plan, c)
	require.NoError(t, err)
	return extSet, c
}

// A type declaration followed by type variation declarations used to make
// ToProto dereference a nil ExtensionTypeVariation, because the type variation
// sub-sort indexed the whole declaration slice with offsets valid only for the
// type variation sub-slice.
func TestSetToProtoWithTypesAndTypeVariationsDoesNotPanic(t *testing.T) {
	extSet, c := newSortTestSet(t,
		typeDecl(1, "point"),
		typeVariationDecl(1, "var_a"),
		typeVariationDecl(2, "var_b"),
	)

	require.NotPanics(t, func() {
		_, decls := extSet.ToProto(c)
		assert.Len(t, decls, 3)
	})
}

// Type variations must come out ordered by anchor. A preceding type
// declaration shifts the sub-slice, which is the case the old indexing got
// wrong.
func TestSetToProtoSortsTypeVariationsByAnchor(t *testing.T) {
	decls := []*extensionspb.SimpleExtensionDeclaration{typeDecl(1, "point")}
	const nVariations = 8
	for i := nVariations; i >= 1; i-- {
		decls = append(decls, typeVariationDecl(uint32(i), fmt.Sprintf("var_%d", i)))
	}

	extSet, c := newSortTestSet(t, decls...)
	_, got := extSet.ToProto(c)

	require.Len(t, got, nVariations+1)
	anchors := make([]uint32, 0, nVariations)
	for _, d := range got[1:] {
		tv := d.GetExtensionTypeVariation()
		require.NotNil(t, tv, "expected a type variation declaration after the type declarations")
		anchors = append(anchors, tv.TypeVariationAnchor)
	}
	assert.Equal(t, []uint32{1, 2, 3, 4, 5, 6, 7, 8}, anchors)
}

// Functions must come out ordered by anchor. There is deliberately no type
// declaration here so that the type variation sub-sort is a no-op and this
// test isolates the function sub-sort. The single type variation is what
// shifts the function sub-slice.
func TestSetToProtoSortsFunctionsByAnchor(t *testing.T) {
	decls := []*extensionspb.SimpleExtensionDeclaration{typeVariationDecl(1, "var_a")}
	const nFunctions = 8
	for i := nFunctions; i >= 1; i-- {
		decls = append(decls, functionDecl(uint32(i), fmt.Sprintf("func_%d", i)))
	}

	extSet, c := newSortTestSet(t, decls...)
	_, got := extSet.ToProto(c)

	require.Len(t, got, nFunctions+1)
	anchors := make([]uint32, 0, nFunctions)
	for _, d := range got[1:] {
		fn := d.GetExtensionFunction()
		require.NotNil(t, fn, "expected a function declaration after the type variation declarations")
		anchors = append(anchors, fn.FunctionAnchor)
	}
	assert.Equal(t, []uint32{1, 2, 3, 4, 5, 6, 7, 8}, anchors)
}
