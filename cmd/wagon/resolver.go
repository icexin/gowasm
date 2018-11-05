package main

import (
	"reflect"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/wasm"
	"github.com/icexin/gowasm"
)

func gomodule(r *gowasm.Resolver) *wasm.Module {
	methods := []string{
		"debug",
		"runtime.wasmExit",
		"runtime.wasmWrite",
		"runtime.nanotime",
		"runtime.walltime",
		"runtime.scheduleCallback",
		"runtime.clearScheduledCallback",
		"runtime.getRandomData",
		"syscall/js.valueGet",
		"syscall/js.valueSet",
		"syscall/js.valueIndex",
		"syscall/js.valueSetIndex",
		"syscall/js.valueLength",
		"syscall/js.valueNew",
		"syscall/js.valuePrepareString",
		"syscall/js.valueCall",
		"syscall/js.valueInvoke",
		"syscall/js.valueVal",
		"syscall/js.stringVal",
		"syscall/js.valueLoadString",
		"syscall/js.valueInstanceOf",
	}

	m := wasm.NewModule()

	m.Export.Entries = map[string]wasm.ExportEntry{}

	for i, method := range methods {
		method := method
		sig := wasm.FunctionSig{
			Form:        0,
			ParamTypes:  []wasm.ValueType{wasm.ValueTypeI32},
			ReturnTypes: nil,
		}
		m.Types.Entries = append(m.Types.Entries, sig)

		fun := wasm.Function{
			Sig: &sig,
			Host: reflect.ValueOf(func(proc *exec.Process, sp int32) {
				r.CallMethod("go", method, int64(sp))
			}),
			Body: &wasm.FunctionBody{},
		}
		m.FunctionIndexSpace = append(m.FunctionIndexSpace, fun)

		m.Export.Entries[method] = wasm.ExportEntry{
			FieldStr: method,
			Kind:     wasm.ExternalFunction,
			Index:    uint32(i),
		}
	}
	return m
}
