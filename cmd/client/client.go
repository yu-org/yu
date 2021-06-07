package main

import (
	"github.com/Lawliet-Chan/yu/common"
	"github.com/Lawliet-Chan/yu/keypair"
	. "github.com/Lawliet-Chan/yu/node"
	"github.com/gorilla/websocket"
	"net/url"
)

func main() {

	pubkey, privkey, err := keypair.GenKeyPair(keypair.Sr25519)
	if err != nil {
		panic("generate key error: " + err.Error())
	}

	ecall := &common.Ecall{
		TripodName: "pow",
		ExecName:   "Transfer",
		Params:     "",
	}
	signByt, err := privkey.SignData(ecall.Bytes())
	if err != nil {
		panic("sign data error: " + err.Error())
	}

	u := url.URL{Scheme: "ws", Host: "localhost:8999", Path: ExecApiPath}
	u.Query().Add(TripodNameKey, ecall.TripodName)
	u.Query().Add(CallNameKey, ecall.ExecName)
	u.Query().Add(AddressKey, pubkey.Address().String())
	u.Query().Add(SignatureKey, common.ToHex(signByt))
	u.Query().Add(PubkeyKey, pubkey.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic("dial chain error: " + err.Error())
	}
	defer c.Close()

}
