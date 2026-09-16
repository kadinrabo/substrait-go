// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"fmt"

	"github.com/substrait-io/substrait-go/v9/expr"
	"github.com/substrait-io/substrait-go/v9/extensions"
	"github.com/substrait-io/substrait-go/v9/plan"
	"github.com/substrait-io/substrait-go/v9/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
	"google.golang.org/protobuf/types/known/anypb"
)

// RelToProto encodes a relation as its protobuf message.
func RelToProto(rel plan.Rel) *proto.Rel {
	switch r := rel.(type) {
	case *plan.NamedTableReadRel:
		return namedTableReadRelToProto(r)
	case *plan.VirtualTableReadRel:
		return virtualTableReadRelToProto(r)
	case *plan.ExtensionTableReadRel:
		return extensionTableReadRelToProto(r)
	case *plan.IcebergTableReadRel:
		return icebergTableReadRelToProto(r)
	case *plan.LocalFileReadRel:
		return localFileReadRelToProto(r)
	case *plan.FilterRel:
		return filterRelToProto(r)
	case *plan.FetchRel:
		return fetchRelToProto(r)
	case *plan.ProjectRel:
		return projectRelToProto(r)
	case *plan.AggregateRel:
		return aggregateRelToProto(r)
	case *plan.SortRel:
		return sortRelToProto(r)
	case *plan.SetRel:
		return setRelToProto(r)
	default:
		panic(fmt.Sprintf("wire: unhandled relation %T", rel))
	}
}

type readRelReader interface {
	BaseSchema() types.NamedStruct
	Filter() expr.Expression
	BestEffortFilter() expr.Expression
	Projection() *expr.MaskExpression
}

func baseReadRelToProto(rc *plan.RelCommon, advExt *extensions.AdvancedExtension, r readRelReader) *proto.ReadRel {
	out := &proto.ReadRel{
		Common:            relCommonToProto(rc),
		BaseSchema:        NamedStructToProto(r.BaseSchema()),
		AdvancedExtension: advExt,
	}
	if f := r.Filter(); f != nil {
		out.Filter = ExprToProto(f)
	}
	if f := r.BestEffortFilter(); f != nil {
		out.BestEffortFilter = ExprToProto(f)
	}
	if p := r.Projection(); p != nil {
		out.Projection = MaskExpressionToProto(p)
	}
	return out
}

func namedTableReadRelToProto(n *plan.NamedTableReadRel) *proto.Rel {
	readRel := baseReadRelToProto(&n.RelCommon, n.GetAdvancedExtension(), n)
	readRel.ReadType = &proto.ReadRel_NamedTable_{
		NamedTable: &proto.ReadRel_NamedTable{
			Names:             n.Names(),
			AdvancedExtension: n.NamedTableAdvancedExtension(),
		},
	}
	return &proto.Rel{RelType: &proto.Rel_Read{Read: readRel}}
}

func virtualTableReadRelToProto(v *plan.VirtualTableReadRel) *proto.Rel {
	readRel := baseReadRelToProto(&v.RelCommon, v.GetAdvancedExtension(), v)
	values := make([]*proto.Expression_Nested_Struct, len(v.Values()))
	for i, val := range v.Values() {
		values[i] = VirtualTableExpressionValueToProto(val)
	}
	readRel.ReadType = &proto.ReadRel_VirtualTable_{
		VirtualTable: &proto.ReadRel_VirtualTable{Expressions: values},
	}
	return &proto.Rel{RelType: &proto.Rel_Read{Read: readRel}}
}

func extensionTableReadRelToProto(e *plan.ExtensionTableReadRel) *proto.Rel {
	readRel := baseReadRelToProto(&e.RelCommon, e.GetAdvancedExtension(), e)
	readRel.ReadType = &proto.ReadRel_ExtensionTable_{
		ExtensionTable: &proto.ReadRel_ExtensionTable{Detail: e.Detail()},
	}
	return &proto.Rel{RelType: &proto.Rel_Read{Read: readRel}}
}

func icebergTableReadRelToProto(n *plan.IcebergTableReadRel) *proto.Rel {
	readRel := baseReadRelToProto(&n.RelCommon, n.GetAdvancedExtension(), n)

	if directTableType, ok := n.TableType().(*plan.Direct); ok {
		direct := &proto.ReadRel_IcebergTable_MetadataFileRead{
			MetadataUri: directTableType.MetadataUri,
		}
		if directTableType.SnapshotId != "" {
			direct.Snapshot = &proto.ReadRel_IcebergTable_MetadataFileRead_SnapshotId{
				SnapshotId: string(directTableType.SnapshotId),
			}
		} else if directTableType.SnapshotTimestamp != 0 {
			direct.Snapshot = &proto.ReadRel_IcebergTable_MetadataFileRead_SnapshotTimestamp{
				SnapshotTimestamp: int64(directTableType.SnapshotTimestamp),
			}
		}
		readRel.ReadType = &proto.ReadRel_IcebergTable_{
			IcebergTable: &proto.ReadRel_IcebergTable{
				TableType: &proto.ReadRel_IcebergTable_Direct{Direct: direct},
			},
		}
	}

	return &proto.Rel{RelType: &proto.Rel_Read{Read: readRel}}
}

func localFileReadRelToProto(lf *plan.LocalFileReadRel) *proto.Rel {
	items := make([]*proto.ReadRel_LocalFiles_FileOrFiles, len(lf.Items()))
	for i := range lf.Items() {
		item := lf.Item(i)
		items[i] = fileOrFilesToProto(&item)
	}

	readRel := baseReadRelToProto(&lf.RelCommon, lf.ReadRelAdvancedExtension(), lf)
	readRel.ReadType = &proto.ReadRel_LocalFiles_{
		LocalFiles: &proto.ReadRel_LocalFiles{
			Items:             items,
			AdvancedExtension: lf.GetAdvancedExtension(),
		},
	}
	return &proto.Rel{RelType: &proto.Rel_Read{Read: readRel}}
}

func fileOrFilesToProto(f *plan.FileOrFiles) *proto.ReadRel_LocalFiles_FileOrFiles {
	ret := &proto.ReadRel_LocalFiles_FileOrFiles{
		PartitionIndex: f.PartIndex,
		Start:          f.Start,
		Length:         f.Len,
	}
	switch f.PathType {
	case plan.URIPath:
		ret.PathType = &proto.ReadRel_LocalFiles_FileOrFiles_UriPath{UriPath: f.Path}
	case plan.URIPathGlob:
		ret.PathType = &proto.ReadRel_LocalFiles_FileOrFiles_UriPathGlob{UriPathGlob: f.Path}
	case plan.URIFile:
		ret.PathType = &proto.ReadRel_LocalFiles_FileOrFiles_UriFile{UriFile: f.Path}
	case plan.URIFolder:
		ret.PathType = &proto.ReadRel_LocalFiles_FileOrFiles_UriFolder{UriFolder: f.Path}
	}

	switch fm := f.Format.(type) {
	case *plan.ParquetReadOptions:
		ret.FileFormat = &proto.ReadRel_LocalFiles_FileOrFiles_Parquet{
			Parquet: (*proto.ReadRel_LocalFiles_FileOrFiles_ParquetReadOptions)(fm),
		}
	case *plan.ArrowReadOptions:
		ret.FileFormat = &proto.ReadRel_LocalFiles_FileOrFiles_Arrow{
			Arrow: (*proto.ReadRel_LocalFiles_FileOrFiles_ArrowReadOptions)(fm),
		}
	case *plan.OrcReadOptions:
		ret.FileFormat = &proto.ReadRel_LocalFiles_FileOrFiles_Orc{
			Orc: (*proto.ReadRel_LocalFiles_FileOrFiles_OrcReadOptions)(fm),
		}
	case *plan.DwrfReadOptions:
		ret.FileFormat = &proto.ReadRel_LocalFiles_FileOrFiles_Dwrf{
			Dwrf: (*proto.ReadRel_LocalFiles_FileOrFiles_DwrfReadOptions)(fm),
		}
	case *plan.ExtensionReadOptions:
		ret.FileFormat = &proto.ReadRel_LocalFiles_FileOrFiles_Extension{
			Extension: (*anypb.Any)(fm),
		}
	}
	return ret
}

func filterRelToProto(fr *plan.FilterRel) *proto.Rel {
	return &proto.Rel{
		RelType: &proto.Rel_Filter{
			Filter: &proto.FilterRel{
				Common:            relCommonToProto(&fr.RelCommon),
				Input:             RelToProto(fr.Input()),
				Condition:         ExprToProto(fr.Condition()),
				AdvancedExtension: fr.GetAdvancedExtension(),
			},
		},
	}
}

func fetchRelToProto(f *plan.FetchRel) *proto.Rel {
	return &proto.Rel{
		RelType: &proto.Rel_Fetch{
			Fetch: &proto.FetchRel{
				Common:            relCommonToProto(&f.RelCommon),
				Input:             RelToProto(f.Input()),
				OffsetMode:        &proto.FetchRel_Offset{Offset: f.Offset()},
				CountMode:         &proto.FetchRel_Count{Count: f.Count()},
				AdvancedExtension: f.GetAdvancedExtension(),
			},
		},
	}
}

func projectRelToProto(p *plan.ProjectRel) *proto.Rel {
	exprs := make([]*proto.Expression, len(p.Expressions()))
	for i, e := range p.Expressions() {
		exprs[i] = ExprToProto(e)
	}
	return &proto.Rel{
		RelType: &proto.Rel_Project{
			Project: &proto.ProjectRel{
				Common:            relCommonToProto(&p.RelCommon),
				Input:             RelToProto(p.Input()),
				Expressions:       exprs,
				AdvancedExtension: p.GetAdvancedExtension(),
			},
		},
	}
}

func aggregateRelToProto(ar *plan.AggregateRel) *proto.Rel {
	groupingExprs := make([]*proto.Expression, len(ar.GroupingExpressions()))
	for i, e := range ar.GroupingExpressions() {
		groupingExprs[i] = ExprToProto(e)
	}

	refs := ar.GroupingReferences()
	groupings := make([]*proto.AggregateRel_Grouping, len(refs))
	for i := range refs {
		groupings[i] = &proto.AggregateRel_Grouping{ExpressionReferences: refs[i]}
	}

	measures := make([]*proto.AggregateRel_Measure, len(ar.Measures()))
	for i := range ar.Measures() {
		m := ar.Measures()[i]
		measures[i] = aggRelMeasureToProto(&m)
	}

	return &proto.Rel{
		RelType: &proto.Rel_Aggregate{
			Aggregate: &proto.AggregateRel{
				Common:              relCommonToProto(&ar.RelCommon),
				Input:               RelToProto(ar.Input()),
				GroupingExpressions: groupingExprs,
				Groupings:           groupings,
				Measures:            measures,
				AdvancedExtension:   ar.GetAdvancedExtension(),
			},
		},
	}
}

func aggRelMeasureToProto(am *plan.AggRelMeasure) *proto.AggregateRel_Measure {
	ret := &proto.AggregateRel_Measure{
		Measure: AggregateFunctionToProto(am.Measure()),
	}
	if f := am.RawFilter(); f != nil {
		ret.Filter = ExprToProto(f)
	}
	return ret
}

func sortRelToProto(sr *plan.SortRel) *proto.Rel {
	sorts := make([]*proto.SortField, len(sr.Sorts()))
	for i := range sr.Sorts() {
		s := sr.Sorts()[i]
		sorts[i] = SortFieldToProto(&s)
	}
	return &proto.Rel{
		RelType: &proto.Rel_Sort{
			Sort: &proto.SortRel{
				Common:            relCommonToProto(&sr.RelCommon),
				Input:             RelToProto(sr.Input()),
				Sorts:             sorts,
				AdvancedExtension: sr.GetAdvancedExtension(),
			},
		},
	}
}

func setRelToProto(s *plan.SetRel) *proto.Rel {
	inputs := make([]*proto.Rel, len(s.Inputs()))
	for i, in := range s.Inputs() {
		inputs[i] = RelToProto(in)
	}
	return &proto.Rel{
		RelType: &proto.Rel_Set{
			Set: &proto.SetRel{
				Common:            relCommonToProto(&s.RelCommon),
				Inputs:            inputs,
				Op:                proto.SetRel_SetOp(s.Op()),
				AdvancedExtension: s.GetAdvancedExtension(),
			},
		},
	}
}
