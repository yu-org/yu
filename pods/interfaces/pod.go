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

	Name() string

	PodHeader() *PodHeader

	OnInitialize(blockNum BlockNum) error

	OnFinalize(blockNum BlockNum) error

}

type PodHeader struct {
	callFns map[string]Execution
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