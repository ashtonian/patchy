package patchy

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"TitleCaseString", "title_case_string"},
		// {"HTTPStatusCode", "http_status_code"},
		{"", ""},
		{"a", "a"},
		{"A", "a"},
		{"_", "_"},
	}

	for _, tt := range tests {
		got := ToSnakeCase(tt.input)
		if got != tt.want {
			t.Errorf("toSnakeCase(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

func TestGetFieldMetadata(t *testing.T) {

	type EmbeddedAddress struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Person struct {
		EmbeddedAddress `json:"embeddedAddress"`
		Name            string            `json:"name"`
		Age             int               `json:"age"`
		Address         *Address          `json:"address"`
		Hobbies         []string          `json:"hobbies"`
		Pets            map[string]string `json:"pets"`
	}

	patchy, err := NewPatchy(reflect.TypeOf(Person{}))
	if err != nil {
		t.Fatal(err)
	}

	type TestCase struct {
		name          string
		pointer       string
		expectedValue *FieldMetadata
		expectedError error
	}

	testCases := []TestCase{
		{
			name:    "regular field",
			pointer: "/name",
			expectedValue: &FieldMetadata{
				Type:            reflect.String,
				StructFieldName: "Name",
			},
			expectedError: nil,
		},
		{
			name:    "embedded field",
			pointer: "/embeddedAddress/street",
			expectedValue: &FieldMetadata{
				Type: reflect.String,
				// Tags:            "json:\"street\"",
				StructFieldName: "Street",
			},
			expectedError: nil,
		},
		{
			name:    "slice index",
			pointer: "/hobbies/0",
			expectedValue: &FieldMetadata{
				Type:        reflect.Slice,
				SubElemType: reflect.String,
				TargetStr:   "0",
				// Tags:            "json:\"hobbies\"",
				StructFieldName: "Hobbies",
			},
			expectedError: nil,
		},
		{
			name:    "map key",
			pointer: "/pets/dog",
			expectedValue: &FieldMetadata{
				Type:        reflect.TypeOf(map[string]string{}).Kind(),
				SubElemType: reflect.String,
				TargetStr:   "dog",
				// Tags:            "json:\"pets\"",
				StructFieldName: "Pets",
			},
			expectedError: nil,
		},
		{
			name:    "root",
			pointer: "",
			expectedValue: &FieldMetadata{
				Type:            reflect.Struct,
				StructFieldName: "",
			},
			expectedError: nil,
		},
		{
			name:          "nonexistent field",
			pointer:       "/nonexistent",
			expectedValue: nil,
			expectedError: errors.New("field not found"),
		},
		{
			name:          "invalid slice index",
			pointer:       "/hobbies/invalid",
			expectedValue: nil,
			expectedError: errors.New("invalid array index"),
		},
		{
			name:          "nested nil pointer",
			pointer:       "/address/notfound",
			expectedValue: nil,
			expectedError: errors.New("field not found"),
		},
	}
	for _, tc := range testCases {
		actualValue, actualError := patchy.getFieldMetadataRec(patchy.entityType, strings.Split(tc.pointer, "/")[1:])
		if actualError != nil && tc.expectedError == nil {
			t.Errorf("%s: unexpected error found '%v'", tc.name, actualError)
			continue
		}
		if tc.expectedError != nil && actualError == nil {
			t.Errorf("%s: expected error '%v' not found", tc.name, tc.expectedError)
			continue
		}
		if tc.expectedError != nil && actualError != nil && tc.expectedError.Error() != actualError.Error() {
			t.Errorf("%s: expected error '%v', actual error '%v'", tc.name, tc.expectedError, actualError)
			continue
		}
		if tc.expectedError != nil && actualError != nil && tc.expectedError.Error() == actualError.Error() {
			continue
		}
		if actualValue.Type != tc.expectedValue.Type {
			t.Errorf("%s: expected type %v, actual type %v", tc.name, tc.expectedValue.Type, actualValue.Type)
		}
		if actualValue.SubElemType != tc.expectedValue.SubElemType {
			t.Errorf("%s: expected SliceElemType %v, actual SliceElemType %v", tc.name, tc.expectedValue.SubElemType, actualValue.SubElemType)
		}
		if actualValue.TargetStr != tc.expectedValue.TargetStr {
			t.Errorf("%s: expected TargetStr %v, actual TargetStr %v", tc.name, tc.expectedValue.TargetStr, actualValue.TargetStr)
		}
		if actualValue.StructFieldName != tc.expectedValue.StructFieldName {
			t.Errorf("%s: expected StructFieldName %s, actual StructFieldName %s", tc.name, tc.expectedValue.StructFieldName, actualValue.StructFieldName)
		}

	}

}
