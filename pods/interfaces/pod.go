package interfaces

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	. "yu/common"
)

type Pod interface {

	PodHeader() *PodHeader

	OnInitialize(blockNum BlockNum) error

	OnFinalize(blockNum BlockNum) error

}

type PodHeader struct {
	name string
	callFns map[string]Execution
}

func NewPodHeader(name string) *PodHeader {
	return &PodHeader{
		name: name,
		callFns: make(map[string]Execution),
	}
}

func (ph *PodHeader) Name() string {
	return ph.name
}

func(ph *PodHeader) SetCallFns(fns ...Execution) {
	for _, fn := range fns {
		ptr := reflect.ValueOf(fn).Pointer()
		nameFull := runtime.FuncForPC(ptr).Name()
		nameEnd:= filepath.Ext(nameFull)
		name := strings.TrimPrefix(nameEnd, ".")
		ph.callFns[name] = fn
		fmt.Printf("register CallFn (%s) into PodHeader \n", name)
	}
}