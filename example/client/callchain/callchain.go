package callchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/apps/metamask"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/protocol"
	. "github.com/yu-org/yu/core/types"
	"go.uber.org/atomic"
	"io"
	"net/http"
	"net/url"
)

const (
	Http = iota
	Websocket
)

func CallChainByReading(rdCall *RdCall) ([]byte, error) {
	u := url.URL{Scheme: "http", Host: "localhost:7999", Path: RdApiPath}
	bodyByt, err := json.Marshal(rdCall)
	if err != nil {
		return nil, err
	}

	logrus.Debug("rdCall: ", u.String())

	resp, err := http.Post(u.String(), "application/json", bytes.NewReader(bodyByt))
	if err != nil {
		panic("post rdCall message to chain error: " + err.Error())
	}
	return io.ReadAll(resp.Body)
}

func CallChainByWriting(postWriting *WritingPostBody) error {
	u := url.URL{Scheme: "http", Host: "localhost:7999", Path: WrApiPath}
	bodyByt, err := json.Marshal(postWriting)
	if err != nil {
		return err
	}

	logrus.Debug("wrCall: ", u.String())

	_, err = http.Post(u.String(), "application/json", bytes.NewReader(bodyByt))
	return err
}

func CallChainByWritingWithECDSA(privKey *ecdsa.PrivateKey, wrCall *WrCall) error {
	msgByt, err := json.Marshal(wrCall)
	if err != nil {
		return err
	}
	mmHash := metamask.MetamaskMsgHash(msgByt)
	sig, err := crypto.Sign(mmHash, privKey)
	if err != nil {
		return err
	}

	pubkey := crypto.FromECDSAPub(&privKey.PublicKey)

	recoverPub, err := crypto.Ecrecover(mmHash, sig)
	if err != nil {
		return err
	}
	if !bytes.Equal(pubkey, recoverPub) {
		return errors.New("pubkey != recover pubkey")
	}

	u := url.URL{Scheme: "http", Host: "localhost:7999", Path: WrApiPath}
	postBody := WritingPostBody{
		Pubkey:    hexutil.Encode(pubkey),
		Signature: hexutil.Encode(sig),
		Call:      wrCall,
	}
	bodyByt, err := json.Marshal(postBody)
	if err != nil {
		return err
	}

	logrus.Debug("wrCall: ", u.String())

	_, err = http.Post(u.String(), "application/json", bytes.NewReader(bodyByt))
	return err
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

func (s *Subscriber) SubEvent(ch chan *Receipt) {
	for {
		if s.closed.Load() {
			return
		}
		_, msg, err := s.conn.ReadMessage()
		if err != nil {
			panic("sub event msg from chain error: " + err.Error())
		}
		result := new(Receipt)
		err = result.Decode(msg)
		if err != nil {
			logrus.Panicf("decode receipt error: %s", err.Error())
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
