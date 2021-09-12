package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/yu-altar/yu/apps/asset"
	. "github.com/yu-altar/yu/common"
	. "github.com/yu-altar/yu/keypair"
	. "github.com/yu-altar/yu/node"
	. "github.com/yu-altar/yu/result"
	"net/url"
	"time"
)

func main() {

	pubkey, privkey, err := GenKeyPair(Sr25519)
	if err != nil {
		panic("generate key error: " + err.Error())
	}

	toPubkey, _, err := GenKeyPair(Sr25519)
	if err != nil {
		panic("generate To Address key error: " + err.Error())
	}

	go subEvent()

	createAccount(privkey, pubkey)
	time.Sleep(4 * time.Second)
	transfer(privkey, pubkey, toPubkey.Address())
	time.Sleep(4 * time.Second)

	transfer(privkey, pubkey, toPubkey.Address())
	time.Sleep(6 * time.Second)

	queryAccount(pubkey)
	queryAccount(toPubkey)

	select {}
}

type QryAccount struct {
	Account string `json:"account"`
}

func queryAccount(pubkey PubKey) {
	qa := &QryAccount{Account: pubkey.Address().String()}
	paramByt, err := json.Marshal(qa)
	if err != nil {
		panic("json encode qryAccount error: " + err.Error())
	}
	qcall := &Qcall{
		TripodName: "asset",
		QueryName:  "QueryBalance",
		BlockHash:  Hash{},
		Params:     JsonString(paramByt),
	}
	callChainByQry(pubkey.Address(), qcall)
}

type CreateAccountInfo struct {
	Amount uint64 `json:"amount"`
}

func createAccount(privkey PrivKey, pubkey PubKey) {
	paramsByt, err := json.Marshal(CreateAccountInfo{
		Amount: 500,
	})
	if err != nil {
		panic("create-account params marshal error: " + err.Error())
	}
	ecall := &Ecall{
		TripodName: "asset",
		ExecName:   "CreateAccount",
		Params:     JsonString(paramsByt),
	}
	callChainByExec(privkey, pubkey, ecall)
}

type TransferInfo struct {
	To     string `json:"to"`
	Amount uint64 `json:"amount"`
}

func transfer(privkey PrivKey, pubkey PubKey, to Address) {
	params := TransferInfo{
		To:     to.String(),
		Amount: 100,
	}
	paramsByt, err := json.Marshal(params)
	if err != nil {
		panic("transfer params marshal error: " + err.Error())
	}
	ecall := &Ecall{
		TripodName: "asset",
		ExecName:   "Transfer",
		Params:     JsonString(paramsByt),
	}
	callChainByExec(privkey, pubkey, ecall)
}

func callChainByQry(addr Address, qcall *Qcall) {
	u := url.URL{Scheme: "ws", Host: "localhost:8999", Path: QryApiPath}
	q := u.Query()
	q.Set(TripodNameKey, qcall.TripodName)
	q.Set(CallNameKey, qcall.QueryName)
	q.Set(BlockHashKey, qcall.BlockHash.String())

	u.RawQuery = q.Encode()

	//logrus.Info("qcall: ", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic("qcall dial chain error: " + err.Error())
	}
	err = c.WriteMessage(websocket.TextMessage, []byte(qcall.Params))
	if err != nil {
		panic("write qcall message to chain error: " + err.Error())
	}
	_, resp, err := c.ReadMessage()
	if err != nil {
		fmt.Println("get qcall response error: " + err.Error())
	}
	var amount asset.Amount
	err = json.Unmarshal(resp, &amount)
	if err != nil {
		panic("json decode response error: " + err.Error())
	}
	logrus.Infof("get account(%s) balance(%d)", addr.String(), amount)
}

func callChainByExec(privkey PrivKey, pubkey PubKey, ecall *Ecall) {
	signByt, err := privkey.SignData(ecall.Bytes())
	if err != nil {
		panic("sign data error: " + err.Error())
	}

	u := url.URL{Scheme: "ws", Host: "localhost:8999", Path: ExecApiPath}
	q := u.Query()
	q.Set(TripodNameKey, ecall.TripodName)
	q.Set(CallNameKey, ecall.ExecName)
	q.Set(AddressKey, pubkey.Address().String())
	q.Set(SignatureKey, ToHex(signByt))
	q.Set(PubkeyKey, pubkey.StringWithType())

	u.RawQuery = q.Encode()

	// logrus.Info("ecall: ", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic("ecall dial chain error: " + err.Error())
	}

	err = c.WriteMessage(websocket.TextMessage, []byte(ecall.Params))
	if err != nil {
		panic("write ecall message to chain error: " + err.Error())
	}
}

func subEvent() {
	u := url.URL{Scheme: "ws", Host: "localhost:8999", Path: SubResultsPath}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic("dial chain error: " + err.Error())
	}

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			panic("sub event msg from chain error: " + err.Error())
		}
		result, err := DecodeResult(msg)
		if err != nil {
			logrus.Panicf("decode result error: %s", err.Error())
		}
		switch result.Type() {
		case EventType:
			logrus.Info(result.(*Event).Sprint())
		case ErrorType:
			logrus.Error(result.(*Error).Error())
		}
	}
}
