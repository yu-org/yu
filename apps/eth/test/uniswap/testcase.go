package uniswap

import (
	"context"

	"github.com/yu-org/yu/apps/eth/test/pkg"
)

type TestCase interface {
	Prepare(ctx context.Context, m *pkg.WalletManager) error
	Run(ctx context.Context, m *pkg.WalletManager) error
	Name() string
}
