// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// protoConstAliases parses the non-test sources of this package and returns the
// set of protobuf constant names that are re-exported here, mapped to the name
// this package gives them. It only looks for the "Exported = proto.Something"
// shape that every enum re-export in this package uses.
func protoConstAliases(t *testing.T) map[string]string {
	t.Helper()

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	require.NoError(t, err)
	require.Contains(t, pkgs, "plan")

	aliases := make(map[string]string)
	for _, file := range pkgs["plan"].Files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST {
				continue
			}
			for _, spec := range gen.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok || len(vs.Names) != len(vs.Values) {
					continue
				}
				for i, value := range vs.Values {
					sel, ok := value.(*ast.SelectorExpr)
					if !ok {
						continue
					}
					pkgIdent, ok := sel.X.(*ast.Ident)
					if !ok || pkgIdent.Name != "proto" {
						continue
					}
					aliases[sel.Sel.Name] = vs.Names[i].Name
				}
			}
		}
	}
	return aliases
}

// enumConstPrefix returns the prefix protoc-gen-go gives the generated value
// constants of an enum, which is the underscore-joined path of the messages the
// enum is nested in: substrait.RelCommon.Hint.ComputationType -> RelCommon_Hint_
func enumConstPrefix(enum protoreflect.EnumDescriptor) string {
	path := strings.TrimPrefix(string(enum.FullName()), string(enum.ParentFile().Package())+".")
	path = strings.TrimSuffix(path, string(enum.Name()))
	return strings.ReplaceAll(path, ".", "_")
}

// TestEveryEnumValueHasExportedConstant asserts that for every value the live
// protobuf descriptor declares, this package exports a corresponding constant.
// A value the spec defines but substrait-go never re-exports is unreachable for
// consumers, who then have to import the protobuf package to name it.
func TestEveryEnumValueHasExportedConstant(t *testing.T) {
	aliases := protoConstAliases(t)

	enums := []protoreflect.EnumDescriptor{
		proto.JoinRel_JoinType(0).Descriptor(),
		proto.SetRel_SetOp(0).Descriptor(),
		proto.RelCommon_Hint_ComputationType(0).Descriptor(),
		proto.ComparisonJoinKey_SimpleComparisonType(0).Descriptor(),
		proto.WriteRel_WriteOp(0).Descriptor(),
		proto.WriteRel_OutputMode(0).Descriptor(),
		proto.Expression_Subquery_SetPredicate_PredicateOp(0).Descriptor(),
		proto.Expression_Subquery_SetComparison_ReductionOp(0).Descriptor(),
		proto.Expression_Subquery_SetComparison_ComparisonOp(0).Descriptor(),
	}

	for _, enum := range enums {
		t.Run(string(enum.Name()), func(t *testing.T) {
			prefix := enumConstPrefix(enum)
			values := enum.Values()
			for i := 0; i < values.Len(); i++ {
				value := values.Get(i)
				constName := prefix + string(value.Name())
				_, ok := aliases[constName]
				assert.Truef(t, ok,
					"the spec declares %s = %d but package plan exports no constant for proto.%s",
					value.FullName(), value.Number(), constName)
			}
		})
	}
}

// TestEnumConstantWireNumbers pins the wire number of every enum constant this
// package re-exports, so a re-export can never silently point at the wrong
// value and so a renumbering upstream is caught here.
func TestEnumConstantWireNumbers(t *testing.T) {
	for name, tt := range map[string]struct {
		got  int32
		want int32
	}{
		"JoinTypeUnspecified":                   {int32(JoinTypeUnspecified), 0},
		"JoinTypeInner":                         {int32(JoinTypeInner), 1},
		"JoinTypeOuter":                         {int32(JoinTypeOuter), 2},
		"JoinTypeLeft":                          {int32(JoinTypeLeft), 3},
		"JoinTypeRight":                         {int32(JoinTypeRight), 4},
		"JoinTypeLeftSemi":                      {int32(JoinTypeLeftSemi), 5},
		"JoinTypeLeftAnti":                      {int32(JoinTypeLeftAnti), 6},
		"JoinTypeLeftSingle":                    {int32(JoinTypeLeftSingle), 7},
		"JoinTypeRightSemi":                     {int32(JoinTypeRightSemi), 8},
		"JoinTypeRightAnti":                     {int32(JoinTypeRightAnti), 9},
		"JoinTypeRightSingle":                   {int32(JoinTypeRightSingle), 10},
		"JoinTypeLeftMark":                      {int32(JoinTypeLeftMark), 11},
		"JoinTypeRightMark":                     {int32(JoinTypeRightMark), 12},
		"SetOpUnspecified":                      {int32(SetOpUnspecified), 0},
		"SetOpMinusPrimary":                     {int32(SetOpMinusPrimary), 1},
		"SetOpMinusMultiset":                    {int32(SetOpMinusMultiset), 2},
		"SetOpIntersectionPrimary":              {int32(SetOpIntersectionPrimary), 3},
		"SetOpIntersectionMultiset":             {int32(SetOpIntersectionMultiset), 4},
		"SetOpUnionDistinct":                    {int32(SetOpUnionDistinct), 5},
		"SetOpUnionAll":                         {int32(SetOpUnionAll), 6},
		"SetOpMinusPrimaryAll":                  {int32(SetOpMinusPrimaryAll), 7},
		"SetOpIntersectionMultisetAll":          {int32(SetOpIntersectionMultisetAll), 8},
		"ComputationTypeUnspecified":            {int32(ComputationTypeUnspecified), 0},
		"ComputationTypeHashtable":              {int32(ComputationTypeHashtable), 1},
		"ComputationTypeBloomFilter":            {int32(ComputationTypeBloomFilter), 2},
		"ComputationTypeUnknown":                {int32(ComputationTypeUnknown), 9999},
		"SimpleComparisonTypeUnspecified":       {int32(SimpleComparisonTypeUnspecified), 0},
		"SimpleComparisonTypeEq":                {int32(SimpleComparisonTypeEq), 1},
		"SimpleComparisonTypeIsNotDistinctFrom": {int32(SimpleComparisonTypeIsNotDistinctFrom), 2},
		"SimpleComparisonTypeMightEqual":        {int32(SimpleComparisonTypeMightEqual), 3},
		"WriteOpUnspecified":                    {int32(WriteOpUnspecified), 0},
		"WriteOpInsert":                         {int32(WriteOpInsert), 1},
		"WriteOpDelete":                         {int32(WriteOpDelete), 2},
		"WriteOpUpdate":                         {int32(WriteOpUpdate), 3},
		"WriteOpCTAS":                           {int32(WriteOpCTAS), 4},
		"OutputModeUnspecified":                 {int32(OutputModeUnspecified), 0},
		"OutputModeNoOutput":                    {int32(OutputModeNoOutput), 1},
		"OutputModeModifiedRecords":             {int32(OutputModeModifiedRecords), 2},
		"SetPredicateOpUnspecified":             {int32(SetPredicateOpUnspecified), 0},
		"SetPredicateOpExists":                  {int32(SetPredicateOpExists), 1},
		"SetPredicateOpUnique":                  {int32(SetPredicateOpUnique), 2},
		"SetComparisonReductionOpUnspecified":   {int32(SetComparisonReductionOpUnspecified), 0},
		"SetComparisonReductionOpAny":           {int32(SetComparisonReductionOpAny), 1},
		"SetComparisonReductionOpAll":           {int32(SetComparisonReductionOpAll), 2},
		"SetComparisonComparisonOpUnspecified":  {int32(SetComparisonComparisonOpUnspecified), 0},
		"SetComparisonComparisonOpEq":           {int32(SetComparisonComparisonOpEq), 1},
		"SetComparisonComparisonOpNe":           {int32(SetComparisonComparisonOpNe), 2},
		"SetComparisonComparisonOpLt":           {int32(SetComparisonComparisonOpLt), 3},
		"SetComparisonComparisonOpGt":           {int32(SetComparisonComparisonOpGt), 4},
		"SetComparisonComparisonOpLe":           {int32(SetComparisonComparisonOpLe), 5},
		"SetComparisonComparisonOpGe":           {int32(SetComparisonComparisonOpGe), 6},
	} {
		assert.Equalf(t, tt.want, tt.got, "%s must stay on wire number %d", name, tt.want)
	}
}
