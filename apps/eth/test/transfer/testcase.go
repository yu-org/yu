package transfer

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/yu-org/yu/apps/eth/test/pkg"
)

type TestCase interface {
	Run(ctx context.Context, m *pkg.WalletManager) error
	Name() string
}

type RandomTransferTestCase struct {
	CaseName     string
	walletCount  int
	initialCount uint64
	steps        int
	tm           *pkg.TransferManager

	wallets   []*pkg.CaseEthWallet
	transCase *pkg.TransferCase
}

func NewRandomTest(name string, count int, initial uint64, steps int) *RandomTransferTestCase {
	return &RandomTransferTestCase{
		CaseName:     name,
		walletCount:  count,
		initialCount: initial,
		steps:        steps,
		tm:           pkg.NewTransferManager(),
	}
}

func (tc *RandomTransferTestCase) Name() string {
	return tc.CaseName
}

func (tc *RandomTransferTestCase) Run(ctx context.Context, m *pkg.WalletManager) error {
	var wallets []*pkg.EthWallet
	var err error
	wallets, err = m.GenerateRandomWallets(tc.walletCount, tc.initialCount)
	if err != nil {
		return err
	}
	logrus.Infof("%s create wallets finish", tc.CaseName)
	tc.wallets = pkg.GenerateCaseWallets(tc.initialCount, wallets)
	tc.transCase = tc.tm.GenerateRandomTransferSteps(tc.steps, tc.wallets)
	return runAndAssert(tc.transCase, m, wallets)
}

func runAndAssert(transferCase *pkg.TransferCase, m *pkg.WalletManager, wallets []*pkg.EthWallet) error {
	if err := transferCase.Run(m); err != nil {
		return err
	}
	logrus.Info("wait transfer transaction done")
	time.Sleep(5 * time.Second)
	success, err := assert(transferCase, m, wallets)
	if err != nil {
		return err
	}
	if !success {
		return errors.New("transfer manager assert failed")
	}

	bm := pkg.GetDefaultBlockManager()
	block, err := bm.GetCurrentBlock()
	if err != nil {
		return err
	}
	logrus.Infof("Block(%d) StateRoot: %s", block.Height, block.StateRoot.String())
	return nil
}

func assert(transferCase *pkg.TransferCase, walletsManager *pkg.WalletManager, wallets []*pkg.EthWallet) (bool, error) {
	var got map[string]*pkg.CaseEthWallet
	var success bool
	var err error
	for i := 0; i < 20; i++ {
		got, success, err = transferCase.AssertExpect(walletsManager, wallets)
		if err != nil {
			return false, err
		}
		if success {
			return true, nil
		} else {
			// wait block
			time.Sleep(4 * time.Second)
			continue
		}
	}

	printChange(got, transferCase.Expect, transferCase)
	return false, nil
}

func printChange(got, expect map[string]*pkg.CaseEthWallet, transferCase *pkg.TransferCase) {
	for _, step := range transferCase.Steps {
		logrus.Infof("%v transfer %v eth to %v", step.From.Address, step.Count, step.To.Address)
	}
	for k, v := range got {
		ev, ok := expect[k]
		if ok {
			if v.EthCount != ev.EthCount {
				logrus.Infof("%v got:%v expect:%v", k, v.EthCount, ev.EthCount)
			}
		}
	}
}
