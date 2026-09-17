// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substrait-io/substrait-go/v9/expr"
	"github.com/substrait-io/substrait-go/v9/types"
	"google.golang.org/protobuf/types/known/anypb"
)

// customExtDef is a test ExtensionRelDefinition that claims a fixed output schema.
type customExtDef struct {
	detail *anypb.Any
	schema types.RecordType
}

func (d *customExtDef) Schema(inputs []Rel) types.RecordType  { return d.schema }
func (d *customExtDef) Build(_ []Rel) *anypb.Any              { return d.detail }
func (d *customExtDef) Expressions(_ []Rel) []expr.Expression { return nil }

func TestIsRecordTypeSupported(t *testing.T) {
	fixedSchema := *types.NewRecordTypeFromTypes([]types.Type{
		&types.Int64Type{Nullability: types.NullabilityRequired},
	})
	decoded := &customExtDef{schema: fixedSchema}
	undecoded := &UndecodedExtension{}

	assert.True(t, isRecordTypeSupported(&ExtensionSingleRel{definition: decoded}))
	assert.True(t, isRecordTypeSupported(&ExtensionLeafRel{definition: decoded}))
	assert.True(t, isRecordTypeSupported(&ExtensionMultiRel{definition: decoded}))

	assert.False(t, isRecordTypeSupported(&ExtensionSingleRel{definition: undecoded}))
	assert.False(t, isRecordTypeSupported(&ExtensionLeafRel{definition: undecoded}))
	assert.False(t, isRecordTypeSupported(&ExtensionMultiRel{definition: undecoded}))
}
