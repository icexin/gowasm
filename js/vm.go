package js

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"unsafe"
)

type Getter interface {
	Get(property string) (interface{}, bool)
}

type VM struct {
	valueid Ref
	values  map[Ref]*Value
	mem     *Memory
	Log     *log.Logger
	//refs    map[reflect.Value]Ref
}

func NewVM(mem *Memory) *VM {
	vm := &VM{
		valueid: ValueGo + 1,
		values:  make(map[Ref]*Value),
		mem:     mem,
	}
	vm.initDefaultValue()
	return vm
}

func (vm *VM) initDefaultValue() {
	vm.values[ValueNaN] = defaultValue(ValueNaN, "NaN")
	vm.values[ValueUndefined] = defaultValue(ValueUndefined, "Undefined")
	vm.values[ValueNull] = defaultValue(ValueNull, "Null")
	vm.values[ValueTrue] = defaultValue(ValueTrue, "True")
	vm.values[ValueFalse] = defaultValue(ValueFalse, "False")
	vm.values[ValueMemory] = &Value{
		name:  "Memory",
		ref:   ValueMemory,
		value: reflect.ValueOf(vm.mem),
	}

	goruntime := &Value{
		name:  "Go",
		ref:   ValueGo,
		value: reflect.ValueOf(struct{}{}),
	}
	vm.values[ValueGo] = goruntime
	DefaultGlobal.Register("Go", goruntime)

	vm.values[ValueGlobal] = &Value{
		name:  "Gloabl",
		ref:   ValueGlobal,
		value: reflect.ValueOf(DefaultGlobal),
	}
}

const (
	tagString = 1
	tagSymbol = 2
	tagFunc   = 3
	tagObject = 4
)

func floatValue(f float64) Ref {
	if f != f {
		return ValueNaN
	}
	return *(*Ref)(unsafe.Pointer(&f))
}

func (vm *VM) storeValue(name string, x interface{}) Ref {
	if x == nil {
		return ValueNull
	}
	switch xx := x.(type) {
	case int8, int16, int32, int64, int:
		return floatValue(float64(reflect.ValueOf(x).Int()))
	case uint8, uint16, uint32, uint64, uint:
		return floatValue(float64(reflect.ValueOf(x).Uint()))
	case float32, float64:
		return floatValue(reflect.ValueOf(x).Float())
	case bool:
		if xx {
			return ValueTrue
		} else {
			return ValueFalse
		}
	}

	var tag int64
	v := reflect.ValueOf(x)
	t := v.Type()
	switch t.Kind() {
	case reflect.String:
		tag = tagString
	case reflect.Func:
		tag = tagFunc
	default:
		tag = tagObject
	}
	vm.valueid++
	ref := vm.valueid | Ref(tag<<32)
	vm.values[ref] = &Value{
		name:  name,
		value: v,
		ref:   ref,
	}
	return ref
}

func (vm *VM) loadValue(ref Ref) (*Value, bool) {
	n, ok := ref.Number()
	if ok {
		return &Value{
			name:  "number",
			value: reflect.ValueOf(n),
			ref:   ref,
		}, true
	}
	v, ok := vm.values[ref]
	return v, ok
}

func (vm *VM) Property(ref Ref, name string) Ref {
	parent, ok := vm.values[ref]
	if !ok {
		// log.Printf("ref %x not found", ref)
		return ValueUndefined
	}
	v, ok := vm.property(parent.value, name)
	if !ok {
		// log.Printf("ref %s property %s not found", ref, name)
		return ValueUndefined
	}
	if value, ok := v.(*Value); ok {
		return value.ref
	}
	fullname := fmt.Sprintf("%s.%s", parent.name, name)
	return vm.storeValue(fullname, v)
}

func (vm *VM) property(p reflect.Value, name string) (interface{}, bool) {
	name = strings.Title(name)
	// Getter interface
	if g, ok := p.Interface().(Getter); ok {
		prop, ok := g.Get(name)
		return prop, ok
	}

	// Map
	if p.Kind() == reflect.Map {
		g := p.MapIndex(reflect.ValueOf(name))
		if g.IsValid() {
			return g.Interface(), true
		}
		return nil, false
	}

	// Method
	prop := p.MethodByName(name)
	if prop.IsValid() {
		return prop.Interface(), true
	}

	// FieldByName must not be a ptr
	if p.Kind() == reflect.Ptr {
		p = p.Elem()
	}
	prop = p.FieldByName(name)
	if prop.IsValid() {
		return prop.Interface(), true
	}
	return nil, false
}

func (vm *VM) Exception(err error) Ref {
	e, ok := err.(*Exception)
	if !ok {
		e = &Exception{"EINVAL", err.Error()}
	}
	return vm.storeValue("error", e)
}

func (vm *VM) call(name string, f reflect.Value, args []Ref) (ret Ref, err error) {
	retv := f.Call(vm.parseArgs(args))
	if len(retv) == 0 {
		return ValueUndefined, nil
	}
	errv := retv[len(retv)-1]
	var ok bool
	if err, ok = errv.Interface().(error); ok {
		if _, ok := err.(*Exception); ok {
			return
		}
		err = NewException("", err.Error())
		return
	}
	name = fmt.Sprintf("%s(%s)", name, args)
	return vm.storeValue(name, retv[0].Interface()), nil
}

func (vm *VM) New(ref Ref, args []Ref) (Ref, error) {
	v, ok := vm.loadValue(ref)
	if !ok {
		return 0, ErrNotfound
	}
	return vm.call(v.name, v.value, args)
}

func (vm *VM) Call(ref Ref, method string, args []Ref) (Ref, error) {
	v, ok := vm.loadValue(ref)
	if !ok {
		return 0, ErrNotfound
	}
	name := fmt.Sprintf("%s.%s", v.name, method)
	method = strings.Title(method)
	f := v.value.MethodByName(method)
	if !f.IsValid() {
		return 0, ErrNotfound
	}
	// log.Printf("call %s, args: %v", name, args)
	return vm.call(name, f, args)
}

func (vm *VM) Store(x interface{}) Ref {
	return vm.storeValue("store", x)
}

func (vm *VM) Value(ref Ref) *Value {
	v, ok := vm.loadValue(ref)
	if !ok {
		return nil
	}
	return v
}

func (vm *VM) parseArgs(args []Ref) []reflect.Value {
	var ret []reflect.Value
	for _, arg := range args {
		v, ok := vm.loadValue(arg)
		if ok {
			ret = append(ret, v.value)
		} else {
			panic("bad ref " + arg.String())
		}
	}
	return ret
}

func (vm *VM) DebugStr(ref Ref) string {
	v, ok := vm.loadValue(ref)
	if !ok {
		return "<undefined>"
	}
	return fmt.Sprintf("<%s,%s,%s>", v.name, v.ref, v.value.Kind())
}
