package main

import (
	"encoding/json"
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/keypair"
	. "github.com/Lawliet-Chan/yu/node"
	. "github.com/Lawliet-Chan/yu/result"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/url"
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

	transfer(privkey, pubkey, toPubkey.Address())

	select {}

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
	callChain(privkey, pubkey, ecall)
}

type TransferInfo struct {
	To     []byte `json:"to"`
	Amount uint64 `json:"amount"`
}

func transfer(privkey PrivKey, pubkey PubKey, to Address) {
	params := TransferInfo{
		To:     to.Bytes(),
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
	callChain(privkey, pubkey, ecall)
}

func callChain(privkey PrivKey, pubkey PubKey, ecall *Ecall) {
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
	q.Set(KeyTypeKey, Sr25519)
	q.Set(PubkeyKey, pubkey.String())

	u.RawQuery = q.Encode()

	logrus.Info(u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic("dial chain error: " + err.Error())
	}

	err = c.WriteMessage(websocket.TextMessage, []byte(ecall.Params))
	if err != nil {
		panic("write message to chain error: " + err.Error())
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
