package tripod

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/core/chain_env"
	"github.com/yu-org/yu/core/tripod/dev"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

type TripodHeader struct {
	*ChainEnv
	*Land

	name string
	// Key: Execution Name
	execs map[string]ExecAndLei
	// Key: Query Name
	queries map[string]dev.Query
	// key: p2p-handler type code
	P2pHandlers map[int]dev.P2pHandler
}

func NewTripodHeader(name string) *TripodHeader {
	return &TripodHeader{
		name:        name,
		execs:       make(map[string]ExecAndLei),
		queries:     make(map[string]dev.Query),
		P2pHandlers: make(map[int]dev.P2pHandler),
	}
}

func (t *TripodHeader) Name() string {
	return t.name
}

func (t *TripodHeader) SetChainEnv(env *ChainEnv) {
	t.ChainEnv = env
}

func (t *TripodHeader) SetLand(land *Land) {
	t.Land = land
}

func (t *TripodHeader) SetExec(fn dev.Execution, lei uint64) *TripodHeader {
	name := getFuncName(fn)
	t.execs[name] = ExecAndLei{
		exec: fn,
		lei:  lei,
	}
	logrus.Infof("register Execution(%s) into Tripod(%s) \n", name, t.name)
	return t
}

func (t *TripodHeader) SetQueries(queries ...dev.Query) {
	for _, q := range queries {
		name := getFuncName(q)
		t.queries[name] = q
		logrus.Infof("register Query(%s) into Tripod(%s) \n", name, t.name)
	}
}

func (t *TripodHeader) SetP2pHandler(code int, handler dev.P2pHandler) *TripodHeader {
	t.P2pHandlers[code] = handler
	logrus.Infof("register P2pHandler(%d) into Tripod(%s) \n", code, t.name)
	return t
}

func getFuncName(i interface{}) string {
	ptr := reflect.ValueOf(i).Pointer()
	nameFull := runtime.FuncForPC(ptr).Name()
	nameEnd := filepath.Ext(nameFull)
	funcName := strings.TrimPrefix(nameEnd, ".")
	return strings.TrimSuffix(funcName, "-fm")
}

func (t *TripodHeader) ExistExec(execName string) bool {
	_, ok := t.execs[execName]
	return ok
}

func (t *TripodHeader) GetExec(name string) (dev.Execution, uint64) {
	execEne, ok := t.execs[name]
	if ok {
		return execEne.exec, execEne.lei
	}
	return nil, 0
}

func (t *TripodHeader) GetQuery(name string) dev.Query {
	return t.queries[name]
}

func (t *TripodHeader) AllQueryNames() []string {
	allNames := make([]string, 0)
	for name, _ := range t.queries {
		allNames = append(allNames, name)
	}
	return allNames
}

func (t *TripodHeader) AllExecNames() []string {
	allNames := make([]string, 0)
	for name, _ := range t.execs {
		allNames = append(allNames, name)
	}
	return allNames
}

type ExecAndLei struct {
	exec dev.Execution
	lei  uint64
}
