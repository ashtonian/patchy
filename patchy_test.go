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
		// {
		// 	name:    "regular field",
		// 	pointer: "/name",
		// 	expectedValue: &FieldMetadata{
		// 		Type:          reflect.String,
		// 		IsSlice:       false,
		// 		IsMap:         false,
		// 		IsStruct:      false,
		// 		IsPrimitive:   true,
		// 		SliceElemType: reflect.Invalid,
		// 		MapValueType:  reflect.Invalid,
		// 		// Tags:            "json:\"name\"",
		// 		StructFieldName: "Name",
		// 	},
		// 	expectedError: nil,
		// },
		// {
		// 	name:    "embedded field",
		// 	pointer: "/embeddedAddress/street",
		// 	expectedValue: &FieldMetadata{
		// 		Type:          reflect.String,
		// 		IsSlice:       false,
		// 		IsMap:         false,
		// 		IsStruct:      false,
		// 		IsPrimitive:   true,
		// 		SliceElemType: reflect.Invalid,
		// 		MapValueType:  reflect.Invalid,
		// 		// Tags:            "json:\"street\"",
		// 		StructFieldName: "Street",
		// 	},
		// 	expectedError: nil,
		// },
		// {
		// 	name:    "slice index",
		// 	pointer: "/hobbies/0",
		// 	expectedValue: &FieldMetadata{
		// 		Type:          reflect.Slice,
		// 		IsSlice:       true,
		// 		IsMap:         false,
		// 		IsStruct:      false,
		// 		IsPrimitive:   false,
		// 		SliceElemType: reflect.String,
		// 		MapValueType:  reflect.Invalid,
		// 		// Tags:            "json:\"hobbies\"",
		// 		StructFieldName: "Hobbies",
		// 	},
		// 	expectedError: nil,
		// },
		// {
		// 	name:    "map key",
		// 	pointer: "/pets/dog",
		// 	expectedValue: &FieldMetadata{
		// 		Type:         reflect.TypeOf(map[string]string{}).Kind(),
		// 		IsMap:        true,
		// 		MapValueType: reflect.String,
		// 		// Tags:            "json:\"pets\"",
		// 		StructFieldName: "Pets",
		// 	},
		// 	expectedError: nil,
		// },
		// {
		// 	name:    "root",
		// 	pointer: "",
		// 	expectedValue: &FieldMetadata{
		// 		Type:        reflect.Struct,
		// 		IsSlice:     false,
		// 		IsMap:       false,
		// 		IsStruct:    true,
		// 		IsPrimitive: false,
		// 		// SliceElemType: reflect.Invalid,
		// 		// MapValueType:  reflect.Invalid,
		// 		// Tags:            "",
		// 		StructFieldName: "",
		// 	},
		// 	expectedError: nil,
		// },
		// {
		// 	name:          "nonexistent field",
		// 	pointer:       "/nonexistent",
		// 	expectedValue: nil,
		// 	expectedError: errors.New("field not found"),
		// },
		{
			name:          "invalid slice index",
			pointer:       "/hobbies/invalid",
			expectedValue: nil,
			expectedError: errors.New("invalid index: invalid"),
		},
		// {
		// 	name:          "invalid map key",
		// 	pointer:       "/pets/invalid",
		// 	expectedValue: nil,
		// 	expectedError: errors.New("invalid key: invalid"),
		// },
		// {
		// 	name:    "nil pointer",
		// 	pointer: "/address/city",
		// 	expectedValue: &FieldMetadata{
		// 		Type:          reflect.String,
		// 		IsSlice:       false,
		// 		IsMap:         false,
		// 		IsStruct:      false,
		// 		IsPrimitive:   true,
		// 		SliceElemType: reflect.Invalid,
		// 		MapValueType:  reflect.Invalid,
		// 		// Tags:            "json:\"city\"",
		// 		StructFieldName: "City",
		// 	},
		// 	expectedError: nil,
		// },
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
		if actualValue.Type != tc.expectedValue.Type {
			t.Errorf("%s: expected type %v, actual type %v", tc.name, tc.expectedValue.Type, actualValue.Type)
		}
		if actualValue.IsSlice != tc.expectedValue.IsSlice {
			t.Errorf("%s: expected IsSlice %t, actual IsSlice %t", tc.name, tc.expectedValue.IsSlice, actualValue.IsSlice)
		}
		if actualValue.IsMap != tc.expectedValue.IsMap {
			t.Errorf("%s: expected IsMap %t, actual IsMap %t", tc.name, tc.expectedValue.IsMap, actualValue.IsMap)
		}
		if actualValue.IsStruct != tc.expectedValue.IsStruct {
			t.Errorf("%s: expected IsStruct %t, actual IsStruct %t", tc.name, tc.expectedValue.IsStruct, actualValue.IsStruct)
		}
		if actualValue.IsPrimitive != tc.expectedValue.IsPrimitive {
			t.Errorf("%s: expected IsPrimitive %t, actual IsPrimitive %t", tc.name, tc.expectedValue.IsPrimitive, actualValue.IsPrimitive)
		}
		if actualValue.SliceElemType != tc.expectedValue.SliceElemType {
			t.Errorf("%s: expected SliceElemType %v, actual SliceElemType %v", tc.name, tc.expectedValue.SliceElemType, actualValue.SliceElemType)
		}
		if actualValue.MapValueType != tc.expectedValue.MapValueType {
			t.Errorf("%s: expected MapValueType %v, actual MapValueType %v", tc.name, tc.expectedValue.MapValueType, actualValue.MapValueType)
		}
		// if actualValue.Tags != tc.expectedValue.Tags {
		// 	t.Errorf("%s: expected Tags %s, actual Tags %s", tc.name, tc.expectedValue.Tags, actualValue.Tags)
		// }
		if actualValue.StructFieldName != tc.expectedValue.StructFieldName {
			t.Errorf("%s: expected StructFieldName %s, actual StructFieldName %s", tc.name, tc.expectedValue.StructFieldName, actualValue.StructFieldName)
		}

	}

}
