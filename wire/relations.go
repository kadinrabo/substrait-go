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
