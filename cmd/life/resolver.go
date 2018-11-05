package main

import (
	"github.com/icexin/gowasm"
	"github.com/perlin-network/life/exec"
)

type Resolver struct {
	*gowasm.Resolver
}

func (r *Resolver) ResolveFunc(module, field string) exec.FunctionImport {
	return func(vm *exec.VirtualMachine) int64 {
		frame := vm.GetCurrentFrame()
		sp := frame.Locals[0]
		return r.CallMethod(module, field, sp)
	}

}

func (r *Resolver) ResolveGlobal(module, field string) int64 {
	return 0
}
