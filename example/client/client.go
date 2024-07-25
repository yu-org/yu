package main

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/core/keypair"
	"github.com/yu-org/yu/core/types"
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

	sub, err := NewSubscriber()
	if err != nil {
		panic("new subscriber failed: " + err.Error())
	}

	resultCh := make(chan *types.Receipt)
	go sub.SubEvent(resultCh)

	logrus.Info("--- send Creating Account ---")
	CreateAccount(privkey, pubkey, 500)
	time.Sleep(4 * time.Second)

	logrus.Info("--- send Transferring 1 ---")
	TransferBalance(privkey, pubkey, toPubkey.Address(), 50, 0)
	time.Sleep(4 * time.Second)

	logrus.Info("--- send Transferring 2 ---")
	TransferBalance(privkey, pubkey, toPubkey.Address(), 100, 0)
	time.Sleep(6 * time.Second)

	QueryAccount(pubkey)
	QueryAccount(toPubkey)

	select {}
}
