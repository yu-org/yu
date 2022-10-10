package tripod

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/core/chain_env"
	"github.com/yu-org/yu/core/tripod/dev"
	. "github.com/yu-org/yu/core/types"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

type Tripod struct {
	*ChainEnv
	*Land

	BlockVerifier
	TxnChecker

	Init
	BlockCycle

	Instance interface{}

	name string
	// Key: Execution Name
	execs map[string]dev.Execution
	// Key: Query Name
	queries map[string]dev.Query
	// key: p2p-handler type code
	P2pHandlers map[int]dev.P2pHandler
}

func NewTripod(name string) *Tripod {
	return &Tripod{
		name:        name,
		execs:       make(map[string]dev.Execution),
		queries:     make(map[string]dev.Query),
		P2pHandlers: make(map[int]dev.P2pHandler),

		BlockVerifier: &DefaultBlockVerifier{},
		TxnChecker:    &DefaultTxnChecker{},
		Init:          &DefaultInit{},
		BlockCycle:    &DefaultBlockCycle{},
	}
}

func (t *Tripod) SetInstance(instance interface{}) {
	t.Instance = instance
}

func (t *Tripod) Name() string {
	return t.name
}

func (t *Tripod) SetChainEnv(env *ChainEnv) {
	t.ChainEnv = env
}

func (t *Tripod) SetLand(land *Land) {
	t.Land = land
}

func (t *Tripod) SetInit(init Init) {
	t.Init = init
}

func (t *Tripod) SetBlockCycle(bc BlockCycle) {
	t.BlockCycle = bc
}

func (t *Tripod) SetBlockVerifier(bv BlockVerifier) {
	t.BlockVerifier = bv
}

func (t *Tripod) SetTxnChecker(tc TxnChecker) {
	t.TxnChecker = tc
}

func (t *Tripod) SetExec(fn dev.Execution) *Tripod {
	name := getFuncName(fn)
	t.execs[name] = fn
	logrus.Debugf("register Execution(%s) into Tripod(%s) \n", name, t.name)
	return t
}

func (t *Tripod) SetQueries(queries ...dev.Query) {
	for _, q := range queries {
		name := getFuncName(q)
		t.queries[name] = q
		logrus.Debugf("register Query(%s) into Tripod(%s) \n", name, t.name)
	}
}

func (t *Tripod) SetP2pHandler(code int, handler dev.P2pHandler) *Tripod {
	t.P2pHandlers[code] = handler
	logrus.Debugf("register P2pHandler(%d) into Tripod(%s) \n", code, t.name)
	return t
}

func getFuncName(i interface{}) string {
	ptr := reflect.ValueOf(i).Pointer()
	nameFull := runtime.FuncForPC(ptr).Name()
	nameEnd := filepath.Ext(nameFull)
	funcName := strings.TrimPrefix(nameEnd, ".")
	return strings.TrimSuffix(funcName, "-fm")
}

func (t *Tripod) ExistExec(execName string) bool {
	_, ok := t.execs[execName]
	return ok
}

func (t *Tripod) GetExec(name string) dev.Execution {
	return t.execs[name]
}

func (t *Tripod) GetQuery(name string) dev.Query {
	return t.queries[name]
}

func (t *Tripod) AllQueryNames() []string {
	allNames := make([]string, 0)
	for name, _ := range t.queries {
		allNames = append(allNames, name)
	}
	return allNames
}

func (t *Tripod) AllExecNames() []string {
	allNames := make([]string, 0)
	for name, _ := range t.execs {
		allNames = append(allNames, name)
	}
	return allNames
}
