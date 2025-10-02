package pkg

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/yu-org/yu/apps/eth/test/contracts"
)

const (
	nodeUrl                 = "http://localhost:9092"
	waitForConfirmationTime = 1 * time.Second
	maxRetries              = 300
)

type ERC20Wallet struct {
	PK         string         `json:"pk"`
	Address    common.Address `json:"address"`
	Balance    uint64         `json:"balance"`
	TokenCount uint64         `json:"tokenCount"`
}

func (w *ERC20Wallet) Copy() *ERC20Wallet {
	return &ERC20Wallet{
		PK:         w.PK,
		Address:    w.Address,
		Balance:    w.Balance,
		TokenCount: w.TokenCount,
	}
}

type Erc20TransferManager struct {
	ContractAddr common.Address
}

func NewErc20TransferManager(contractAddr common.Address) *Erc20TransferManager {
	return &Erc20TransferManager{
		ContractAddr: contractAddr,
	}
}

func GenerateERC20Wallets(initialTokenCount uint64, wallets []*ERC20Wallet) []*ERC20Wallet {
	for _, w := range wallets {
		w.TokenCount = initialTokenCount
	}
	return wallets
}

func (m *Erc20TransferManager) GenerateRandomErc20TransferSteps(stepCount int, wallets []*ERC20Wallet, contractAddress common.Address, chainID int64) *Erc20TransferCase {
	t := &Erc20TransferCase{
		ChainID:      chainID,
		Original:     getCopyERC20(wallets),
		Expect:       getCopyERC20(wallets),
		ContractAddr: contractAddress,
	}
	steps := make([]*Erc20Step, 0)
	r := rand.New(rand.NewSource(time.Now().Unix()))
	curTransfer := 1
	for i := 0; i < stepCount; i++ {
		steps = append(steps, generateRandomERC20Step(r, wallets, m.ContractAddr, curTransfer))
		curTransfer++
	}
	t.Steps = steps
	calculateExpectERC20(t)
	return t
}

func (m *Erc20TransferManager) GenerateErc20TransferSteps(wallets []*ERC20Wallet) *Erc20TransferCase {
	t := &Erc20TransferCase{
		Original: getCopyERC20(wallets),
		Expect:   getCopyERC20(wallets),
	}
	steps := make([]*Erc20Step, 0)
	curTransfer := 1
	for i := 0; i < len(wallets); i += 2 {
		steps = append(steps, generateERC20Step(wallets[i], wallets[i+1], m.ContractAddr, curTransfer))
		curTransfer++
	}
	t.Steps = steps
	calculateExpectERC20(t)
	return t
}

func (m *Erc20TransferManager) GenerateSameTargetErc20TransferSteps(stepCount int, wallets []*ERC20Wallet, target *ERC20Wallet) *Erc20TransferCase {
	t := &Erc20TransferCase{
		Original: getCopyERC20(wallets),
		Expect:   getCopyERC20(wallets),
	}
	steps := make([]*Erc20Step, 0)
	cur := 0
	curTransfer := 1
	for i := 0; i < stepCount; i++ {
		from := wallets[cur]
		steps = append(steps, generateERC20TransferStep(from, target, m.ContractAddr, curTransfer))
		cur++
		if cur >= len(wallets) {
			cur = 0
		}
		curTransfer++
	}
	t.Steps = steps
	calculateExpectERC20(t)
	return t
}

func (tc *Erc20TransferCase) Run(m *WalletManager) error {
	nonceMap := make(map[string]uint64)
	client, err := ethclient.Dial(nodeUrl)
	if err != nil {
		return fmt.Errorf("Failed to connect to the Ethereum client: %v", err)
	}

	defer client.Close()

	for _, step := range tc.Steps {
		fromAddress := step.From.Address.Hex()
		toAddress := step.To.Address.Hex()
		//contractAddress := step.ContractAddr.Hex()

		if _, ok := nonceMap[fromAddress]; ok {
			nonceMap[fromAddress]++
		}
		privateKey, err := crypto.HexToECDSA(step.From.PK)
		if err != nil {
			return err
		}

		ownerAuth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(tc.ChainID))
		if err != nil {
			return err
		}

		amount := new(big.Int).SetUint64(step.Count)

		// Get suggested gas price to ensure it's >= base fee
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			// Fallback to a reasonable gas price if suggestion fails
			gasPrice = big.NewInt(2e9) // 2 gwei
		}
		ownerAuth.GasPrice = gasPrice
		ownerAuth.GasLimit = uint64(6e7)

		if err := tc.TransferERC20(client, step.ContractAddr, *ownerAuth, fromAddress, toAddress, amount); err != nil {
			return err
		}
	}
	return nil
}

func (tc *Erc20TransferCase) TransferERC20(client *ethclient.Client, contractAddress common.Address, ownerAuth bind.TransactOpts, fromAddress string, toAddress string, amount *big.Int) error {
	var lastTxHash common.Hash

	tokenInstance, err := contracts.NewToken(contractAddress, client)
	if err != nil {
		return err
	}

	toAddr := common.HexToAddress(toAddress)
	tx, err := tokenInstance.Transfer(&ownerAuth, toAddr, amount)
	if err != nil {
		return err
	}

	lastTxHash = tx.Hash()
	isConfirmed, err := waitForConfirmation(client, lastTxHash)
	if err != nil {
		return err
	}
	if !isConfirmed {
		return fmt.Errorf("transaction %s was not confirmed", lastTxHash.Hex())
	}

	// callOpts := &bind.CallOpts{
	// 	Pending: false,
	// 	Context: context.Background(),
	// }
	//balance, err := tokenInstance.BalanceOf(callOpts, toAddr)
	if err != nil {
		return err
	}

	return nil
}

func (tc *Erc20TransferCase) QueryERC20(contractAddress common.Address, address string) (uint64, error) {
	client, err := ethclient.Dial(nodeUrl)
	if err != nil {
		return 0, fmt.Errorf("Failed to connect to the Ethereum client: %v", err)
	}

	defer client.Close()

	tokenInstance, err := contracts.NewToken(contractAddress, client)
	if err != nil {
		return 0, err
	}

	toAddr := common.HexToAddress(address)

	callOpts := &bind.CallOpts{
		Pending: false,
		Context: context.Background(),
	}
	balance, err := tokenInstance.BalanceOf(callOpts, toAddr)
	if err != nil {
		return 0, err
	}

	return balance.Uint64(), nil
}

func waitForConfirmation(client *ethclient.Client, txHash common.Hash) (bool, error) {
	for i := 0; i < maxRetries; i++ {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			if receipt.Status == types.ReceiptStatusSuccessful {
				return true, nil
			}
			return false, fmt.Errorf("transaction failed with status: %v", receipt.Status)
		}
		time.Sleep(waitForConfirmationTime)
	}
	return false, fmt.Errorf("transaction was not confirmed after %d retries", maxRetries)
}

func (tc *Erc20TransferCase) AssertExpect(m *WalletManager, wallets []*ERC20Wallet) (map[string]*ERC20Wallet, bool, error) {
	got := make(map[string]*ERC20Wallet)
	for _, w := range wallets {
		addressStr := w.Address.Hex()

		c, err := tc.QueryERC20(tc.ContractAddr, addressStr)
		if err != nil {
			return nil, false, err
		}

		got[addressStr] = &ERC20Wallet{
			PK:         w.PK,
			Address:    w.Address,
			Balance:    w.Balance,
			TokenCount: c,
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
		if e.Balance != value.Balance {
			return got, false, nil
		}
	}
	return got, true, nil
}

func calculateExpectERC20(tc *Erc20TransferCase) {
	for _, step := range tc.Steps {
		calculateERC20(step, tc.Expect)
	}
}

func calculateERC20(step *Erc20Step, expect map[string]*ERC20Wallet) {
	fromAddress := step.From.Address.Hex()
	toAddress := step.To.Address.Hex()

	fromWallet := expect[fromAddress]
	toWallet := expect[toAddress]

	fromWallet.TokenCount = fromWallet.TokenCount - step.Count
	toWallet.TokenCount = toWallet.TokenCount + step.Count
	expect[fromAddress] = fromWallet
	expect[toAddress] = toWallet
}

func generateRandomERC20Step(r *rand.Rand, wallets []*ERC20Wallet, contractAddr common.Address, transfer int) *Erc20Step {
	from := r.Intn(len(wallets))
	to := from + 1
	if to >= len(wallets) {
		to = 0
	}
	return &Erc20Step{
		From:         wallets[from],
		To:           wallets[to],
		Count:        uint64(transfer),
		ContractAddr: contractAddr,
	}
}

func generateERC20Step(from, to *ERC20Wallet, contractAddr common.Address, transfer int) *Erc20Step {
	return &Erc20Step{
		From:         from,
		To:           to,
		Count:        uint64(transfer),
		ContractAddr: contractAddr,
	}
}

func generateERC20TransferStep(from, to *ERC20Wallet, contractAddr common.Address, transferCount int) *Erc20Step {
	return &Erc20Step{
		From:         from,
		To:           to,
		Count:        uint64(transferCount),
		ContractAddr: contractAddr,
	}
}

func getCopyERC20(wallets []*ERC20Wallet) map[string]*ERC20Wallet {
	m := make(map[string]*ERC20Wallet)
	for _, w := range wallets {
		addressStr := w.Address.Hex()
		m[addressStr] = w.Copy()
	}
	return m
}

type Erc20TransferCase struct {
	ChainID int64
	Steps   []*Erc20Step `json:"steps"`
	// address to wallet
	Original     map[string]*ERC20Wallet `json:"original"`
	Expect       map[string]*ERC20Wallet `json:"expect"`
	ContractAddr common.Address          `json:"contractAddress"`
}

type Erc20Step struct {
	From         *ERC20Wallet   `json:"from"`
	To           *ERC20Wallet   `json:"to"`
	Count        uint64         `json:"count"`
	ContractAddr common.Address `json:"contractAddress"`
}
