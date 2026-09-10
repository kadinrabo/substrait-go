// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"github.com/substrait-io/substrait-go/v9/expr"
	ext "github.com/substrait-io/substrait-go/v9/extensions"
	"github.com/substrait-io/substrait-go/v9/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
	"google.golang.org/protobuf/testing/protocmp"
)

// ExprToProto round-trips the nested expression kinds through the decoder: encode,
// decode, encode again, and the two encodings must match.
func TestExprToProtoRoundTrips(t *testing.T) {
	req := types.NullabilityRequired
	i32 := &types.Int32Type{Nullability: req}
	prim := expr.NewPrimitiveLiteral[int32](5, false)
	boolLit := expr.NewPrimitiveLiteral(true, false)
	reg := expr.NewEmptyExtensionRegistry(ext.GetDefaultCollectionWithNoError())
	schema := types.NewRecordTypeFromTypes([]types.Type{i32})

	ifThen, err := expr.NewIfThen(expr.IfThenPair{If: boolLit, Then: prim}, prim)
	require.NoError(t, err)
	sw, err := expr.NewSwitch(prim, prim, struct {
		If   expr.Literal
		Then expr.Expression
	}{If: prim, Then: prim})
	require.NoError(t, err)

	cases := []struct {
		name string
		e    expr.Expression
	}{
		{"cast", &expr.Cast{Type: &types.Int64Type{Nullability: req}, Input: prim, FailureBehavior: types.CastFailBehaviorThrowException}},
		{"dynamic_parameter", &expr.DynamicParameter{OutputType: i32, ParameterReference: 3}},
		{"if_then", ifThen},
		{"switch", sw},
		{"singular_or_list", &expr.SingularOrList{Value: prim, Options: []expr.Expression{prim, boolLit}}},
		{"multi_or_list", &expr.MultiOrList{Value: []expr.Expression{prim}, Options: [][]expr.Expression{{prim}, {boolLit}}}},
		{"map_expr", &expr.MapExpr{Nullable: true, TypeVariationRef: 2, KeyValues: []struct {
			Key   expr.Expression
			Value expr.Expression
		}{{Key: prim, Value: boolLit}}}},
		{"struct_expr", &expr.StructExpr{Nullable: false, TypeVariationRef: 1, Fields: []expr.Expression{prim, boolLit}}},
		{"list_expr", expr.NewListExpr(true, prim, prim)},
		{"lambda", &expr.Lambda{Parameters: &types.StructType{Nullability: req, Types: []types.Type{i32}}, Body: prim}},
		{"literal", prim},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ExprToProto(tc.e)
			back, err := expr.ExprFromProto(got, schema, reg)
			require.NoError(t, err)
			if diff := cmp.Diff(got, ExprToProto(back), protocmp.Transform()); diff != "" {
				t.Errorf("ExprToProto(%s) did not round-trip (-want +got):\n%s", tc.name, diff)
			}
		})
	}
}

// Field references encode a reference kind and a root kind. Decoding the exotic
// roots needs surrounding context, so assert the encoded arms directly.
func TestFieldReferenceToProtoArms(t *testing.T) {
	i32 := &types.Int32Type{Nullability: types.NullabilityRequired}
	prim := expr.NewPrimitiveLiteral[int32](5, false)

	directRoot, err := expr.NewRootFieldRefFromType(expr.NewStructFieldRef(0), i32)
	require.NoError(t, err)
	sel := FieldReferenceToProto(directRoot).GetSelection()
	require.NotNil(t, sel.GetDirectReference())
	require.NotNil(t, sel.GetRootReference())

	maskProto := &proto.Expression_MaskExpression{Select: &proto.Expression_MaskExpression_StructSelect{
		StructItems: []*proto.Expression_MaskExpression_StructItem{{Field: 0}}}}
	masked := &expr.FieldReference{Reference: expr.MaskExpressionFromProto(maskProto), Root: expr.RootReference}
	require.NotNil(t, FieldReferenceToProto(masked).GetSelection().GetMaskedReference())

	outer := FieldReferenceToProto(&expr.FieldReference{Reference: expr.NewStructFieldRef(1), Root: expr.OuterReference(2)})
	require.Equal(t, uint32(2), outer.GetSelection().GetOuterReference().GetStepsOut())

	lambda := FieldReferenceToProto(&expr.FieldReference{Reference: expr.NewStructFieldRef(0), Root: expr.LambdaParameterReference{StepsOut: 1}})
	require.Equal(t, uint32(1), lambda.GetSelection().GetLambdaParameterReference().GetStepsOut())

	exprRoot := FieldReferenceToProto(&expr.FieldReference{Reference: expr.NewStructFieldRef(0), Root: prim})
	require.NotNil(t, exprRoot.GetSelection().GetExpression())
}
