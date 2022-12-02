package tripod

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/core/env"
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
	// Key: Writing Name
	writings map[string]dev.Writing
	// Key: Reading Name
	readings map[string]dev.Reading
	// key: p2p-handler type code
	P2pHandlers map[int]dev.P2pHandler
}

func NewTripod(name string) *Tripod {
	return &Tripod{
		name:        name,
		writings:    make(map[string]dev.Writing),
		readings:    make(map[string]dev.Reading),
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

func (t *Tripod) SetWritings(wrs ...dev.Writing) {
	for _, wr := range wrs {
		name := getFuncName(wr)
		t.writings[name] = wr
		logrus.Debugf("register Writing(%s) into Tripod(%s) \n", name, t.name)
	}
}

func (t *Tripod) SetReadings(readings ...dev.Reading) {
	for _, r := range readings {
		name := getFuncName(r)
		t.readings[name] = r
		logrus.Debugf("register Reading(%s) into Tripod(%s) \n", name, t.name)
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

func (t *Tripod) ExistWriting(name string) bool {
	_, ok := t.writings[name]
	return ok
}

func (t *Tripod) GetWriting(name string) dev.Writing {
	return t.writings[name]
}

func (t *Tripod) GetReading(name string) dev.Reading {
	return t.readings[name]
}

func (t *Tripod) AllReadingNames() []string {
	allNames := make([]string, 0)
	for name, _ := range t.readings {
		allNames = append(allNames, name)
	}
	return allNames
}

func (t *Tripod) AllWritingNames() []string {
	allNames := make([]string, 0)
	for name, _ := range t.writings {
		allNames = append(allNames, name)
	}
	return allNames
}
