package main

import (
	"container/list"
	"errors"
	"reflect"
)

type Op struct {
	Operation string
	From      string
	Path      string
	Value     interface{}
}

var (
	TagName = "patchy"
)

func NewPatchSpec(model interface{}) (*PatchSpec, error) {
	if model == nil {
		return nil, errors.New("patchy: 'Model' is a required field")
	}

	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Struct {
		return nil, errors.New("patchy: 'Model' must be a struct type")
	}
	fields := list.New()
	for i := 0; i < t.NumField(); i++ {
		l.PushFront(t.Field(i))
	}
	for l.Len() > 0 {
		f := l.Remove(l.Front()).(reflect.StructField)
		_, ok := f.Tag.Lookup(TagName)
		// TODO: loop through fields check types create spec, enable field nesting

	}
	return nil, nil
}

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
