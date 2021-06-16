package tripod

import (
	"github.com/sirupsen/logrus"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

type TripodMeta struct {
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

func (t *TripodMeta) Name() string {
	return t.name
}

func (t *TripodMeta) SetExecs(fns ...Execution) {
	for _, fn := range fns {
		name := getFuncName(fn)
		t.execs[name] = fn
		logrus.Infof("register Execution(%s) into Tripod(%s) \n", name, t.name)
	}
}

func (t *TripodMeta) SetQueries(queries ...Query) {
	for _, q := range queries {
		name := getFuncName(q)
		t.queries[name] = q
		logrus.Infof("register Query(%s) into Tripod(%s) \n", name, t.name)
	}
}

func getFuncName(i interface{}) string {
	ptr := reflect.ValueOf(i).Pointer()
	nameFull := runtime.FuncForPC(ptr).Name()
	nameEnd := filepath.Ext(nameFull)
	funcName := strings.TrimPrefix(nameEnd, ".")
	return strings.TrimSuffix(funcName, "-fm")
}

func (t *TripodMeta) ExistExec(execName string) bool {
	_, ok := t.execs[execName]
	return ok
}

func (t *TripodMeta) GetExec(name string) Execution {
	return t.execs[name]
}

func (t *TripodMeta) GetQuery(name string) Query {
	return t.queries[name]
}

func (t *TripodMeta) AllQueryNames() []string {
	allNames := make([]string, 0)
	for name, _ := range t.queries {
		allNames = append(allNames, name)
	}
	return allNames
}

func (t *TripodMeta) AllExecNames() []string {
	allNames := make([]string, 0)
	for name, _ := range t.execs {
		allNames = append(allNames, name)
	}
	return allNames
}
