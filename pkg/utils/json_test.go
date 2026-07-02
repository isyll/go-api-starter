package utils_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/isyll/go-grpc-starter/pkg/utils"
)

type TestStruct struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"-"`
	NoTag     string
	Nested    NestedStruct `json:"nested"`
}

type NestedStruct struct {
	Field    string `json:"field"`
	SubField string `json:"sub_field"`
}

func TestGetJSONFieldName(t *testing.T) {
	typ := reflect.TypeOf(TestStruct{})

	tests := []struct {
		name      string
		fieldName string
		expected  string
	}{
		{
			name:      "simple json tag",
			fieldName: "ID",
			expected:  "id",
		},
		{
			name:      "snake_case json tag",
			fieldName: "FirstName",
			expected:  "first_name",
		},
		{
			name:      "ignored field uses struct name",
			fieldName: "Email",
			expected:  "Email",
		},
		{
			name:      "no tag uses struct name",
			fieldName: "NoTag",
			expected:  "NoTag",
		},
		{
			name:      "nested field path",
			fieldName: "Nested.Field",
			expected:  "nested.field",
		},
		{
			name:      "nested field with snake case",
			fieldName: "Nested.SubField",
			expected:  "nested.sub_field",
		},
		{
			name:      "non-existent field",
			fieldName: "NonExistent",
			expected:  "NonExistent",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.GetJSONFieldName(typ, tc.fieldName)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetJSONFieldNameWithPointerType(t *testing.T) {
	typ := reflect.TypeOf(&TestStruct{})

	result := utils.GetJSONFieldName(typ, "FirstName")
	assert.Equal(t, "first_name", result)
}
