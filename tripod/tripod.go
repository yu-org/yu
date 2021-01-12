package tripod

import (
	"github.com/sirupsen/logrus"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	. "yu/blockchain"
	. "yu/common"
)

type Tripod interface {
	TripodMeta() *TripodMeta

	OnInitialize() (IBlock, error)

	OnFinalize(block IBlock) error
}

type TripodMeta struct {
	name   string
	exeFns map[string]Execution
}

func NewTripodMeta(name string) *TripodMeta {
	return &TripodMeta{
		name:   name,
		exeFns: make(map[string]Execution),
	}
}

func (ph *TripodMeta) Name() string {
	return ph.name
}

func (ph *TripodMeta) SetExecFns(fns ...Execution) {
	for _, fn := range fns {
		ptr := reflect.ValueOf(fn).Pointer()
		nameFull := runtime.FuncForPC(ptr).Name()
		nameEnd := filepath.Ext(nameFull)
		name := strings.TrimPrefix(nameEnd, ".")
		ph.exeFns[name] = fn
		logrus.Infof("register CallFn (%s) into TripodMeta \n", name)
	}
}

func (ph *TripodMeta) GetExecFn(name string) Execution {
	return ph.exeFns[name]
}
