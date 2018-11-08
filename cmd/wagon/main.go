// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/validate"
	"github.com/go-interpreter/wagon/wasm"
	"github.com/icexin/gowasm"
)

var (
	verbose    = flag.Bool("v", false, "enable/disable verbose mode")
	verify     = flag.Bool("verify-module", false, "run module verification")
	cpuprofile = flag.String("cpuprofile", "cpu.pprof", "write cpu profile to file")
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	if *cpuprofile != "" {
		pf, _ := os.Create("cpu.profile")
		defer pf.Close()
		pprof.StartCPUProfile(pf)
		defer pprof.StopCPUProfile()
	}

	wasm.SetDebugMode(*verbose)

	run(flag.Arg(0), *verify)
}

func run(fname string, verify bool) {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := gowasm.NewResolver()
	rt := gowasm.NewRuntime()
	rt.Register(r)

	m, err := wasm.ReadModule(f, func(name string) (*wasm.Module, error) {
		if name == "go" {
			return gomodule(r), nil
		}
		return nil, fmt.Errorf("module %s not found", name)
	})

	if err != nil {
		log.Fatalf("could not read module: %v", err)
	}

	if verify {
		err = validate.VerifyModule(m)
		if err != nil {
			log.Fatalf("could not verify module: %v", err)
		}
	}

	if m.Export == nil {
		log.Fatalf("module has no export section")
	}

	vm, err := exec.NewVM(m)
	if err != nil {
		log.Fatalf("could not create VM: %v", err)
	}

	rt.SetMemory(vm.Memory())

	entry := m.Export.Entries["run"]
	entryid := entry.Index

	argc, argv := gowasm.PrepareArgs(vm.Memory(), flag.Args(), os.Environ())
	for !rt.Exited() {
		_, err = vm.ExecCode(int64(entryid), uint64(argc), uint64(argv))
		if err != nil {
			panic(err)
		}
		if !rt.Exited() {
			rt.WaitTimer()
		}
	}
}
