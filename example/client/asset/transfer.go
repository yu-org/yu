package asset

import (
	"encoding/json"
	"github.com/HyperService-Consortium/go-hexutil"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/keypair"
	"github.com/yu-org/yu/core/protocol"
	. "github.com/yu-org/yu/example/client/callchain"
	"math/big"
)

func QueryAccount(pubkey PubKey) uint64 {
	addr := pubkey.Address()
	params := map[string]string{"account": addr.String()}
	paramsByt, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}
	rdCall := &RdCall{
		TripodName: "asset",
		FuncName:   "QueryBalance",
		Params:     string(paramsByt),
	}
	resp, err := CallChainByReading(rdCall)
	if err != nil {
		panic(err)
	}
	respMap := make(map[string]*big.Int)
	err = json.Unmarshal(resp, &respMap)
	if err != nil {
		panic("json decode qryAccount response error: " + err.Error())
	}
	//amount := new(big.Int)
	//err = amount.UnmarshalText(resp)
	//if err != nil {
	//	panic(err)
	//}
	return respMap["amount"].Uint64()
}

type CreateAccountInfo struct {
	Amount uint64 `json:"amount"`
}

func CreateAccount(privkey PrivKey, pubkey PubKey, amount uint64) {
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
	byt, err := json.Marshal(wrCall)
	if err != nil {
		panic(err)
	}
	msgHash := BytesToHash(byt)
	sig, err := privkey.SignData(msgHash.Bytes())
	if err != nil {
		panic(err)
	}
	postBody := &protocol.WritingPostBody{
		Pubkey:    pubkey.StringWithType(),
		Address:   pubkey.Address().String(),
		Signature: hexutil.Encode(sig),
		Call:      wrCall,
	}
	err = CallChainByWriting(postBody)
	if err != nil {
		panic(err)
	}
}

type TransferInfo struct {
	To     string `json:"to"`
	Amount uint64 `json:"amount"`
}

func TransferBalance(privkey PrivKey, pubkey PubKey, to Address, amount, leiPrice uint64) {
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
	byt, err := json.Marshal(wrCall)
	if err != nil {
		panic(err)
	}
	msgHash := BytesToHash(byt)
	sig, err := privkey.SignData(msgHash.Bytes())
	if err != nil {
		panic(err)
	}
	postBody := &protocol.WritingPostBody{
		Pubkey:    pubkey.StringWithType(),
		Signature: hexutil.Encode(sig),
		Address:   pubkey.Address().String(),
		Call:      wrCall,
	}
	err = CallChainByWriting(postBody)
	if err != nil {
		panic(err)
	}
}
