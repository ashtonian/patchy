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

type ColumnNamer func(field reflect.StructField) string
type AllowedOps func(field reflect.StructField) []string
type FieldLocator func(t reflect.Type, pathName string) (reflect.StructField, error)
type Validator func(value interface{}) error
type Converter func(value interface{}) (interface{}, error)
type OptionFunc func(*Patchy) error

type Patchy struct {
	entityType   reflect.Type
	tableName    string
	colNameFunc  ColumnNamer
	validator    Validator
	converter    Converter
	allowedOps   AllowedOps
	fieldLocator FieldLocator
}

func AllowOpsFromTag(field reflect.StructField) []string {
	patchRaw, found := field.Tag.Lookup(TagName)
	if !found {
		return []string{}
	}
	patchyConfig := strings.Split(patchRaw, ",")
	if patchyConfig[0] == "-" {
		return []string{}
	}
	// TODO: validate ops / convert to opt enums.
	// for i, v := range patchyConfig {
	// 	// validate
	// }
	return patchyConfig
}

func FieldFromTag(t reflect.Type, pathName string) (reflect.StructField, error) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if strings.Split(jsonTag, ",")[0] == pathName {
			return field, nil
		}
	}

	return reflect.StructField{}, errors.New("field not found")
}

func ColNameFromTag(field reflect.StructField) string {
	db, found := field.Tag.Lookup("db")
	if !found {
		ToSnakeCase(field.Name)
	}
	dbTags := strings.Split(db, ",")
	return dbTags[0]
}

func NewPatchy(entityType reflect.Type, options ...OptionFunc) (*Patchy, error) {
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	if entityType.Kind() != reflect.Struct {
		return nil, errors.New("input type should be a struct or a pointer to a struct")
	}

	p := &Patchy{
		entityType:   entityType,
		tableName:    ToSnakeCase(entityType.Name()),
		colNameFunc:  ColNameFromTag,
		allowedOps:   AllowOpsFromTag,
		fieldLocator: FieldFromTag,
	}
	for _, option := range options {
		err := option(p)
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}

func WithAllowedOps(allowed AllowedOps) OptionFunc {
	return func(p *Patchy) error {
		if allowed == nil {
			return errors.New("allowedOps cannot be nil")
		}
		p.allowedOps = allowed
		return nil
	}
}

func WithValidator(validator Validator) OptionFunc {
	return func(p *Patchy) error {
		if validator == nil {
			return errors.New("validator cannot be nil")
		}
		p.validator = validator
		return nil
	}
}

func WithFieldLocator(fieldLocator FieldLocator) OptionFunc {
	return func(p *Patchy) error {
		if fieldLocator == nil {
			return errors.New("fieldLocator cannot be nil")
		}
		p.fieldLocator = fieldLocator
		return nil
	}
}

func WithConverter(converter Converter) OptionFunc {
	return func(p *Patchy) error {
		if converter == nil {
			return errors.New("converter cannot be nil")
		}
		p.converter = converter
		return nil
	}
}

func WithColumnNamer(colNameFunc ColumnNamer) OptionFunc {
	return func(p *Patchy) error {
		if colNameFunc == nil {
			return errors.New("ColumnNamer cannot be nil")
		}
		p.colNameFunc = colNameFunc
		return nil
	}
}

func WithTableName(tableName string) OptionFunc {
	return func(p *Patchy) error {
		p.tableName = tableName
		return nil
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

// NOTE: this is a very naive implementation, it does not handle all cases, specifically acronyms like HTTPStatusCode are not handled well.
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
	StructFieldName string
	Type            reflect.Kind
	SubElemType     reflect.Kind
	AllowedOps      []string
	ColumnName      string
	PathTarget      string // This can be an array index, '-' meaning after last key, or a map key depending on the subtype.
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
		return FieldMetadata{}, errors.New("invalid JSON pointer")
	}

	field, err := p.fieldLocator(t, parts[0])
	if err != nil {
		return FieldMetadata{}, err
	}

	if len(parts) == 1 {
		return p.buildMetadata(field), nil
	}

	switch field.Type.Kind() {
	case reflect.Struct:
		return p.getFieldMetadataRec(field.Type, parts[1:])
	case reflect.Ptr:
		return p.getFieldMetadataRec(field.Type.Elem(), parts[1:])
	case reflect.Slice:
		// if index == "-" {
		meta := p.buildMetadata(field)

		if len(parts) > 1 {
			indexStr, err := validateArrayIndex(parts[1])
			if err != nil {
				return FieldMetadata{}, err
			}
			meta.PathTarget = indexStr
		}
		return meta, nil

	case reflect.Map:
		meta := p.buildMetadata(field)
		if len(parts) > 1 {
			meta.PathTarget = parts[1]
		}
		return meta, nil
	default:
		return FieldMetadata{}, errors.New("invalid JSON pointer")
	}
}

func (p *Patchy) buildMetadata(field reflect.StructField) FieldMetadata {
	meta := FieldMetadata{
		StructFieldName: field.Name,
		Type:            field.Type.Kind(),
		ColumnName:      p.colNameFunc(field),
		AllowedOps:      p.allowedOps(field),
	}

	switch field.Type.Kind() {
	case reflect.Slice:
		meta.SubElemType = field.Type.Elem().Kind()
	case reflect.Map:
		meta.SubElemType = field.Type.Elem().Kind()
	case reflect.Ptr:
		meta.Type = field.Type.Elem().Kind()
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

func validateArrayIndex(s string) (string, error) {
	if s == "-" {
		return s, nil
	}
	_, err := strconv.Atoi(s)
	if err != nil {
		return "", errors.New("invalid array index")
	}

	return s, nil

}
