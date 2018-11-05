// Package gowasm implements the wasm go runtime
package gowasm

import (
	"crypto/rand"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/icexin/gowasm/js"
	_ "github.com/icexin/gowasm/js/fs"
)

var (
	logger = log.New(ioutil.Discard, "gowasm", log.LstdFlags)
)

// Runtime implements the runtime needed to run wasm code compiled by go toolchain
type Runtime struct {
	exited bool
	mem    []byte
	jsvm   *js.VM

	timeOrigin time.Time
	timerid    int32
	timers     map[int32]*time.Timer
	wakeupch   chan int32
}

func NewRuntime() *Runtime {
	rt := &Runtime{
		jsvm:       js.NewVM(),
		timeOrigin: time.Now(),
		timers:     make(map[int32]*time.Timer),
		wakeupch:   make(chan int32, 1000),
	}
	return rt
}

// SetMemory set vm's memory
func (rt *Runtime) SetMemory(b []byte) {
	rt.jsvm.SetMemory(b)
	rt.mem = b
}

func (rt *Runtime) wasmExit(code int32) {
	rt.exited = true
}

func (rt *Runtime) wasmWrite(fd int64, p int64, n int32) {
	os.Stderr.Write(rt.mem[p : p+int64(n)])
}

func (rt *Runtime) nanotime() int64 {
	return int64(time.Since(rt.timeOrigin).Nanoseconds())
}

func (rt *Runtime) walltime() (int64, int32) {
	nsec := time.Now().UnixNano()
	secs := nsec / 1e9
	nsec = nsec - (secs * 1e9)
	return secs, int32(nsec)
}

// Exited will be true if runtime.wasmExit has been called
func (rt *Runtime) Exited() bool {
	return rt.exited
}

// WaitTimer waiting for timeout of timers set by go runtime in wasm
func (rt *Runtime) WaitTimer() {
	<-rt.wakeupch

}

func (rt *Runtime) scheduleCallback(delay int64) int32 {
	rt.timerid++
	id := rt.timerid
	rt.timers[id] = time.AfterFunc(time.Millisecond*time.Duration(delay+1), func() {
		rt.wakeupch <- id
	})
	return id
}

func (rt *Runtime) clearScheduleCallback(id int32) {
	timer, ok := rt.timers[id]
	if !ok {
		return
	}
	timer.Stop()
	delete(rt.timers, id)
}

func (rt *Runtime) getRandomData(r []byte) {
	rand.Read(r)
}

func (rt *Runtime) debug(v int64) {
	log.Print(v)
}

func (rt *Runtime) syscallJsValueGet(ref js.Ref, name string) js.Ref {
	ret := rt.jsvm.Property(ref, name)
	return ret
}

func (rt *Runtime) syscallJsValueSet(ref js.Ref, name string, value js.Ref) {

}

func (rt *Runtime) syscallJsValueNew(ref js.Ref, args []js.Ref) (ret js.Ref, ok bool) {
	defer func() {
		err := recover()
		if err != nil {
			ret = rt.jsvm.Exception(err.(error))
			ok = false
		}
	}()

	ret, err := rt.jsvm.New(ref, args)
	if err != nil {
		return rt.jsvm.Exception(err), false
	}
	return ret, true
}

func (rt *Runtime) syscallJsValueCall(ref js.Ref, method string, args []js.Ref) (ret js.Ref, ok bool) {
	defer func() {
		err := recover()
		if err != nil {
			ret = rt.jsvm.Exception(err.(error))
			ok = false
		}
	}()

	ret, err := rt.jsvm.Call(ref, method, args)
	if err != nil {
		return rt.jsvm.Exception(err), false
	}
	return ret, true
}

func (rt *Runtime) syscallJsValuePrepareString(ref js.Ref) (js.Ref, int64) {
	v := rt.jsvm.Value(ref)
	if v == nil {
		return js.ValueUndefined, 0
	}
	str := v.String()
	return rt.jsvm.Store(str), int64(len(str))
}

func (rt *Runtime) syscallJsValueLoadString(ref js.Ref, b []byte) {
	v := rt.jsvm.Value(ref)
	if v == nil {
		return
	}
	str := v.String()
	copy(b, str)
}

func (rt *Runtime) syscallJsStringVal(value string) js.Ref {
	return rt.jsvm.Store(value)
}

// Register register the go runtime functions to Registry
func (rt *Runtime) Register(r Registry) {
	r.Register("go", "runtime.wasmExit", rt.wasmExit)
	r.Register("go", "runtime.wasmWrite", rt.wasmWrite)
	r.Register("go", "runtime.nanotime", rt.nanotime)
	r.Register("go", "runtime.walltime", rt.walltime)
	r.Register("go", "runtime.scheduleCallback", rt.scheduleCallback)
	r.Register("go", "runtime.clearScheduledCallback", rt.clearScheduleCallback)
	r.Register("go", "runtime.getRandomData", rt.getRandomData)
	r.Register("go", "runtime.debug", rt.debug)
	r.Register("go", "syscall/js.valueGet", rt.syscallJsValueGet)
	r.Register("go", "syscall/js.valueSet", rt.syscallJsValueSet)
	r.Register("go", "syscall/js.valueNew", rt.syscallJsValueNew)
	r.Register("go", "syscall/js.valuePrepareString", rt.syscallJsValuePrepareString)
	r.Register("go", "syscall/js.valueCall", rt.syscallJsValueCall)
	r.Register("go", "syscall/js.stringVal", rt.syscallJsStringVal)
	r.Register("go", "syscall/js.valueLoadString", rt.syscallJsValueLoadString)
}
