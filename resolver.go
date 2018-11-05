package gowasm

import (
	"fmt"
	"reflect"
)

type Registry interface {
	Register(module, field string, f interface{})
}

type method struct {
	Type reflect.Type
	Func reflect.Value
}

type Resolver struct {
	mem     []byte
	modules map[string]*method
}

func NewResolver() *Resolver {
	return &Resolver{
		modules: make(map[string]*method),
	}
}

func (r *Resolver) SetMemory(b []byte) {
	r.mem = b
}

func (r *Resolver) Register(module, field string, f interface{}) {
	key := module + "." + field
	r.modules[key] = &method{
		Type: reflect.TypeOf(f),
		Func: reflect.ValueOf(f),
	}
}

func (r *Resolver) CallMethod(module, field string, sp int64) int64 {
	if field != "runtime.wasmWrite" {
		logger.Printf("call %s.%s", module, field)
	}
	m, ok := r.modules[module+"."+field]
	if !ok {
		panic(fmt.Sprintf("%s.%s not found", module, field))
	}
	return r.callMethod(m, sp)
}

func (r *Resolver) callMethod(m *method, sp int64) int64 {
	mem := r.mem
	dec := NewDecoder(mem, sp+8)
	mtype := m.Type
	args := []reflect.Value{}
	for i := 0; i < mtype.NumIn(); i++ {
		argtype := mtype.In(i)
		ref := reflect.New(argtype)
		dec.Decode(ref)
		args = append(args, ref.Elem())
	}
	rets := m.Func.Call(args)
	enc := NewEncoder(mem, dec.Offset())
	for i := 0; i < len(rets); i++ {
		ret := rets[i]
		enc.Encode(ret)
	}
	return 0
}
