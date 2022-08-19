package main

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/core/keypair"
	. "github.com/yu-org/yu/example/client/asset"
	. "github.com/yu-org/yu/example/client/callchain"
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

	go SubEvent()

	logrus.Info("--- send Creating Account ---")
	CreateAccount(Websocket, privkey, pubkey, 500)
	time.Sleep(4 * time.Second)

	logrus.Info("--- send Transfering 1 ---")
	TransferBalance(Websocket, privkey, pubkey, toPubkey.Address(), 50, 0)
	time.Sleep(4 * time.Second)

	logrus.Info("--- send Transfering 2 ---")
	TransferBalance(Websocket, privkey, pubkey, toPubkey.Address(), 100, 0)
	time.Sleep(6 * time.Second)

	QueryAccount(pubkey)
	QueryAccount(toPubkey)

	select {}
}
