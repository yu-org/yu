package core

import (
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/keypair"
	"net/http"
	"path/filepath"
)

const (
	DownloadUpdatedPath     = "/download/upgrade"
	RegisterNodeKeepersPath = "/nodekeeper/register"
	RegisterWorkersPath     = "/worker/register"
	HeartbeatPath           = "/heartbeat"

	// Worker accept block from p2p network.
	// Master forwards this request to Worker.
	// Deprecated
	BlockFromP2P = "/p2p/block"

	// Worker accept txns from p2p network.
	// Master forwards this request to Worker.
	// Deprecated
	TxnsFromP2P = "/p2p/txns"

	StartBlockPath    = "/block/start"
	EndBlockPath      = "/block/end"
	FinalizeBlockPath = "/block/finalize"

	CheckTxnsPath   = "/txns/check"
	ExecuteTxnsPath = "/txns/execute"

	// For developers, every customized Execution and Query of tripods
	// will base on '/api'.
	RootApiPath = "/api"

	TripodNameKey = "tripod"
	CallNameKey   = "call_name"
	AddressKey    = "address"
	BlockHashKey  = "block_hash"
	PubkeyKey     = "pubkey"
	SignatureKey  = "signature"
)

var (
	ExecApiPath    = filepath.Join(RootApiPath, ExecCallType)
	QryApiPath     = filepath.Join(RootApiPath, QryCallType)
	SubResultsPath = "/subscribe/results"
)

// return (Tripod Name, Execution/Query Name)
func GetTripodCallName(req *http.Request) (string, string) {
	query := req.URL.Query()
	return query.Get(TripodNameKey), query.Get(CallNameKey)
}

// return the Address of Txn-Sender
func GetAddress(req *http.Request) Address {
	return HexToAddress(req.URL.Query().Get(AddressKey))
}

func GetBlockHash(req *http.Request) Hash {
	return HexToHash(req.URL.Query().Get(BlockHashKey))
}

func GetPubkey(req *http.Request) (keypair.PubKey, error) {
	pubkeyStr := req.URL.Query().Get(PubkeyKey)
	return keypair.PubkeyFromStr(pubkeyStr)
}

func GetSignature(req *http.Request) []byte {
	signStr := req.URL.Query().Get(SignatureKey)
	return FromHex(signStr)
}
