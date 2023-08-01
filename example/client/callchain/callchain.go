package callchain

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core"
	. "github.com/yu-org/yu/core/keypair"
	. "github.com/yu-org/yu/core/result"
	"go.uber.org/atomic"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	Http = iota
	Websocket
)

func CallChainByReading(reqTyp int, rdCall *RdCall) []byte {
	var (
		scheme, port string
	)
	switch reqTyp {
	case Http:
		scheme = "http"
		port = "7999"
	case Websocket:
		scheme = "ws"
		port = "8999"
	}
	u := url.URL{Scheme: scheme, Host: fmt.Sprintf("localhost:%s", port), Path: RdApiPath}
	q := u.Query()
	q.Set(TripodNameKey, rdCall.TripodName)
	q.Set(CallNameKey, rdCall.ReadingName)
	q.Set(BlockHashKey, rdCall.BlockHash.String())

	u.RawQuery = q.Encode()

	logrus.Debug("rdCall: ", u.String())

	switch reqTyp {
	case Http:
		resp, err := http.Get(u.String())
		if err != nil {
			panic("post rdCall message to chain error: " + err.Error())
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic("read rdCall response body error: " + err.Error())
		}
		return body
	case Websocket:
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			panic("rdCall dial chain error: " + err.Error())
		}
		defer c.Close()
		err = c.WriteMessage(websocket.TextMessage, []byte(rdCall.Params))
		if err != nil {
			panic("write rdCall message to chain error: " + err.Error())
		}
		_, resp, err := c.ReadMessage()
		if err != nil {
			panic("get rdCall response error: " + err.Error())
		}
		return resp
	}
	return nil
}

func CallChainByWriting(reqType int, privkey PrivKey, pubkey PubKey, wrCall *WrCall) {
	hash, err := wrCall.Hash()
	if err != nil {
		panic("wrCall hash error: " + err.Error())
	}
	signByt, err := privkey.SignData(hash)
	if err != nil {
		panic("sign data error: " + err.Error())
	}

	var (
		scheme, port string
	)
	switch reqType {
	case Http:
		scheme = "http"
		port = "7999"
	case Websocket:
		scheme = "ws"
		port = "8999"
	}

	u := url.URL{Scheme: scheme, Host: fmt.Sprintf("localhost:%s", port), Path: WrApiPath}
	q := u.Query()
	q.Set(TripodNameKey, wrCall.TripodName)
	q.Set(CallNameKey, wrCall.WritingName)
	q.Set(AddressKey, pubkey.Address().String())
	q.Set(SignatureKey, ToHex(signByt))
	q.Set(PubkeyKey, pubkey.StringWithType())
	q.Set(LeiPriceKey, hexutil.EncodeUint64(wrCall.LeiPrice))

	u.RawQuery = q.Encode()

	logrus.Debug("wrCall: ", u.String())

	switch reqType {
	case Http:
		_, err = http.Post(u.String(), "application/json", strings.NewReader(wrCall.Params))
		if err != nil {
			panic("post wrCall message to chain error: " + err.Error())
		}
	case Websocket:
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			panic("wrCall dial chain error: " + err.Error())
		}
		defer c.Close()
		err = c.WriteMessage(websocket.TextMessage, []byte(wrCall.Params))
		if err != nil {
			panic("write wrCall message to chain error: " + err.Error())
		}
	}
}

type Subscriber struct {
	conn   *websocket.Conn
	closed atomic.Bool
}

func NewSubscriber() (*Subscriber, error) {
	u := url.URL{Scheme: "ws", Host: "localhost:8999", Path: SubResultsPath}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic("dial chain error: " + err.Error())
	}
	return &Subscriber{
		conn: conn,
	}, nil
}

func (s *Subscriber) SubEvent(ch chan *Result) {
	for {
		if s.closed.Load() {
			return
		}
		_, msg, err := s.conn.ReadMessage()
		if err != nil {
			panic("sub event msg from chain error: " + err.Error())
		}
		result := new(Result)
		err = result.Decode(msg)
		if err != nil {
			logrus.Panicf("decode result error: %s", err.Error())
		}

		if ch != nil {
			ch <- result
		}
	}
}

func (s *Subscriber) CloseSub() error {
	s.closed.Store(true)
	return s.conn.Close()
}
