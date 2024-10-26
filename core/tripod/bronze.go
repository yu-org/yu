package tripod

import "github.com/yu-org/yu/core/env"

type Bronze struct {
	*env.ChainEnv
	*Land

	name string
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
