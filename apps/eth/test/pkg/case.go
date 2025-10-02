package pkg

import (
	"math/rand"
	"time"
)

type CaseEthWallet struct {
	*EthWallet
	EthCount uint64 `json:"ethCount"`
}

func (c *CaseEthWallet) Copy() *CaseEthWallet {
	return &CaseEthWallet{
		EthWallet: c.EthWallet.Copy(),
		EthCount:  c.EthCount,
	}
}

type TransferManager struct{}

func NewTransferManager() *TransferManager {
	return &TransferManager{}
}

func GenerateCaseWallets(initialEthCount uint64, wallets []*EthWallet) []*CaseEthWallet {
	c := make([]*CaseEthWallet, 0)
	for _, w := range wallets {
		c = append(c, &CaseEthWallet{
			EthWallet: w,
			EthCount:  initialEthCount,
		})
	}
	return c
}

func (m *TransferManager) GenerateRandomTransferSteps(stepCount int, wallets []*CaseEthWallet) *TransferCase {
	t := &TransferCase{
		Original: getCopy(wallets),
		Expect:   getCopy(wallets),
	}
	steps := make([]*Step, 0)
	r := rand.New(rand.NewSource(time.Now().Unix()))
	curTransfer := 1
	for i := 0; i < stepCount; i++ {
		steps = append(steps, generateRandomStep(r, wallets, curTransfer))
		curTransfer++
	}
	t.Steps = steps
	calculateExpect(t)
	return t
}

func (m *TransferManager) GenerateTransferSteps(wallets []*CaseEthWallet) *TransferCase {
	t := &TransferCase{
		Original: getCopy(wallets),
		Expect:   getCopy(wallets),
	}
	steps := make([]*Step, 0)
	curTransfer := 1
	for i := 0; i < len(wallets); i += 2 {
		steps = append(steps, generateStep(wallets[i], wallets[i+1], curTransfer))
		curTransfer++
	}
	t.Steps = steps
	calculateExpect(t)
	return t
}

func (m *TransferManager) GenerateSameTargetTransferSteps(stepCount int, wallets []*CaseEthWallet, target *CaseEthWallet) *TransferCase {
	t := &TransferCase{
		Original: getCopy(wallets),
		Expect:   getCopy(wallets),
	}
	steps := make([]*Step, 0)
	cur := 0
	curTransfer := 1
	for i := 0; i < stepCount; i++ {
		from := wallets[cur]
		steps = append(steps, generateTransferStep(from, target, curTransfer))
		cur++
		if cur >= len(wallets) {
			cur = 0
		}
		curTransfer++
	}
	t.Steps = steps
	calculateExpect(t)
	return t
}

func (tc *TransferCase) Run(m *WalletManager) error {
	nonceMap := make(map[string]uint64)
	for _, step := range tc.Steps {
		if _, ok := nonceMap[step.From.Address]; ok {
			nonceMap[step.From.Address]++
		}
		if err := m.TransferEth(step.From, step.To, step.Count, nonceMap[step.From.Address]); err != nil {
			return err
		}
	}
	return nil
}

func (tc *TransferCase) AssertExpect(m *WalletManager, wallets []*EthWallet) (map[string]*CaseEthWallet, bool, error) {
	got := make(map[string]*CaseEthWallet)
	for _, w := range wallets {
		c, err := m.QueryEth(w)
		if err != nil {
			return nil, false, err
		}
		got[w.Address] = &CaseEthWallet{
			EthWallet: w,
			EthCount:  c,
		}
	}
	if len(tc.Expect) != len(got) {
		return got, false, nil
	}
	for key, value := range got {
		e, ok := tc.Expect[key]
		if !ok {
			return got, false, nil
		}
		if e.EthCount != value.EthCount {
			return got, false, nil
		}
	}
	return got, true, nil
}

func calculateExpect(tc *TransferCase) {
	for _, step := range tc.Steps {
		calculate(step, tc.Expect)
	}
}

func calculate(step *Step, expect map[string]*CaseEthWallet) {
	fromWallet := expect[step.From.Address]
	toWallet := expect[step.To.Address]
	fromWallet.EthCount = fromWallet.EthCount - step.Count
	toWallet.EthCount = toWallet.EthCount + step.Count
	expect[step.From.Address] = fromWallet
	expect[step.To.Address] = toWallet
}

func generateRandomStep(r *rand.Rand, wallets []*CaseEthWallet, transfer int) *Step {
	from := r.Intn(len(wallets))
	to := from + 1
	if to >= len(wallets) {
		to = 0
	}
	return &Step{
		From:  wallets[from].EthWallet,
		To:    wallets[to].EthWallet,
		Count: uint64(transfer),
	}
}

func generateStep(from, to *CaseEthWallet, transfer int) *Step {
	return &Step{
		From:  from.EthWallet,
		To:    to.EthWallet,
		Count: uint64(transfer),
	}
}

func generateTransferStep(from, to *CaseEthWallet, transferCount int) *Step {
	return &Step{
		From:  from.EthWallet,
		To:    to.EthWallet,
		Count: uint64(transferCount),
	}
}

func getCopy(wallets []*CaseEthWallet) map[string]*CaseEthWallet {
	m := make(map[string]*CaseEthWallet)
	for _, w := range wallets {
		m[w.Address] = w.Copy()
	}
	return m
}

type TransferCase struct {
	Steps []*Step `json:"steps"`
	// address to wallet
	Original map[string]*CaseEthWallet `json:"original"`
	Expect   map[string]*CaseEthWallet `json:"expect"`
}

type Step struct {
	From  *EthWallet `json:"from"`
	To    *EthWallet `json:"to"`
	Count uint64     `json:"count"`
}
