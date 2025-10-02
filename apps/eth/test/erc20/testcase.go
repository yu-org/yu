package erc20

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/yu-org/yu/apps/eth/test/contracts"
	"github.com/yu-org/yu/apps/eth/test/pkg"
)

const (
	accountInitialFunds      = 1e18
	waitForConfirmationTime  = 1 * time.Second
	maxRetries               = 300
	accountInitialERC20Token = 1e18
)

type ERC20DeployedContract struct {
	tokenAddress     common.Address
	tokenTransaction *types.Transaction
	tokenInstance    *contracts.Token
}

type TestCase interface {
	Prepare(ctx context.Context, m *pkg.WalletManager, client *ethclient.Client, wallets []*pkg.ERC20Wallet) (common.Address, error)
	Run(ctx context.Context, m *pkg.WalletManager) error
	Name() string
}

type TestData struct {
	TestContracts common.Address
}

type RandomTransferTestCase struct {
	nodeURL      string
	ChainID      int64
	CaseName     string
	walletCount  int
	initialCount uint64
	steps        int
	tm           *pkg.Erc20TransferManager

	wallets      []*pkg.CaseEthWallet
	transCase    *pkg.Erc20TransferCase
	erc20Wallets []*pkg.ERC20Wallet
}

func NewRandomTest(name, nodeUrl string, count int, initial uint64, steps int, chainID int64) *RandomTransferTestCase {
	return &RandomTransferTestCase{
		nodeURL:      nodeUrl,
		CaseName:     name,
		walletCount:  count,
		initialCount: initial,
		steps:        steps,
		tm:           pkg.NewErc20TransferManager(common.Address{}),
		ChainID:      chainID,
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
	log.Println(fmt.Sprintf("%s create wallets finish", tc.CaseName))
	tc.wallets = pkg.GenerateCaseWallets(tc.initialCount, wallets)

	client, err := ethclient.Dial(tc.nodeURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	defer client.Close()
	if err != nil {
		log.Fatalf("Failed to Close  the Ethereum client: %v", err)
	}

	var erc20Wallets []*pkg.ERC20Wallet

	for _, ethWallet := range tc.wallets {
		erc20Wallets = append(erc20Wallets, convertEthWalletToERC20Wallet(ethWallet))
	}

	contractAddress, err := tc.Prepare(ctx, m, client, erc20Wallets)
	fmt.Println("Contract Address:", contractAddress)
	if err != nil {
		return err
	}

	tc.transCase = tc.tm.GenerateRandomErc20TransferSteps(tc.steps, erc20Wallets, contractAddress, tc.ChainID)
	return runAndAssert(tc.transCase, m, erc20Wallets)
}

func (tc *RandomTransferTestCase) Prepare(ctx context.Context, m *pkg.WalletManager, client *ethclient.Client, wallets []*pkg.ERC20Wallet) (common.Address, error) {
	deployerUsers, err := m.GenerateRandomWallets(1, accountInitialFunds)
	fmt.Println(deployerUsers)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to generate deployer user: %v", err.Error())
	}

	// get gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to suggest gas price: %v", err)
	}

	contractAddress, err := tc.prepareDeployerContract(deployerUsers[0], gasPrice, client, wallets)
	fmt.Println(contractAddress)
	if err != nil {
		return common.Address{}, fmt.Errorf("prepare contract failed, err:%v", err)
	}

	return contractAddress, nil
}

func (tc *RandomTransferTestCase) prepareDeployerContract(deployerUser *pkg.EthWallet, gasPrice *big.Int, client *ethclient.Client, wallets []*pkg.ERC20Wallet) (contractAddress common.Address, err error) {
	privateKey, err := crypto.HexToECDSA(deployerUser.PK)
	if err != nil {
		return common.Address{}, nil
	}

	depolyerAuth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(tc.ChainID))
	depolyerAuth.GasPrice = gasPrice
	depolyerAuth.GasLimit = uint64(6e7)
	//depolyerNonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(deployerUser.Address))

	if err != nil {
		return common.Address{}, nil
	}

	//depolyerAuth.Nonce = big.NewInt(int64(depolyerNonce))

	deployedToken, err := deployERC20Contracts(depolyerAuth, client)

	tc.tm = pkg.NewErc20TransferManager(deployedToken.tokenAddress)

	ERC20DeployedContracts := []*ERC20DeployedContract{deployedToken}

	err = dispatchTestToken(client, depolyerAuth, ERC20DeployedContracts, wallets, big.NewInt(accountInitialERC20Token))
	if err != nil {
		log.Fatalf("failed to dispatch test tokens: %v", err)
	}

	return deployedToken.tokenAddress, nil
}

func runAndAssert(transferCase *pkg.Erc20TransferCase, m *pkg.WalletManager, wallets []*pkg.ERC20Wallet) error {
	if err := transferCase.Run(m); err != nil {
		return err
	}
	log.Println("wait transfer transaction done")
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
	log.Printf("Block(%d) StateRoot: %s", block.Height, block.StateRoot.String())
	return nil
}

func assert(transferCase *pkg.Erc20TransferCase, walletsManager *pkg.WalletManager, wallets []*pkg.ERC20Wallet) (bool, error) {
	var got map[string]*pkg.ERC20Wallet
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

func printChange(got, expect map[string]*pkg.ERC20Wallet, transferCase *pkg.Erc20TransferCase) {
	for _, step := range transferCase.Steps {
		log.Println(fmt.Sprintf("%v transfer %v erc20 token to %v", step.From.Address, step.Count, step.To.Address))
	}
	for k, v := range got {
		ev, ok := expect[k]
		if ok {
			if v.Balance != ev.Balance {
				log.Println(fmt.Sprintf("%v got:%v expect:%v", k, v.Balance, ev.Balance))
			}
		}
	}
}

// deploy Erc20 token contracts
func deployERC20Contracts(auth *bind.TransactOpts, client *ethclient.Client) (*ERC20DeployedContract, error) {
	var err error

	deployedToken := &ERC20DeployedContract{}
	deployedToken.tokenAddress, deployedToken.tokenTransaction, deployedToken.tokenInstance, err = contracts.DeployToken(auth, client)

	if err != nil {
		return nil, err
	}

	return deployedToken, nil
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

func dispatchTestToken(client *ethclient.Client, ownerAuth *bind.TransactOpts, ERC20DeployedContracts []*ERC20DeployedContract, testUsers []*pkg.ERC20Wallet, accountInitialERC20Token *big.Int) error {
	var lastTxHash common.Hash
	for _, contract := range ERC20DeployedContracts {
		for _, user := range testUsers {
			amount := accountInitialERC20Token
			tx, err := contract.tokenInstance.Transfer(ownerAuth, user.Address, amount)
			if err != nil {
				return err
			}
			lastTxHash = tx.Hash()
		}
	}

	isConfirmed, err := waitForConfirmation(client, lastTxHash)
	if err != nil {
		return err
	}
	if !isConfirmed {
		return fmt.Errorf("transaction %s was not confirmed", lastTxHash.Hex())
	}

	// for _, contract := range ERC20DeployedContracts {
	// 	for _, user := range testUsers {
	// 		callOpts := &bind.CallOpts{
	// 			Pending: false,
	// 			Context: context.Background(),
	// 		}
	// 		balance, err := contract.tokenInstance.BalanceOf(callOpts, user.Address)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	return nil
}

func convertEthWalletToERC20Wallet(ethWallet *pkg.CaseEthWallet) *pkg.ERC20Wallet {
	return &pkg.ERC20Wallet{
		PK:         ethWallet.PK,
		Address:    common.HexToAddress(ethWallet.Address),
		Balance:    0,
		TokenCount: 0,
	}
}
