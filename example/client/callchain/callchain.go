package callchain

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core"
	. "github.com/yu-org/yu/core/keypair"
	. "github.com/yu-org/yu/core/result"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	Http = iota
	Websocket
)

func CallChainByQry(reqtyp int, qcall *Rdcall) []byte {
	u := url.URL{Scheme: "ws", Host: "localhost:8999", Path: QryApiPath}
	q := u.Query()
	q.Set(TripodNameKey, qcall.TripodName)
	q.Set(CallNameKey, qcall.QueryName)
	q.Set(BlockHashKey, qcall.BlockHash.String())

	u.RawQuery = q.Encode()

	// logrus.Info("qcall: ", u.String())
	switch reqtyp {
	case Http:
		resp, err := http.Post(u.String(), "application/json", strings.NewReader(qcall.Params))
		if err != nil {
			panic("post qcall message to chain error: " + err.Error())
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic("read qcall response body error: " + err.Error())
		}
		return body
	case Websocket:
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			panic("qcall dial chain error: " + err.Error())
		}
		defer c.Close()
		err = c.WriteMessage(websocket.TextMessage, []byte(qcall.Params))
		if err != nil {
			panic("write qcall message to chain error: " + err.Error())
		}
		_, resp, err := c.ReadMessage()
		if err != nil {
			panic("get qcall response error: " + err.Error())
		}
		return resp
	}
	return nil
}

func CallChainByExec(reqType int, privkey PrivKey, pubkey PubKey, ecall *WrCall) {
	hash, err := ecall.Hash()
	if err != nil {
		panic("ecall hash error: " + err.Error())
	}
	signByt, err := privkey.SignData(hash)
	if err != nil {
		panic("sign data error: " + err.Error())
	}

	var scheme string
	switch reqType {
	case Http:
		scheme = "http"
	case Websocket:
		scheme = "ws"
	}

	u := url.URL{Scheme: scheme, Host: "localhost:8999", Path: ExecApiPath}
	q := u.Query()
	q.Set(TripodNameKey, ecall.TripodName)
	q.Set(CallNameKey, ecall.ExecName)
	q.Set(AddressKey, pubkey.Address().String())
	q.Set(SignatureKey, ToHex(signByt))
	q.Set(PubkeyKey, pubkey.StringWithType())
	q.Set(LeiPriceKey, hexutil.EncodeUint64(ecall.LeiPrice))

	u.RawQuery = q.Encode()

	// logrus.Info("ecall: ", u.String())

	switch reqType {
	case Http:
		_, err := http.Post(u.String(), "application/json", strings.NewReader(ecall.Params))
		if err != nil {
			panic("post ecall message to chain error: " + err.Error())
		}
	case Websocket:
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			panic("ecall dial chain error: " + err.Error())
		}
		defer c.Close()
		err = c.WriteMessage(websocket.TextMessage, []byte(ecall.Params))
		if err != nil {
			panic("write ecall message to chain error: " + err.Error())
		}
	}
}

func SubEvent() {
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
