package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	yucommon "github.com/yu-org/yu/common"

	"github.com/yu-org/yu/apps/eth/test/pkg"
)

var resultJson = "stateRootTestResult.json"

type StateRootTestCase struct {
	*RandomTransferTestCase
}

func NewStateRootTestCase(name string, count int, initial uint64, steps int) *StateRootTestCase {
	return &StateRootTestCase{
		RandomTransferTestCase: NewRandomTest(name, count, initial, steps),
	}
}

func (st *StateRootTestCase) Name() string {
	return "StateRootTestCase"
}

func (st *StateRootTestCase) Run(ctx context.Context, m *pkg.WalletManager) error {
	if err := st.RandomTransferTestCase.Run(ctx, m); err != nil {
		return err
	}
	hash, err := getStateRoot()
	if err != nil {
		return err
	}
	result := StateRootTestResult{
		Wallets:      st.wallets,
		TransferCase: st.transCase,
		StateRoot:    hash,
	}
	content, _ := json.Marshal(result)

	if _, err = os.Stat("stateRootTestResult.json"); err == nil {
		if err = os.Remove("stateRootTestResult.json"); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	file, err := os.Create(resultJson)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	_, err = file.Write(content)
	if err != nil {
		return err
	}
	return nil
}

type StateRootTestResult struct {
	Wallets      []*pkg.CaseEthWallet `json:"wallets"`
	TransferCase *pkg.TransferCase    `json:"transferCase"`
	StateRoot    yucommon.Hash        `json:"stateRoot"`
}

func getStateRoot() (yucommon.Hash, error) {
	b, err := pkg.GetDefaultBlockManager().GetBlockByIndex(3)
	if err != nil {
		return [32]byte{}, err
	}
	return b.StateRoot, nil
}

type StateRootAssertTestCase struct {
	content []byte
	initial uint64
}

func NewStateRootAssertTestCase(content []byte, initial uint64) *StateRootAssertTestCase {
	return &StateRootAssertTestCase{content: content, initial: initial}
}

func (s *StateRootAssertTestCase) Run(ctx context.Context, m *pkg.WalletManager) error {
	result := &StateRootTestResult{}
	if err := json.Unmarshal(s.content, result); err != nil {
		return err
	}
	var lastWallet *pkg.EthWallet
	var err error
	for _, wallet := range result.Wallets {
		lastWallet, err = m.CreateEthWalletByAddress(s.initial, wallet.PK, wallet.Address)
		if err != nil {
			return err
		}
	}
	m.AssertWallet(lastWallet, s.initial)
	if err := runAndAssert(result.TransferCase, m, getWallets(result.Wallets)); err != nil {
		return err
	}
	stateRoot, err := getStateRoot()
	if err != nil {
		return err
	}
	if result.StateRoot != stateRoot {
		return fmt.Errorf("expected stateRoot %s, got %s", stateRoot.String(), result.StateRoot.String())
	}
	return nil
}

func (s *StateRootAssertTestCase) Name() string {
	return "StateRootAssertTestCase"
}

func getWallets(ws []*pkg.CaseEthWallet) []*pkg.EthWallet {
	got := make([]*pkg.EthWallet, 0)
	for _, w := range ws {
		got = append(got, w.EthWallet)
	}
	return got
}
