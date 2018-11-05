package gowasm

import (
	"bytes"
	"encoding/binary"
)

func PrepareArgs(mem []byte, args []string, envs []string) (int, int) {
	argc := len(args)
	offset := 4096
	strdup := func(s string) int {
		copy(mem[offset:], s+"\x00")
		ptr := offset
		offset += len(s) + (8 - len(s)%8)
		return ptr
	}
	var argvAddr []int
	for _, arg := range args {
		argvAddr = append(argvAddr, strdup(arg))
	}

	argvAddr = append(argvAddr, len(envs))
	for _, env := range envs {
		argvAddr = append(argvAddr, strdup(env))
	}

	argv := offset
	buf := bytes.NewBuffer(mem[offset:offset])
	for _, addr := range argvAddr {
		binary.Write(buf, binary.LittleEndian, int64(addr))
	}
	return argc, argv
}
