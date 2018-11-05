package js

import (
	"fmt"
	"unsafe"
)

const (
	nanHead = 0x7FF80000
)

const (
	ValueNaN Ref = nanHead<<32 | iota
	ValueUndefined
	ValueNull
	ValueTrue
	ValueFalse
	ValueGlobal
	ValueMemory
	ValueGo
)

type Ref int64

func (r Ref) Number() (int64, bool) {
	f := *(*float64)(unsafe.Pointer(&r))
	if f == f {
		return int64(f), true
	}
	return 0, false
}

func (r Ref) ID() int64 {
	id := uint32(r)
	return int64(id)
}

func (r Ref) String() string {
	n, ok := r.Number()
	if ok {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("0x%x", int64(r))
}
