package transfer

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/keypair"
	. "github.com/yu-org/yu/example/client/callchain"
	"math/big"
)

type QryAccount struct {
	Account string `json:"account"`
}

func QueryAccount(pubkey PubKey) {
	qa := &QryAccount{Account: pubkey.Address().String()}
	paramByt, err := json.Marshal(qa)
	if err != nil {
		panic("json encode qryAccount error: " + err.Error())
	}
	qcall := &Qcall{
		TripodName: "asset",
		QueryName:  "QueryBalance",
		BlockHash:  Hash{},
		Params:     string(paramByt),
	}
	resp := CallChainByQry(Websocket, qcall)
	amount := new(big.Int)
	err = amount.UnmarshalText(resp)
	if err != nil {
		panic(err)
	}
	logrus.Infof("get account(%s) balance(%d)", pubkey.Address().String(), amount)
}

type CreateAccountInfo struct {
	Amount uint64 `json:"amount"`
}

func CreateAccount(privkey PrivKey, pubkey PubKey) {
	paramsByt, err := json.Marshal(CreateAccountInfo{
		Amount: 500,
	})
	if err != nil {
		panic("create-account params marshal error: " + err.Error())
	}
	ecall := &Ecall{
		TripodName: "asset",
		ExecName:   "CreateAccount",
		Params:     string(paramsByt),
		LeiPrice:   0,
	}
	CallChainByExec(Websocket, privkey, pubkey, ecall)
}

type TransferInfo struct {
	To     string `json:"to"`
	Amount uint64 `json:"amount"`
}

func TransferBalance(reqType int, privkey PrivKey, pubkey PubKey, to Address) {
	params := TransferInfo{
		To:     to.String(),
		Amount: 100,
	}
	paramsByt, err := json.Marshal(params)
	if err != nil {
		panic("TransferBalance params marshal error: " + err.Error())
	}
	ecall := &Ecall{
		TripodName: "asset",
		ExecName:   "Transfer",
		Params:     string(paramsByt),
		LeiPrice:   0,
	}
	CallChainByExec(reqType, privkey, pubkey, ecall)
}
