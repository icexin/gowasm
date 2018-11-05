package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/icexin/gowasm"
	"github.com/perlin-network/life/exec"
	"github.com/perlin-network/life/utils"
)

func run(vm *exec.VirtualMachine, entry int, argc, argv int64, rt *gowasm.Runtime) (int64, error) {
	vm.Ignite(entry, argc, argv)
	for !vm.Exited {
		for !vm.Exited {
			vm.Execute()
			if vm.Delegate != nil {
				vm.Delegate()
				vm.Delegate = nil
			}
		}

		if !rt.Exited() {
			rt.WaitTimer()
			vm.Ignite(entry, argc, argv)
		}
	}

	if vm.ExitError != nil {
		return -1, utils.UnifyError(vm.ExitError)
	}
	return vm.ReturnValue, nil

}

func main() {
	flag.Parse()

	// Read WebAssembly *.wasm file.
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	io.Copy(buf, f)
	f.Close()
	input := buf.Bytes()

	r := &Resolver{gowasm.NewResolver()}
	rt := gowasm.NewRuntime()
	rt.Register(r)

	// Instantiate a new WebAssembly VM with a few resolved imports.
	vm, err := exec.NewVirtualMachine(input, exec.VMConfig{
		DefaultMemoryPages: 12800,
		DefaultTableSize:   65536,
	}, r, nil)

	if err != nil {
		panic(err)
	}

	r.SetMemory(vm.Memory)
	rt.SetMemory(vm.Memory)

	// Get the function ID of the entry function to be executed.
	entryID, ok := vm.GetFunctionExport("run")
	if !ok {
		fmt.Printf("Entry function run not found;\n")
		entryID = 0
	}

	// If any function prior to the entry function was declared to be
	// called by the module, run it first.
	if vm.Module.Base.Start != nil {
		startID := int(vm.Module.Base.Start.Index)
		_, err := vm.Run(startID)
		if err != nil {
			vm.PrintStackTrace()
			panic(err)
		}
	}

	argc, argv := gowasm.PrepareArgs(vm.Memory, flag.Args(), os.Environ())
	// Run the WebAssembly module's entry function.
	_, err = run(vm, entryID, int64(argc), int64(argv), rt)
	if err != nil {
		vm.PrintStackTrace()
		panic(err)
	}
}
