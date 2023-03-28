package patchy

import (
	"reflect"
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

type ValidatorFunc func(value interface{}) error
type ConverterFunc func(value interface{}) (interface{}, error)
type OptionFunc func(*Patchy)

type Patchy struct {
	entityType  reflect.Type
	tableName   string
	columnNames []string
	validator   ValidatorFunc
	converter   ConverterFunc
	allowedOps  map[string]bool
	// fieldInfoCache map[string]fieldInfo
}

func NewPatchy(entityType reflect.Type, options ...OptionFunc) *Patchy {
	p := &Patchy{
		entityType: entityType,
		tableName:  entityType.Name(),
	}
	for _, option := range options {
		option(p)
	}
	return p
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

func WithTableName(tableName string) OptionFunc {
	return func(p *Patchy) {
		p.tableName = tableName
	}
}

var (
	TagName = "patchy"
)

// func NewPatchSpec(model interface{}) (*PatchSpec, error) {
// 	if model == nil {
// 		return nil, errors.New("patchy: 'Model' is a required field")
// 	}

// 	t := reflect.TypeOf(model)
// 	if t.Kind() != reflect.Struct {
// 		return nil, errors.New("patchy: 'Model' must be a struct type")
// 	}
// 	fields := list.New()
// 	for i := 0; i < t.NumField(); i++ {
// 		l.PushFront(t.Field(i))
// 	}
// 	for l.Len() > 0 {
// 		f := l.Remove(l.Front()).(reflect.StructField)
// 		_, ok := f.Tag.Lookup(TagName)
// 		// TODO: loop through fields check types create spec, enable field nesting

// 	}
// 	return nil, nil
// }

// created using reflection once, then used every query to generate sql.
type PatchSpec struct {
	TargetResource string
	Fields         []PatchField
}

type PatchField struct {
	AllowedOps string
	TargetType string
	// Validator func // only need two funcs
	// Converter func
	// Filter func
}

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
