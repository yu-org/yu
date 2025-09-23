package transfer

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github.com/yu-org/yu/apps/eth/test/pkg"
)

type RandomBenchmarkTest struct {
	CaseName     string
	initialCount uint64
	tm           *pkg.TransferManager
	wallets      []*pkg.EthWallet
	rm           *rate.Limiter
}

func NewRandomBenchmarkTest(name string, initial uint64, wallets []*pkg.EthWallet, rm *rate.Limiter) *RandomBenchmarkTest {
	return &RandomBenchmarkTest{
		CaseName:     name,
		initialCount: initial,
		tm:           pkg.NewTransferManager(),
		wallets:      wallets,
		rm:           rm,
	}
}

func (tc *RandomBenchmarkTest) Name() string {
	return tc.CaseName
}

func (tc *RandomBenchmarkTest) Run(ctx context.Context, m *pkg.WalletManager) error {
	transferCase := tc.tm.GenerateTransferSteps(pkg.GenerateCaseWallets(tc.initialCount, tc.wallets))
	for i, step := range transferCase.Steps {
		if err := tc.rm.Wait(ctx); err != nil {
			return err
		}
		if err := m.TransferEth(step.From, step.To, step.Count, uint64(i)+uint64(time.Now().UnixNano())); err != nil {
			logrus.Error("Failed to transfer step: from:%v, to:%v", step.From, step.To)
		}
	}
	return nil
}
