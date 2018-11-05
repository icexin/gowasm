// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/validate"
	"github.com/go-interpreter/wagon/wasm"
	"github.com/icexin/gowasm"
)

func main() {
	log.SetPrefix("wasm-run: ")
	log.SetFlags(0)

	verbose := flag.Bool("v", false, "enable/disable verbose mode")
	verify := flag.Bool("verify-module", false, "run module verification")

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	wasm.SetDebugMode(*verbose)

	run(os.Stdout, flag.Arg(0), *verify)
}

func run(w io.Writer, fname string, verify bool) {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	log.Printf(fname)
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

	r.SetMemory(vm.Memory())
	rt.SetMemory(vm.Memory())

	entry := m.Export.Entries["run"]
	entryid := entry.Index

	argc, argv := prepareArgs(vm.Memory())
	log.SetOutput(ioutil.Discard)
	_, err = vm.ExecCode(int64(entryid), uint64(argc), uint64(argv))
	if err != nil {
		panic(err)
	}
	log.Printf("done")
}

func prepareArgs(mem []byte) (int, int) {
	hostArgs := flag.Args()
	argc := len(hostArgs)
	offset := 4096
	strdup := func(s string) int {
		copy(mem[offset:], s+"\x00")
		ptr := offset
		offset += len(s) + (8 - len(s)%8)
		return ptr
	}
	var argvAddr []int
	for _, arg := range hostArgs {
		argvAddr = append(argvAddr, strdup(arg))
	}

	argv := offset
	buf := bytes.NewBuffer(mem[offset:offset])
	for _, addr := range argvAddr {
		binary.Write(buf, binary.LittleEndian, int64(addr))
	}
	return argc, argv
}
