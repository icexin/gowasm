package js

import (
	"fmt"
	"reflect"
)

type Value struct {
	name  string // for debug
	value reflect.Value
	ref   Ref
}

func defaultValue(ref Ref, name string) *Value {
	return &Value{
		name:  name,
		ref:   ref,
		value: reflect.ValueOf(name),
	}
}

func (v *Value) String() string {
	return fmt.Sprintf("%s", v.value.Interface())
}
