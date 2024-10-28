package tripod

import (
	"github.com/yu-org/yu/core/env"
	"reflect"
	"strings"
)

type Bronze struct {
	*env.ChainEnv
	*Land

	name string

	Instance any
}

func NewBronze() *Bronze {
	return NewBronzeWithName("")
}

func NewBronzeWithName(name string) *Bronze {
	return &Bronze{name: name}
}

func (b *Bronze) Name() string {
	return b.name
}

func (b *Bronze) SetChainEnv(env *env.ChainEnv) {
	b.ChainEnv = env
}

func (b *Bronze) SetLand(land *Land) {
	b.Land = land
}

func (b *Bronze) SetInstance(bronzeInstance any) {
	if b.name == "" {
		pkgStruct := reflect.TypeOf(bronzeInstance).String()
		strArr := strings.Split(pkgStruct, ".")
		bronzeName := strings.ToLower(strArr[len(strArr)-1])
		b.name = bronzeName
	}
	b.Instance = bronzeInstance
}
