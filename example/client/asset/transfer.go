package asset

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/context"
	. "github.com/yu-org/yu/core/keypair"
	. "github.com/yu-org/yu/example/client/callchain"
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
	qcall := &RdCall{
		TripodName: "asset",
		FuncName:   "QueryBalance",
		BlockHash:  Hash{},
		Params:     string(paramByt),
	}
	resp := CallChainByReading(Websocket, qcall)
	respMap := make(context.H)
	err = json.Unmarshal(resp, &respMap)
	if err != nil {
		panic("json decode qryAccount response error: " + err.Error())
	}
	//amount := new(big.Int)
	//err = amount.UnmarshalText(resp)
	//if err != nil {
	//	panic(err)
	//}
	logrus.Infof("get account(%s) balance(%v)", pubkey.Address().String(), respMap["amount"])
}

type CreateAccountInfo struct {
	Amount uint64 `json:"amount"`
}

func CreateAccount(reqType int, privkey PrivKey, pubkey PubKey, amount uint64) {
	paramsByt, err := json.Marshal(CreateAccountInfo{
		Amount: amount,
	})
	if err != nil {
		panic("create-account params marshal error: " + err.Error())
	}
	wrCall := &WrCall{
		TripodName: "asset",
		FuncName:   "CreateAccount",
		Params:     string(paramsByt),
		LeiPrice:   0,
	}
	CallChainByWriting(reqType, privkey, pubkey, wrCall)
}

type TransferInfo struct {
	To     string `json:"to"`
	Amount uint64 `json:"amount"`
}

func TransferBalance(reqType int, privkey PrivKey, pubkey PubKey, to Address, amount, leiPrice uint64) {
	params := TransferInfo{
		To:     to.String(),
		Amount: amount,
	}
	paramsByt, err := json.Marshal(params)
	if err != nil {
		panic("TransferBalance params marshal error: " + err.Error())
	}
	wrCall := &WrCall{
		TripodName: "asset",
		FuncName:   "Transfer",
		Params:     string(paramsByt),
		LeiPrice:   leiPrice,
	}
	CallChainByWriting(reqType, privkey, pubkey, wrCall)
}
