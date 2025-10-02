package uniswap

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"

	"github.com/yu-org/yu/apps/eth/test/contracts"
	"github.com/yu-org/yu/apps/eth/test/pkg"
)

const (
	waitForConfirmationTime = 1 * time.Second
	maxRetries              = 300
)

type ERC20DeployedContract struct {
	tokenAddress     common.Address
	tokenTransaction *types.Transaction
	tokenInstance    *contracts.Token
}

type UniswapV2DeployedContracts struct {
	weth9Address                 common.Address
	uniswapV2FactoryAddress      common.Address
	uniswapV2Router01Address     common.Address
	weth9Transaction             *types.Transaction
	uniswapV2FactoryTransaction  *types.Transaction
	uniswapV2Router01Transaction *types.Transaction
	weth9Instance                *contracts.WETH9
	uniswapV2FactoryInstance     *contracts.UniswapV2Factory
	uniswapV2RouterInstance      *contracts.UniswapV2Router01
}
type TestData struct {
	TestUsers     []*pkg.EthWallet `json:"testUsers"`
	TestContracts []TestContract
}

type TestContract struct {
	UniswapV2Router common.Address      `json:"uniswapV2Router"`
	TokenPairs      [][2]common.Address `json:"tokenPairs"`
}

// deploy Erc20 token contracts
func deployERC20Contracts(auth *bind.TransactOpts, client *ethclient.Client, deployNum int) ([]*ERC20DeployedContract, error) {
	var err error
	deployedTokens := make([]*ERC20DeployedContract, 0)

	for i := 0; i < deployNum; i++ {
		deployedToken := &ERC20DeployedContract{}
		deployedToken.tokenAddress, deployedToken.tokenTransaction, deployedToken.tokenInstance, err = contracts.DeployToken(auth, client)
		if err != nil {
			return nil, err
		}

		deployedTokens = append(deployedTokens, deployedToken)
		auth.Nonce.Add(auth.Nonce, big.NewInt(1))

	}

	return deployedTokens, nil
}

// deploy UniswapV2 Contracts
/*
   Deploy WETH
   Deploy UniswapV2Factory (FeeToSetter)
   Deploy UniswapV2Router01 (WETH addresse, factory addresse)
*/
func deployUniswapV2Contracts(auth *bind.TransactOpts, client *ethclient.Client) (*UniswapV2DeployedContracts, error) {
	var err error
	deployed := &UniswapV2DeployedContracts{}

	// Deploy WETH9
	deployed.weth9Address, deployed.weth9Transaction, deployed.weth9Instance, err = contracts.DeployWETH9(auth, client)
	if err != nil {
		return nil, err
	}
	auth.Nonce.Add(auth.Nonce, big.NewInt(1))

	// Deploy UniswapV2Factory
	deployed.uniswapV2FactoryAddress, deployed.uniswapV2FactoryTransaction, deployed.uniswapV2FactoryInstance, err = contracts.DeployUniswapV2Factory(auth, client, auth.From)
	if err != nil {
		return nil, err
	}
	auth.Nonce.Add(auth.Nonce, big.NewInt(1))

	// Deploy UniswapV2Router01
	deployed.uniswapV2Router01Address, deployed.uniswapV2Router01Transaction, deployed.uniswapV2RouterInstance, err = contracts.DeployUniswapV2Router01(auth, client, deployed.uniswapV2FactoryAddress, deployed.weth9Address)
	if err != nil {
		return nil, err
	}
	auth.Nonce.Add(auth.Nonce, big.NewInt(1))

	return deployed, nil
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

func generateTestAuth(client *ethclient.Client, user *pkg.EthWallet, chainID int64, gasPrice *big.Int, gasLimit uint64) (*bind.TransactOpts, error) {
	privateKey, err := crypto.HexToECDSA(user.PK)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainID))
	if err != nil {
		return nil, fmt.Errorf("failed to create authorized transactor: %v", err)
	}

	auth.GasPrice = gasPrice
	auth.GasLimit = gasLimit

	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(user.Address))
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))

	return auth, nil
}

func saveTestDataToFile(filename string, data TestData) {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		logrus.Fatalf("Failed to create directory: %v", err)
	}
	file, err := os.Create(filename)
	if err != nil {
		logrus.Fatalf("Error creating file: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	writer := bufio.NewWriter(file)

	encoder := json.NewEncoder(writer)
	if err := encoder.Encode(&data); err != nil {
		logrus.Fatalf("Failed to encode data to JSON: %v", err)
	}

	if err := writer.Flush(); err != nil {
		logrus.Fatalf("Failed to flush writer: %v", err)
	}

	fmt.Println("Data successfully written to", filename)
}

func loadTestDataFromFile(filename string) (TestData, error) {
	var data TestData

	file, err := os.Open(filename)
	if err != nil {
		return data, fmt.Errorf("error opening file: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	reader := bufio.NewReader(file)

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&data); err != nil {
		return data, fmt.Errorf("failed to decode JSON data: %v", err)
	}

	return data, nil
}

func dispatchTestToken(client *ethclient.Client, ownerAuth *bind.TransactOpts, ERC20DeployedContracts []*ERC20DeployedContract, testUsers []*pkg.EthWallet, accountInitialERC20Token *big.Int) error {
	var lastTxHash common.Hash
	for _, contract := range ERC20DeployedContracts {
		for _, user := range testUsers {
			amount := accountInitialERC20Token
			tx, err := contract.tokenInstance.Transfer(ownerAuth, common.HexToAddress(user.Address), amount)
			if err != nil {
				return err
			}
			lastTxHash = tx.Hash()
			ownerAuth.Nonce = ownerAuth.Nonce.Add(ownerAuth.Nonce, big.NewInt(1))
		}
	}

	isConfirmed, err := waitForConfirmation(client, lastTxHash)
	if err != nil {
		return err
	}
	if !isConfirmed {
		return fmt.Errorf("transaction %s was not confirmed", lastTxHash.Hex())
	}
	return nil
}
