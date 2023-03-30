package patchy

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type Op struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	From  string      `json:"from,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type EntityID interface{}

type ColumnNameFunc func(field reflect.Type) string
type ValidatorFunc func(value interface{}) error
type ConverterFunc func(value interface{}) (interface{}, error)
type OptionFunc func(*Patchy)

type Patchy struct {
	entityType  reflect.Type
	tableName   string
	colNameFunc ColumnNameFunc
	validator   ValidatorFunc
	converter   ConverterFunc
}

func NewPatchy(entityType reflect.Type, options ...OptionFunc) (*Patchy, error) {
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	if entityType.Kind() != reflect.Struct {
		return nil, errors.New("input type should be a struct or a pointer to a struct")
	}

	p := &Patchy{
		entityType: entityType,
		tableName:  ToSnakeCase(entityType.Name()),
		colNameFunc: func(field reflect.Type) string {
			// TODO: parse db tag
			return ToSnakeCase(field.Name())
		},
	}
	for _, option := range options {
		option(p)
	}
	return p, nil
}

func WithValidator(validator ValidatorFunc) OptionFunc {
	return func(p *Patchy) {
		p.validator = validator
	}
}

func WithConverter(converter ConverterFunc) OptionFunc {
	return func(p *Patchy) {
		p.converter = converter
	}
}

func WithColumnNameFunc(colNameFunc ColumnNameFunc) OptionFunc {
	return func(p *Patchy) {
		p.colNameFunc = colNameFunc
	}
}

func WithTableName(tableName string) OptionFunc {
	return func(p *Patchy) {
		p.tableName = tableName
	}
}

var (
	TagName = "patchy"
)

func (op *Op) GetValue() interface{} {
	return nil
}

type Error struct {
	Attempted Op
	Reason    string
	Code      int // custom
}

// NOTE: this is a very naive implementation, it does not handle all cases specifically acronyms like HTTPStatusCode
func ToSnakeCase(s string) string {
	var result strings.Builder

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 && !unicode.IsUpper(rune(s[i-1])) && s[i-1] != '_' {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

type FieldMetadata struct {
	Type            reflect.Kind
	IsSlice         bool
	IsMap           bool
	IsStruct        bool
	IsPrimitive     bool
	SliceElemType   reflect.Kind
	MapValueType    reflect.Kind
	StructFieldName string
	AllowedOps      []string
	IsIgnored       bool
	ColumnName      string
	TargetIndex     int
	TargetKey       string
}

// func (p *Patchy) GetFieldMetadata(jsonPointer string) (FieldMetadata, error) {

// 	parts := strings.Split(jsonPointer, "/")
// 	if len(parts) < 2 {
// 		return FieldMetadata{}, errors.New("invalid JSON pointer format")
// 	}

// 	return p.getFieldMetadataRec(t, parts[1:])
// }

func (p *Patchy) getFieldMetadataRec(t reflect.Type, parts []string) (FieldMetadata, error) {
	if len(parts) == 0 {
		return p.buildMetadata(t, ""), nil
	}

	fieldName, fieldType, err := getFieldByJsonTag(t, parts[0])
	if err != nil {
		return FieldMetadata{}, err
	}

	if len(parts) == 1 {
		return p.buildMetadata(fieldType, fieldName), nil
	}

	switch fieldType.Kind() {
	case reflect.Struct:
		return p.getFieldMetadataRec(fieldType, parts[1:])
	case reflect.Ptr:
		return p.getFieldMetadataRec(fieldType.Elem(), parts[1:])
	case reflect.Slice:
		// index, err := parseArrayIndex(parts[1])
		// if err != nil {
		// 	return FieldMetadata{}, err
		// }

		// if index == "-" {
		return p.buildMetadata(fieldType, fieldName), nil
		// }

		// return p.getFieldMetadataRec(fieldType.Elem(), parts[2:])
	case reflect.Map:
		// get the key type of the map and verify its a string
		return p.buildMetadata(fieldType, fieldName), nil
	default:
		return FieldMetadata{}, errors.New("invalid JSON pointer")
	}
}

// getFieldByJsonTag returns the field name and type of the field with the given json tag.
func getFieldByJsonTag(t reflect.Type, tagName string) (string, reflect.Type, error) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if strings.Split(jsonTag, ",")[0] == tagName {
			return field.Name, field.Type, nil
		}
	}

	return "", nil, errors.New("field not found")
}

func (p *Patchy) buildMetadata(field reflect.Type, fieldName string) FieldMetadata {
	meta := FieldMetadata{
		StructFieldName: fieldName,
		Type:            field.Kind(),
		ColumnName:      p.colNameFunc(field),
	}

	switch field.Kind() {
	case reflect.Slice:
		meta.IsSlice = true
		meta.SliceElemType = field.Elem().Kind()
	case reflect.Map:
		meta.IsMap = true
		meta.MapValueType = field.Elem().Kind()
	case reflect.Struct:
		meta.IsStruct = true
	case reflect.Ptr:
		meta.Type = field.Elem().Kind()
		meta.IsPrimitive = isPrimitiveType(field.Elem().Kind())
	default:
		meta.Type = field.Kind()
		meta.IsPrimitive = isPrimitiveType(field.Kind())
	}

	return meta

}

func isPrimitiveType(k reflect.Kind) bool {
	switch k {
	case reflect.Bool, reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func parseArrayIndex(s string) (string, error) {
	if s == "-" {
		return s, nil
	}
	_, err := strconv.Atoi(s)
	if err != nil {
		return "", errors.New("invalid array index")
	}

	return s, nil

}
