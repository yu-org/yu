package node

import (
	"net/http"
	"path/filepath"
	. "yu/common"
	"yu/keypair"
)

const (
	DownloadUpdatedPath     = "/download/upgrade"
	RegisterNodeKeepersPath = "/nodekeeper/register"
	RegisterWorkersPath     = "/worker/register"
	HeartbeatPath           = "/heartbeat"

	// Worker accept block from p2p network.
	// Master forwards this request to Worker.
	BlockFromP2P = "/p2p/block"

	// Worker accept txns from p2p network.
	// Master forwards this request to Worker.
	TxnsFromP2P = "/p2p/txns"

	// For developers, every customized Execution and Query of tripods
	// will base on '/api'.
	RootApiPath = "/api"

	TripodNameKey = "tripod"
	CallNameKey   = "call_name"
	AddressKey    = "address"
	BlockNumKey   = "block_num"
	KeyTypeKey    = "key_type"
	PubkeyKey     = "pubkey"
	SignatureKey  = "signature"
)

var (
	ExecApiPath = filepath.Join(RootApiPath, ExecCallType)
	QryApiPath  = filepath.Join(RootApiPath, QryCallType)
)

// return (Tripod Name, Execution/Query Name)
func GetTripodCallName(req *http.Request) (string, string) {
	query := req.URL.Query()
	return query.Get(TripodNameKey), query.Get(CallNameKey)
}

// return the Address of Txn-Sender
func GetAddress(req *http.Request) Address {
	return StrToAddress(req.URL.Query().Get(AddressKey))
}

func GetBlockNumber(req *http.Request) (BlockNum, error) {
	bnstr := req.URL.Query().Get(BlockNumKey)
	return StrToBlockNum(bnstr)
}

func GetPubkeyAndSignature(req *http.Request) (keypair.PubKey, []byte, error) {
	keyType := req.URL.Query().Get(KeyTypeKey)
	pubkeyStr := req.URL.Query().Get(PubkeyKey)
	pubkey, err := keypair.PubKeyFromBytes(keyType, []byte(pubkeyStr))
	if err != nil {
		return nil, nil, err
	}
	signatureStr := req.URL.Query().Get(SignatureKey)
	return pubkey, []byte(signatureStr), nil
}
