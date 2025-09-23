package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/sirupsen/logrus"

	"github.com/yu-org/yu/apps/eth/ethrpc"
)

type EthWallet struct {
	PK      string `json:"pk"`
	Address string `json:"address"`
}

func (e *EthWallet) Copy() *EthWallet {
	return &EthWallet{
		PK:      e.PK,
		Address: e.Address,
	}
}

type WalletManager struct {
	nodeUrl string
	pk      string
	chainID int64
}

func NewWalletManager(chainID int64, nodeUrl, pk string) *WalletManager {
	return &WalletManager{
		nodeUrl: nodeUrl,
		pk:      pk,
		chainID: chainID,
	}
}

func (m *WalletManager) GenerateRandomWallets(count int, initialEthCount uint64) ([]*EthWallet, error) {
	wallets := make([]*EthWallet, 0)
	for i := 1; i <= count; i++ {
		wallet, err := m.createEthWallet(initialEthCount)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
		if i%2000 == 0 {
			m.AssertWallet(wallet, initialEthCount)
			logrus.Infof("assert %v/%v wallet done", i, count)
		}
	}
	m.AssertWallet(wallets[len(wallets)-1], initialEthCount)
	return wallets, nil
}

func (m *WalletManager) AssertWallet(w *EthWallet, count uint64) {
	for {
		got, err := m.QueryEth(w)
		if err == nil && got >= count {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (m *WalletManager) BatchGenerateRandomWallets(count int, initialEthCount uint64) ([]*EthWallet, error) {
	wallets := make([]*EthWallet, 0)
	for i := 1; i <= count; i++ {
		wallet, err := m.createEthWallet(initialEthCount)
		if err != nil {
			return nil, err
		}
		if i%5000 == 0 {
			m.AssertWallet(wallet, initialEthCount)
			fmt.Println(fmt.Sprintf("assert %v  wallet success", i))
		}
		wallets = append(wallets, wallet)
		fmt.Println(fmt.Sprintf("create %v/%v wallet", i, count))
	}
	return wallets, nil
}

func (m *WalletManager) createEthWallet(initialEthCount uint64) (*EthWallet, error) {
	privateKey, address := generatePrivateKey()
	return m.CreateEthWalletByAddress(initialEthCount, privateKey, address)
}

func (m *WalletManager) CreateEthWalletByAddress(initialEthCount uint64, privateKey, address string) (*EthWallet, error) {
	if err := m.transferEth(m.pk, address, initialEthCount); err != nil {
		return nil, err
	}
	// logrus.Infof("create wallet %v", address))
	return &EthWallet{PK: privateKey, Address: address}, nil
}

func (m *WalletManager) TransferEth(from, to *EthWallet, amount, nonce uint64) error {
	// logrus.Infof("transfer %v eth from %v to %v", amount, from.Address, to.Address))
	if err := m.transferEth(from.PK, to.Address, amount); err != nil {
		return err
	}
	return nil
}

func (m *WalletManager) QueryEth(wallet *EthWallet) (uint64, error) {
	client, err := ethclient.Dial(m.nodeUrl)
	if err != nil {
		return 0, err
	}
	defer client.Close()

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(wallet.Address), nil)
	if err != nil {
		return 0, err
	}

	return balance.Uint64(), nil
}

func parse(v string) (uint64, error) {
	if !strings.HasPrefix(v, "0x") {
		return 0, fmt.Errorf("%v should start with 0v", v)
	}
	value, err := strconv.ParseUint(v[2:], 16, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

type queryResponse struct {
	Result string `json:"result"`
}

func (m *WalletManager) transferEth(privateKeyHex string, toAddress string, amount uint64) error {
	return m.sendRawTx(privateKeyHex, toAddress, amount)
}

// sendRawTx is used by transferring and contract creation/invocation.
func (m *WalletManager) sendRawTx(privateKeyHex string, toAddress string, amount uint64) error {
	to := common.HexToAddress(toAddress)
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		logrus.Fatal(err)
	}

	gasLimit := uint64(21000)

	client, err := ethclient.Dial(m.nodeUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(int64(0)),
		Gas:      gasLimit,
		To:       &to,
		Value:    big.NewInt(int64(amount)),
		Data:     nil,
	})

	chainID := big.NewInt(int64(m.chainID))
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		logrus.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)

	return err
}

type RawTxReq struct {
	privateKeyHex string
	toAddress     string
	amount        uint64
	data          []byte
	nonce         uint64
}

func (m *WalletManager) sendBatchRawTxs(rawTxs []*RawTxReq) error {
	client, err := ethclient.Dial(m.nodeUrl)
	if err != nil {
		logrus.Fatal(err)
	}
	defer client.Close()

	batchTx := new(ethrpc.BatchTx)

	for _, rawTx := range rawTxs {
		to := common.HexToAddress(rawTx.toAddress)
		gasLimit := uint64(21000)

		privateKey, err := crypto.HexToECDSA(rawTx.privateKeyHex)
		if err != nil {
			logrus.Fatal(err)
		}

		fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
		nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
		if err != nil {
			logrus.Fatal(err)
		}

		tx := types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			GasPrice: big.NewInt(int64(0)),
			Gas:      gasLimit,
			To:       &to,
			Value:    big.NewInt(int64(rawTx.amount)),
			Data:     rawTx.data,
		})

		chainID := big.NewInt(int64(m.chainID))
		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
		if err != nil {
			logrus.Fatal(err)
		}
		rawTxBytes, err := rlp.EncodeToBytes(signedTx)
		if err != nil {
			logrus.Fatal(err)
		}
		batchTx.TxsBytes = append(batchTx.TxsBytes, rawTxBytes)
	}

	batchTxBytes, err := json.Marshal(batchTx)
	if err != nil {
		return err
	}

	requestBody := fmt.Sprintf(
		`	{
		"jsonrpc": "2.0",
		"id": 0,
		"method": "eth_sendBatchRawTransactions",
		"params": ["0x%x"] 
	}`, batchTxBytes)
	_, err = sendRequest(m.nodeUrl, requestBody)
	return err
}
