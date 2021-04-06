package tripod

import (
	"github.com/sirupsen/logrus"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	. "yu/tripod_store"
)

type TripodMeta struct {
	*TripodStore

	name string
	// Key: Execution Name
	execs map[string]Execution
	// Key: Query Name
	queries map[string]Query
}

func NewTripodMeta(name string) *TripodMeta {
	return &TripodMeta{
		name:    name,
		execs:   make(map[string]Execution),
		queries: make(map[string]Query),
	}
}

func (th *TripodMeta) Name() string {
	return th.name
}

func (th *TripodMeta) SetExecs(fns ...Execution) {
	for _, fn := range fns {
		ptr := reflect.ValueOf(fn).Pointer()
		nameFull := runtime.FuncForPC(ptr).Name()
		nameEnd := filepath.Ext(nameFull)
		name := strings.TrimPrefix(nameEnd, ".")
		th.execs[name] = fn
		logrus.Infof("register Execution (%s) into TripodMeta \n", name)
	}
}

func (th *TripodMeta) SetQueries(queries ...Query) {
	for _, q := range queries {
		ptr := reflect.ValueOf(q).Pointer()
		nameFull := runtime.FuncForPC(ptr).Name()
		nameEnd := filepath.Ext(nameFull)
		name := strings.TrimPrefix(nameEnd, ".")
		th.queries[name] = q
		logrus.Infof("register Query (%s) into TripodMeta \n", name)
	}
}

func (th *TripodMeta) ExistExec(execName string) bool {
	_, ok := th.execs[execName]
	return ok
}

func (th *TripodMeta) GetExec(name string) Execution {
	return th.execs[name]
}

func (th *TripodMeta) GetQuery(name string) Query {
	return th.queries[name]
}

func (th *TripodMeta) AllQueryNames() []string {
	allNames := make([]string, 0)
	for name, _ := range th.queries {
		allNames = append(allNames, name)
	}
	return allNames
}

func (th *TripodMeta) AllExecNames() []string {
	allNames := make([]string, 0)
	for name, _ := range th.execs {
		allNames = append(allNames, name)
	}
	return allNames
}
