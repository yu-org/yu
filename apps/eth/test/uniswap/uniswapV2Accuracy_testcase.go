package uniswap

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"

	"github.com/yu-org/yu/apps/eth/test/contracts"
	"github.com/yu-org/yu/apps/eth/test/pkg"
)

type UniswapV2AccuracyTestCase struct {
	NodeURL       string
	ChainID       int64
	CaseName      string
	walletCount   int
	initialCount  uint64
	testUsers     int
	deployedUsers int
}

func (ca *UniswapV2AccuracyTestCase) Name() string {
	return ca.CaseName
}

func NewUniswapV2AccuracyTestCase(name, nodeURL string, count int, initial uint64, chainID int64) *UniswapV2AccuracyTestCase {
	return &UniswapV2AccuracyTestCase{
		NodeURL:       nodeURL,
		ChainID:       chainID,
		CaseName:      name,
		walletCount:   count,
		initialCount:  initial,
		deployedUsers: 1,
		testUsers:     2,
	}
}

const swapTimes = 200

func (ca *UniswapV2AccuracyTestCase) Run(ctx context.Context, m *pkg.WalletManager) error {
	preparedTestData, err := ca.Prepare(ctx, m)
	if err != nil {
		return err
	}

	client, err := ethclient.Dial("http://localhost:9092")
	if err != nil {
		logrus.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	defer client.Close()
	if err != nil {
		logrus.Fatalf("Failed to Close  the Ethereum client: %v", err)
	}

	// get gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		logrus.Fatalf("Failed to suggest gas price: %v", err)
	}

	// Arrange :
	testUser := preparedTestData.TestUsers
	if err != nil {
		return err
	}
	testUserPK, err := crypto.HexToECDSA(testUser[0].PK)
	if err != nil {
		logrus.Fatalf("Failed to parse private key: %v", err)
	}
	testUserAuth, err := bind.NewKeyedTransactorWithChainID(testUserPK, big.NewInt(ca.ChainID))
	if err != nil {
		logrus.Fatalf("Failed to create authorized transactor: %v", err)
	}
	testUserAuth.GasPrice = gasPrice
	testUserAuth.GasLimit = uint64(6e7)

	// Perform  swaps
	const maxRetries = 300
	const retryDelay = 10 * time.Millisecond
	var retryErrors []struct {
		Nonce int
		Err   error
	}
	// Act
	routeraddress := preparedTestData.TestContracts[0].UniswapV2Router
	uniswapV2RouterInstance, err := contracts.NewUniswapV2Router01(routeraddress, client)
	if err != nil {
		logrus.Fatalf("Failed to get UniswapV2Router instance: %v", err)
	}
	swapPath := []common.Address{
		preparedTestData.TestContracts[0].TokenPairs[0][0],
		preparedTestData.TestContracts[0].TokenPairs[0][1],
	}
	for i := 0; i < swapTimes; i++ {
		// Execute swap operation
		nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(testUser[0].Address))
		if err != nil {
			logrus.Fatalf("Failed to get nonce: %v", err)
		}

		for j := 0; j < maxRetries; j++ {
			//(testUserAuth, amountOut, []common.Address{preparedTestData.TestContracts[0].TokenPairs[0][0], preparedTestData.TestContracts[0].TokenPairs[0][1]}, common.HexToAddress(testUser[0].Address), big.NewInt(time.Now().Unix()+1000))
			swapETHForExactTokensTx, err := uniswapV2RouterInstance.SwapExactTokensForTokens(testUserAuth, big.NewInt(100), big.NewInt(0), swapPath, common.HexToAddress(testUser[0].Address), big.NewInt(time.Now().Unix()+1000))
			if err == nil {
				// Wait for transaction confirmation
				if i == (swapTimes - 1) {
					isConfirmed, err := waitForConfirmation(client, swapETHForExactTokensTx.Hash())
					if err != nil {
						logrus.Fatalf("Failed to confirm swapETHForExactTokensTx transaction: %v", err)
					}
					if !isConfirmed {
						logrus.Fatalf("SwapETHForExactTokens transaction was not confirmed")
					}
				}
				break
			}
			retryErrors = append(retryErrors, struct {
				Nonce int
				Err   error
			}{Nonce: int(nonce), Err: err}) // record和 nonce
			time.Sleep(retryDelay)
		}

		if err != nil {
			logrus.Fatalf("Failed to swapETHForExactTokensTx transaction after %d attempts: %v", maxRetries, err)
		}

	}

	// Get account balance
	token1ddress := preparedTestData.TestContracts[0].TokenPairs[0][0]
	token1AddressInstance, err := contracts.NewToken(token1ddress, client)
	if err != nil {
		logrus.Fatalf("Failed to get Token1 instance: %v", err)
	}
	token2ddress := preparedTestData.TestContracts[0].TokenPairs[0][1]
	token2AddressInstance, err := contracts.NewToken(token2ddress, client)
	if err != nil {
		logrus.Fatalf("Failed to get Token2 instance: %v", err)
	}
	token1Balance, err := token1AddressInstance.BalanceOf(nil, common.HexToAddress(testUser[0].Address))
	if err != nil {
		logrus.Fatalf("Failed to get Token1 balance: %v", err)
	}
	fmt.Printf("Token2 balance: %s\n", token1Balance.String())

	token2Balance, err := token2AddressInstance.BalanceOf(nil, common.HexToAddress(testUser[0].Address))
	if err != nil {
		logrus.Fatalf("Failed to get Token2 balance: %v", err)
	}
	fmt.Printf("Token2 balance: %s\n", token2Balance.String())
	// Expect results
	// Initial state
	token1Reserve := big.NewInt(1e18)
	token2Reserve := big.NewInt(1e18)
	k := new(big.Int).Mul(token1Reserve, token2Reserve)

	// User initial state
	expectedToken1Balance := big.NewInt(1e18)
	expectedToken2Balance := big.NewInt(1e18)

	swapToken := big.NewInt(100)

	for i := 0; i < swapTimes; i++ {
		token1toToken2Price := 99
		token2Received := big.NewInt(int64(token1toToken2Price))

		// Update reserves
		token1Reserve.Add(token1Reserve, swapToken)
		token2Reserve.Sub(k, token2Reserve)

		// Update user balance
		expectedToken1Balance.Sub(expectedToken1Balance, swapToken)
		expectedToken2Balance.Add(expectedToken2Balance, token2Received)
	}

	// Print final reserves
	fmt.Printf("Final token1Reserve: %s\n", token1Reserve.String())
	fmt.Printf("Final token2Reserve: %s\n", token2Reserve.String())

	if token1Balance.Cmp(expectedToken1Balance) != 0 {
		logrus.Fatalf("Expected user TokenA balance to be %s, but got %s", expectedToken1Balance.String(), token1Balance.String())
	}

	if token2Balance.Cmp(expectedToken2Balance) != 0 {
		logrus.Fatalf("Expected user token2Balance balance to be %s, but got %s", expectedToken2Balance.String(), token2Balance.String())
	}
	return err
}

func (ca *UniswapV2AccuracyTestCase) prepareDeployerContract(deployerUser *pkg.EthWallet, testUsers []*pkg.EthWallet, gasPrice *big.Int, client *ethclient.Client) (UniswapV2Router common.Address, TokenPairs [][2]common.Address, err error) {
	// set tx auth
	privateKey, err := crypto.HexToECDSA(deployerUser.PK)
	if err != nil {
		return [20]byte{}, nil, fmt.Errorf("failed to parse private key: %v", err)
	}
	depolyerAuth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(ca.ChainID))
	if err != nil {
		return [20]byte{}, nil, fmt.Errorf("failed to create authorized transactor: %v", err)
	}
	depolyerAuth.GasPrice = gasPrice
	depolyerAuth.GasLimit = uint64(6e7)
	depolyerNonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(deployerUser.Address))
	if err != nil {
		return [20]byte{}, nil, fmt.Errorf("failed to get nonce: %v", err)
	}
	depolyerAuth.Nonce = big.NewInt(int64(depolyerNonce))
	// deploy contracts
	uniswapV2Contract, err := deployUniswapV2Contracts(depolyerAuth, client)
	if err != nil {
		return [20]byte{}, nil, fmt.Errorf("Failed to deploy contract: %v", err)
	}
	ERC20DeployedContracts, err := deployERC20Contracts(depolyerAuth, client, tokenContractNum)
	if err != nil {
		return [20]byte{}, nil, fmt.Errorf("Failed to deploy ERC20 contracts: %v", err)
	}
	lastIndex := len(ERC20DeployedContracts) - 1
	isConfirmed, err := waitForConfirmation(client, ERC20DeployedContracts[lastIndex].tokenTransaction.Hash())
	if err != nil {
		return [20]byte{}, nil, fmt.Errorf("Failed to confirm approve transaction: %v", err)
	}
	if !isConfirmed {
		return [20]byte{}, nil, fmt.Errorf("transaction was not confirmed")
	}
	err = dispatchTestToken(client, depolyerAuth, ERC20DeployedContracts, testUsers, big.NewInt(accountInitialERC20Token))
	if err != nil {
		return [20]byte{}, nil, fmt.Errorf("failed to dispatch test tokens: %v", err)
	}
	var lastTxHash common.Hash
	for _, contract := range ERC20DeployedContracts {
		_, err := contract.tokenInstance.Approve(depolyerAuth, uniswapV2Contract.uniswapV2Router01Address, big.NewInt(approveAmount))
		if err != nil {
			return [20]byte{}, nil, fmt.Errorf("failed to create approve transaction for user %s: %v", deployerUser.Address, err)
		}

		depolyerAuth.Nonce = depolyerAuth.Nonce.Add(depolyerAuth.Nonce, big.NewInt(1))
		for _, user := range testUsers {
			testAuth, err := generateTestAuth(client, user, ca.ChainID, gasPrice, gasLimit)
			if err != nil {
				return [20]byte{}, nil, fmt.Errorf("failed to generate test auth for user %s: %v", user.Address, err)
			}
			tx, err := contract.tokenInstance.Approve(testAuth, uniswapV2Contract.uniswapV2Router01Address, big.NewInt(approveAmount))
			if err != nil {
				return [20]byte{}, nil, fmt.Errorf("failed to create approve transaction for user %s: %v", user.Address, err)
			}
			lastTxHash = tx.Hash()
			// logrus.Infof("Approve transaction hash for user %s: %s", user.Address, tx.Hash().Hex())
			testAuth.Nonce = testAuth.Nonce.Add(testAuth.Nonce, big.NewInt(1))
		}
	}
	isConfirmed, err = waitForConfirmation(client, lastTxHash)
	if err != nil {
		return [20]byte{}, nil, err
	}
	if !isConfirmed {
		return [20]byte{}, nil, fmt.Errorf("transaction %s was not confirmed", lastTxHash.Hex())
	}
	tokenPairs := generateTokenPairs(ERC20DeployedContracts)
	// add liquidity
	for _, pair := range tokenPairs {
		addLiquidityTx, err := uniswapV2Contract.uniswapV2RouterInstance.AddLiquidity(
			depolyerAuth,
			pair[0],
			pair[1],
			big.NewInt(amountADesired),
			big.NewInt(amountBDesired),
			big.NewInt(0),
			big.NewInt(0),
			common.HexToAddress(deployerUser.Address),
			big.NewInt(time.Now().Unix()+1000),
		)
		if err != nil {
			return [20]byte{}, nil, fmt.Errorf("failed to create add liquidity transaction for pair %s - %s: %v", pair[0].Hex(), pair[1].Hex(), err)
		}
		depolyerAuth.Nonce = depolyerAuth.Nonce.Add(depolyerAuth.Nonce, big.NewInt(1))
		lastTxHash = addLiquidityTx.Hash()
	}
	isConfirmed, err = waitForConfirmation(client, lastTxHash)
	if err != nil {
		return [20]byte{}, nil, fmt.Errorf("failed to confirm add liquidity transaction: %v", err)
	}
	if !isConfirmed {
		return [20]byte{}, nil, errors.New("add liquidity transaction was not confirmed")
	}
	return uniswapV2Contract.uniswapV2Router01Address, tokenPairs, nil
}

func (ca *UniswapV2AccuracyTestCase) Prepare(ctx context.Context, m *pkg.WalletManager) (TestData, error) {
	deployerUsers, err := m.GenerateRandomWallets(ca.deployedUsers, accountInitialFunds)
	if err != nil {
		return TestData{}, fmt.Errorf("failed to generate deployer user: %v", err.Error())
	}
	testUsers, err := m.GenerateRandomWallets(ca.testUsers, accountInitialFunds)
	if err != nil {
		return TestData{}, fmt.Errorf("failed to generate test users: %v", err)
	}
	client, err := ethclient.Dial(ca.NodeURL)
	if err != nil {
		return TestData{}, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	// get gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return TestData{}, fmt.Errorf("failed to suggest gas price: %v", err)
	}
	preparedTestData := TestData{
		TestUsers:     testUsers,
		TestContracts: make([]TestContract, 0),
	}
	for _, deployerUser := range deployerUsers {
		router, tokenPairs, err := ca.prepareDeployerContract(deployerUser, testUsers, gasPrice, client)
		if err != nil {
			return TestData{}, fmt.Errorf("prepare contract failed, err:%v", err)
		}
		preparedTestData.TestContracts = append(preparedTestData.TestContracts, TestContract{router, tokenPairs})
	}

	return preparedTestData, nil
}
