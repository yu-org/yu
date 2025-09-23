package erc20

import (
	"context"
	"fmt"
	"log"

	"github.com/yu-org/yu/apps/eth/test/conf"
	"github.com/yu-org/yu/apps/eth/test/pkg"
)

type EthManager struct {
	config *conf.EthCaseConf
	wm     *pkg.WalletManager
	// tm     *pkg.TransferManager
	testcases []TestCase
}

func (m *EthManager) Configure(cfg *conf.EthCaseConf, nodeUrl, pk string, chainID int64) {
	m.config = cfg
	m.wm = pkg.NewWalletManager(chainID, nodeUrl, pk)
	m.testcases = []TestCase{}
}

func (m *EthManager) PreCreateWallets(walletCount int, initCount uint64) ([]*pkg.EthWallet, error) {
	wallets, err := m.wm.GenerateRandomWallets(walletCount, initCount)
	if err != nil {
		return nil, err
	}
	return wallets, nil
}

func (m *EthManager) AddTestCase(tc ...TestCase) {
	m.testcases = append(m.testcases, tc...)
}

func (m *EthManager) Run(ctx context.Context) error {
	for _, tc := range m.testcases {
		log.Println(fmt.Sprintf("start to test %v", tc.Name()))
		if err := tc.Run(ctx, m.wm); err != nil {
			return fmt.Errorf("%s failed, err:%v", tc.Name(), err)
		}
		log.Println(fmt.Sprintf("test %v success", tc.Name()))
	}
	return nil
}

func (m *EthManager) Prepare(ctx context.Context) error {
	// for _, tc := range m.testcases {
	// 	log.Println(fmt.Sprintf("start to prepare %v", tc.Name()))
	// 	if _, err := tc.Prepare(ctx, m.wm); err != nil {
	// 		return fmt.Errorf("%s failed, err:%v", tc.Name(), err)
	// 	}
	// 	log.Println(fmt.Sprintf("prepare %v success", tc.Name()))
	// }
	return nil
}
