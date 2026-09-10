// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"github.com/substrait-io/substrait-go/v9/expr"
	ext "github.com/substrait-io/substrait-go/v9/extensions"
	"github.com/substrait-io/substrait-go/v9/plan"
	"github.com/substrait-io/substrait-go/v9/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
)

var relSchema = types.NamedStruct{
	Names: []string{"a", "b"},
	Struct: types.StructType{
		Nullability: types.NullabilityRequired,
		Types: []types.Type{
			&types.Int32Type{Nullability: types.NullabilityRequired},
			&types.BooleanType{Nullability: types.NullabilityRequired},
		},
	},
}

type stubExtDef struct{ detail *anypb.Any }

func (d *stubExtDef) Schema(inputs []plan.Rel) types.RecordType {
	return *types.NewRecordTypeFromStruct(relSchema.Struct)
}
func (d *stubExtDef) Build(_ []plan.Rel) *anypb.Any              { return d.detail }
func (d *stubExtDef) Expressions(_ []plan.Rel) []expr.Expression { return nil }

func relTestReg() expr.ExtensionRegistry {
	return expr.NewEmptyExtensionRegistry(ext.GetDefaultCollectionWithNoError())
}

// assertRel checks that a relation encodes to protobuf and survives a decode and
// re-encode unchanged, decoding against reg (which must know any functions the
// relation references).
func assertRelReg(t *testing.T, name string, r plan.Rel, reg expr.ExtensionRegistry) {
	t.Helper()
	got := RelToProto(r)
	rt, err := plan.RelFromProto(got, reg)
	require.NoError(t, err, name)
	if diff := cmp.Diff(RelToProto(rt), got, protocmp.Transform()); diff != "" {
		t.Errorf("RelToProto(%s) did not round-trip (-want +got):\n%s", name, diff)
	}
}

func assertRel(t *testing.T, name string, r plan.Rel) {
	t.Helper()
	assertRelReg(t, name, r, relTestReg())
}

// RelToProto must match each relation's own encoding. Rels are built through the
// public builder so the corpus mirrors real usage.
func TestRelToProtoAllRels(t *testing.T) {
	b := plan.NewBuilderDefault()
	scan := b.NamedScan([]string{"t"}, relSchema)
	scan2 := b.NamedScan([]string{"t2"}, relSchema)

	boolRef, err := b.RootFieldRef(scan, 1)
	require.NoError(t, err)
	intRef, err := b.RootFieldRef(scan, 0)
	require.NoError(t, err)

	filter, err := b.Filter(scan, boolRef)
	require.NoError(t, err)
	assertRel(t, "named_scan", scan)
	assertRel(t, "filter", filter)

	fetch, err := b.Fetch(scan, 2, 5)
	require.NoError(t, err)
	assertRel(t, "fetch", fetch)

	project, err := b.Project(scan, intRef)
	require.NoError(t, err)
	assertRel(t, "project", project)

	sorts, err := b.SortFields(scan, 0)
	require.NoError(t, err)
	sort, err := b.Sort(scan, sorts...)
	require.NoError(t, err)
	assertRel(t, "sort", sort)

	set, err := b.Set(plan.SetOpUnionAll, scan, scan2)
	require.NoError(t, err)
	assertRel(t, "set", set)

	cross, err := b.Cross(scan, scan2)
	require.NoError(t, err)
	assertRel(t, "cross", cross)

	joinCond, err := b.RootFieldRef(scan, 1)
	require.NoError(t, err)
	join, err := b.Join(scan, scan2, joinCond, plan.JoinTypeInner)
	require.NoError(t, err)
	assertRel(t, "join_no_filter", join)

	postFilter, err := b.RootFieldRef(scan, 1)
	require.NoError(t, err)
	joinCond2, err := b.RootFieldRef(scan, 1)
	require.NoError(t, err)
	joinFiltered, err := b.JoinAndFilter(scan, scan2, joinCond2, postFilter, plan.JoinTypeInner)
	require.NoError(t, err)
	assertRel(t, "join_with_filter", joinFiltered)

	aggReg := relTestReg()
	agg, err := expr.NewCustomAggregateFunc(aggReg,
		ext.NewAggFuncVariant(ext.FunctionID{URN: "extension:x", Name: "sum:i32"}),
		&types.Int64Type{Nullability: types.NullabilityRequired}, nil,
		types.AggInvocationAll, types.AggPhaseInitialToResult, nil, intRef)
	require.NoError(t, err)
	aggNoFilter, err := b.AggregateColumns(scan, []plan.AggRelMeasure{b.Measure(agg, nil)}, 0)
	require.NoError(t, err)
	assertRelReg(t, "aggregate_measure_no_filter", aggNoFilter, aggReg)

	measureFilter, err := b.RootFieldRef(scan, 1)
	require.NoError(t, err)
	aggWithFilter, err := b.AggregateColumns(scan, []plan.AggRelMeasure{b.Measure(agg, measureFilter)}, 0)
	require.NoError(t, err)
	assertRelReg(t, "aggregate_measure_with_filter", aggWithFilter, aggReg)

	write, err := b.NamedInsert(scan, []string{"dest"}, relSchema)
	require.NoError(t, err)
	assertRel(t, "named_write", write)

	extTable := b.ExtensionTable(&anypb.Any{TypeUrl: "ext-table"}, relSchema)
	assertRel(t, "extension_table", extTable)

	iceberg, err := b.IcebergTableFromMetadataFile("s3://meta.json", plan.SnapshotId("snap-1"), relSchema)
	require.NoError(t, err)
	assertRel(t, "iceberg", iceberg)

	vt, err := b.VirtualTable([]string{"a"}, expr.StructLiteralValue{expr.NewPrimitiveLiteral[int32](3, false)})
	require.NoError(t, err)
	assertRel(t, "virtual_table", vt)

	def := &stubExtDef{detail: &anypb.Any{TypeUrl: "ext-rel"}}
	extSingle, err := b.ExtensionSingle(scan, def)
	require.NoError(t, err)
	assertRel(t, "extension_single", extSingle)

	extLeaf, err := b.ExtensionLeaf(def)
	require.NoError(t, err)
	assertRel(t, "extension_leaf", extLeaf)

	extMulti, err := b.ExtensionMulti([]plan.Rel{scan, scan2}, def)
	require.NoError(t, err)
	assertRel(t, "extension_multi", extMulti)

	// A remapped rel exercises the Emit branch of the common encoding.
	remapped, err := filter.Remap(0)
	require.NoError(t, err)
	assertRel(t, "filter_emit", remapped)
}

// HashJoin, MergeJoin, and LocalFile have no public builder; build them from
// protobuf. This also covers the legacy-key branch (equality vs custom) and the
// omit-when-nil post-join filter.
func TestRelToProtoFromProtoOnlyRels(t *testing.T) {
	reg := expr.NewEmptyExtensionRegistry(ext.GetDefaultCollectionWithNoError())
	b := plan.NewBuilderDefault()
	scan := b.NamedScan([]string{"t"}, relSchema)
	scanProto := RelToProto(scan)

	fieldRef := func(field int32) *proto.Expression_FieldReference {
		return &proto.Expression_FieldReference{
			ReferenceType: &proto.Expression_FieldReference_DirectReference{
				DirectReference: &proto.Expression_ReferenceSegment{
					ReferenceType: &proto.Expression_ReferenceSegment_StructField_{
						StructField: &proto.Expression_ReferenceSegment_StructField{Field: field}}}},
			RootType: &proto.Expression_FieldReference_RootReference_{
				RootReference: &proto.Expression_FieldReference_RootReference{}},
		}
	}
	eqKey := &proto.ComparisonJoinKey{
		Left:  fieldRef(0),
		Right: fieldRef(0),
		Comparison: &proto.ComparisonJoinKey_ComparisonType{
			InnerType: &proto.ComparisonJoinKey_ComparisonType_Simple{
				Simple: proto.ComparisonJoinKey_SIMPLE_COMPARISON_TYPE_EQ}},
	}
	customKey := &proto.ComparisonJoinKey{
		Left:  fieldRef(0),
		Right: fieldRef(0),
		Comparison: &proto.ComparisonJoinKey_ComparisonType{
			InnerType: &proto.ComparisonJoinKey_ComparisonType_CustomFunctionReference{
				CustomFunctionReference: 3}},
	}
	postFilter := &proto.Expression{RexType: &proto.Expression_Literal_{
		Literal: &proto.Expression_Literal{LiteralType: &proto.Expression_Literal_Boolean{Boolean: true}}}}

	directCommon := func() *proto.RelCommon {
		return &proto.RelCommon{EmitKind: &proto.RelCommon_Direct_{Direct: &proto.RelCommon_Direct{}}}
	}
	build := func(name string, r *proto.Rel) {
		rel, err := plan.RelFromProto(r, reg)
		require.NoError(t, err, name)
		assertRel(t, name, rel)
	}

	build("hashjoin_equality", &proto.Rel{RelType: &proto.Rel_HashJoin{HashJoin: &proto.HashJoinRel{
		Common: directCommon(), Left: scanProto, Right: scanProto, Keys: []*proto.ComparisonJoinKey{eqKey},
		Type: proto.HashJoinRel_JOIN_TYPE_INNER}}})
	build("hashjoin_custom_with_filter", &proto.Rel{RelType: &proto.Rel_HashJoin{HashJoin: &proto.HashJoinRel{
		Common: directCommon(), Left: scanProto, Right: scanProto, Keys: []*proto.ComparisonJoinKey{customKey},
		Type: proto.HashJoinRel_JOIN_TYPE_INNER, PostJoinFilter: postFilter}}})
	build("mergejoin_equality", &proto.Rel{RelType: &proto.Rel_MergeJoin{MergeJoin: &proto.MergeJoinRel{
		Common: directCommon(), Left: scanProto, Right: scanProto, Keys: []*proto.ComparisonJoinKey{eqKey},
		Type: proto.MergeJoinRel_JOIN_TYPE_INNER}}})
	build("mergejoin_custom_with_filter", &proto.Rel{RelType: &proto.Rel_MergeJoin{MergeJoin: &proto.MergeJoinRel{
		Common: directCommon(), Left: scanProto, Right: scanProto, Keys: []*proto.ComparisonJoinKey{customKey},
		Type: proto.MergeJoinRel_JOIN_TYPE_INNER, PostJoinFilter: postFilter}}})

	// Distinct advanced extensions on the read relation and the local-files
	// message so the two slots cannot be confused.
	readRelExt := &ext.AdvancedExtension{Enhancement: &anypb.Any{TypeUrl: "read-rel"}}
	localFilesExt := &ext.AdvancedExtension{Enhancement: &anypb.Any{TypeUrl: "local-files"}}
	localFile := &proto.Rel{RelType: &proto.Rel_Read{Read: &proto.ReadRel{
		Common:            directCommon(),
		BaseSchema:        NamedStructToProto(relSchema),
		AdvancedExtension: readRelExt,
		ReadType: &proto.ReadRel_LocalFiles_{LocalFiles: &proto.ReadRel_LocalFiles{
			AdvancedExtension: localFilesExt,
			Items: []*proto.ReadRel_LocalFiles_FileOrFiles{{
				PathType: &proto.ReadRel_LocalFiles_FileOrFiles_UriFile{UriFile: "file:///data.parquet"},
				FileFormat: &proto.ReadRel_LocalFiles_FileOrFiles_Parquet{
					Parquet: &proto.ReadRel_LocalFiles_FileOrFiles_ParquetReadOptions{}},
			}}}}}}}
	build("local_file", localFile)
}
