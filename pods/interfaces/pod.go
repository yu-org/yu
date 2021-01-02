package interfaces

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	. "yu/blockchain"
	. "yu/common"
)

type Pod interface {
	PodHeader() *PodHeader

	OnInitialize(block *Block) error

	OnFinalize(block *Block) error
}

type PodHeader struct {
	name   string
	exeFns map[string]Execution
}

func NewPodHeader(name string) *PodHeader {
	return &PodHeader{
		name:   name,
		exeFns: make(map[string]Execution),
	}
}

func (ph *PodHeader) Name() string {
	return ph.name
}

func (ph *PodHeader) SetExecFns(fns ...Execution) {
	for _, fn := range fns {
		ptr := reflect.ValueOf(fn).Pointer()
		nameFull := runtime.FuncForPC(ptr).Name()
		nameEnd := filepath.Ext(nameFull)
		name := strings.TrimPrefix(nameEnd, ".")
		ph.exeFns[name] = fn
		fmt.Printf("register CallFn (%s) into PodHeader \n", name)
	}
}

func(ph *PodHeader) GetExecFn(name string) Execution {
	return ph.exeFns[name]
}