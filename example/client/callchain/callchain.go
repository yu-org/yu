package callchain

import (
	"bytes"
	"encoding/json"
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
)

const (
	Http = iota
	Websocket
)

func CallChainByReading(rdCall *RdCall, params map[string]string) []byte {
	u := url.URL{Scheme: "http", Host: "localhost:7999", Path: RdApiPath}
	q := u.Query()
	q.Set(TripodKey, rdCall.TripodName)
	q.Set(FuncNameKey, rdCall.FuncName)
	for key, value := range params {
		q.Set(key, value)
	}

	u.RawQuery = q.Encode()

	logrus.Debug("rdCall: ", u.String())

	resp, err := http.Get(u.String())
	if err != nil {
		panic("post rdCall message to chain error: " + err.Error())
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("read rdCall response body error: " + err.Error())
	}
	return body

}

func CallChainByWriting(privkey PrivKey, pubkey PubKey, wrCall *WrCall) {
	hash, err := wrCall.Hash()
	if err != nil {
		panic("wrCall hash error: " + err.Error())
	}
	signByt, err := privkey.SignData(hash)
	if err != nil {
		panic("sign data error: " + err.Error())
	}

	u := url.URL{Scheme: "http", Host: "localhost:7999", Path: WrApiPath}
	postBody := WritingPostBody{
		Pubkey:    pubkey.StringWithType(),
		Signature: ToHex(signByt),
		Call:      wrCall,
	}
	bodyByt, err := json.Marshal(postBody)
	if err != nil {
		panic("marshal post body failed: " + err.Error())
	}

	logrus.Debug("wrCall: ", u.String())

	_, err = http.Post(u.String(), "application/json", bytes.NewReader(bodyByt))
	if err != nil {
		panic("post wrCall message to chain error: " + err.Error())
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
